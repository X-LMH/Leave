package admin

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"errors"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
)

const (
	maxRecordCourseLength      = 255
	maxRecordTravelWayLength   = 32
	maxRecordDestinationLength = 255
	maxRecordReasonBytes       = 65535 // MySQL TEXT stores at most 65535 bytes, not characters.
)

func normalizeAdminRecord(p *dto.AdminRecordUpdateRequest) error {
	p.LeaveReason = strings.TrimSpace(p.LeaveReason)
	p.AffectedCourse = strings.TrimSpace(p.AffectedCourse)
	p.TravelWay = strings.TrimSpace(p.TravelWay)
	p.Destination = strings.TrimSpace(p.Destination)
	if p.LeaveTypeID == 0 || p.IsLeaveSchool == nil || p.LeaveReason == "" ||
		p.StartTime.IsZero() || p.EndTime.IsZero() || p.AppliedAt.IsZero() || p.ApprovedAt.IsZero() ||
		!p.EndTime.After(p.StartTime) || p.ApprovedAt.Before(p.AppliedAt) {
		return ErrorInvalidAdminInput
	}
	if !*p.IsLeaveSchool {
		p.TravelWay = ""
		p.Destination = ""
	} else if p.Destination == "" {
		return ErrorInvalidAdminInput
	}
	if utf8.RuneCountInString(p.AffectedCourse) > maxRecordCourseLength ||
		utf8.RuneCountInString(p.TravelWay) > maxRecordTravelWayLength ||
		utf8.RuneCountInString(p.Destination) > maxRecordDestinationLength || len(p.LeaveReason) > maxRecordReasonBytes {
		return ErrorInvalidAdminInput
	}
	return nil
}

func adminRecordResponse(row *models.Record) *dto.AdminRecord {
	return &dto.AdminRecord{
		AdminRecordListItem: dto.AdminRecordListItem{
			ID:            row.ID,
			StudentID:     row.StudentID,
			Name:          row.Name,
			ClassName:     row.ClassName,
			LeaveTypeID:   row.LeaveTypeID,
			LeaveTypeName: row.LeaveTypeName,
			StartTime:     row.StartTime,
			EndTime:       row.EndTime,
			IsLeaveSchool: row.IsLeaveSchool,
			AppliedAt:     row.AppliedAt,
		},
		Gender:         row.Gender,
		College:        row.College,
		Major:          row.Major,
		ParentName:     row.ParentName,
		ParentPhone:    row.ParentPhone,
		TeacherName:    row.TeacherName,
		AffectedCourse: row.AffectedCourse,
		LeaveReason:    row.LeaveReason,
		TravelWay:      row.TravelWay,
		Destination:    row.Destination,
		ApprovedAt:     row.ApprovedAt,
		CreatedAt:      row.CreatedAt,
	}
}

func GetAdminRecords(query dto.AdminRecordQuery) (*dto.AdminRecordListResponse, error) {
	rows, total, err := mysql.GetAdminRecords(query)
	if err != nil {
		return nil, err
	}
	return &dto.AdminRecordListResponse{
		Items:    rows,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func GetAdminRecord(id uint) (*dto.AdminRecord, error) {
	row, err := mysql.GetAdminRecord(id)
	if err != nil {
		return nil, err
	}
	return adminRecordResponse(row), nil
}

func UpdateAdminRecord(id uint, input *dto.AdminRecordUpdateRequest) (*dto.AdminRecord, error) {
	if err := normalizeAdminRecord(input); err != nil {
		return nil, err
	}
	var result *dto.AdminRecord
	err := mysql.Transaction(func(tx *gorm.DB) error {
		row, err := mysql.GetAdminRecordForUpdate(tx, id)
		if err != nil {
			return err
		}
		// Unchanged types retain the historical name even after renaming or deletion.
		if input.LeaveTypeID != row.LeaveTypeID {
			leaveType, err := mysql.GetRecordLeaveTypeForShare(tx, input.LeaveTypeID)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrorInvalidAdminInput
			}
			if err != nil {
				return err
			}
			if !leaveType.IsEnabled {
				return ErrorInvalidAdminInput
			}
			row.LeaveTypeID = leaveType.ID
			row.LeaveTypeName = leaveType.Name
		}
		row.StartTime = input.StartTime
		row.EndTime = input.EndTime
		row.AffectedCourse = input.AffectedCourse
		row.IsLeaveSchool = *input.IsLeaveSchool
		row.LeaveReason = input.LeaveReason
		row.TravelWay = input.TravelWay
		row.Destination = input.Destination
		row.AppliedAt = input.AppliedAt
		row.ApprovedAt = &input.ApprovedAt
		if err := mysql.UpdateAdminRecordContent(tx, row); err != nil {
			return err
		}
		result = adminRecordResponse(row)
		return nil
	})
	return result, err
}
