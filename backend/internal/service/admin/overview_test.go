package admin

import (
	"backend/internal/dao/mysql"
	"backend/internal/models"
	"testing"
	"time"
)

func TestOverviewWeeks(t *testing.T) {
	for _, test := range []struct {
		now    string
		monday string
		sunday string
	}{
		{"2026-10-05T00:00:00+08:00", "2026-10-05", "2026-10-11"},
		{"2026-10-11T23:59:59+08:00", "2026-10-05", "2026-10-11"},
		{"2026-10-01T12:00:00+08:00", "2026-09-28", "2026-10-04"},
		{"2027-01-01T00:00:00+08:00", "2026-12-28", "2027-01-03"},
		{"2026-10-04T16:00:00Z", "2026-10-05", "2026-10-11"},
	} {
		t.Run(test.now, func(t *testing.T) {
			now, err := time.Parse(time.RFC3339, test.now)
			if err != nil {
				t.Fatal(err)
			}
			weeks := overviewWeeks(now)
			if len(weeks) != 9 || weeks[8].Start.Format(time.DateOnly) != test.monday || weeks[8].End.Format(time.DateOnly) != test.sunday {
				t.Fatalf("incorrect ranges: %+v", weeks)
			}
			for i, week := range weeks {
				if week.Start.Weekday() != time.Monday || week.End.Weekday() != time.Sunday || week.Count != 0 || week.IsCurrent != (i == 8) {
					t.Fatalf("invalid week: %+v", week)
				}
				if week.Start.Hour() != 0 || week.End.Sub(week.Start) != 6*24*time.Hour || weeks[8].Start.Sub(week.Start) != time.Duration(8-i)*7*24*time.Hour {
					t.Fatalf("invalid spacing: %+v", week)
				}
			}
		})
	}
}

func TestOverviewVersionUsage(t *testing.T) {
	versions := []models.AppVersion{
		{ID: 1, VersionCode: 10, VersionName: "0.0.10", Status: models.AppVersionStatusPublished},
		{ID: 2, VersionCode: 9, VersionName: "0.0.5", Status: models.AppVersionStatusArchived},
		{ID: 3, VersionCode: 8, VersionName: "0.0.8", Status: models.AppVersionStatusArchived},
		{ID: 4, VersionCode: 7, VersionName: "0.0.7", Status: models.AppVersionStatusArchived},
	}
	usage, latest := overviewVersionUsage([]mysql.OverviewReportedVersion{
		{Version: " 0.0.10 ", Count: 48},
		{Version: "0.0.5", Count: 12},
		{Version: "0.0.8", Count: 2},
		{Version: "0.0.7", Count: 2},
		{Version: " ", Count: 8},
		{Version: "0.0.6", Count: 1},
		{Version: "0.0.11", Count: 1},
	}, versions)
	if latest == nil || latest.VersionCode != 10 || usage.Total != 74 || usage.LatestCount != 48 || len(usage.Groups) != 5 {
		t.Fatalf("invalid summary: %+v, %+v", usage, latest)
	}
	var total int64
	var percentage float64
	for i, count := range []int64{48, 12, 4, 8, 2} {
		if usage.Groups[i].Count != count {
			t.Fatalf("invalid group: %+v", usage.Groups[i])
		}
		total += usage.Groups[i].Count
		percentage += usage.Groups[i].Percentage
	}
	if total != usage.Total || percentage < 99.999 || percentage > 100.001 {
		t.Fatalf("inconsistent totals: %+v", usage)
	}
	versions = append(versions, models.AppVersion{ID: 5, VersionCode: 10, VersionName: "newest-id", Status: models.AppVersionStatusPublished})
	_, latest = overviewVersionUsage(nil, versions)
	if latest.ID != 5 {
		t.Fatal("same code must choose highest ID")
	}
}

func TestOverviewVersionUsageEmptyAndMissingRelease(t *testing.T) {
	usage, latest := overviewVersionUsage(nil, nil)
	if latest != nil || usage.Total != 0 || usage.Groups == nil || len(usage.Groups) != 0 {
		t.Fatalf("invalid empty state: %+v", usage)
	}
	usage, latest = overviewVersionUsage([]mysql.OverviewReportedVersion{{Version: "old", Count: 2}, {Version: "", Count: 3}}, nil)
	if latest != nil || usage.Groups[3].Count != 3 || usage.Groups[4].Count != 2 {
		t.Fatalf("missing release must not guess version age: %+v", usage)
	}
	usage, _ = overviewVersionUsage([]mysql.OverviewReportedVersion{{Count: 3}}, nil)
	if len(usage.Groups) != 4 || usage.Groups[3].Percentage != 100 {
		t.Fatalf("hide unused other group: %+v", usage)
	}
}
