package admin

import (
	"backend/internal/dto"
	"backend/internal/models"
	"reflect"
	"testing"
)

func TestNormalizeAppVersionQuery(t *testing.T) {
	for _, tc := range []struct {
		platform     string
		wantPlatform string
		status       string
		valid        bool
	}{
		{"", "", "", true},
		{" Android ", "android", " published ", true},
		{"android", "android", "archived", true},
		{" iOS ", "ios", "", true},
		{"android", "android", "draft", false},
	} {
		query := dto.AppVersionListQuery{
			Platform: tc.platform,
			Status:   tc.status,
			Keyword:  " 0.0.2 ",
		}
		err := normalizeAppVersionQuery(&query)
		if (err == nil) != tc.valid {
			t.Errorf("%+v: %v", tc, err)
		}
		if tc.valid && (query.Platform != tc.wantPlatform || query.Keyword != "0.0.2") {
			t.Errorf("query not normalized: %+v", query)
		}
	}
}

func TestAdminAppVersionNotes(t *testing.T) {
	for _, tc := range []struct {
		notes string
		want  []string
		valid bool
	}{
		{"", []string{}, true},
		{"null", []string{}, true},
		{"[]", []string{}, true},
		{`["修复登录","优化交互"]`, []string{"修复登录", "优化交互"}, true},
		{"broken", nil, false},
		{`{"note":"修复"}`, nil, false},
		{`[123]`, nil, false},
	} {
		row := models.AppVersion{
			ID:           12,
			VersionCode:  20,
			PackageFile:  "leave.apk",
			ReleaseNotes: tc.notes,
		}
		item, err := adminAppVersion(&row)
		if (err == nil) != tc.valid {
			t.Errorf("%q: %v", tc.notes, err)
			continue
		}
		if tc.valid && (!reflect.DeepEqual(item.ReleaseNotes, tc.want) || item.ID != row.ID || item.PackageFile != row.PackageFile || item.VersionCode != row.VersionCode) {
			t.Errorf("unexpected record: %+v", item)
		}
	}
}
