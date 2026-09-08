package dto

// AppVersionResponse is returned to a client before it enters the application.
type AppVersionResponse struct {
	Platform     string   `json:"platform"`
	VersionCode  int      `json:"version_code"`
	VersionName  string   `json:"version_name"`
	DownloadURL  string   `json:"download_url"`
	APKSHA256    string   `json:"apk_sha256"`
	ReleaseNotes []string `json:"release_notes"`
}
