package service

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
)

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
	profile, err := mysql.GetProfileByStuID(studentID)
	if err != nil || profile == nil {
		return nil, err
	}

	class, err := mysql.GetClassByID(profile.ClassID)
	if err != nil {
		return nil, err
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

// CreateRecord 待请假记录接口适配新版 Record 模型后实现。
func CreateRecord(_ string, _ *models.ParamRecord) error {
	return mysql.ErrorSubmitRecord
}

func GetRecord(recordID int) (*models.Record, error) {
	return mysql.GetRecordByID(recordID)
}

// GetRecordsList 待请假记录列表响应适配新版 Record 模型后实现。
func GetRecordsList(_ string) ([]*ResRecord, error) {
	return nil, mysql.ErrorRecordNotExist
}

func DeleteRecord(recordID int) error {
	return mysql.DeleteRecordByID(recordID)
}
