package mysql

import (
	"strings"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestAppVersionUpdatesNewestFirst(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "root:password@tcp(localhost:3306)/leave",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var query string
	if err := db.Callback().Query().After("gorm:query").Register("test:capture_versions", func(tx *gorm.DB) {
		query = tx.Statement.SQL.String()
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := (&DAO{db: db}).GetAppVersionUpdates("android", 1); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(query, "ORDER BY version_code DESC") {
		t.Fatalf("updates are not newest first: %s", query)
	}
}
