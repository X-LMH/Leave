package app

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	commonservice "backend/internal/service/common"
)

type OptionsService struct {
	dao     *mysql.DAO
	options *commonservice.OptionsService
}

func NewOptionsService(dao *mysql.DAO, options *commonservice.OptionsService) *OptionsService {
	return &OptionsService{
		dao:     dao,
		options: options,
	}
}

// Include only this student's existing assignment when it has been disabled.
// Other disabled options remain unavailable for new assignments.
func (o *OptionsService) GetProfileClassOptions(studentID string) ([]*dto.ClassOption, error) {
	options, err := o.options.GetClassOptions(true)
	if err != nil {
		return nil, err
	}
	profile, err := o.dao.GetProfileByStuID(studentID)
	if err != nil {
		return nil, err
	}
	if profile == nil || profile.ClassID == 0 {
		return options, nil
	}
	for _, option := range options {
		if option.ID == profile.ClassID {
			return options, nil
		}
	}
	class, err := o.dao.GetClassByID(profile.ClassID)
	if err != nil {
		return nil, err
	}
	return append(options, &dto.ClassOption{
		ID:        class.ID,
		College:   class.College,
		Major:     class.Major,
		ClassName: class.ClassName,
		IsEnabled: class.IsEnabled,
	}), nil
}

func (o *OptionsService) GetProfileApartmentOptions(studentID, gender string) ([]*dto.ApartmentOption, error) {
	options, err := o.options.GetApartmentOptions(gender, true)
	if err != nil {
		return nil, err
	}
	profile, err := o.dao.GetProfileByStuID(studentID)
	if err != nil {
		return nil, err
	}
	if profile == nil || profile.ApartmentID == 0 {
		return options, nil
	}
	for _, option := range options {
		if option.ID == profile.ApartmentID {
			return options, nil
		}
	}
	apartment, err := o.dao.GetApartmentByID(profile.ApartmentID)
	if err != nil {
		return nil, err
	}
	// Changing gender must not reintroduce an incompatible old apartment.
	if gender != "" && apartment.Gender != gender {
		return options, nil
	}
	return append(options, &dto.ApartmentOption{
		ID:        apartment.ID,
		Name:      apartment.Name,
		Gender:    apartment.Gender,
		IsEnabled: apartment.IsEnabled == 1,
	}), nil
}
