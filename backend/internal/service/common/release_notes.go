package common

import (
	"encoding/json"
	"errors"
	"strings"

	"backend/internal/dto"
)

// ParseReleaseNotes accepts historical arrays and categorized release objects.
func ParseReleaseNotes(value string) ([]string, *dto.ReleaseNotesGrouped, error) {
	notes := make([]string, 0)
	value = strings.TrimSpace(value)
	if value == "" || value == "null" {
		return notes, nil, nil
	}
	// TODO(release-notes-history): 历史数组记录全部迁移为分类对象后才能删除此分支；
	// 用户升级 App 不会改变数据库中的历史记录，不能随客户端兼容字段一起删除。
	if strings.HasPrefix(value, "[") {
		if err := json.Unmarshal([]byte(value), &notes); err != nil {
			return nil, nil, err
		}
		return notes, nil, nil
	}
	if !strings.HasPrefix(value, "{") || !json.Valid([]byte(value)) {
		return nil, nil, errors.New("invalid release notes JSON")
	}
	grouped := &dto.ReleaseNotesGrouped{}
	decoder := json.NewDecoder(strings.NewReader(value))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(grouped); err != nil {
		return nil, nil, err
	}

	// Missing or null categories must serialize as arrays for both clients.
	if grouped.Features == nil {
		grouped.Features = []string{}
	}
	if grouped.Improvements == nil {
		grouped.Improvements = []string{}
	}
	if grouped.BugFixes == nil {
		grouped.BugFixes = []string{}
	}
	if grouped.Notices == nil {
		grouped.Notices = []string{}
	}

	// TODO(release-notes-compat): App 兼容字段的删除门槛见 dto.AppVersionUpdate。
	// 管理端表格和首页摘要也依赖展开数组，需同步适配后才能移除此展开逻辑。
	notes = append(notes, grouped.Features...)
	notes = append(notes, grouped.Improvements...)
	notes = append(notes, grouped.BugFixes...)
	notes = append(notes, grouped.Notices...)
	return notes, grouped, nil
}
