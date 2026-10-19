package admin

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/service"
	"backend/internal/utils/file"
	passwordutil "backend/internal/utils/password"
	"backend/internal/utils/validator"
	"errors"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
)

type StudentService struct {
	storage config.StorageConfig
	dao     *mysql.DAO
}

func NewStudentService(dao *mysql.DAO, storage config.StorageConfig) *StudentService {
	return &StudentService{
		storage: storage,
		dao:     dao,
	}
}

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
		!models.IsValidGender(p.Gender) || p.ClassID == 0 ||
		utf8.RuneCountInString(p.DormitoryNumber) > 32 {
		return ErrorInvalidAdminInput
	}
	return nil
}

func (s *StudentService) studentResponse(user *models.User, profile *models.Profile) (*dto.Student, error) {
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
	url, err := file.AccessURL(s.storage.BaseURL, avatarPath)
	if err != nil {
		return nil, err
	}
	row.AvatarURL = url
	return row, nil
}

func (s *StudentService) GetAdminStudents(query dto.StudentQuery) (*dto.StudentListResponse, error) {
	items, total, err := s.dao.GetStudents(query)
	if err != nil {
		return nil, err
	}
	versions, err := s.dao.GetAppVersions("android", nil)
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

func (s *StudentService) GetAdminStudent(id uint) (*dto.Student, error) {
	user, err := s.dao.GetStudent(id)
	if err != nil {
		return nil, err
	}
	profile, err := s.dao.GetProfileByStuID(user.StudentID)
	if err != nil {
		return nil, err
	}
	row, err := s.studentResponse(user, profile)
	if err != nil {
		return nil, err
	}
	versions, err := s.dao.GetAppVersions("android", nil)
	if err != nil {
		return nil, err
	}
	row.AppVersionStatus = appVersionState(row.AppVersion, versions)
	return row, nil
}

func (s *StudentService) UpdateAdminStudent(id uint, input *dto.ProfileRequest) (*dto.Student, error) {
	if err := normalizeAdminStudent(input); err != nil {
		return nil, err
	}
	// Validate locked reference data inside the transaction so concurrent changes
	// cannot invalidate the rules between validation and saving.
	if err := s.dao.Transaction(func(tx *gorm.DB) error {
		user, err := s.dao.GetStudentForUpdate(tx, id)
		if err != nil {
			return err
		}
		current, err := s.dao.GetStudentProfileForUpdate(tx, user.StudentID)
		if err != nil {
			return err
		}
		class, err := s.dao.GetStudentClassForShare(tx, input.ClassID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorInvalidAdminInput
		}
		if err != nil {
			return err
		}
		var apartment *models.Apartment
		if input.ApartmentID != 0 {
			apartment, err = s.dao.GetStudentApartmentForShare(tx, input.ApartmentID)
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
		return s.dao.SaveStudentProfile(tx, profile)
	}); err != nil {
		return nil, err
	}
	return s.GetAdminStudent(id)
}

func (s *StudentService) UpdateAdminStudentStatus(id uint, input *dto.StudentStatusRequest) (*dto.Student, error) {
	if input.Status == nil || *input.Status > models.UserStatusActive {
		return nil, ErrorInvalidAdminInput
	}
	if err := s.dao.UpdateStudentStatus(id, *input.Status); err != nil {
		return nil, err
	}
	return s.GetAdminStudent(id)
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

const defaultStudentPassword = "12345678"

// ResetAdminStudentPassword only changes the explicitly selected student account.
func (s *StudentService) ResetAdminStudentPassword(id uint) error {
	hash, err := passwordutil.Hash(defaultStudentPassword)
	if err != nil {
		return err
	}
	return s.dao.Transaction(func(tx *gorm.DB) error {
		user, err := s.dao.GetStudentForUpdate(tx, id)
		if err != nil {
			return err
		}
		return s.dao.UpdateUserPassword(tx, user.ID, hash)
	})
}
