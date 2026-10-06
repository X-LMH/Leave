package mysql

import (
	"backend/internal/dto"
	"backend/internal/models"
)

func GetClassByID(classID uint) (*models.Class, error) {
	class := new(models.Class)
	if err := db.Where("id = ?", classID).First(class).Error; err != nil {
		return nil, err
	}
	return class, nil
}

func GetClasses() ([]*models.Class, error) {
	classes := make([]*models.Class, 0)
	err := db.Select("id", "college", "major", "class_name").
		Where("is_enabled = ?", true).
		Order("college ASC, major ASC, class_name ASC, id ASC").
		Find(&classes).Error
	return classes, err
}

func GetAdminClasses(query dto.AdminClassListQuery) ([]*models.Class, int64, error) {
	classes := make([]*models.Class, 0)
	dbQuery := db.Model(&models.Class{})
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
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&classes).Error
	return classes, total, err
}
