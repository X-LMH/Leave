package logger

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestErrorHidesDatabaseValues(t *testing.T) {
	core, entries := observer.New(zap.InfoLevel)
	log := zap.New(core)
	err := fmt.Errorf("保存失败: %w", &mysql.MySQLError{
		Number:   1062,
		SQLState: [5]byte{'2', '3', '0', '0', '0'},
		Message:  "Duplicate entry 'private-student-id' for key 'users.student_id'",
	})
	log.Error("写入失败", Error(err))
	message := entries.All()[0].ContextMap()["error"].(string)
	if strings.Contains(message, "private-student-id") || !strings.Contains(message, "1062") {
		t.Fatalf("unexpected database error: %s", message)
	}
	log.Error("文件失败", Error(errors.New("permission denied")))
	if entries.All()[1].ContextMap()["error"] != "permission denied" {
		t.Fatal("non-database cause should be retained")
	}
}
