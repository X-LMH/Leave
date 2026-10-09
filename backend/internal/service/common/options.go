package common

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"errors"
)

var ErrorInvalidApartmentGender = errors.New("invalid apartment gender")

// App callers request enabled options; admin callers also need disabled values
// to display existing assignments and prevent selecting them for new assignments.
func GetClassOptions(enabledOnly bool) ([]*dto.ClassOption, error) {
	classes, err := mysql.GetClassOptions(enabledOnly)
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

func GetApartmentOptions(gender string, enabledOnly bool) ([]*dto.ApartmentOption, error) {
	if gender != "" && !models.IsValidGender(gender) {
		return nil, ErrorInvalidApartmentGender
	}
	apartments, err := mysql.GetApartmentOptions(gender, enabledOnly)
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

func GetLeaveTypeOptions(enabledOnly bool) ([]*dto.LeaveTypeOption, error) {
	leaveTypes, err := mysql.GetLeaveTypeOptions(enabledOnly)
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
