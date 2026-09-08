package service

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"encoding/json"
	"strings"
)

// GetCurrentAppVersion returns the currently usable version for a platform.
func GetCurrentAppVersion(platform string) (*dto.AppVersionResponse, error) {
	platform = strings.ToLower(strings.TrimSpace(platform))
	version, err := mysql.GetCurrentAppVersion(platform)
	if err != nil {
		return nil, err
	}
	return appVersionResponse(version), nil
}

// IsCurrentAppVersion reports whether a client is running the current published version.
func IsCurrentAppVersion(platform string, versionCode int) (*dto.AppVersionResponse, bool, error) {
	current, err := GetCurrentAppVersion(platform)
	if err != nil {
		return nil, false, err
	}
	return current, current.VersionCode == versionCode, nil
}

func appVersionResponse(version *models.AppVersion) *dto.AppVersionResponse {
	var releaseNotes []string
	if version.ReleaseNotes != "" {
		_ = json.Unmarshal([]byte(version.ReleaseNotes), &releaseNotes)
	}
	return &dto.AppVersionResponse{
		Platform:     version.Platform,
		VersionCode:  version.VersionCode,
		VersionName:  version.VersionName,
		DownloadURL:  version.DownloadURL,
		APKSHA256:    version.APKSHA256,
		ReleaseNotes: releaseNotes,
	}
}
