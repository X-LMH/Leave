package app

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"encoding/json"
	"errors"
	"strings"

	"go.uber.org/zap"
)

type AppVersionService struct {
	logger *zap.Logger
	dao    *mysql.DAO
}

func NewAppVersionService(dao *mysql.DAO, logger *zap.Logger) *AppVersionService {
	return &AppVersionService{
		logger: logger,
		dao:    dao,
	}
}

var ErrAppVersionNotFound = errors.New("未找到已发布的应用版本")

// GetCurrentAppVersion returns the currently usable version for a platform.
func (a *AppVersionService) GetCurrentAppVersion(platform string, currentVersionCode int) (*dto.AppVersionResponse, error) {
	version, err := a.GetCurrentAppPackage(platform)
	if err != nil {
		return nil, err
	}
	updates, err := a.dao.GetAppVersionUpdates(platform, currentVersionCode)
	if err != nil {
		return nil, err
	}
	response := &dto.AppVersionResponse{
		Platform:          version.Platform,
		LatestVersionCode: version.VersionCode,
		LatestVersionName: version.VersionName,
		Updates:           make([]dto.AppVersionUpdate, 0, len(updates)),
	}
	for _, update := range updates {
		response.Updates = append(response.Updates, dto.AppVersionUpdate{
			VersionCode:  update.VersionCode,
			VersionName:  update.VersionName,
			ReleaseNotes: releaseNotes(update.ReleaseNotes),
		})
	}
	return response, nil
}

// GetCurrentAppPackage returns the current version record used to serve its package.
func (a *AppVersionService) GetCurrentAppPackage(platform string) (*models.AppVersion, error) {
	platform = strings.ToLower(strings.TrimSpace(platform))
	version, err := a.dao.GetCurrentAppVersion(platform)
	if errors.Is(err, mysql.ErrorAppVersionNotFound) {
		return nil, ErrAppVersionNotFound
	}
	return version, err
}

// IsCurrentAppVersion reports whether a client meets the minimum supported version.
func (a *AppVersionService) IsCurrentAppVersion(platform string, versionCode int) (*dto.AppVersionResponse, bool, error) {
	current, err := a.GetCurrentAppVersion(platform, versionCode)
	if err != nil {
		return nil, false, err
	}
	return current, versionCode >= current.LatestVersionCode, nil
}

func releaseNotes(value string) []string {
	var notes []string
	if value != "" {
		_ = json.Unmarshal([]byte(value), &notes)
	}
	return notes
}
