package admin

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
)

func feedbackItem(row *models.Feedback) *dto.AdminFeedbackItem {
	studentID := row.StudentID
	// Also mask identities in historical or inconsistent anonymous rows.
	if row.IsAnonymous {
		studentID = nil
	}
	return &dto.AdminFeedbackItem{
		ID:          row.ID,
		StudentID:   studentID,
		IsAnonymous: row.IsAnonymous,
		Content:     row.Content,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func GetAdminFeedbacks(query dto.FeedbackListQuery) (*dto.FeedbackListResponse, error) {
	rows, total, err := mysql.GetFeedbackList(query)
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

func GetAdminFeedback(id uint) (*dto.AdminFeedbackItem, error) {
	row, err := mysql.GetFeedback(id)
	if err != nil {
		return nil, err
	}
	return feedbackItem(row), nil
}

func DeleteAdminFeedback(id uint) error { return mysql.DeleteFeedback(id) }
