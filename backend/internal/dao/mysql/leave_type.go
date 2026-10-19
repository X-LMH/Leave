package mysql

import (
	"backend/internal/dto"
	"backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *DAO) GetLeaveTypeList(query dto.LeaveTypeListQuery) ([]*models.LeaveType, int64, error) {
	rows := make([]*models.LeaveType, 0)
	q := s.db.Model(&models.LeaveType{})
	if query.Name != "" {
		q = q.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.IsEnabled != nil {
		q = q.Where("is_enabled = ?", *query.IsEnabled)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("sort_order ASC, id ASC").
		Scopes(Paginate(query.Page, query.PageSize)).
		Find(&rows).Error
	return rows, total, err
}

func (s *DAO) CreateLeaveType(row *models.LeaveType) error { return s.db.Create(row).Error }
func (s *DAO) UpdateLeaveType(row *models.LeaveType) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var current models.LeaveType
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, row.ID).Error; err != nil {
			return err
		}
		if err := tx.Model(&current).Updates(map[string]any{"name": row.Name, "sort_order": row.SortOrder, "is_enabled": row.IsEnabled}).Error; err != nil {
			return err
		}
		return tx.First(row, row.ID).Error
	})
}

func (s *DAO) DeleteLeaveType(id uint) error {
	result := s.db.Delete(&models.LeaveType{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *DAO) GetLeaveTypeOptions(enabledOnly bool) ([]*models.LeaveType, error) {
	rows := make([]*models.LeaveType, 0)
	query := s.db.Select("id", "name", "is_enabled")
	if enabledOnly {
		query = query.Where("is_enabled = ?", true)
	}
	err := query.Order("sort_order ASC, id ASC").Find(&rows).Error
	return rows, err
}
