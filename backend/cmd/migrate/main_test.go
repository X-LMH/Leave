package main

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

// TestApplyMigrations 使用 CI 提供的临时 MySQL，验证迁移程序确实能执行迁移。
// 本地没有设置 MYSQL_TEST_DSN 时跳过，避免测试依赖开发者本机数据库。
func TestApplyMigrations(t *testing.T) {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("MYSQL_TEST_DSN is not set")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE schema_migrations (
			version BIGINT NOT NULL,
			dirty BOOLEAN NOT NULL
		);
		INSERT INTO schema_migrations (version, dirty) VALUES (0, 0);
	`)
	if err != nil {
		t.Fatalf("create migration fixtures: %v", err)
	}
	defer db.Exec("DROP TABLE IF EXISTS migration_test, users, records, profiles, leave_types, feedbacks, classes, app_versions, apartments, schema_migrations")

	if err := applyMigrations(db, "../../database/migrations"); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	var version int
	var dirty bool
	if err := db.QueryRow("SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty); err != nil {
		t.Fatalf("read migration state: %v", err)
	}
	if version != 4 || dirty {
		t.Fatalf("unexpected migration state: version=%d dirty=%t", version, dirty)
	}

	var columnCount int
	if err := db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = 'records' AND column_name = 'duration'
	`).Scan(&columnCount); err != nil {
		t.Fatalf("check migrated schema: %v", err)
	}
	if columnCount != 0 {
		t.Fatal("duration column still exists after migration")
	}

	if err := db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = 'records' AND column_name = 'teacher_name'
	`).Scan(&columnCount); err != nil {
		t.Fatalf("check teacher name column: %v", err)
	}
	if columnCount != 1 {
		t.Fatal("teacher_name column does not exist after migration")
	}
}
