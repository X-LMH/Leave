package models

import "time"

const (
	AppVersionStatusPublished = "published"
	AppVersionStatusArchived  = "archived"
)

// AppVersion 是一个平台已发布客户端版本的记录。
// ReleaseNotes 以 JSON 数组形式存储，序列化和反序列化由 Service 层处理。
type AppVersion struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	Platform     string    `gorm:"column:platform"`
	VersionCode  int       `gorm:"column:version_code"`
	VersionName  string    `gorm:"column:version_name"`
	PackageFile  string    `gorm:"column:package_file"`
	ReleaseNotes string    `gorm:"column:release_notes"`
	Status       string    `gorm:"column:status"`
	PublishedAt  time.Time `gorm:"column:published_at"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (AppVersion) TableName() string { return "app_versions" }
