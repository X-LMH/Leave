package app

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"errors"
	"strings"
	"unicode/utf8"
)

const maxFeedbackContentLength = 500

var ErrorInvalidFeedback = errors.New("invalid feedback")

func CreateFeedback(studentID string, request *dto.FeedbackCreateRequest) (*dto.FeedbackResponse, error) {
	if !validateAndNormalizeFeedbackRequest(request) {
		return nil, ErrorInvalidFeedback
	}

	feedback := &models.Feedback{
		IsAnonymous: request.IsAnonymous,
		Content:     request.Content,
	}
	if !request.IsAnonymous {
		feedback.StudentID = &studentID
	}
	if err := mysql.InsertFeedback(feedback); err != nil {
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
