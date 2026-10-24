package admin

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/utils/file"
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type FeedbackService struct {
	logger  *zap.Logger
	dao     *mysql.DAO
	storage config.StorageConfig
}

func NewFeedbackService(dao *mysql.DAO, storage config.StorageConfig, logger *zap.Logger) *FeedbackService {
	return &FeedbackService{
		logger:  logger,
		dao:     dao,
		storage: storage,
	}
}

func feedbackItem(row *models.Feedback) *dto.AdminFeedbackItem {
	studentID := row.StudentID
	// Also mask identities in historical or inconsistent anonymous rows.
	if row.IsAnonymous {
		studentID = nil
	}
	return &dto.AdminFeedbackItem{
		ID:          row.ID,
		IsHandled:   row.IsHandled,
		StudentID:   studentID,
		IsAnonymous: row.IsAnonymous,
		Content:     row.Content,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func (f *FeedbackService) GetAdminFeedbacks(query dto.FeedbackListQuery) (*dto.FeedbackListResponse, error) {
	rows, total, err := f.dao.GetFeedbackList(query)
	if err != nil {
		return nil, err
	}
	items := make([]*dto.AdminFeedbackItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, feedbackItem(row))
	}
	return &dto.FeedbackListResponse{
		Items:    items,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func (f *FeedbackService) GetAdminFeedback(id uint) (*dto.AdminFeedbackDetail, error) {
	row, err := f.dao.GetFeedback(id)
	if err != nil {
		return nil, err
	}
	return feedbackDetail(row)
}

func (f *FeedbackService) DeleteAdminFeedback(id uint) error { return f.dao.DeleteFeedback(id) }

// 只允许服务端生成的反馈图片文件名，不能通过数据库路径读取其他业务或任意文件。
var feedbackAttachmentKey = regexp.MustCompile(`^feedbacks/[0-9a-f]{32}\.(jpg|png)$`)

func feedbackAttachmentKeys(row *models.Feedback) ([]string, error) {
	var keys []string
	if row.Attachments == nil {
		return keys, nil
	}
	if err := json.Unmarshal([]byte(*row.Attachments), &keys); err != nil {
		return nil, err
	}
	for _, key := range keys {
		if !feedbackAttachmentKey.MatchString(key) {
			return nil, fmt.Errorf("invalid feedback attachment path")
		}
	}
	return keys, nil
}

func feedbackDetail(row *models.Feedback) (*dto.AdminFeedbackDetail, error) {
	keys, err := feedbackAttachmentKeys(row)
	if err != nil {
		return nil, err
	}
	attachments := make([]dto.FeedbackAttachment, 0, len(keys))
	for index := range keys {
		attachments = append(attachments, dto.FeedbackAttachment{
			Index: index,
			URL:   fmt.Sprintf("/admin/feedbacks/%d/attachments/%d", row.ID, index),
		})
	}
	return &dto.AdminFeedbackDetail{
		AdminFeedbackItem: *feedbackItem(row),
		Attachments:       attachments,
	}, nil
}

func (f *FeedbackService) UpdateAdminFeedbackHandling(id uint, request *dto.FeedbackHandlingRequest) (*dto.AdminFeedbackDetail, error) {
	if request.IsHandled == nil {
		return nil, ErrorInvalidAdminInput
	}
	var detail *dto.AdminFeedbackDetail
	err := f.dao.Transaction(func(tx *gorm.DB) error {
		row, err := f.dao.GetFeedbackForUpdate(tx, id)
		if err != nil {
			return err
		}
		if row.IsHandled != *request.IsHandled {
			if err := f.dao.UpdateFeedbackHandling(tx, row, *request.IsHandled); err != nil {
				return err
			}
		}
		detail, err = feedbackDetail(row)
		return err
	})
	return detail, err
}

func (f *FeedbackService) OpenAdminFeedbackAttachment(id uint, index int) (*os.File, error) {
	row, err := f.dao.GetFeedback(id)
	if err != nil {
		return nil, err
	}
	keys, err := feedbackAttachmentKeys(row)
	if err != nil {
		return nil, err
	}
	if index < 0 || index >= len(keys) {
		return nil, gorm.ErrRecordNotFound
	}
	image, err := file.Open(f.storage.RootDir, keys[index])
	if os.IsNotExist(err) {
		return nil, gorm.ErrRecordNotFound
	}
	return image, err
}
