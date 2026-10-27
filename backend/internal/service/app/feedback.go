package app

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	logutil "backend/internal/logger"
	"backend/internal/models"
	"backend/internal/utils/file"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"strings"
	"unicode/utf8"

	"go.uber.org/zap"
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

const maxFeedbackContentLength = 500
const MaxFeedbackAttachments = 3
const MaxFeedbackAttachmentSize = 5 << 20

var ErrorInvalidFeedback = errors.New("invalid feedback")

func (f *FeedbackService) CreateFeedback(studentID string, request *dto.FeedbackCreateRequest, uploads ...*multipart.FileHeader) (*dto.FeedbackResponse, error) {
	if len(uploads) > MaxFeedbackAttachments || !validateAndNormalizeFeedbackRequest(request) {
		return nil, ErrorInvalidFeedback
	}

	images, err := readFeedbackImages(uploads)
	if err != nil {
		return nil, err
	}
	feedback := &models.Feedback{
		IsAnonymous: request.IsAnonymous,
		Content:     request.Content,
	}
	if !request.IsAnonymous {
		feedback.StudentID = &studentID
	}
	if err := f.saveFeedbackWithAttachments(feedback, images, f.dao.InsertFeedback); err != nil {
		return nil, err
	}
	return toFeedbackResponse(feedback), nil
}

func validateAndNormalizeFeedbackRequest(request *dto.FeedbackCreateRequest) bool {
	request.Content = strings.TrimSpace(request.Content)
	return request.Content != "" && utf8.RuneCountInString(request.Content) <= maxFeedbackContentLength
}

func toFeedbackResponse(feedback *models.Feedback) *dto.FeedbackResponse {
	return &dto.FeedbackResponse{
		ID:          feedback.ID,
		IsAnonymous: feedback.IsAnonymous,
		Content:     feedback.Content,
		CreatedAt:   feedback.CreatedAt,
	}
}

type feedbackImage struct {
	data      []byte
	extension string
}

func readFeedbackImages(uploads []*multipart.FileHeader) ([]feedbackImage, error) {
	images := make([]feedbackImage, 0, len(uploads))
	for _, upload := range uploads {
		if upload == nil || upload.Size > MaxFeedbackAttachmentSize {
			return nil, ErrorInvalidFeedback
		}
		content, err := upload.Open()
		if err != nil {
			return nil, err
		}
		// 独立限制读取长度，不能仅依赖客户端或 multipart 元数据。
		data, extension, readErr := file.ReadImage(io.LimitReader(content, MaxFeedbackAttachmentSize), 4096)
		closeErr := content.Close()
		if errors.Is(readErr, file.ErrInvalidImage) || len(data) > MaxFeedbackAttachmentSize {
			return nil, ErrorInvalidFeedback
		}
		if err := errors.Join(readErr, closeErr); err != nil {
			return nil, err
		}
		images = append(images, feedbackImage{data: data, extension: extension})
	}
	return images, nil
}

// 文件与数据库无法组成原子事务；写库前的任何失败都清理本次文件。
func (f *FeedbackService) saveFeedbackWithAttachments(feedback *models.Feedback, images []feedbackImage, insert func(*models.Feedback) error) (err error) {
	root := f.storage.RootDir
	files := make([]file.Input, 0, len(images))
	for _, image := range images {
		files = append(files, file.Input{
			Extension: image.extension,
			Content:   bytes.NewReader(image.data),
		})
	}
	keys, err := file.Save(root, "feedbacks", files...)
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			return
		}
		for _, key := range keys {
			if cleanupErr := file.Delete(root, key); cleanupErr != nil {
				f.logger.Warn("清理反馈附件失败",
					zap.String("operation", "app.feedback.cleanup_attachments"),
					zap.String("file_key", key),
					logutil.Error(cleanupErr),
				)
			}
		}
	}()
	if len(keys) > 0 {
		data, marshalErr := json.Marshal(keys)
		if marshalErr != nil {
			return marshalErr
		}
		encoded := string(data)
		feedback.Attachments = &encoded
	}
	return insert(feedback)
}
