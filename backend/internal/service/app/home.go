package app

import (
	"backend/internal/dto"
	"backend/internal/service"
	"backend/internal/utils/file"
	"time"

	"gorm.io/gorm"
)

func (r *RecordService) GetHome(studentID string) (*dto.HomeResponse, error) {
	return r.getHomeAt(studentID, time.Now())
}

func (r *RecordService) getHomeAt(studentID string, now time.Time) (*dto.HomeResponse, error) {
	// 运行环境统一使用北京时间，按传入时间的时区确定本月范围。
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	result := new(dto.HomeResponse)
	// 用户信息、统计与最新假条共用快照，避免并发修改时返回相互矛盾的数据。
	err := r.dao.OverviewReadTransaction(func(tx *gorm.DB) error {
		var err error
		result.User, err = r.dao.GetHomeUser(tx, studentID)
		if err != nil {
			return err
		}
		result.User.StudentID = studentID
		avatarPath := result.User.AvatarURL
		if avatarPath == "" {
			avatarPath = service.DefaultAvatarPath
		}
		result.User.AvatarURL, err = file.AccessURL(r.profile.storage.BaseURL, avatarPath)
		if err != nil {
			return err
		}
		result.Statistics, err = r.dao.GetHomeStatistics(tx, studentID, monthStart, monthStart.AddDate(0, 1, 0))
		if err != nil {
			return err
		}
		record, err := r.dao.GetHomeLatestRecord(tx, studentID)
		if err != nil {
			return err
		}
		if record != nil {
			days, hours := durationParts(record.StartTime, record.EndTime)
			result.LatestRecord = &dto.HomeRecord{
				ID:            record.ID,
				LeaveType:     record.LeaveTypeName,
				StartTime:     record.StartTime,
				EndTime:       record.EndTime,
				DurationDays:  days,
				DurationHours: hours,
				IsLeaveSchool: record.IsLeaveSchool,
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
