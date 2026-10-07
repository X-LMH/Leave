package admin

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/service"
	"backend/internal/utils/file"
	"backend/internal/utils/validator"
	"errors"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
)

func normalizeAdminStudent(p *dto.ProfileRequest) error {
	p.Name = strings.TrimSpace(p.Name)
	p.Phone = strings.TrimSpace(p.Phone)
	p.Gender = strings.TrimSpace(p.Gender)
	p.ParentName = strings.TrimSpace(p.ParentName)
	p.ParentPhone = strings.TrimSpace(p.ParentPhone)
	p.TeacherName = strings.TrimSpace(p.TeacherName)
	p.DormitoryNumber = strings.TrimSpace(p.DormitoryNumber)
	if !validAdminText(p.Name, 64) || !validAdminText(p.ParentName, 64) || !validAdminText(p.TeacherName, 64) ||
		!validator.IsMainlandMobile(p.Phone) || !validator.IsMainlandMobile(p.ParentPhone) ||
		(p.Gender != models.GenderMale && p.Gender != models.GenderFemale) || p.ClassID == 0 ||
		utf8.RuneCountInString(p.DormitoryNumber) > 32 {
		return ErrorInvalidAdminInput
	}
	return nil
}

func studentResponse(user *models.User, profile *models.Profile) (*dto.Student, error) {
	row := &dto.Student{
		ID:             user.ID,
		StudentID:      user.StudentID,
		Status:         user.Status,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
		LastSeenAt:     user.LastSeenAt,
		LastSeenDevice: user.LastSeenDevice,
		AppVersion:     user.AppVersion,
	}
	avatarPath := service.DefaultAvatarPath
	if profile != nil {
		row.ProfileRequest = dto.ProfileRequest{
			Name:            profile.Name,
			Phone:           profile.Phone,
			Gender:          profile.Gender,
			ParentName:      profile.ParentName,
			ParentPhone:     profile.ParentPhone,
			ClassID:         profile.ClassID,
			ApartmentID:     profile.ApartmentID,
			DormitoryNumber: profile.DormitoryNumber,
			TeacherName:     profile.TeacherName,
		}
		if profile.AvatarURL != "" {
			avatarPath = profile.AvatarURL
		}
	}
	url, err := file.AccessURL(config.Cfg.Storage.BaseURL, avatarPath)
	if err != nil {
		return nil, err
	}
	row.AvatarURL = url
	return row, nil
}

func GetAdminStudents(query dto.StudentQuery) (*dto.StudentListResponse, error) {
	items, total, err := mysql.GetStudents(query)
	if err != nil {
		return nil, err
	}
	versions, err := mysql.GetAppVersions("android", nil)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		item.AppVersionStatus = appVersionState(item.AppVersion, versions)
	}
	return &dto.StudentListResponse{
		Items:    items,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func GetAdminStudent(id uint) (*dto.Student, error) {
	user, err := mysql.GetStudent(id)
	if err != nil {
		return nil, err
	}
	profile, err := mysql.GetProfileByStuID(user.StudentID)
	if err != nil {
		return nil, err
	}
	row, err := studentResponse(user, profile)
	if err != nil {
		return nil, err
	}
	versions, err := mysql.GetAppVersions("android", nil)
	if err != nil {
		return nil, err
	}
	row.AppVersionStatus = appVersionState(row.AppVersion, versions)
	return row, nil
}

func UpdateAdminStudent(id uint, input *dto.ProfileRequest) (*dto.Student, error) {
	if err := normalizeAdminStudent(input); err != nil {
		return nil, err
	}
	// Validate locked reference data inside the transaction so concurrent changes
	// cannot invalidate the rules between validation and saving.
	if err := mysql.Transaction(func(tx *gorm.DB) error {
		user, err := mysql.GetStudentForUpdate(tx, id)
		if err != nil {
			return err
		}
		current, err := mysql.GetStudentProfileForUpdate(tx, user.StudentID)
		if err != nil {
			return err
		}
		class, err := mysql.GetStudentClassForShare(tx, input.ClassID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorInvalidAdminInput
		}
		if err != nil {
			return err
		}
		var apartment *models.Apartment
		if input.ApartmentID != 0 {
			apartment, err = mysql.GetStudentApartmentForShare(tx, input.ApartmentID)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrorInvalidAdminInput
			}
			if err != nil {
				return err
			}
		}
		if err := validateStudentReferences(input, current, class, apartment); err != nil {
			return err
		}
		profile := studentProfile(user.StudentID, input)
		return mysql.SaveStudentProfile(tx, profile)
	}); err != nil {
		return nil, err
	}
	return GetAdminStudent(id)
}

func UpdateAdminStudentStatus(id uint, input *dto.StudentStatusRequest) (*dto.Student, error) {
	if input.Status == nil || *input.Status > models.UserStatusActive {
		return nil, ErrorInvalidAdminInput
	}
	if err := mysql.UpdateStudentStatus(id, *input.Status); err != nil {
		return nil, err
	}
	return GetAdminStudent(id)
}

func validateStudentReferences(input *dto.ProfileRequest, current *models.Profile, class *models.Class, apartment *models.Apartment) error {
	if !class.IsEnabled && class.ID != current.ClassID {
		return ErrorInvalidAdminInput
	}
	if apartment == nil {
		return nil
	}
	if apartment.Gender != input.Gender {
		return ErrorInvalidAdminInput
	}
	if apartment.IsEnabled != 1 && apartment.ID != current.ApartmentID {
		return ErrorInvalidAdminInput
	}
	return nil
}

func studentProfile(studentID string, input *dto.ProfileRequest) *models.Profile {
	return &models.Profile{
		StudentID:       studentID,
		Name:            input.Name,
		Phone:           input.Phone,
		Gender:          input.Gender,
		ClassID:         input.ClassID,
		ParentName:      input.ParentName,
		ParentPhone:     input.ParentPhone,
		TeacherName:     input.TeacherName,
		ApartmentID:     input.ApartmentID,
		DormitoryNumber: input.DormitoryNumber,
	}
}
