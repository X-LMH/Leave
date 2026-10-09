package mysql

import (
	"backend/internal/dto"
	"backend/internal/models"
	"errors"

	"gorm.io/gorm"
)

var ErrorAppVersionNotFound = errors.New("未找到已发布的应用版本")

func GetAppVersionByID(id uint64) (*models.AppVersion, error) {
	version := new(models.AppVersion)
	if err := db.Where("id = ?", id).First(version).Error; err != nil {
		return nil, err
	}
	return version, nil
}

func GetAppVersionList(query dto.AppVersionListQuery) ([]models.AppVersion, int64, error) {
	dbQuery := db.Model(&models.AppVersion{})
	if query.Platform != "" {
		dbQuery = dbQuery.Where("platform = ?", query.Platform)
	}
	if query.Status != "" {
		dbQuery = dbQuery.Where("status = ?", query.Status)
	}
	if query.Keyword != "" {
		pattern := "%" + query.Keyword + "%"
		dbQuery = dbQuery.Where("(version_name LIKE ? OR CAST(version_code AS CHAR) LIKE ?)", pattern, pattern)
	}
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	versions := make([]models.AppVersion, 0)
	err := dbQuery.Order("version_code DESC, id DESC").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).Find(&versions).Error
	return versions, total, err
}

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

// GetAppVersions returns complete version records, newest first.
// Empty platform or statuses leaves that filter unrestricted for version management.
func GetAppVersions(platform string, statuses []string) ([]models.AppVersion, error) {
	query := db.Model(&models.AppVersion{})
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}

	var versions []models.AppVersion
	err := query.Order("version_code DESC").Order("id DESC").Find(&versions).Error
	return versions, err
}
