package admin

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"errors"
	"strings"

	"gorm.io/gorm"
)

const maxApartmentNameLength = 100

var ErrorApartmentInUse = errors.New("apartment is referenced by profiles")

func normalizeApartment(p *dto.ApartmentRequest) error {
	p.Name = strings.TrimSpace(p.Name)
	p.Gender = strings.TrimSpace(p.Gender)
	if !validAdminText(p.Name, maxApartmentNameLength) || !models.IsValidGender(p.Gender) || p.SortOrder == nil || p.IsEnabled == nil {
		return ErrorInvalidAdminInput
	}
	return nil
}

func apartmentItem(row *models.Apartment) *dto.ApartmentListItem {
	return &dto.ApartmentListItem{
		ID:        row.ID,
		Name:      row.Name,
		Gender:    row.Gender,
		SortOrder: row.SortOrder,
		IsEnabled: row.IsEnabled == 1,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func GetAdminApartments(query dto.ApartmentListQuery) (*dto.ApartmentListResponse, error) {
	if query.Gender != "" && !models.IsValidGender(query.Gender) {
		return nil, ErrorInvalidAdminInput
	}
	rows, total, err := mysql.GetApartmentList(query)
	if err != nil {
		return nil, err
	}
	items := make([]*dto.ApartmentListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, apartmentItem(row))
	}
	return &dto.ApartmentListResponse{
		Items:    items,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func applyApartmentRequest(row *models.Apartment, p *dto.ApartmentRequest) {
	row.Name = p.Name
	row.Gender = p.Gender
	row.SortOrder = uint(*p.SortOrder)
	row.IsEnabled = 0
	if *p.IsEnabled {
		row.IsEnabled = 1
	}
}

func CreateAdminApartment(p *dto.ApartmentRequest) (*dto.ApartmentListItem, error) {
	if err := normalizeApartment(p); err != nil {
		return nil, err
	}
	row := new(models.Apartment)
	applyApartmentRequest(row, p)
	if err := mysql.CreateApartment(row); err != nil {
		return nil, err
	}
	return apartmentItem(row), nil
}

func UpdateAdminApartment(id uint, p *dto.ApartmentRequest) (*dto.ApartmentListItem, error) {
	if err := normalizeApartment(p); err != nil {
		return nil, err
	}
	var row *models.Apartment
	err := mysql.Transaction(func(tx *gorm.DB) error {
		var err error
		row, err = mysql.GetApartmentForUpdate(tx, id)
		if err != nil {
			return err
		}
		if row.Gender != p.Gender {
			used, err := mysql.ApartmentHasProfiles(tx, id)
			if err != nil {
				return err
			}
			if used {
				return ErrorApartmentInUse
			}
		}
		applyApartmentRequest(row, p)
		return mysql.SaveApartment(tx, row)
	})
	if err != nil {
		return nil, err
	}
	return apartmentItem(row), nil
}

func DeleteAdminApartment(id uint) error {
	return mysql.Transaction(func(tx *gorm.DB) error {
		row, err := mysql.GetApartmentForUpdate(tx, id)
		if err != nil {
			return err
		}
		used, err := mysql.ApartmentHasProfiles(tx, id)
		if err != nil {
			return err
		}
		if used {
			return ErrorApartmentInUse
		}
		return mysql.DeleteApartment(tx, row)
	})
}
