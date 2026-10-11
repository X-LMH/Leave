package mysql

import (
	"backend/internal/dto"
	"backend/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetApartmentList(query dto.ApartmentListQuery) ([]*models.Apartment, int64, error) {
	rows := make([]*models.Apartment, 0)
	q := db.Model(&models.Apartment{})
	if query.Name != "" {
		q = q.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.Gender != "" {
		q = q.Where("gender = ?", query.Gender)
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

func CreateApartment(row *models.Apartment) error { return db.Create(row).Error }

func GetApartmentForUpdate(tx *gorm.DB, id uint) (*models.Apartment, error) {
	var row models.Apartment
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func ApartmentHasProfiles(tx *gorm.DB, id uint) (bool, error) {
	var count int64
	// Include deleted profiles so restoring them cannot leave a dangling reference.
	err := tx.Unscoped().
		Model(&models.Profile{}).
		Where("apartment_id = ?", id).
		Count(&count).Error
	return count > 0, err
}

func SaveApartment(tx *gorm.DB, row *models.Apartment) error {
	if err := tx.Model(row).Updates(map[string]any{
		"name":       row.Name,
		"gender":     row.Gender,
		"sort_order": row.SortOrder,
		"is_enabled": row.IsEnabled,
	}).Error; err != nil {
		return err
	}
	return tx.First(row, row.ID).Error
}

func DeleteApartment(tx *gorm.DB, row *models.Apartment) error { return tx.Delete(row).Error }
