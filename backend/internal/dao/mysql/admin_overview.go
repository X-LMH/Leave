package mysql

import (
	"backend/internal/dto"
	"backend/internal/models"
	"database/sql"
	"time"

	"gorm.io/gorm"
)

// 使用可重复读的只读事务，让总量、最近记录和版本分组读取同一数据库快照。
func (s *DAO) OverviewReadTransaction(fn func(*gorm.DB) error) error {
	return s.db.Transaction(fn, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
}

type OverviewReportedVersion struct {
	Version string
	Count   int64
}

func (s *DAO) GetOverviewStatistics(tx *gorm.DB, now time.Time) (dto.OverviewStatistics, error) {
	// 学生从账号表统计，包含停用和未完善资料的账号；班级也包含停用项。
	// 带 DeletedAt 的模型由 GORM 自动排除软删除记录。
	var counts dto.OverviewStatistics
	if err := tx.Model(&models.User{}).
		Where("role = ?", models.RoleStudent).
		Count(&counts.StudentCount).Error; err != nil {
		return counts, err
	}
	if err := tx.Model(&models.Class{}).Count(&counts.ClassCount).Error; err != nil {
		return counts, err
	}
	if err := tx.Model(&models.Record{}).Count(&counts.RecordCount).Error; err != nil {
		return counts, err
	}
	if err := tx.Model(&models.Feedback{}).
		Where("is_handled = ?", false).
		Count(&counts.UnhandledFeedbackCount).Error; err != nil {
		return counts, err
	}
	return counts, nil
}

type OverviewWeekCount struct {
	WeekIndex int
	Count     int64
}

// 北京时间没有夏令时变化，一周固定为 168 小时。
const overviewWeekDuration = 7 * 24 * time.Hour

// Service 按北京时间计算周边界，SQL 参数转换为现有 DATETIME 存储使用的本地时区。
func (s *DAO) GetOverviewWeekCounts(tx *gorm.DB, start, next, now time.Time) ([]OverviewWeekCount, error) {
	rows := make([]OverviewWeekCount, 0)
	// 按记录创建时间归周，周范围左闭右开；排除统计时间之后创建的记录。
	err := tx.Model(&models.Record{}).
		Select(
			"FLOOR(TIMESTAMPDIFF(SECOND, ?, created_at) / ?) AS week_index, COUNT(*) AS count",
			start.In(time.Local),
			int64(overviewWeekDuration/time.Second),
		).
		Where("created_at >= ? AND created_at < ? AND created_at <= ?", start.In(time.Local), next.In(time.Local), now.In(time.Local)).
		Group("week_index").
		Scan(&rows).Error
	return rows, err
}

func (s *DAO) GetOverviewRecentRecords(tx *gorm.DB, now time.Time) ([]dto.OverviewRecord, error) {
	// 使用记录中的姓名、班级和类型快照，不关联当前资料，避免历史内容随资料修改而变化。
	rows := make([]dto.OverviewRecord, 0)
	err := tx.Model(&models.Record{}).
		Select("id", "name", "student_id", "class_name", "leave_type_id", "leave_type_name", "applied_at").
		Where("created_at <= ?", now.In(time.Local)).
		Order("created_at DESC, id DESC").
		Limit(5).
		Scan(&rows).Error
	return rows, err
}

func (s *DAO) GetOverviewReportedVersions(tx *gorm.DB) ([]OverviewReportedVersion, error) {
	rows := make([]OverviewReportedVersion, 0)
	// 使用二进制分组，避免数据库忽略大小写而把未知版本合并到已知版本；空白由 Service 处理。
	err := tx.Model(&models.User{}).
		Select("BINARY app_version AS version, COUNT(*) AS count").
		Where("role = ?", models.RoleStudent).
		Group("BINARY app_version").
		Scan(&rows).Error
	return rows, err
}

func (s *DAO) GetOverviewAndroidVersions(tx *gorm.DB, now time.Time) ([]models.AppVersion, error) {
	rows := make([]models.AppVersion, 0)
	err := tx.Model(&models.AppVersion{}).
		Select("id", "version_code", "version_name", "status", "published_at", "release_notes").
		Where("platform = ? AND status IN ? AND published_at <= ?", models.AppVersionPlatformAndroid,
			[]string{models.AppVersionStatusPublished, models.AppVersionStatusArchived}, now.In(time.Local)).
		Order("version_code DESC, id DESC").
		Find(&rows).Error
	return rows, err
}
