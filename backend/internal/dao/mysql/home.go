package mysql

import (
	"backend/internal/dto"
	"backend/internal/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

func (s *DAO) GetHomeUser(tx *gorm.DB, studentID string) (dto.HomeUser, error) {
	var user dto.HomeUser
	err := tx.Model(&models.Profile{}).
		Select("profiles.name", "profiles.avatar_url", "COALESCE(classes.class_name, '') AS class_name").
		Joins("LEFT JOIN classes ON classes.id = profiles.class_id").
		Where("profiles.student_id = ?", studentID).
		Scan(&user).Error
	return user, err
}

func (s *DAO) GetHomeStatistics(tx *gorm.DB, studentID string, monthStart, nextMonth time.Time) (dto.HomeStatistics, error) {
	var result dto.HomeStatistics
	// DATETIME 沿用连接的 Local 时区，本月次数按请假开始时间统计。
	err := tx.Model(&models.Record{}).
		Select(
			`COUNT(*) AS total_count,
			 COALESCE(SUM(CASE WHEN start_time >= ? AND start_time < ? THEN 1 ELSE 0 END), 0) AS month_count`,
			monthStart.In(time.Local),
			nextMonth.In(time.Local),
		).
		Where("student_id = ?", studentID).
		Scan(&result).Error
	return result, err
}

func (s *DAO) GetHomeLatestRecord(tx *gorm.DB, studentID string) (*models.Record, error) {
	var record models.Record
	err := tx.
		Select("id", "leave_type_name", "start_time", "end_time", "is_leave_school").
		Where("student_id = ?", studentID).
		Order("created_at DESC, id DESC").
		Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}
