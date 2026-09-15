package mysql

import (
	"backend/internal/models"
	"errors"

	"gorm.io/gorm"
)

var ErrorAppVersionNotFound = errors.New("未找到已发布的应用版本")

// GetCurrentAppVersion returns the highest published version for one platform.
func GetCurrentAppVersion(platform string) (*models.AppVersion, error) {
	version := new(models.AppVersion)
	err := db.Where("platform = ? AND status = ?", platform, models.AppVersionStatusPublished).
		Order("version_code DESC").
		First(version).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrorAppVersionNotFound
	}
	if err != nil {
		return nil, err
	}
	return version, nil
}

func GetAppVersionUpdates(platform string, versionCode int) ([]models.AppVersion, error) {
	var versions []models.AppVersion
	err := db.Where("platform = ? AND version_code > ? AND status IN ?", platform, versionCode, []string{models.AppVersionStatusPublished, models.AppVersionStatusArchived}).
		Order("version_code ASC").Find(&versions).Error
	return versions, err
}
