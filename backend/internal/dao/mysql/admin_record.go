package mysql

import (
	"backend/internal/dto"
	"backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func GetAdminRecords(query dto.AdminRecordQuery) ([]*dto.AdminRecordListItem, int64, error) {
	q := db.Model(&models.Record{})
	if query.StudentID != "" {
		q = q.Where("student_id LIKE ?", "%"+query.StudentID+"%")
	}
	if query.Name != "" {
		q = q.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if classFilter := query.ClassFilter; classFilter != nil {
		q = q.Where("college = ? AND major = ? AND class_name = ?", classFilter.College, classFilter.Major, classFilter.ClassName)
	}
	if query.LeaveTypeID != 0 {
		q = q.Where("leave_type_id = ?", query.LeaveTypeID)
	}
	if query.IsLeaveSchool != nil {
		q = q.Where("is_leave_school = ?", *query.IsLeaveSchool)
	}
	if query.StartTimeFrom != nil {
		q = q.Where("start_time >= ?", query.StartTimeFrom.In(time.Local))
	}
	if query.StartTimeTo != nil {
		q = q.Where("start_time < ?", query.StartTimeTo.In(time.Local))
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]*dto.AdminRecordListItem, 0)
	err := q.Select(
		"id",
		"student_id",
		"name",
		"class_name",
		"leave_type_id",
		"leave_type_name",
		"start_time",
		"end_time",
		"is_leave_school",
		"applied_at",
	).
		Order("applied_at DESC, id DESC").
		Scopes(Paginate(query.Page, query.PageSize)).
		Scan(&rows).Error
	return rows, total, err
}

func GetAdminRecord(id uint) (*models.Record, error) {
	var row models.Record
	if err := db.First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func GetAdminRecordForUpdate(tx *gorm.DB, id uint) (*models.Record, error) {
	var row models.Record
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func GetRecordLeaveTypeForShare(tx *gorm.DB, id uint) (*models.LeaveType, error) {
	var row models.LeaveType
	if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func UpdateAdminRecordContent(tx *gorm.DB, record *models.Record) error {
	return tx.Model(record).Updates(map[string]any{
		"leave_type_id":   record.LeaveTypeID,
		"leave_type_name": record.LeaveTypeName,
		"start_time":      record.StartTime,
		"end_time":        record.EndTime,
		"affected_course": record.AffectedCourse,
		"is_leave_school": record.IsLeaveSchool,
		"leave_reason":    record.LeaveReason,
		"travel_way":      record.TravelWay,
		"destination":     record.Destination,
		"applied_at":      record.AppliedAt,
		"approved_at":     record.ApprovedAt,
	}).Error
}
