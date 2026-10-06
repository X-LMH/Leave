package mysql

import (
	"backend/internal/dto"
	"backend/internal/models"
	"errors"
	"fmt"
	driver "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The integration suite creates and drops its own schema on a loopback MySQL server.
// It never loads the application's database configuration.
func TestAdminManagementIntegration(t *testing.T) {
	dsn := os.Getenv("ADMIN_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set ADMIN_TEST_MYSQL_DSN for an isolated loopback MySQL integration test")
	}
	cfg, err := driver.ParseDSN(dsn)
	if err != nil {
		t.Fatal("invalid test MySQL DSN")
	}
	host, _, err := net.SplitHostPort(cfg.Addr)
	if err != nil || cfg.Net != "tcp" || (host != "localhost" && host != "127.0.0.1" && host != "::1") {
		t.Fatal("integration tests require a loopback TCP MySQL server")
	}
	cfg.DBName = ""
	cfg.ParseTime = true
	cfg.MultiStatements = true
	server, err := gorm.Open(gormmysql.Open(cfg.FormatDSN()), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatal("cannot connect to isolated MySQL server")
	}
	serverSQL, err := server.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer serverSQL.Close()
	schema := fmt.Sprintf("leave_admin_test_%d", time.Now().UnixNano())
	if err := server.Exec("CREATE DATABASE `" + schema + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci").Error; err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := server.Exec("DROP DATABASE `" + schema + "`").Error; err != nil {
			t.Errorf("drop test schema: %v", err)
		}
	}()
	cfg.DBName = schema
	testDB, err := gorm.Open(gormmysql.Open(cfg.FormatDSN()), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatal(err)
	}
	connection, err := testDB.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	original := db
	db = testDB
	defer func() { db = original }()
	for _, name := range []string{"leave_types", "classes", "profiles"} {
		sql, err := os.ReadFile(filepath.Join("..", "..", "..", "database", "tables", name+".sql"))
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(string(sql)).Error; err != nil {
			t.Fatal(err)
		}
	}
	first := &models.LeaveType{Name: "病假", SortOrder: 2, IsEnabled: true}
	second := &models.LeaveType{Name: "事假", SortOrder: 1, IsEnabled: false}
	for _, row := range []*models.LeaveType{first, second} {
		if err := CreateAdminLeaveType(row); err != nil {
			t.Fatal(err)
		}
	}
	if err := CreateAdminLeaveType(&models.LeaveType{Name: first.Name}); !IsDuplicateKeyError(err) {
		t.Fatalf("duplicate error: %v", err)
	}
	items, total, err := GetAdminLeaveTypes(dto.AdminLeaveTypeListQuery{Page: 1, PageSize: 1})
	if err != nil || total != 2 || len(items) != 1 || items[0].ID != second.ID {
		t.Fatalf("pagination/order: %v %d %+v", err, total, items)
	}
	first.SortOrder = 0
	first.IsEnabled = false
	if err := UpdateAdminLeaveType(first); err != nil {
		t.Fatal(err)
	}
	if first.SortOrder != 0 || first.IsEnabled || first.CreatedAt.IsZero() || first.UpdatedAt.IsZero() {
		t.Fatal("zero values/timestamps lost")
	}
	enabled, err := GetEnabledLeaveTypes()
	if err != nil || len(enabled) != 0 {
		t.Fatalf("disabled options: %v %+v", err, enabled)
	}
	if err := DeleteAdminLeaveType(first.ID); err != nil {
		t.Fatal(err)
	}
	if err := DeleteAdminLeaveType(first.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("repeat delete: %v", err)
	}
	if err := CreateAdminLeaveType(&models.LeaveType{Name: first.Name}); !IsDuplicateKeyError(err) {
		t.Fatalf("deleted name uniqueness: %v", err)
	}
	first.Name = "other"
	if err := UpdateAdminLeaveType(first); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("update deleted: %v", err)
	}
	items, total, err = GetAdminLeaveTypes(dto.AdminLeaveTypeListQuery{Page: 1, PageSize: 10, Name: "事"})
	if err != nil || total != 1 || len(items) != 1 {
		t.Fatalf("name filter: %v %d", err, total)
	}
	class := &models.Class{College: "学院", Major: "专业", ClassName: "一班", IsEnabled: true}
	if err := CreateAdminClass(class); err != nil {
		t.Fatal(err)
	}
	if err := CreateAdminClass(&models.Class{College: class.College, Major: class.Major, ClassName: class.ClassName}); !IsDuplicateKeyError(err) {
		t.Fatalf("class duplicate: %v", err)
	}
	class.IsEnabled = false
	if err := UpdateAdminClass(class); err != nil {
		t.Fatal(err)
	}
	options, err := GetClasses()
	if err != nil || len(options) != 0 {
		t.Fatalf("disabled classes: %v %+v", err, options)
	}
	disabled := false
	classes, total, err := GetAdminClasses(dto.AdminClassListQuery{Page: 1, PageSize: 10, ClassName: "一", IsEnabled: &disabled})
	if err != nil || total != 1 || len(classes) != 1 {
		t.Fatalf("class filters: %v %d", err, total)
	}
	profile := &models.Profile{StudentID: "test", ClassID: class.ID, Name: "学生", Gender: "male"}
	if err := FinishProfile(profile); err != nil {
		t.Fatal(err)
	}
	if err := DeleteAdminClass(class.ID); !errors.Is(err, ErrorClassInUse) {
		t.Fatalf("referenced class: %v", err)
	}
	if err := db.Delete(profile).Error; err != nil {
		t.Fatal(err)
	}
	if err := DeleteAdminClass(class.ID); !errors.Is(err, ErrorClassInUse) {
		t.Fatalf("soft deleted profile: %v", err)
	}
	if err := db.Unscoped().Delete(profile).Error; err != nil {
		t.Fatal(err)
	}
	if err := DeleteAdminClass(class.ID); err != nil {
		t.Fatal(err)
	}
	if err := DeleteAdminClass(class.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("missing class: %v", err)
	}
	profile.StudentID = "new"
	profile.ID = 0
	if err := FinishProfile(profile); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("assignment to deleted class: %v", err)
	}
	// All data and connections above are confined to the generated test schema.
	if !strings.HasPrefix(schema, "leave_admin_test_") {
		t.Fatal("unexpected test schema")
	}
}
