package admin

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	driver "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// This suite creates its own loopback schema and never loads application config.
func TestAdminRecordsIntegration(t *testing.T) {
	dsn := os.Getenv("ADMIN_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set ADMIN_TEST_MYSQL_DSN for isolated loopback MySQL coverage")
	}
	cfg, err := driver.ParseDSN(dsn)
	if err != nil {
		t.Fatal("invalid test MySQL DSN")
	}
	host, port, err := net.SplitHostPort(cfg.Addr)
	if err != nil || cfg.Net != "tcp" || (host != "localhost" && host != "127.0.0.1" && host != "::1") {
		t.Fatal("integration tests require loopback TCP MySQL")
	}
	cfg.DBName = ""
	cfg.ParseTime = true
	cfg.Loc = time.Local
	server, err := gorm.Open(gormmysql.Open(cfg.FormatDSN()), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatal("cannot connect to isolated MySQL server")
	}
	serverSQL, err := server.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer serverSQL.Close()
	schema := fmt.Sprintf("leave_records_test_%d", time.Now().UnixNano())
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
	for _, name := range []string{"records", "classes", "leave_types"} {
		sql, err := os.ReadFile(filepath.Join("..", "..", "..", "database", "tables", name+".sql"))
		if err != nil {
			t.Fatal(err)
		}
		if err := testDB.Exec(string(sql)).Error; err != nil {
			t.Fatal(err)
		}
	}
	previous := config.Cfg
	defer func() { config.Cfg = previous }()
	portNumber, err := strconv.Atoi(port)
	if err != nil {
		t.Fatal(err)
	}
	config.Cfg.Mysql.Host = host
	config.Cfg.Mysql.Port = portNumber
	config.Cfg.Mysql.Database = schema
	config.Cfg.Mysql.User = cfg.User
	config.Cfg.Mysql.Password = cfg.Passwd
	if err := mysql.Init(); err != nil {
		t.Fatal("cannot initialize isolated test database")
	}

	class := models.Class{
		College:   "原学院",
		Major:     "原专业",
		ClassName: "原班级",
		IsEnabled: false,
	}
	oldType := models.LeaveType{
		Name:      "已改名的原因",
		IsEnabled: false,
	}
	newType := models.LeaveType{
		Name:      "新原因",
		IsEnabled: true,
	}
	disabledType := models.LeaveType{
		Name:      "其他停用类型",
		IsEnabled: false,
	}
	for _, row := range []any{&class, &oldType, &newType, &disabledType} {
		if err := testDB.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	start := time.Date(2026, 10, 10, 0, 0, 0, 0, time.Local)
	rows := []models.Record{
		{
			StudentID:     "20260001",
			Name:          "历史姓名",
			Gender:        "female",
			College:       class.College,
			Major:         class.Major,
			ClassName:     class.ClassName,
			ParentName:    "家长",
			ParentPhone:   "13800138000",
			TeacherName:   "辅导员",
			LeaveTypeID:   oldType.ID,
			LeaveTypeName: "历史原因",
			StartTime:     start,
			EndTime:       start.Add(time.Hour),
			IsLeaveSchool: true,
			LeaveReason:   "旧内容",
			TravelWay:     "公交",
			Destination:   "医院",
			AppliedAt:     start,
			CreatedAt:     start,
		},
		{
			StudentID:     "20260002",
			Name:          "历史姓名",
			College:       class.College,
			Major:         class.Major,
			ClassName:     class.ClassName,
			LeaveTypeID:   oldType.ID,
			LeaveTypeName: "历史原因",
			StartTime:     start.Add(time.Hour),
			EndTime:       start.Add(2 * time.Hour),
			AppliedAt:     start,
		},
		{
			StudentID:   "20260003",
			Name:        "历史姓名",
			College:     class.College,
			Major:       class.Major,
			ClassName:   class.ClassName,
			LeaveTypeID: oldType.ID,
			StartTime:   start.Add(24 * time.Hour),
			EndTime:     start.Add(25 * time.Hour),
			AppliedAt:   start,
		},
		{
			StudentID:   "deleted",
			Name:        "历史姓名",
			College:     "已删除记录的学院",
			Major:       class.Major,
			ClassName:   class.ClassName,
			LeaveTypeID: oldType.ID,
			StartTime:   start,
			EndTime:     start.Add(time.Hour),
			AppliedAt:   start.Add(time.Hour),
		},
	}
	for i := range rows {
		if err := testDB.Create(&rows[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := testDB.Delete(&rows[3]).Error; err != nil {
		t.Fatal(err)
	}
	end := start.Add(24 * time.Hour)
	query := dto.AdminRecordQuery{
		Page:     1,
		PageSize: 1,
		ClassFilter: &dto.RecordClassFilter{
			College:   class.College,
			Major:     class.Major,
			ClassName: class.ClassName,
		},
		LeaveTypeID:   oldType.ID,
		Name:          "历史",
		StudentID:     "2026",
		StartTimeFrom: &start,
		StartTimeTo:   &end,
	}
	// Renaming and deleting current classes must not affect snapshot filtering.
	if err := testDB.Model(&class).Update("class_name", "新名称").Error; err != nil {
		t.Fatal(err)
	}
	if err := testDB.Delete(&class).Error; err != nil {
		t.Fatal(err)
	}
	list, err := GetAdminRecords(query)
	if err != nil {
		t.Fatal(err)
	}
	if list.Total != 2 || len(list.Items) != 1 || list.Items[0].ID != rows[1].ID {
		t.Fatalf("wrong count, range or stable ordering: %+v", list)
	}
	query.Page = 2
	list, err = GetAdminRecords(query)
	if err != nil || len(list.Items) != 1 || list.Items[0].ID != rows[0].ID {
		t.Fatalf("wrong second page: %+v, %v", list, err)
	}
	leaving := true
	query.IsLeaveSchool = &leaving
	query.Page = 1
	list, err = GetAdminRecords(query)
	if err != nil || list.Total != 1 {
		t.Fatalf("school filter: %+v, %v", list, err)
	}
	query.StudentID = "no-match"
	list, err = GetAdminRecords(query)
	if err != nil || list.Total != 0 || list.Items == nil || len(list.Items) != 0 {
		t.Fatalf("empty list must be an array: %+v, %v", list, err)
	}
	if _, err := GetAdminRecord(rows[3].ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatal("deleted detail visible")
	}
	if _, err := GetAdminRecord(999999); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatal("missing detail accepted")
	}

	input := validRecordInput()
	input.LeaveTypeID = oldType.ID
	*input.IsLeaveSchool = false
	result, err := UpdateAdminRecord(rows[0].ID, &input)
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != rows[0].Name || result.College != class.College || result.ParentName != rows[0].ParentName || result.TeacherName != rows[0].TeacherName || !result.CreatedAt.Equal(rows[0].CreatedAt) || result.LeaveTypeName != "历史原因" || result.TravelWay != "" || result.Destination != "" || result.IsLeaveSchool {
		t.Fatalf("snapshots or cleared fields changed: %+v", result)
	}
	if err := testDB.Delete(&oldType).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := UpdateAdminRecord(rows[0].ID, &input); err != nil {
		t.Fatalf("deleted original type must remain valid: %v", err)
	}
	input.LeaveTypeID = newType.ID
	result, err = UpdateAdminRecord(rows[0].ID, &input)
	if err != nil || result.LeaveTypeName != newType.Name {
		t.Fatalf("switch type: %+v, %v", result, err)
	}
	input.LeaveTypeID = disabledType.ID
	if _, err := UpdateAdminRecord(rows[0].ID, &input); !errors.Is(err, ErrorInvalidAdminInput) {
		t.Fatal("disabled new type accepted")
	}
	input.LeaveTypeID = oldType.ID
	input.LeaveReason = "should not persist"
	if _, err := UpdateAdminRecord(rows[0].ID, &input); !errors.Is(err, ErrorInvalidAdminInput) {
		t.Fatal("deleted new type accepted")
	}
	result, err = GetAdminRecord(rows[0].ID)
	if err != nil || result.LeaveReason == input.LeaveReason || result.LeaveTypeID != newType.ID {
		t.Fatal("failed update was persisted")
	}
	if _, err := UpdateAdminRecord(999999, &input); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatal("missing update accepted")
	}
	// Force a persistence failure after a permitted type switch and verify rollback.
	otherType := models.LeaveType{
		Name:      "另一个原因",
		IsEnabled: true,
	}
	if err := testDB.Create(&otherType).Error; err != nil {
		t.Fatal(err)
	}
	if err := testDB.Exec("CREATE TRIGGER reject_record_update BEFORE UPDATE ON records FOR EACH ROW SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'test rollback'").Error; err != nil {
		t.Fatal(err)
	}
	input.LeaveTypeID = otherType.ID
	if _, err := UpdateAdminRecord(rows[0].ID, &input); err == nil {
		t.Fatal("expected save failure")
	}
	result, err = GetAdminRecord(rows[0].ID)
	if err != nil || result.LeaveTypeID != newType.ID {
		t.Fatal("save failure did not roll back")
	}
}
