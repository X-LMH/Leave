package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	passwordutil "backend/internal/utils/password"
	driver "github.com/go-sql-driver/mysql"
)

func TestIsHashed(t *testing.T) {
	hash, err := passwordutil.Hash(" 密码 ")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		value   string
		hashed  bool
		invalid bool
	}{
		{"plain-password", false, false},
		{" 密码 ", false, false},
		{hash, true, false},
		{"$argon2id$broken", false, true},
		{"$2a$other-algorithm", false, true},
		{"", false, true},
	} {
		got, err := isHashed(tc.value)
		if got != tc.hashed || (err != nil) != tc.invalid {
			t.Errorf("classification mismatch: hashed=%v err=%v", got, err)
		}
	}
}

func TestDatabaseConfig(t *testing.T) {
	t.Setenv("MYSQL_PASSWORD", "environment-password")
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("mysql:\n  host: 127.0.0.1\n  port: 3306\n  user: test\n  database: isolated\n  password: yaml-password\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := readDatabaseConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Passwd != "environment-password" || cfg.DBName != "isolated" {
		t.Fatal("wrong configuration precedence")
	}
	if _, err := readDatabaseConfig(path + ".missing"); err == nil {
		t.Fatal("missing config accepted")
	}
	if err := run(nil, &bytes.Buffer{}); err == nil {
		t.Fatal("implicit database selection accepted")
	}
}

func TestMigrationIntegration(t *testing.T) {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("set MYSQL_TEST_DSN for isolated loopback MySQL migration tests")
	}
	cfg, err := driver.ParseDSN(dsn)
	if err != nil {
		t.Fatal("invalid test DSN")
	}
	host, _, err := net.SplitHostPort(cfg.Addr)
	if err != nil || cfg.Net != "tcp" || (host != "localhost" && host != "127.0.0.1" && host != "::1") {
		t.Fatal("requires loopback TCP MySQL")
	}
	cfg.DBName = ""
	server, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal("open test MySQL failed")
	}
	defer server.Close()
	schema := fmt.Sprintf("leave_hash_test_%d", time.Now().UnixNano())
	if _, err := server.Exec("CREATE DATABASE `" + schema + "`"); err != nil {
		t.Fatal("create isolated schema failed")
	}
	defer func() {
		if _, err := server.Exec("DROP DATABASE `" + schema + "`"); err != nil {
			t.Error("drop isolated schema failed")
		}
	}()
	cfg.DBName = schema
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal("open isolated schema failed")
	}
	defer db.Close()
	if _, err := db.Exec("CREATE TABLE users (id INT UNSIGNED PRIMARY KEY, password VARCHAR(255) NOT NULL) ENGINE=InnoDB"); err != nil {
		t.Fatal(err)
	}
	existing, err := passwordutil.Hash("existing-password")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO users VALUES (1, ?), (2, ?)", " plain-password ", existing); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := migrate(db, false, &out); err != nil {
		t.Fatal(err)
	}
	var stored string
	if err := db.QueryRow("SELECT password FROM users WHERE id=1").Scan(&stored); err != nil || stored != " plain-password " {
		t.Fatal("preview wrote data")
	}
	if err := migrate(db, true, &out); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT password FROM users WHERE id=1").Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if ok, err := passwordutil.Verify(stored, " plain-password "); err != nil || !ok {
		t.Fatal("password changed during migration")
	}
	firstHash := stored
	if err := migrate(db, true, &out); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT password FROM users WHERE id=1").Scan(&stored); err != nil || stored != firstHash {
		t.Fatal("repeat migration changed hash")
	}
	if err := db.QueryRow("SELECT password FROM users WHERE id=2").Scan(&stored); err != nil || stored != existing {
		t.Fatal("existing hash modified")
	}
	// 第二条写入被触发器拒绝，第一条已更新的数据也必须回滚。
	if _, err := db.Exec("UPDATE users SET password='short'"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("CREATE TRIGGER reject_second BEFORE UPDATE ON users FOR EACH ROW BEGIN IF NEW.id = 2 THEN SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'test rejection'; END IF; END"); err != nil {
		t.Fatal(err)
	}
	if err := migrate(db, true, &out); err == nil {
		t.Fatal("failed write accepted")
	}
	if err := db.QueryRow("SELECT password FROM users WHERE id=1").Scan(&stored); err != nil || stored != "short" {
		t.Fatal("failed migration was not rolled back")
	}
	if strings.Contains(out.String(), "plain-password") || strings.Contains(out.String(), existing) {
		t.Fatal("credentials leaked in output")
	}
}
