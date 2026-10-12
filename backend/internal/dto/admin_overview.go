package dto

import "time"

type OverviewStatistics struct {
	StudentCount  int64 `json:"student_count"`
	ClassCount    int64 `json:"class_count"`
	RecordCount   int64 `json:"record_count"`
	FeedbackCount int64 `json:"feedback_count"`
}

type OverviewWeek struct {
	Start time.Time `json:"start"`
	// End 表示用于展示的周日日期；查询上界使用下一周周一零点。
	End       time.Time `json:"end"`
	Count     int64     `json:"count"`
	IsCurrent bool      `json:"is_current"`
}

type OverviewWeeklyApplications struct {
	Weeks []OverviewWeek `json:"weeks"`
	Total int64          `json:"total"`
}

type OverviewVersionGroup struct {
	Key        string  `json:"key"`
	Version    string  `json:"version"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

// OverviewVersionUsage 按学生账号最近上报的版本统计，不代表实时在线或安装数量。
type OverviewVersionUsage struct {
	Total       int64                  `json:"total"`
	Groups      []OverviewVersionGroup `json:"groups"`
	LatestCount int64                  `json:"latest_count"`
}

type OverviewRecord struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	StudentID     string    `json:"student_id"`
	ClassName     string    `json:"class_name"`
	LeaveTypeID   uint      `json:"leave_type_id"`
	LeaveTypeName string    `json:"leave_type_name"`
	AppliedAt     time.Time `json:"applied_at"`
}

type OverviewVersion struct {
	VersionName  string    `json:"version_name"`
	VersionCode  int       `json:"version_code"`
	PublishedAt  time.Time `json:"published_at"`
	ReleaseNotes []string  `json:"release_notes"`
}

type AdminOverview struct {
	GeneratedAt          time.Time                  `json:"generated_at"`
	Statistics           OverviewStatistics         `json:"statistics"`
	WeeklyApplications   OverviewWeeklyApplications `json:"weekly_applications"`
	VersionUsage         OverviewVersionUsage       `json:"version_usage"`
	RecentRecords        []OverviewRecord           `json:"recent_records"`
	LatestAndroidVersion *OverviewVersion           `json:"latest_android_version"`
}
