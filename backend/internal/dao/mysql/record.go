package mysql

import (
	"backend/internal/models"
	"errors"

	"gorm.io/gorm"
)

func InsertRecord(record *models.Record) error {
	if err := db.Create(record).Error; err != nil {
		return ErrorSubmitRecord
	}
	return nil
}

func GetLeaveTypeByID(leaveTypeID uint) (*models.LeaveType, error) {
	leaveType := new(models.LeaveType)
	if err := db.Where("id = ?", leaveTypeID).First(leaveType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrorLeaveTypeNotExist
		}
		return nil, err
	}
	return leaveType, nil
}

func GetEnabledLeaveTypes() ([]*models.LeaveType, error) {
	leaveTypes := make([]*models.LeaveType, 0)
	err := db.Select("id", "name", "sort_order").
		Where("is_enabled = ?", 1).
		Order("sort_order ASC, id ASC").
		Find(&leaveTypes).Error
	return leaveTypes, err
}

func GetRecordByID(studentID string, recordID int) (*models.Record, error) {
	record := new(models.Record)
	err := db.Where("id = ? AND student_id = ?", recordID, studentID).First(record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrorRecordNotExist
		}
		return nil, err
	}
	return record, nil
}

func GetRecordsByStuID(studentID string, page, pageSize int, leaveTypeID uint) ([]*models.Record, error) {
	records := make([]*models.Record, 0)
	query := db.Where("student_id = ?", studentID)
	if leaveTypeID > 0 {
		query = query.Where("leave_type_id = ?", leaveTypeID)
	}
	err := query.
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&records).Error
	return records, err
}

func DeleteRecordByID(studentID string, recordID int) error {
	result := db.Where("id = ? AND student_id = ?", recordID, studentID).Delete(&models.Record{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrorRecordNotExist
	}
	return nil
}
