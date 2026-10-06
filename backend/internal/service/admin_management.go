package service

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"errors"
	"math"
	"strings"
	"unicode/utf8"
)

var ErrorInvalidAdminInput = errors.New("invalid management input")

const (
	maxLeaveTypeNameLength = 32
	maxCollegeNameLength   = 100
	maxMajorNameLength     = 100
	maxClassNameLength     = 64
)

func validAdminText(text string, max int) bool {
	return text != "" && utf8.RuneCountInString(text) <= max
}
func normalizeLeaveType(p *dto.AdminLeaveTypeRequest) error {
	p.Name = strings.TrimSpace(p.Name)
	if !validAdminText(p.Name, maxLeaveTypeNameLength) ||
		p.SortOrder == nil ||
		*p.SortOrder > math.MaxUint32 ||
		p.IsEnabled == nil {
		return ErrorInvalidAdminInput
	}
	return nil
}
func normalizeClass(p *dto.AdminClassRequest) error {
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
func leaveTypeItem(row *models.LeaveType) *dto.AdminLeaveTypeItem {
	return &dto.AdminLeaveTypeItem{
		ID:        row.ID,
		Name:      row.Name,
		SortOrder: row.SortOrder,
		IsEnabled: row.IsEnabled,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
func classItem(row *models.Class) *dto.AdminClassListItem {
	return &dto.AdminClassListItem{
		ID:        row.ID,
		College:   row.College,
		Major:     row.Major,
		ClassName: row.ClassName,
		IsEnabled: row.IsEnabled,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
func GetAdminLeaveTypes(query dto.AdminLeaveTypeListQuery) (*dto.AdminLeaveTypeListResponse, error) {
	rows, total, err := mysql.GetAdminLeaveTypes(query)
	if err != nil {
		return nil, err
	}
	items := make([]*dto.AdminLeaveTypeItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, leaveTypeItem(row))
	}
	return &dto.AdminLeaveTypeListResponse{
		Items:    items,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}
func CreateAdminLeaveType(p *dto.AdminLeaveTypeRequest) (*dto.AdminLeaveTypeItem, error) {
	if err := normalizeLeaveType(p); err != nil {
		return nil, err
	}
	row := &models.LeaveType{
		Name:      p.Name,
		SortOrder: uint(*p.SortOrder),
		IsEnabled: *p.IsEnabled,
	}
	if err := mysql.CreateAdminLeaveType(row); err != nil {
		return nil, err
	}
	return leaveTypeItem(row), nil
}

func UpdateAdminLeaveType(p *dto.AdminLeaveTypeRequest, id uint) (*dto.AdminLeaveTypeItem, error) {
	if err := normalizeLeaveType(p); err != nil {
		return nil, err
	}
	row := &models.LeaveType{
		ID:        id,
		Name:      p.Name,
		SortOrder: uint(*p.SortOrder),
		IsEnabled: *p.IsEnabled,
	}
	if err := mysql.UpdateAdminLeaveType(row); err != nil {
		return nil, err
	}
	return leaveTypeItem(row), nil
}

func CreateAdminClass(p *dto.AdminClassRequest) (*dto.AdminClassListItem, error) {
	if err := normalizeClass(p); err != nil {
		return nil, err
	}
	row := &models.Class{
		College:   p.College,
		Major:     p.Major,
		ClassName: p.ClassName,
		IsEnabled: *p.IsEnabled,
	}
	if err := mysql.CreateAdminClass(row); err != nil {
		return nil, err
	}
	return classItem(row), nil
}

func UpdateAdminClass(p *dto.AdminClassRequest, id uint) (*dto.AdminClassListItem, error) {
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
	if err := mysql.UpdateAdminClass(row); err != nil {
		return nil, err
	}
	return classItem(row), nil
}

func DeleteAdminLeaveType(id uint) error { return mysql.DeleteAdminLeaveType(id) }
func DeleteAdminClass(id uint) error     { return mysql.DeleteAdminClass(id) }
