package app

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	driver "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestHomeIntegration(t *testing.T) {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("set MYSQL_TEST_DSN for isolated loopback MySQL home coverage")
	}
	cfg, err := driver.ParseDSN(dsn)
	if err != nil {
		t.Fatal("invalid test MySQL DSN")
	}
	host, _, err := net.SplitHostPort(cfg.Addr)
	if err != nil || cfg.Net != "tcp" || (host != "localhost" && host != "127.0.0.1" && host != "::1") {
		t.Fatal("integration tests require loopback TCP MySQL")
	}
	cfg.DBName = ""
	cfg.ParseTime = true
	cfg.Loc = time.Local
	server, err := gorm.Open(gormmysql.Open(cfg.FormatDSN()), &gorm.Config{})
	if err != nil {
		t.Fatal("cannot connect to isolated MySQL server")
	}
	serverSQL, err := server.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer serverSQL.Close()
	schema := fmt.Sprintf("leave_home_test_%d", time.Now().UnixNano())
	if err := server.Exec("CREATE DATABASE `" + schema + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci").Error; err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := server.Exec("DROP DATABASE `" + schema + "`").Error; err != nil {
			t.Errorf("drop isolated schema: %v", err)
		}
	}()
	cfg.DBName = schema
	db, err := gorm.Open(gormmysql.Open(cfg.FormatDSN()), &gorm.Config{})
	if err != nil {
		t.Fatal("cannot open isolated schema")
	}
	connection, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	for _, name := range []string{"users", "records", "profiles", "classes"} {
		definition, err := os.ReadFile(filepath.Join("..", "..", "..", "database", "tables", name+".sql"))
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(string(definition)).Error; err != nil {
			t.Fatal(err)
		}
	}
	dao := mysql.NewDAO(db, zap.NewNop())
	if err := db.Create(&[]models.User{
		{StudentID: "student-a", Role: models.RoleStudent, Status: models.UserStatusActive},
		{StudentID: "student-b", Role: models.RoleStudent, Status: models.UserStatusActive},
		{StudentID: "admin", Role: models.RoleAdmin, Status: models.UserStatusActive},
	}).Error; err != nil {
		t.Fatal(err)
	}
	service := NewRecordService(dao, NewProfileService(dao, config.StorageConfig{BaseURL: "https://example.com/api/v1/uploads"}, zap.NewNop()), zap.NewNop())
	// 与运行环境一致，使用本地北京时间覆盖年初的月份边界。
	now := time.Date(2026, 1, 1, 0, 30, 0, 0, time.Local)
	monthStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	adminHome, err := service.getHomeAt("admin", now)
	if err != nil {
		t.Fatal(err)
	}
	if adminHome.User.Role != models.RoleAdmin || adminHome.User.Name != "" {
		t.Fatalf("unexpected admin home without profile: %+v", adminHome.User)
	}
	empty, err := service.getHomeAt("student-a", now)
	if err != nil {
		t.Fatal(err)
	}
	if empty.Statistics != (dto.HomeStatistics{}) || empty.LatestRecord != nil {
		t.Fatalf("unexpected empty home: %+v", empty)
	}
	if empty.User.Name != "" || empty.User.ClassName != "" || empty.User.AvatarURL != "https://example.com/api/v1/uploads/avatars/default.png" {
		t.Fatalf("unexpected default home user: %+v", empty.User)
	}
	if empty.User.StudentID != "student-a" || empty.User.Role != models.RoleStudent {
		t.Fatalf("missing authenticated student ID without profile: %+v", empty.User)
	}
	class := models.Class{College: "计算机学院", Major: "计算机科学与技术", ClassName: "计算机1班", IsEnabled: true}
	if err := db.Create(&class).Error; err != nil {
		t.Fatal(err)
	}
	profiles := []models.Profile{
		{StudentID: "student-a", ClassID: class.ID, Name: "张三", AvatarURL: "avatars/student-a.png"},
		{StudentID: "student-b", Name: "李四", AvatarURL: "avatars/student-b.png"},
	}
	if err := db.Create(&profiles).Error; err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(empty)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &payload); err != nil {
		t.Fatal(err)
	}
	if string(payload["latest_record"]) != "null" {
		t.Fatalf("empty latest_record must be null: %s", encoded)
	}
	starts := []time.Time{monthStart.Add(-time.Second), monthStart, monthStart.AddDate(0, 1, 0)}
	durations := []time.Duration{30 * time.Minute, 25*time.Hour + time.Second, time.Hour}
	var rows []models.Record
	for i, start := range starts {
		rows = append(rows, models.Record{
			StudentID:     "student-a",
			LeaveTypeName: "事假-本科生",
			StartTime:     start,
			EndTime:       start.Add(durations[i]),
			AppliedAt:     monthStart,
			CreatedAt:     monthStart,
			IsLeaveSchool: i == 2,
		})
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	other := models.Record{
		StudentID: "student-b",
		StartTime: monthStart,
		EndTime:   monthStart.Add(100 * time.Hour),
		AppliedAt: monthStart,
		CreatedAt: monthStart.Add(time.Hour),
	}
	if err := db.Create(&other).Error; err != nil {
		t.Fatal(err)
	}
	deleted := other
	deleted.ID = 0
	deleted.StudentID = "student-a"
	if err := db.Create(&deleted).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Delete(&deleted).Error; err != nil {
		t.Fatal(err)
	}
	home, err := service.getHomeAt("student-a", now)
	if err != nil {
		t.Fatal(err)
	}
	if home.Statistics != (dto.HomeStatistics{TotalCount: 3, MonthCount: 1}) {
		t.Fatalf("wrong statistics: %+v", home.Statistics)
	}
	if home.User.Name != "张三" || home.User.ClassName != "计算机1班" || home.User.AvatarURL != "https://example.com/api/v1/uploads/avatars/student-a.png" {
		t.Fatalf("wrong home user: %+v", home.User)
	}
	if home.User.StudentID != "student-a" {
		t.Fatalf("wrong authenticated student ID: %+v", home.User)
	}
	if home.LatestRecord == nil || home.LatestRecord.ID != rows[2].ID || !home.LatestRecord.IsLeaveSchool {
		t.Fatalf("wrong latest record: %+v", home.LatestRecord)
	}
	if err := service.DeleteRecord("student-a", int(rows[2].ID)); err != nil {
		t.Fatal(err)
	}
	home, err = service.getHomeAt("student-a", now)
	if err != nil {
		t.Fatal(err)
	}
	if home.Statistics != (dto.HomeStatistics{TotalCount: 2, MonthCount: 1}) {
		t.Fatalf("statistics did not refresh after deletion: %+v", home.Statistics)
	}
	if home.LatestRecord == nil || home.LatestRecord.ID != rows[1].ID || home.LatestRecord.DurationDays != 1 || home.LatestRecord.DurationHours != 2 {
		t.Fatalf("duration or latest record did not refresh: %+v", home.LatestRecord)
	}
}
