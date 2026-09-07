package models

type VersionInfo struct {
	Latest      string   `json:"latest"`
	DownloadURL string   `json:"downloadURL"`
	UpdateLog   []string `json:"update_log"`
}
