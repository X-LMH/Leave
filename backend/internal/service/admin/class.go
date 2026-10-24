package admin

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"strings"

	"go.uber.org/zap"
)

type ClassService struct {
	logger *zap.Logger
	dao    *mysql.DAO
}

func NewClassService(dao *mysql.DAO, logger *zap.Logger) *ClassService {
	return &ClassService{
		logger: logger,
		dao:    dao,
	}
}

const (
	maxCollegeNameLength = 100
	maxMajorNameLength   = 100
	maxClassNameLength   = 64
)

func (c *ClassService) GetAdminClasses(query dto.ClassListQuery) (*dto.ClassListResponse, error) {
	classes, total, err := c.dao.GetClassList(query)
	if err != nil {
		return nil, err
	}

	items := make([]*dto.ClassListItem, 0, len(classes))
	for _, class := range classes {
		items = append(items, &dto.ClassListItem{
			ID:        class.ID,
			College:   class.College,
			Major:     class.Major,
			ClassName: class.ClassName,
			IsEnabled: class.IsEnabled,
			CreatedAt: class.CreatedAt,
			UpdatedAt: class.UpdatedAt,
		})
	}
	return &dto.ClassListResponse{
		Items:    items,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func normalizeClass(p *dto.ClassRequest) error {
	p.College = strings.TrimSpace(p.College)
	p.Major = strings.TrimSpace(p.Major)
	p.ClassName = strings.TrimSpace(p.ClassName)
	if !validAdminText(p.College, maxCollegeNameLength) ||
		!validAdminText(p.Major, maxMajorNameLength) ||
		!validAdminText(p.ClassName, maxClassNameLength) ||
		p.IsEnabled == nil {
		return ErrorInvalidAdminInput
	}
	return nil
}

func classItem(row *models.Class) *dto.ClassListItem {
	return &dto.ClassListItem{
		ID:        row.ID,
		College:   row.College,
		Major:     row.Major,
		ClassName: row.ClassName,
		IsEnabled: row.IsEnabled,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func (c *ClassService) CreateAdminClass(p *dto.ClassRequest) (*dto.ClassListItem, error) {
	if err := normalizeClass(p); err != nil {
		return nil, err
	}
	row := &models.Class{
		College:   p.College,
		Major:     p.Major,
		ClassName: p.ClassName,
		IsEnabled: *p.IsEnabled,
	}
	if err := c.dao.CreateClass(row); err != nil {
		return nil, err
	}
	return classItem(row), nil
}

func (c *ClassService) UpdateAdminClass(p *dto.ClassRequest, id uint) (*dto.ClassListItem, error) {
	if err := normalizeClass(p); err != nil {
		return nil, err
	}
	row := &models.Class{
		ID:        id,
		College:   p.College,
		Major:     p.Major,
		ClassName: p.ClassName,
		IsEnabled: *p.IsEnabled,
	}
	if err := c.dao.UpdateClass(row); err != nil {
		return nil, err
	}
	return classItem(row), nil
}

func (c *ClassService) DeleteAdminClass(id uint) error {
	return c.dao.DeleteClass(id)
}
