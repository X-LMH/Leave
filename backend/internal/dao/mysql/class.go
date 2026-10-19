package mysql

import (
	"backend/internal/dto"
	"backend/internal/models"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrorClassInUse = errors.New("class is referenced by profiles")

func (s *DAO) GetClassByID(classID uint) (*models.Class, error) {
	class := new(models.Class)
	if err := s.db.Where("id = ?", classID).First(class).Error; err != nil {
		return nil, err
	}
	return class, nil
}

func (s *DAO) GetClassOptions(enabledOnly bool) ([]*models.Class, error) {
	classes := make([]*models.Class, 0)
	query := s.db.Select("id", "college", "major", "class_name", "is_enabled")
	if enabledOnly {
		query = query.Where("is_enabled = ?", true)
	}
	err := query.Order("college ASC, major ASC, class_name ASC, id ASC").Find(&classes).Error
	return classes, err
}

func (s *DAO) GetClassList(query dto.ClassListQuery) ([]*models.Class, int64, error) {
	classes := make([]*models.Class, 0)
	dbQuery := s.db.Model(&models.Class{})
	if query.College != "" {
		dbQuery = dbQuery.Where("college LIKE ?", "%"+query.College+"%")
	}
	if query.Major != "" {
		dbQuery = dbQuery.Where("major LIKE ?", "%"+query.Major+"%")
	}
	if query.ClassName != "" {
		dbQuery = dbQuery.Where("class_name LIKE ?", "%"+query.ClassName+"%")
	}
	if query.IsEnabled != nil {
		dbQuery = dbQuery.Where("is_enabled = ?", *query.IsEnabled)
	}

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := dbQuery.Order("college ASC, major ASC, class_name ASC, id ASC").
		Scopes(Paginate(query.Page, query.PageSize)).
		Find(&classes).Error
	return classes, total, err
}

func (s *DAO) CreateClass(row *models.Class) error { return s.db.Create(row).Error }

func (s *DAO) UpdateClass(row *models.Class) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
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

func (s *DAO) DeleteClass(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
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
