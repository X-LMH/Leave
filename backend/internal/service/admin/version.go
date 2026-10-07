package admin

import (
	"backend/internal/models"
	"strings"
)

const (
	appVersionStatusLatest   = "latest"
	appVersionStatusPrevious = "previous"
	appVersionStatusOutdated = "outdated"
	appVersionStatusUnknown  = "unknown"
)

// appVersionState 按实际发布次数判断版本状态，同一平台的版本名应唯一。
func appVersionState(name string, versions []models.AppVersion) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return appVersionStatusUnknown
	}
	var latestCode int
	var matchedCode int
	for _, version := range versions {
		if version.Status == models.AppVersionStatusPublished && version.VersionCode > latestCode {
			latestCode = version.VersionCode
		}
		if strings.TrimSpace(version.VersionName) == name {
			matchedCode = version.VersionCode
		}
	}
	if latestCode == 0 || matchedCode == 0 || matchedCode > latestCode {
		return appVersionStatusUnknown
	}
	if matchedCode == latestCode {
		return appVersionStatusLatest
	}

	behind := 0
	for _, version := range versions {
		if version.VersionCode <= matchedCode || version.VersionCode > latestCode {
			continue
		}
		behind++
		if behind > 5 {
			return appVersionStatusOutdated
		}
	}
	return appVersionStatusPrevious
}
