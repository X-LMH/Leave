package admin

import (
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	driver "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// This suite creates its own schema; it never connects through business config.
func TestAdminOverviewIntegration(t *testing.T) {
	dao := mysql.NewDAO(nil, zap.NewNop())
	dsn := os.Getenv("ADMIN_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set ADMIN_TEST_MYSQL_DSN for isolated loopback MySQL coverage")
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
	server, err := gorm.Open(gormmysql.Open(cfg.FormatDSN()), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatal("cannot connect to isolated MySQL server")
	}
	serverSQL, err := server.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer serverSQL.Close()
	schema := fmt.Sprintf("leave_overview_test_%d", time.Now().UnixNano())
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
	for _, name := range []string{"users", "classes", "records", "feedbacks", "app_versions"} {
		definition, err := os.ReadFile(filepath.Join("..", "..", "..", "database", "tables", name+".sql"))
		if err != nil {
			t.Fatal(err)
		}
		if err := testDB.Exec(string(definition)).Error; err != nil {
			t.Fatal(err)
		}
	}

	dao = mysql.NewDAO(testDB, zap.NewNop())
	now := time.Date(2026, 10, 11, 12, 0, 0, 0, overviewBeijing)
	overviewService := NewOverviewService(dao, zap.NewNop())
	empty, err := overviewService.getAdminOverviewAt(now)
	if err != nil {
		t.Fatal(err)
	}
	if empty.Statistics != (dto.OverviewStatistics{}) || empty.LatestAndroidVersion != nil || len(empty.RecentRecords) != 0 || len(empty.WeeklyApplications.Weeks) != 9 || empty.WeeklyApplications.Total != 0 || len(empty.VersionUsage.Groups) != 0 {
		t.Fatalf("incorrect empty response: %+v", empty)
	}
	for i, version := range []string{"0.0.10", "0.0.9", "0.0.8", "0.0.7", " ", "unknown", "0.0.10", "0.0.10"} {
		role := models.RoleStudent
		if i == 6 {
			role = models.RoleAdmin
		}
		user := models.User{
			StudentID:  fmt.Sprintf("overview-%d", i),
			Password:   "test-only",
			Role:       role,
			Status:     uint8(i % 2),
			AppVersion: version,
		}
		if i == 7 {
			user.DeletedAt = gorm.DeletedAt{Time: now.In(time.Local), Valid: true}
		}
		if err := testDB.Create(&user).Error; err != nil {
			t.Fatal(err)
		}
	}
	for i := range 2 {
		if err := testDB.Create(&models.Class{
			College:   "学院",
			Major:     "专业",
			ClassName: fmt.Sprintf("班级%d", i),
			IsEnabled: i == 0,
		}).Error; err != nil {
			t.Fatal(err)
		}
	}
	for i := range 2 {
		feedback := models.Feedback{Content: "隔离测试", IsAnonymous: true}
		if i == 1 {
			feedback.DeletedAt = gorm.DeletedAt{Time: now.In(time.Local), Valid: true}
		}
		if err := testDB.Create(&feedback).Error; err != nil {
			t.Fatal(err)
		}
	}
	weeks := overviewWeeks(now)
	first := weeks[0].Start
	current := weeks[8].Start
	dates := []time.Time{
		first.Add(-time.Second),
		first,
		first.AddDate(0, 0, 7).Add(-time.Second),
		first.AddDate(0, 0, 7),
		current.Add(-time.Second),
		current,
	}
	for range 6 {
		dates = append(dates, now.Add(-time.Minute))
	}
	dates = append(dates, now, now.Add(time.Second), now.Add(-time.Hour))
	for i, date := range dates {
		record := models.Record{
			ID:            uint(i + 1),
			StudentID:     "snapshot-student",
			Name:          "姓名快照",
			ClassName:     "已停用班级快照",
			LeaveTypeName: "已改名类型快照",
			AppliedAt:     date.In(time.Local),
			CreatedAt:     date.In(time.Local),
			StartTime:     now.In(time.Local),
			EndTime:       now.Add(time.Hour).In(time.Local),
		}
		if i == len(dates)-1 {
			record.DeletedAt = gorm.DeletedAt{Time: now.In(time.Local), Valid: true}
		}
		// 申请时间与创建时间跨周，验证趋势按创建时间归周。
		if record.ID == 6 {
			record.CreatedAt = current.AddDate(0, 0, -7).In(time.Local)
		}
		// 申请时间虽在未来，但记录已创建，仍应计入创建趋势。
		if record.ID == 14 {
			record.CreatedAt = now.In(time.Local)
		}
		// 申请时间在过去，但创建时间在未来，不能计入本周趋势。
		if record.ID == 1 {
			record.CreatedAt = now.Add(time.Second).In(time.Local)
		}
		if err := testDB.Create(&record).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, version := range []models.AppVersion{
		{Platform: "android", VersionCode: 7, VersionName: "0.0.7", Status: "archived", PublishedAt: now.Add(-time.Hour).In(time.Local), ReleaseNotes: "[]"},
		{Platform: "android", VersionCode: 8, VersionName: "0.0.8", Status: "archived", PublishedAt: now.Add(-time.Hour).In(time.Local), ReleaseNotes: "[]"},
		{Platform: "android", VersionCode: 9, VersionName: "0.0.9", Status: "archived", PublishedAt: now.Add(-time.Hour).In(time.Local), ReleaseNotes: "[]"},
		{Platform: "android", VersionCode: 10, VersionName: "0.0.10", Status: "published", PublishedAt: now.In(time.Local), ReleaseNotes: `["优化查询"]`},
		{Platform: "android", VersionCode: 11, VersionName: "0.0.11", Status: "published", PublishedAt: now.Add(time.Second).In(time.Local), ReleaseNotes: "[]"},
		{Platform: "ios", VersionCode: 99, VersionName: "0.0.99", Status: "published", PublishedAt: now.In(time.Local), ReleaseNotes: "[]"},
	} {
		if err := testDB.Create(&version).Error; err != nil {
			t.Fatal(err)
		}
	}
	result, err := overviewService.getAdminOverviewAt(now)
	if err != nil {
		t.Fatal(err)
	}
	if result.Statistics != (dto.OverviewStatistics{StudentCount: 6, ClassCount: 2, RecordCount: 14, FeedbackCount: 1}) {
		t.Fatalf("incorrect counts: %+v", result.Statistics)
	}
	if result.WeeklyApplications.Total != 13 {
		t.Fatalf("incorrect weekly total: %+v", result.WeeklyApplications)
	}
	for i, count := range []int64{2, 1, 0, 0, 0, 0, 0, 2, 8} {
		if result.WeeklyApplications.Weeks[i].Count != count {
			t.Fatalf("week %d count: %+v", i, result.WeeklyApplications.Weeks[i])
		}
	}
	if len(result.RecentRecords) != 5 {
		t.Fatalf("incorrect recent limit: %+v", result.RecentRecords)
	}
	for i, record := range result.RecentRecords {
		if record.ID != uint(14-i) || record.Name != "姓名快照" || record.ClassName != "已停用班级快照" || record.LeaveTypeName != "已改名类型快照" {
			t.Fatalf("incorrect recent snapshot or order: %+v", record)
		}
	}
	if result.VersionUsage.Total != result.Statistics.StudentCount || result.VersionUsage.LatestCount != 1 || len(result.VersionUsage.Groups) != 5 {
		t.Fatalf("incorrect version summary: %+v", result.VersionUsage)
	}
	for i, count := range []int64{1, 1, 2, 1, 1} {
		if result.VersionUsage.Groups[i].Count != count {
			t.Fatalf("incorrect version group: %+v", result.VersionUsage.Groups[i])
		}
	}
	if result.LatestAndroidVersion == nil || result.LatestAndroidVersion.VersionCode != 10 || len(result.LatestAndroidVersion.ReleaseNotes) != 1 {
		t.Fatalf("incorrect current release: %+v", result.LatestAndroidVersion)
	}
	if err := testDB.Model(&models.AppVersion{}).Where("platform = ?", "android").Update("status", "archived").Error; err != nil {
		t.Fatal(err)
	}
	result, err = overviewService.getAdminOverviewAt(now)
	if err != nil || result.LatestAndroidVersion != nil || result.VersionUsage.LatestCount != 0 {
		t.Fatalf("incorrect no-published-release state: %+v, %v", result, err)
	}
	// Any failed subquery must fail the entire response, without partial numbers.
	if err := testDB.Exec("DROP TABLE feedbacks").Error; err != nil {
		t.Fatal(err)
	}
	if partial, err := overviewService.getAdminOverviewAt(now); err == nil || partial != nil {
		t.Fatalf("query failure returned partial data: %+v, %v", partial, err)
	}
}
