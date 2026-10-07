package admin

import (
	"backend/internal/models"
	"fmt"
	"testing"
)

func TestAppVersionState(t *testing.T) {
	versions := make([]models.AppVersion, 7)
	for i := range versions {
		versions[i] = models.AppVersion{VersionCode: (i + 1) * 100, VersionName: fmt.Sprintf("1.0.%d", i), Status: models.AppVersionStatusArchived}
	}
	versions[6].Status = models.AppVersionStatusPublished
	tests := []struct {
		name   string
		status string
	}{
		{"1.0.6", appVersionStatusLatest},
		{"1.0.5", appVersionStatusPrevious},
		{"1.0.1", appVersionStatusPrevious},
		{"1.0.0", appVersionStatusOutdated},
		{"", appVersionStatusUnknown},
		{"9.9.9", appVersionStatusUnknown},
		{" 1.0.6 ", appVersionStatusLatest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := appVersionState(tt.name, versions)
			if status != tt.status {
				t.Fatalf("got %s", status)
			}
		})
	}
	t.Run("no releases", func(t *testing.T) {
		status := appVersionState("1.0.0", nil)
		if status != appVersionStatusUnknown {
			t.Fatal(status)
		}
	})
	t.Run("archived above latest excluded", func(t *testing.T) {
		records := append(append([]models.AppVersion{}, versions...), models.AppVersion{VersionCode: 800, VersionName: "future", Status: models.AppVersionStatusArchived})
		status := appVersionState("1.0.1", records)
		if status != appVersionStatusPrevious {
			t.Fatal(status)
		}
		status = appVersionState("future", records)
		if status != appVersionStatusUnknown {
			t.Fatal(status)
		}
	})
	t.Run("no published release", func(t *testing.T) {
		status := appVersionState("1.0.0", versions[:6])
		if status != appVersionStatusUnknown {
			t.Fatal(status)
		}
	})
}

func TestAppVersionStateEmptyBatch(t *testing.T) {
	status := appVersionState("", nil)
	if status != appVersionStatusUnknown {
		t.Fatalf("empty batch: status=%s", status)
	}
}
