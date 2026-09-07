package service

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"errors"
	"strings"
)

var (
	ErrorInvalidRecord        = errors.New("invalid leave record")
	ErrorLeaveTypeUnavailable = errors.New("leave type unavailable")
	ErrorProfileIncomplete    = errors.New("profile incomplete")
)

func CreateRecord(studentID string, request *dto.RecordCreateRequest) (*dto.RecordResponse, error) {
	if !validateAndNormalizeRecordRequest(request) {
		return nil, ErrorInvalidRecord
	}

	profile, class, err := getProfileAndClass(studentID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrorProfileIncomplete
	}

	leaveType, err := mysql.GetLeaveTypeByID(request.LeaveTypeID)
	if err != nil {
		if errors.Is(err, mysql.ErrorLeaveTypeNotExist) {
			return nil, ErrorLeaveTypeUnavailable
		}
		return nil, err
	}
	if !leaveType.IsEnabled {
		return nil, ErrorLeaveTypeUnavailable
	}
	record := &models.Record{
		StudentID:      studentID,
		Name:           profile.Name,
		LeaveTypeID:    request.LeaveTypeID,
		LeaveTypeName:  leaveType.Name,
		College:        class.College,
		Major:          class.Major,
		ClassName:      class.ClassName,
		StartTime:      request.StartTime,
		EndTime:        request.EndTime,
		Duration:       request.Duration,
		AffectedCourse: request.AffectedCourse,
		IsLeaveSchool:  *request.IsLeaveSchool,
		LeaveReason:    request.LeaveReason,
		TravelWay:      request.TravelWay,
		AppliedAt:      request.AppliedAt,
		ApprovedAt:     &request.ApprovedAt,
	}
	if err := mysql.InsertRecord(record); err != nil {
		return nil, err
	}
	return toRecordResponse(record), nil
}

func GetRecord(studentID string, recordID int) (*dto.RecordResponse, error) {
	record, err := mysql.GetRecordByID(studentID, recordID)
	if err != nil {
		return nil, err
	}
	return toRecordResponse(record), nil
}

func GetRecordsList(studentID string, query dto.RecordListQuery) ([]*dto.RecordListItem, error) {
	records, err := mysql.GetRecordsByStuID(studentID, query.Page, query.PageSize, query.LeaveType)
	if err != nil {
		return nil, err
	}
	items := make([]*dto.RecordListItem, 0, len(records))
	for _, record := range records {
		items = append(items, &dto.RecordListItem{
			ID:          record.ID,
			LeaveType:   record.LeaveTypeName,
			LeaveReason: record.LeaveReason,
			StartTime:   record.StartTime,
			EndTime:     record.EndTime,
			Duration:    record.Duration,
		})
	}
	return items, nil
}

// GetLeaveTypeOptions returns enabled leave types for the leave application form.
func GetLeaveTypeOptions() ([]*dto.LeaveTypeOption, error) {
	leaveTypes, err := mysql.GetEnabledLeaveTypes()
	if err != nil {
		return nil, err
	}

	options := make([]*dto.LeaveTypeOption, 0, len(leaveTypes))
	for _, leaveType := range leaveTypes {
		options = append(options, &dto.LeaveTypeOption{
			ID:   leaveType.ID,
			Name: leaveType.Name,
		})
	}
	return options, nil
}

func DeleteRecord(studentID string, recordID int) error {
	return mysql.DeleteRecordByID(studentID, recordID)
}

func validateAndNormalizeRecordRequest(request *dto.RecordCreateRequest) bool {
	request.AffectedCourse = strings.TrimSpace(request.AffectedCourse)
	request.LeaveReason = strings.TrimSpace(request.LeaveReason)
	request.TravelWay = strings.TrimSpace(request.TravelWay)

	if request.LeaveTypeID == 0 || request.IsLeaveSchool == nil || request.LeaveReason == "" {
		return false
	}
	if request.StartTime.IsZero() || request.EndTime.IsZero() || request.AppliedAt.IsZero() || request.ApprovedAt.IsZero() {
		return false
	}
	if !request.EndTime.After(request.StartTime) || request.ApprovedAt.Before(request.AppliedAt) {
		return false
	}
	return request.Duration > 0
}

func toRecordResponse(record *models.Record) *dto.RecordResponse {
	response := &dto.RecordResponse{
		ID:        record.ID,
		StudentID: record.StudentID,
		Name:      record.Name,
		ClassInfo: dto.ProfileClassInfo{
			College:   record.College,
			Major:     record.Major,
			ClassName: record.ClassName,
		},
		LeaveType:      record.LeaveTypeName,
		StartTime:      record.StartTime,
		EndTime:        record.EndTime,
		Duration:       record.Duration,
		AffectedCourse: record.AffectedCourse,
		IsLeaveSchool:  record.IsLeaveSchool,
		LeaveReason:    record.LeaveReason,
		TravelWay:      record.TravelWay,
		AppliedAt:      record.AppliedAt,
	}
	if record.ApprovedAt != nil {
		response.ApprovedAt = *record.ApprovedAt
	}
	return response
}
