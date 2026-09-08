package service

import (
	"backend/internal/models"
	"reflect"
	"testing"
)

func TestAppVersionResponse(t *testing.T) {
	version := &models.AppVersion{
		Platform:     "android",
		VersionCode:  2,
		VersionName:  "0.0.2",
		DownloadURL:  "https://download.example.com/Leave-0.0.2.apk",
		APKSHA256:    "checksum",
		ReleaseNotes: `["修复登录问题","优化请假流程"]`,
	}

	got := appVersionResponse(version)
	if got.Platform != "android" || got.VersionCode != 2 || got.VersionName != "0.0.2" {
		t.Fatalf("unexpected version response: %#v", got)
	}
	wantNotes := []string{"修复登录问题", "优化请假流程"}
	if !reflect.DeepEqual(got.ReleaseNotes, wantNotes) {
		t.Fatalf("ReleaseNotes = %#v, want %#v", got.ReleaseNotes, wantNotes)
	}
}

func TestAppVersionResponseInvalidReleaseNotes(t *testing.T) {
	got := appVersionResponse(&models.AppVersion{ReleaseNotes: "not-json"})
	if len(got.ReleaseNotes) != 0 {
		t.Fatalf("ReleaseNotes = %#v, want empty slice", got.ReleaseNotes)
	}
}
