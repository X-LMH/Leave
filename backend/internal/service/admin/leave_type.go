package admin

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"math"
	"strings"

	"go.uber.org/zap"
)

type LeaveTypeService struct {
	logger *zap.Logger
	dao    *mysql.DAO
}

func NewLeaveTypeService(dao *mysql.DAO, logger *zap.Logger) *LeaveTypeService {
	return &LeaveTypeService{
		logger: logger,
		dao:    dao,
	}
}

const maxLeaveTypeNameLength = 32

func normalizeLeaveType(p *dto.LeaveTypeRequest) error {
	p.Name = strings.TrimSpace(p.Name)
	if !validAdminText(p.Name, maxLeaveTypeNameLength) ||
		p.SortOrder == nil ||
		*p.SortOrder > math.MaxUint32 ||
		p.IsEnabled == nil {
		return ErrorInvalidAdminInput
	}
	return nil
}

func leaveTypeItem(row *models.LeaveType) *dto.LeaveTypeItem {
	return &dto.LeaveTypeItem{
		ID:        row.ID,
		Name:      row.Name,
		SortOrder: row.SortOrder,
		IsEnabled: row.IsEnabled,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func (l *LeaveTypeService) GetAdminLeaveTypes(query dto.LeaveTypeListQuery) (*dto.LeaveTypeListResponse, error) {
	rows, total, err := l.dao.GetLeaveTypeList(query)
	if err != nil {
		return nil, err
	}
	items := make([]*dto.LeaveTypeItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, leaveTypeItem(row))
	}
	return &dto.LeaveTypeListResponse{
		Items:    items,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func (l *LeaveTypeService) CreateAdminLeaveType(p *dto.LeaveTypeRequest) (*dto.LeaveTypeItem, error) {
	if err := normalizeLeaveType(p); err != nil {
		return nil, err
	}
	row := &models.LeaveType{
		Name:      p.Name,
		SortOrder: uint(*p.SortOrder),
		IsEnabled: *p.IsEnabled,
	}
	if err := l.dao.CreateLeaveType(row); err != nil {
		return nil, err
	}
	return leaveTypeItem(row), nil
}

func (l *LeaveTypeService) UpdateAdminLeaveType(p *dto.LeaveTypeRequest, id uint) (*dto.LeaveTypeItem, error) {
	if err := normalizeLeaveType(p); err != nil {
		return nil, err
	}
	row := &models.LeaveType{
		ID:        id,
		Name:      p.Name,
		SortOrder: uint(*p.SortOrder),
		IsEnabled: *p.IsEnabled,
	}
	if err := l.dao.UpdateLeaveType(row); err != nil {
		return nil, err
	}
	return leaveTypeItem(row), nil
}

func (l *LeaveTypeService) DeleteAdminLeaveType(id uint) error {
	return l.dao.DeleteLeaveType(id)
}
