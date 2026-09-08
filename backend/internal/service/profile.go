package service

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"errors"
)

var ErrorInvalidApartmentGender = errors.New("invalid apartment gender")

func Profile(p *dto.ProfileRequest, studentID string) (err error) {
	student := &models.Profile{
		StudentID:       studentID,
		ClassID:         p.ClassID,
		Name:            p.Name,
		Phone:           p.Phone,
		Gender:          p.Gender,
		ParentName:      p.ParentName,
		ParentPhone:     p.ParentPhone,
		ApartmentID:     p.ApartmentID,
		DormitoryNumber: p.DormitoryNumber,
		TeacherName:     p.TeacherName,
	}
	return mysql.FinishProfile(student)
}

func GetProfile(studentID string) (*dto.ProfileResponse, error) {
	profile, class, err := getProfileAndClass(studentID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrorProfileIncomplete
	}

	var apartment *models.Apartment
	if profile.ApartmentID != 0 {
		apartment, err = mysql.GetApartmentByID(profile.ApartmentID)
		if err != nil {
			return nil, err
		}
	}

	return toProfileResponse(profile, class, apartment), nil
}

func getProfileAndClass(studentID string) (*models.Profile, *models.Class, error) {
	profile, err := mysql.GetProfileByStuID(studentID)
	if err != nil || profile == nil {
		return profile, nil, err
	}
	class, err := mysql.GetClassByID(profile.ClassID)
	if err != nil {
		return nil, nil, err
	}
	return profile, class, nil
}

// GetClasses returns class options for profile selectors.
func GetClasses() ([]*dto.ClassOption, error) {
	classes, err := mysql.GetClasses()
	if err != nil {
		return nil, err
	}

	options := make([]*dto.ClassOption, 0, len(classes))
	for _, class := range classes {
		options = append(options, &dto.ClassOption{
			ID: class.ID, College: class.College, Major: class.Major, ClassName: class.ClassName,
		})
	}
	return options, nil
}

// GetApartments returns enabled apartment options for profile selectors.
func GetApartments(gender string) ([]*dto.ApartmentOption, error) {
	if gender != "" && gender != models.GenderMale && gender != models.GenderFemale {
		return nil, ErrorInvalidApartmentGender
	}

	apartments, err := mysql.GetApartments(gender)
	if err != nil {
		return nil, err
	}

	options := make([]*dto.ApartmentOption, 0, len(apartments))
	for _, apartment := range apartments {
		options = append(options, &dto.ApartmentOption{
			ID: apartment.ID, Name: apartment.Name, Gender: apartment.Gender,
		})
	}
	return options, nil
}

func toProfileResponse(profile *models.Profile, class *models.Class, apartment *models.Apartment) *dto.ProfileResponse {
	apartmentName := ""
	if apartment != nil {
		apartmentName = apartment.Name
	}

	return &dto.ProfileResponse{
		StudentID:   profile.StudentID,
		Name:        profile.Name,
		Phone:       profile.Phone,
		Gender:      profile.Gender,
		ParentName:  profile.ParentName,
		ParentPhone: profile.ParentPhone,
		TeacherName: profile.TeacherName,
		ClassID:     profile.ClassID,
		ApartmentID: profile.ApartmentID,
		ClassInfo: dto.ProfileClassInfo{
			College:   class.College,
			Major:     class.Major,
			ClassName: class.ClassName,
		},
		ApartmentInfo: dto.ProfileApartmentInfo{
			ApartmentName:   apartmentName,
			DormitoryNumber: profile.DormitoryNumber,
		},
	}
}
