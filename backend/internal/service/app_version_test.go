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
		PackageFile:  "leave-android-2-0.0.2.apk",
		ReleaseNotes: `["修复登录问题","优化请假流程"]`,
	}

	wantNotes := []string{"修复登录问题", "优化请假流程"}
	if !reflect.DeepEqual(releaseNotes(version.ReleaseNotes), wantNotes) {
		t.Fatalf("ReleaseNotes = %#v, want %#v", releaseNotes(version.ReleaseNotes), wantNotes)
	}
}

func TestAppVersionResponseInvalidReleaseNotes(t *testing.T) {
	if len(releaseNotes("not-json")) != 0 {
		t.Fatalf("ReleaseNotes should be empty")
	}
}
