package common

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"errors"

	"go.uber.org/zap"
)

type OptionsService struct {
	logger *zap.Logger
	dao    *mysql.DAO
}

func NewOptionsService(dao *mysql.DAO, logger *zap.Logger) *OptionsService {
	return &OptionsService{
		logger: logger,
		dao:    dao,
	}
}

var ErrorInvalidApartmentGender = errors.New("invalid apartment gender")

// App callers request enabled options; admin callers also need disabled values
// to display existing assignments and prevent selecting them for new assignments.
func (o *OptionsService) GetClassOptions(enabledOnly bool) ([]*dto.ClassOption, error) {
	classes, err := o.dao.GetClassOptions(enabledOnly)
	if err != nil {
		return nil, err
	}
	options := make([]*dto.ClassOption, 0, len(classes))
	for _, class := range classes {
		options = append(options, &dto.ClassOption{
			ID:        class.ID,
			College:   class.College,
			Major:     class.Major,
			ClassName: class.ClassName,
			IsEnabled: class.IsEnabled,
		})
	}
	return options, nil
}

func (o *OptionsService) GetApartmentOptions(gender string, enabledOnly bool) ([]*dto.ApartmentOption, error) {
	if gender != "" && !models.IsValidGender(gender) {
		return nil, ErrorInvalidApartmentGender
	}
	apartments, err := o.dao.GetApartmentOptions(gender, enabledOnly)
	if err != nil {
		return nil, err
	}
	options := make([]*dto.ApartmentOption, 0, len(apartments))
	for _, apartment := range apartments {
		options = append(options, &dto.ApartmentOption{
			ID:        apartment.ID,
			Name:      apartment.Name,
			Gender:    apartment.Gender,
			IsEnabled: apartment.IsEnabled == 1,
		})
	}
	return options, nil
}

func (o *OptionsService) GetLeaveTypeOptions(enabledOnly bool) ([]*dto.LeaveTypeOption, error) {
	leaveTypes, err := o.dao.GetLeaveTypeOptions(enabledOnly)
	if err != nil {
		return nil, err
	}
	options := make([]*dto.LeaveTypeOption, 0, len(leaveTypes))
	for _, leaveType := range leaveTypes {
		options = append(options, &dto.LeaveTypeOption{
			ID:        leaveType.ID,
			Name:      leaveType.Name,
			IsEnabled: leaveType.IsEnabled,
		})
	}
	return options, nil
}
