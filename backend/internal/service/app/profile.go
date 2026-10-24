package app

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/service"
	"backend/internal/utils/file"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ProfileService struct {
	logger  *zap.Logger
	dao     *mysql.DAO
	storage config.StorageConfig
}

func NewProfileService(dao *mysql.DAO, storage config.StorageConfig, logger *zap.Logger) *ProfileService {
	return &ProfileService{
		logger:  logger,
		dao:     dao,
		storage: storage,
	}
}

var ErrorInvalidProfileReference = errors.New("invalid profile reference")

func (p *ProfileService) Profile(input *dto.ProfileRequest, studentID string) (err error) {
	student := &models.Profile{
		StudentID:       studentID,
		ClassID:         input.ClassID,
		Name:            input.Name,
		Phone:           input.Phone,
		Gender:          input.Gender,
		ParentName:      input.ParentName,
		ParentPhone:     input.ParentPhone,
		ApartmentID:     input.ApartmentID,
		DormitoryNumber: input.DormitoryNumber,
		TeacherName:     input.TeacherName,
	}
	// Validate class and apartment assignments under the same locks as management edits.
	return p.dao.Transaction(func(tx *gorm.DB) error {
		current, err := p.dao.GetStudentProfileForUpdate(tx, studentID)
		if err != nil {
			return err
		}
		class, err := p.dao.GetStudentClassForShare(tx, input.ClassID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrorInvalidProfileReference
			}
			return err
		}
		if err := validateProfileClass(current, class); err != nil {
			return err
		}
		if input.ApartmentID != 0 {
			apartment, err := p.dao.GetStudentApartmentForShare(tx, input.ApartmentID)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrorInvalidProfileReference
			}
			if err != nil {
				return err
			}
			if err := validateProfileApartment(input, current, apartment); err != nil {
				return err
			}
		}
		return p.dao.SaveAppProfile(tx, student)
	})
}

func (p *ProfileService) GetProfile(studentID string) (*dto.ProfileResponse, error) {
	profile, class, err := p.getProfileAndClass(studentID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrorProfileIncomplete
	}

	var apartment *models.Apartment
	if profile.ApartmentID != 0 {
		apartment, err = p.dao.GetApartmentByID(profile.ApartmentID)
		if err != nil {
			return nil, err
		}
	}

	data := toProfileResponse(profile, class, apartment)
	avatarPath := profile.AvatarURL
	if avatarPath == "" {
		avatarPath = service.DefaultAvatarPath
	}
	data.AvatarURL, err = file.AccessURL(p.storage.BaseURL, avatarPath)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (p *ProfileService) getProfileAndClass(studentID string) (*models.Profile, *models.Class, error) {
	profile, err := p.dao.GetProfileByStuID(studentID)
	if err != nil || profile == nil {
		return profile, nil, err
	}
	class, err := p.dao.GetClassByID(profile.ClassID)
	if err != nil {
		return nil, nil, err
	}
	return profile, class, nil
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

func validateProfileClass(current *models.Profile, class *models.Class) error {
	if !class.IsEnabled && class.ID != current.ClassID {
		return ErrorInvalidProfileReference
	}
	return nil
}

func validateProfileApartment(input *dto.ProfileRequest, current *models.Profile, apartment *models.Apartment) error {
	if apartment.Gender != input.Gender || (apartment.IsEnabled != 1 && apartment.ID != current.ApartmentID) {
		return ErrorInvalidProfileReference
	}
	return nil
}
