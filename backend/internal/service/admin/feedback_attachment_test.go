package admin

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	appservice "backend/internal/service/app"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"image"
	"image/png"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
)

func TestFeedbackAttachmentPaths(t *testing.T) {
	settings := config.Config{}
	dao := mysql.NewDAO(nil, zap.NewNop())
	for _, value := range []string{`["../secret.png"]`, `["avatars/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png"]`, `["feedbacks/../../secret.png"]`, `["feedbacks/bad.jpg"]`, `not json`} {
		row := &models.Feedback{Attachments: &value}
		if _, err := feedbackDetail(row); err == nil {
			t.Fatalf("accepted invalid paths %s", value)
		}
	}
	row := &models.Feedback{ID: 5, IsAnonymous: true}
	detail, err := feedbackDetail(row)
	if err != nil || detail.Attachments == nil || len(detail.Attachments) != 0 {
		t.Fatalf("legacy NULL: %v %v", detail, err)
	}
	value := `["feedbacks/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png"]`
	row.Attachments = &value
	detail, err = feedbackDetail(row)
	if err != nil || len(detail.Attachments) != 1 || detail.Attachments[0].Index != 0 || detail.Attachments[0].URL != "/admin/feedbacks/5/attachments/0" {
		t.Fatalf("detail=%v err=%v", detail, err)
	}
	if _, err := NewFeedbackService(dao, settings.Storage, zap.NewNop()).UpdateAdminFeedbackHandling(1, &dto.FeedbackHandlingRequest{}); !errors.Is(err, ErrorInvalidAdminInput) {
		t.Fatalf("missing handling value: %v", err)
	}
}

// 由已有的隔离 loopback 数据库套件调用，不读取开发业务库配置。
func testFeedbackAttachmentsIntegration(t *testing.T, db *gorm.DB) {
	settings := config.Config{}
	dao := mysql.NewDAO(db, zap.NewNop())
	t.Helper()
	settings.Storage.RootDir = t.TempDir()
	var data bytes.Buffer
	if err := png.Encode(&data, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	feedbackService := NewFeedbackService(dao, settings.Storage, zap.NewNop())
	appFeedbackService := appservice.NewFeedbackService(dao, settings.Storage, zap.NewNop())
	for _, count := range []int{0, 1, 3} {
		t.Run(fmt.Sprintf("submit %d images", count), func(t *testing.T) {
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			for i := 0; i < count; i++ {
				part, err := writer.CreateFormFile(fmt.Sprintf("attachment_%d", i), "image.png")
				if err != nil {
					t.Fatal(err)
				}
				_, _ = part.Write(data.Bytes())
			}
			_ = writer.Close()
			req := httptest.NewRequest("POST", "/", &body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			if err := req.ParseMultipartForm(1 << 20); err != nil {
				t.Fatal(err)
			}
			defer req.MultipartForm.RemoveAll()
			var uploads []*multipart.FileHeader
			for i := 0; i < count; i++ {
				uploads = append(uploads, req.MultipartForm.File[fmt.Sprintf("attachment_%d", i)][0])
			}
			submitted, err := appFeedbackService.CreateFeedback("private-student", &dto.FeedbackCreateRequest{Content: "图片反馈", IsAnonymous: true}, uploads...)
			if err != nil {
				t.Fatal(err)
			}
			detail, err := feedbackService.GetAdminFeedback(submitted.ID)
			if err != nil || detail.StudentID != nil || detail.IsHandled || len(detail.Attachments) != count {
				t.Fatalf("detail=%+v err=%v", detail, err)
			}
			for _, handled := range []bool{true, true, false, false} {
				result, err := feedbackService.UpdateAdminFeedbackHandling(submitted.ID, &dto.FeedbackHandlingRequest{IsHandled: &handled})
				if err != nil || result.IsHandled != handled {
					t.Fatalf("handled=%v result=%+v err=%v", handled, result, err)
				}
			}
			for i := 0; i < count; i++ {
				file, err := feedbackService.OpenAdminFeedbackAttachment(submitted.ID, i)
				if err != nil {
					t.Fatal(err)
				}
				_ = file.Close()
			}
			if _, err := feedbackService.OpenAdminFeedbackAttachment(submitted.ID, count); !errors.Is(err, gorm.ErrRecordNotFound) {
				t.Fatalf("out of range %v", err)
			}
			if err := feedbackService.DeleteAdminFeedback(submitted.ID); err != nil {
				t.Fatal(err)
			}
			handled := true
			if _, err := feedbackService.UpdateAdminFeedbackHandling(submitted.ID, &dto.FeedbackHandlingRequest{IsHandled: &handled}); !errors.Is(err, gorm.ErrRecordNotFound) {
				t.Fatalf("deleted handling %v", err)
			}
			if _, err := feedbackService.OpenAdminFeedbackAttachment(submitted.ID, 0); !errors.Is(err, gorm.ErrRecordNotFound) {
				t.Fatalf("deleted attachment %v", err)
			}
		})
	}
	handled := false
	if _, err := feedbackService.UpdateAdminFeedbackHandling(4294967295, &dto.FeedbackHandlingRequest{IsHandled: &handled}); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("missing handling %v", err)
	}
}

func testFeedbackMigration(t *testing.T, db *gorm.DB) {
	t.Helper()
	read := func(direction string) string {
		data, err := os.ReadFile(filepath.Join("..", "..", "..", "database", "migrations", "000011_add_feedback_handled_state_and_attachments."+direction+".sql"))
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	up, down := read("up"), read("down")
	if err := db.Exec(down).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO feedbacks (content) VALUES (?)", "历史反馈").Error; err != nil {
		t.Fatal(err)
	}
	for _, sql := range []string{up, down, up} {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	var row models.Feedback
	if err := db.Where("content = ?", "历史反馈").First(&row).Error; err != nil {
		t.Fatal(err)
	}
	if row.IsHandled || row.Attachments != nil {
		encoded, _ := json.Marshal(row)
		t.Fatalf("historical defaults %s", encoded)
	}
	if err := db.Unscoped().Delete(&row).Error; err != nil {
		t.Fatal(err)
	}
}
