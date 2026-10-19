package mysql

import (
	"backend/internal/dto"
	"backend/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *DAO) InsertFeedback(feedback *models.Feedback) error {
	return s.db.Create(feedback).Error
}

func (s *DAO) GetFeedbackList(query dto.FeedbackListQuery) ([]*models.Feedback, int64, error) {
	rows := make([]*models.Feedback, 0)
	q := s.db.Model(&models.Feedback{})
	if query.Keyword != "" {
		q = q.Where("content LIKE ?", "%"+query.Keyword+"%")
	}
	if query.StudentID != "" {
		q = q.Where("is_anonymous = ? AND student_id LIKE ?", false, "%"+query.StudentID+"%")
	}
	if query.IsAnonymous != nil {
		q = q.Where("is_anonymous = ?", *query.IsAnonymous)
	}
	if query.IsHandled != nil {
		q = q.Where("is_handled = ?", *query.IsHandled)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("created_at DESC, id DESC").
		Scopes(Paginate(query.Page, query.PageSize)).
		Find(&rows).Error
	return rows, total, err
}

func (s *DAO) GetFeedback(id uint) (*models.Feedback, error) {
	var row models.Feedback
	if err := s.db.First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *DAO) DeleteFeedback(id uint) error {
	result := s.db.Delete(&models.Feedback{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *DAO) GetFeedbackForUpdate(tx *gorm.DB, id uint) (*models.Feedback, error) {
	var row models.Feedback
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *DAO) UpdateFeedbackHandling(tx *gorm.DB, row *models.Feedback, handled bool) error {
	return tx.Model(row).Update("is_handled", handled).Error
}
