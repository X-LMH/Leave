package dto

import "time"

type AppVersionListQuery struct {
	Keyword  string
	Platform string
	Status   string
	Page     int
	PageSize int
}

type AdminAppVersion struct {
	ID           uint64    `json:"id"`
	Platform     string    `json:"platform"`
	VersionCode  int       `json:"version_code"`
	VersionName  string    `json:"version_name"`
	PackageFile  string    `json:"package_file"`
	ReleaseNotes []string  `json:"release_notes"`
	Status       string    `json:"status"`
	PublishedAt  time.Time `json:"published_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AppVersionListResponse struct {
	Items    []*AdminAppVersion `json:"items"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}
