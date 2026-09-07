package mysql

import (
	"backend/internal/models"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestProfileCreateDoesNotWriteZeroTime(t *testing.T) {
	dryRunDB, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "root:password@tcp(localhost:3306)/leave",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatalf("create dry-run database: %v", err)
	}

	stmt := dryRunDB.Create(&models.Profile{}).Statement
	for _, value := range stmt.Vars {
		if timestamp, ok := value.(time.Time); ok && timestamp.IsZero() {
			t.Fatal("profile INSERT includes a zero timestamp")
		}
	}
}
