package mysql

import (
	"backend/internal/models"
	"strings"
	"testing"

	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestUpdateAdminRecordContentPreservesSnapshotsAndWritesFalse(t *testing.T) {
	tx, err := gorm.Open(gormmysql.New(gormmysql.Config{
		DSN:                       "root:password@tcp(localhost:3306)/leave",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		DryRun:                 true,
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var sql string
	if err := tx.Callback().Update().After("gorm:update").Register("capture_record_update", func(db *gorm.DB) {
		sql = db.Statement.SQL.String()
	}); err != nil {
		t.Fatal(err)
	}
	row := &models.Record{
		ID:            42,
		StudentID:     "must not change",
		Name:          "must not change",
		College:       "must not change",
		LeaveReason:   "new content",
		IsLeaveSchool: false,
		TravelWay:     "",
		Destination:   "",
	}
	if err := UpdateAdminRecordContent(tx, row); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"student_id", "name", "gender", "parent_name", "parent_phone", "teacher_name", "college", "major", "class_name", "created_at", "deleted_at"} {
		if strings.Contains(sql, "`"+field+"`=?") {
			t.Fatalf("protected field %s written: %s", field, sql)
		}
	}
	for _, field := range []string{"is_leave_school", "travel_way", "destination", "leave_reason"} {
		if !strings.Contains(sql, "`"+field+"`=?") {
			t.Fatalf("zero-value field %s omitted: %s", field, sql)
		}
	}
	if !strings.Contains(sql, "`records`.`deleted_at` IS NULL") {
		t.Fatalf("soft deletion guard missing: %s", sql)
	}
}
