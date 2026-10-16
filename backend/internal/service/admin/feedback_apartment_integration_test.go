package admin

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	appservice "backend/internal/service/app"
	commonservice "backend/internal/service/common"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	driver "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// This suite creates its own loopback schema and never loads application config.
func TestFeedbackApartmentIntegration(t *testing.T) {
	dsn := os.Getenv("ADMIN_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set ADMIN_TEST_MYSQL_DSN for isolated loopback MySQL coverage")
	}
	cfg, err := driver.ParseDSN(dsn)
	if err != nil {
		t.Fatal("invalid test MySQL DSN")
	}
	host, port, err := net.SplitHostPort(cfg.Addr)
	if err != nil || cfg.Net != "tcp" || (host != "localhost" && host != "127.0.0.1" && host != "::1") {
		t.Fatal("integration tests require loopback TCP MySQL")
	}
	cfg.DBName = ""
	cfg.ParseTime = true
	cfg.Loc = time.Local
	server, err := gorm.Open(gormmysql.Open(cfg.FormatDSN()), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatal("cannot connect to isolated MySQL server")
	}
	serverSQL, err := server.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer serverSQL.Close()
	schema := fmt.Sprintf("leave_feedback_apartment_test_%d", time.Now().UnixNano())
	if err := server.Exec("CREATE DATABASE `" + schema + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci").Error; err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := server.Exec("DROP DATABASE `" + schema + "`").Error; err != nil {
			t.Errorf("drop test schema: %v", err)
		}
	}()
	cfg.DBName = schema
	testDB, err := gorm.Open(gormmysql.Open(cfg.FormatDSN()), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatal(err)
	}
	connection, err := testDB.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	for _, name := range []string{"feedbacks", "apartments", "profiles", "classes", "users"} {
		sql, err := os.ReadFile(filepath.Join("..", "..", "..", "database", "tables", name+".sql"))
		if err != nil {
			t.Fatal(err)
		}
		if err := testDB.Exec(string(sql)).Error; err != nil {
			t.Fatal(err)
		}
	}
	previous := config.Cfg
	defer func() { config.Cfg = previous }()
	portNumber, err := strconv.Atoi(port)
	if err != nil {
		t.Fatal(err)
	}
	config.Cfg.Mysql.Host = host
	config.Cfg.Mysql.Port = portNumber
	config.Cfg.Mysql.Database = schema
	config.Cfg.Mysql.User = cfg.User
	config.Cfg.Mysql.Password = cfg.Passwd
	if err := mysql.Init(); err != nil {
		t.Fatal("cannot initialize isolated test database")
	}

	t.Run("feedback migration", func(t *testing.T) { testFeedbackMigration(t, testDB) })
	t.Run("feedback attachments and handling", func(t *testing.T) { testFeedbackAttachmentsIntegration(t, testDB) })

	order := uint32(0)
	enabled := true
	input := dto.ApartmentRequest{Name: "测试楼", Gender: models.GenderMale, SortOrder: &order, IsEnabled: &enabled}
	row, err := CreateAdminApartment(&input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CreateAdminApartment(&input); !mysql.IsDuplicateKeyError(err) {
		t.Fatalf("duplicate: %v", err)
	}
	input.Gender = models.GenderFemale
	if _, err := CreateAdminApartment(&input); err != nil {
		t.Fatal("different genders should allow same name", err)
	}
	input.Gender = models.GenderMale
	enabled = false
	updated, err := UpdateAdminApartment(row.ID, &input)
	if err != nil || updated.IsEnabled || updated.SortOrder != 0 {
		t.Fatalf("zero and disabled: %+v %v", updated, err)
	}
	list, err := GetAdminApartments(dto.ApartmentListQuery{Page: 1, PageSize: 1, Name: "测试", Gender: models.GenderMale, IsEnabled: &enabled})
	if err != nil || list.Total != 1 || len(list.Items) != 1 {
		t.Fatalf("filter: %+v %v", list, err)
	}
	appOptions, err := commonservice.GetApartmentOptions(models.GenderMale, true)
	if err != nil || len(appOptions) != 0 {
		t.Fatalf("disabled App options: %+v %v", appOptions, err)
	}
	adminOptions, err := commonservice.GetApartmentOptions(models.GenderMale, false)
	if err != nil || len(adminOptions) != 1 || adminOptions[0].IsEnabled {
		t.Fatalf("admin options: %+v %v", adminOptions, err)
	}
	class := models.Class{College: "学院", Major: "专业", ClassName: "班", IsEnabled: true}
	if err := testDB.Create(&class).Error; err != nil {
		t.Fatal(err)
	}
	profileInput := dto.ProfileRequest{ClassID: class.ID, ApartmentID: row.ID, Gender: models.GenderMale}
	if err := appservice.Profile(&profileInput, "new-disabled"); !errors.Is(err, appservice.ErrorInvalidProfileReference) {
		t.Fatalf("new disabled: %v", err)
	}
	enabled = true
	if _, err := UpdateAdminApartment(row.ID, &input); err != nil {
		t.Fatal(err)
	}
	if err := appservice.Profile(&profileInput, "resident"); err != nil {
		t.Fatal(err)
	}
	enabled = false
	if _, err := UpdateAdminApartment(row.ID, &input); err != nil {
		t.Fatal(err)
	}
	if err := appservice.Profile(&profileInput, "resident"); err != nil {
		t.Fatalf("retain disabled: %v", err)
	}
	residentApartments, err := appservice.GetProfileApartmentOptions("resident", models.GenderMale)
	if err != nil || len(residentApartments) != 1 || residentApartments[0].ID != row.ID || residentApartments[0].IsEnabled {
		t.Fatalf("retain disabled apartment option: %+v %v", residentApartments, err)
	}
	newApartments, err := appservice.GetProfileApartmentOptions("new-student", models.GenderMale)
	if err != nil || len(newApartments) != 0 {
		t.Fatalf("disabled apartment exposed to new student: %+v %v", newApartments, err)
	}
	femaleApartments, err := appservice.GetProfileApartmentOptions("resident", models.GenderFemale)
	if err != nil || len(femaleApartments) != 1 || femaleApartments[0].Gender != models.GenderFemale {
		t.Fatalf("old apartment exposed after gender change: %+v %v", femaleApartments, err)
	}
	if err := testDB.Model(&class).Update("is_enabled", false).Error; err != nil {
		t.Fatal(err)
	}
	residentClasses, err := appservice.GetProfileClassOptions("resident")
	if err != nil || len(residentClasses) != 1 || residentClasses[0].ID != class.ID || residentClasses[0].IsEnabled {
		t.Fatalf("retain disabled class option: %+v %v", residentClasses, err)
	}
	newClasses, err := appservice.GetProfileClassOptions("new-student")
	if err != nil || len(newClasses) != 0 {
		t.Fatalf("disabled class exposed to new student: %+v %v", newClasses, err)
	}
	if err := appservice.Profile(&profileInput, "resident"); err != nil {
		t.Fatalf("retain disabled class and apartment: %v", err)
	}
	newProfile := profileInput
	newProfile.ApartmentID = 0
	if err := appservice.Profile(&newProfile, "new-student"); !errors.Is(err, appservice.ErrorInvalidProfileReference) {
		t.Fatalf("new disabled class assignment: %v", err)
	}
	if err := testDB.Model(&class).Update("is_enabled", true).Error; err != nil {
		t.Fatal(err)
	}
	if err := DeleteAdminApartment(row.ID); !errors.Is(err, ErrorApartmentInUse) {
		t.Fatalf("reference: %v", err)
	}
	input.Gender = models.GenderFemale
	if _, err := UpdateAdminApartment(row.ID, &input); !errors.Is(err, ErrorApartmentInUse) {
		t.Fatalf("gender reference: %v", err)
	}
	if err := testDB.Where("student_id = ?", "resident").Delete(&models.Profile{}).Error; err != nil {
		t.Fatal(err)
	}
	if err := DeleteAdminApartment(row.ID); !errors.Is(err, ErrorApartmentInUse) {
		t.Fatalf("deleted profile: %v", err)
	}
	if _, err := UpdateAdminApartment(row.ID, &input); !errors.Is(err, ErrorApartmentInUse) {
		t.Fatalf("deleted gender reference: %v", err)
	}
	if err := testDB.Unscoped().Where("student_id = ?", "resident").Delete(&models.Profile{}).Error; err != nil {
		t.Fatal(err)
	}
	if err := DeleteAdminApartment(row.ID); err != nil {
		t.Fatal(err)
	}
	if err := DeleteAdminApartment(row.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("missing: %v", err)
	}
	if err := appservice.Profile(&profileInput, "deleted-apartment"); !errors.Is(err, appservice.ErrorInvalidProfileReference) {
		t.Fatalf("deleted assignment: %v", err)
	}
	profileInput.ApartmentID = 0
	profileInput.ClassID = 999999
	if err := appservice.Profile(&profileInput, "deleted-class"); !errors.Is(err, appservice.ErrorInvalidProfileReference) {
		t.Fatalf("deleted class: %v", err)
	}

	start := time.Date(2026, 10, 10, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 0, 1)
	student := "20260001"
	feedbacks := []models.Feedback{
		{StudentID: &student, Content: "建议\n第二行", CreatedAt: start},
		{StudentID: &student, IsAnonymous: true, Content: "建议", CreatedAt: start.Add(23 * time.Hour)},
		{Content: "建议", CreatedAt: end},
	}
	for i := range feedbacks {
		if err := testDB.Create(&feedbacks[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	query := dto.FeedbackListQuery{Page: 1, PageSize: 1, Keyword: "建议"}
	feedbackList, err := GetAdminFeedbacks(query)
	if err != nil || feedbackList.Total != 3 || len(feedbackList.Items) != 1 || feedbackList.Items[0].ID != feedbacks[2].ID || feedbackList.Items[0].StudentID != nil {
		t.Fatalf("order/privacy: %+v %v", feedbackList, err)
	}
	anonymous := true
	query.StudentID = student
	query.IsAnonymous = &anonymous
	feedbackList, err = GetAdminFeedbacks(query)
	if err != nil || feedbackList.Total != 0 {
		t.Fatalf("anonymous search: %+v %v", feedbackList, err)
	}
	detail, err := GetAdminFeedback(feedbacks[0].ID)
	if err != nil || detail.Content != feedbacks[0].Content || detail.StudentID == nil {
		t.Fatalf("detail: %+v %v", detail, err)
	}
	if err := DeleteAdminFeedback(feedbacks[0].ID); err != nil {
		t.Fatal(err)
	}
	if _, err := GetAdminFeedback(feedbacks[0].ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("deleted detail: %v", err)
	}
	if err := DeleteAdminFeedback(feedbacks[0].ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("repeat deletion: %v", err)
	}
	query.StudentID = ""
	query.IsAnonymous = nil
	feedbackList, err = GetAdminFeedbacks(query)
	if err != nil || feedbackList.Total != 2 {
		t.Fatalf("soft deletion: %+v %v", feedbackList, err)
	}
	var deleted models.Feedback
	if err := testDB.Unscoped().First(&deleted, feedbacks[0].ID).Error; err != nil || !deleted.DeletedAt.Valid {
		t.Fatal("not soft deleted", err)
	}
	// A profile writer holds a shared apartment lock until its reference is committed.
	// Management must wait, then observe that reference before deleting/changing gender.
	for _, operation := range []string{"delete", "gender"} {
		t.Run("concurrent "+operation, func(t *testing.T) {
			enabled := true
			order := uint32(0)
			p := dto.ApartmentRequest{Name: "并发楼" + operation, Gender: models.GenderMale, SortOrder: &order, IsEnabled: &enabled}
			apartment, err := CreateAdminApartment(&p)
			if err != nil {
				t.Fatal(err)
			}
			tx := testDB.Begin()
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}
			defer tx.Rollback()
			if _, err := mysql.GetStudentApartmentForShare(tx, apartment.ID); err != nil {
				t.Fatal(err)
			}
			started := make(chan struct{})
			result := make(chan error, 1)
			go func() {
				close(started)
				if operation == "delete" {
					result <- DeleteAdminApartment(apartment.ID)
					return
				}
				p.Gender = models.GenderFemale
				_, err := UpdateAdminApartment(apartment.ID, &p)
				result <- err
			}()
			<-started
			select {
			case err := <-result:
				t.Fatalf("management bypassed shared lock: %v", err)
			case <-time.After(100 * time.Millisecond):
			}
			if err := mysql.SaveStudentProfile(tx, &models.Profile{StudentID: "concurrent-" + operation, ClassID: class.ID, ApartmentID: apartment.ID, Gender: models.GenderMale}); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit().Error; err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-result:
				if !errors.Is(err, ErrorApartmentInUse) {
					t.Fatalf("committed reference was ignored: %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("management did not resume after commit")
			}
		})
	}

}
