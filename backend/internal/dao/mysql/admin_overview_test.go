package mysql

import (
	"errors"
	"strings"
	"testing"
	"time"

	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestOverviewWeekQueryUsesCreationTime(t *testing.T) {
	tx, err := gorm.Open(gormmysql.New(gormmysql.Config{
		DSN:                       "root:test-only@tcp(localhost:3306)/leave",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var query string
	var args []any
	if err := tx.Callback().Row().After("gorm:row").Register("capture_overview_week", func(db *gorm.DB) {
		query = db.Statement.SQL.String()
		args = append([]any(nil), db.Statement.Vars...)
	}); err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 8, 10, 0, 0, 0, 0, time.FixedZone("Beijing", 8*60*60))
	next := start.AddDate(0, 0, 63)
	now := next.Add(-12 * time.Hour)
	_, err = GetOverviewWeekCounts(tx, start, next, now)
	// DryRun 只生成 SQL，不读取数据库行，Scan 应返回该特定错误。
	if !errors.Is(err, gorm.ErrDryRunModeUnsupported) {
		t.Fatalf("unexpected dry-run error: %v", err)
	}
	for _, clause := range []string{
		"TIMESTAMPDIFF(SECOND, ?, created_at)",
		"created_at >= ? AND created_at < ? AND created_at <= ?",
		"`records`.`deleted_at` IS NULL",
		"GROUP BY `week_index`",
	} {
		if !strings.Contains(query, clause) {
			t.Fatalf("missing %q: %s", clause, query)
		}
	}
	if strings.Contains(query, "applied_at") {
		t.Fatalf("trend still depends on application time: %s", query)
	}
	if len(args) != 5 || args[1] != int64(604800) {
		t.Fatalf("invalid query parameters: %v", args)
	}
	for _, item := range []struct {
		index int
		want  time.Time
	}{{0, start}, {2, start}, {3, next}, {4, now}} {
		value, ok := args[item.index].(time.Time)
		if !ok || !value.Equal(item.want) || value.Location() != time.Local {
			t.Fatalf("invalid time boundary at %d: %v", item.index, args[item.index])
		}
	}
}
