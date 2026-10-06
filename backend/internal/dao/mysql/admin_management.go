package mysql

import (
	"backend/internal/dto"
	"backend/internal/models"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrorClassInUse = errors.New("class is referenced by profiles")

func GetAdminLeaveTypes(query dto.AdminLeaveTypeListQuery) ([]*models.LeaveType, int64, error) {
	rows := make([]*models.LeaveType, 0)
	q := db.Model(&models.LeaveType{})
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
	err := q.Order("sort_order ASC, id ASC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&rows).Error
	return rows, total, err
}
func CreateAdminLeaveType(row *models.LeaveType) error { return db.Create(row).Error }
func UpdateAdminLeaveType(row *models.LeaveType) error {
	return db.Transaction(func(tx *gorm.DB) error {
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
func DeleteAdminLeaveType(id uint) error {
	result := db.Delete(&models.LeaveType{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func CreateAdminClass(row *models.Class) error { return db.Create(row).Error }

func UpdateAdminClass(row *models.Class) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var current models.Class
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, row.ID).Error; err != nil {
			return err
		}
		if err := tx.Model(&current).Updates(map[string]any{"college": row.College, "major": row.Major, "class_name": row.ClassName, "is_enabled": row.IsEnabled}).Error; err != nil {
			return err
		}
		return tx.First(row, row.ID).Error
	})
}
func DeleteAdminClass(id uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var row models.Class
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, id).Error; err != nil {
			return err
		}
		// Include deleted profiles: restoring a profile must retain its class relation.
		var count int64
		if err := tx.Unscoped().Model(&models.Profile{}).Where("class_id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrorClassInUse
		}
		return tx.Delete(&row).Error
	})
}
