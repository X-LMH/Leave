package mysql

import "backend/internal/models"

func InsertFeedback(feedback *models.Feedback) error {
	return db.Create(feedback).Error
}
