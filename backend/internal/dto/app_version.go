package dto

// AppVersionResponse is returned to a client before it enters the application.
type AppVersionResponse struct {
	Platform     string   `json:"platform"`
	VersionCode  int      `json:"version_code"`
	VersionName  string   `json:"version_name"`
	ReleaseNotes []string `json:"release_notes"`
}
