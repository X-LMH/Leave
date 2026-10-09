package admin

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"encoding/json"
	"errors"
	"strings"
)

func normalizeAppVersionQuery(query *dto.AppVersionListQuery) error {
	query.Keyword = strings.TrimSpace(query.Keyword)
	query.Platform = strings.ToLower(strings.TrimSpace(query.Platform))
	query.Status = strings.TrimSpace(query.Status)
	if query.Platform != "" && !models.IsValidAppVersionPlatform(query.Platform) {
		return ErrorInvalidAdminInput
	}
	if query.Status != "" && !models.IsValidAppVersionStatus(query.Status) {
		return ErrorInvalidAdminInput
	}
	return nil
}

func adminAppVersion(row *models.AppVersion) (*dto.AdminAppVersion, error) {
	notes := make([]string, 0)
	if row.ReleaseNotes != "" {
		if err := json.Unmarshal([]byte(row.ReleaseNotes), &notes); err != nil {
			return nil, err
		}
		if notes == nil {
			notes = make([]string, 0)
		}
	}
	return &dto.AdminAppVersion{
		ID:           row.ID,
		Platform:     row.Platform,
		VersionCode:  row.VersionCode,
		VersionName:  row.VersionName,
		PackageFile:  row.PackageFile,
		ReleaseNotes: notes,
		Status:       row.Status,
		PublishedAt:  row.PublishedAt,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}, nil
}

func GetAdminAppVersions(query dto.AppVersionListQuery) (*dto.AppVersionListResponse, error) {
	if err := normalizeAppVersionQuery(&query); err != nil {
		return nil, err
	}
	rows, total, err := mysql.GetAppVersionList(query)
	if err != nil {
		return nil, err
	}
	items := make([]*dto.AdminAppVersion, 0, len(rows))
	for _, row := range rows {
		item, err := adminAppVersion(&row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return &dto.AppVersionListResponse{
		Items:    items,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}

func GetAdminCurrentAppVersion() (*dto.AdminAppVersion, error) {
	row, err := mysql.GetCurrentAppVersion(models.AppVersionPlatformAndroid)
	if errors.Is(err, mysql.ErrorAppVersionNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return adminAppVersion(row)
}

func GetAdminAppVersion(id uint64) (*dto.AdminAppVersion, error) {
	row, err := mysql.GetAppVersionByID(id)
	if err != nil {
		return nil, err
	}
	return adminAppVersion(row)
}
