package service

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"encoding/json"
	"errors"
	"strings"
)

var ErrAppVersionNotFound = errors.New("未找到已发布的应用版本")

// GetCurrentAppVersion returns the currently usable version for a platform.
func GetCurrentAppVersion(platform string) (*dto.AppVersionResponse, error) {
	version, err := GetCurrentAppPackage(platform)
	if err != nil {
		return nil, err
	}
	return appVersionResponse(version), nil
}

// GetCurrentAppPackage returns the current version record used to serve its package.
func GetCurrentAppPackage(platform string) (*models.AppVersion, error) {
	platform = strings.ToLower(strings.TrimSpace(platform))
	version, err := mysql.GetCurrentAppVersion(platform)
	if errors.Is(err, mysql.ErrorAppVersionNotFound) {
		return nil, ErrAppVersionNotFound
	}
	return version, err
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
		ReleaseNotes: releaseNotes,
	}
}
