package dto

// AppVersionResponse is returned to a client before it enters the application.
type AppVersionResponse struct {
	Platform          string             `json:"platform"`
	LatestVersionCode int                `json:"latest_version_code"`
	LatestVersionName string             `json:"latest_version_name"`
	Updates           []AppVersionUpdate `json:"updates"`
}

type AppVersionUpdate struct {
	VersionCode int    `json:"version_code"`
	VersionName string `json:"version_name"`
	// TODO(release-notes-compat): 所有用户均已升级到正式版 1.1.1 或更高版本，
	// 且不再支持 1.1.1 之前的旧安装包直接升级后，
	// 可移除 App 响应的 release_notes，并同步移除 App 对该字段的读取与回退。
	ReleaseNotes        []string             `json:"release_notes"`
	ReleaseNotesGrouped *ReleaseNotesGrouped `json:"release_notes_grouped"`
}

// ReleaseNotesGrouped preserves the author-specified categories.
type ReleaseNotesGrouped struct {
	Features     []string `json:"features"`
	Improvements []string `json:"improvements"`
	BugFixes     []string `json:"bug_fixes"`
	Notices      []string `json:"notices"`
}
