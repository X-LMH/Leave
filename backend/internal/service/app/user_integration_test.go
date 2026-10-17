package app

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/utils/jwt"
	passwordutil "backend/internal/utils/password"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	driver "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestRegisterIntegration(t *testing.T) {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("set MYSQL_TEST_DSN for isolated loopback MySQL registration coverage")
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
	server, err := gorm.Open(gormmysql.Open(cfg.FormatDSN()), &gorm.Config{})
	if err != nil {
		t.Fatal("cannot connect to isolated MySQL server")
	}
	connection, err := server.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := connection.Close(); err != nil {
			t.Errorf("close test connection: %v", err)
		}
	})
	schema := fmt.Sprintf("leave_register_test_%d", time.Now().UnixNano())
	if err := server.Exec("CREATE DATABASE `" + schema + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci").Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := server.Exec("DROP DATABASE `" + schema + "`").Error; err != nil {
			t.Errorf("drop isolated schema: %v", err)
		}
	})
	definition, err := os.ReadFile(filepath.Join("..", "..", "..", "database", "tables", "users.sql"))
	if err != nil {
		t.Fatal(err)
	}
	// Always target the isolated schema, regardless of the pooled connection.
	qualified := "CREATE TABLE `" + schema + "`.`users`"
	definitionSQL := strings.Replace(string(definition), "CREATE TABLE `users`", qualified, 1)
	if err := server.Exec(definitionSQL).Error; err != nil {
		t.Fatal(err)
	}
	previous := config.Cfg
	t.Cleanup(func() { config.Cfg = previous })
	portNumber, err := strconv.Atoi(port)
	if err != nil {
		t.Fatal(err)
	}
	config.Cfg.Mysql.Host = host
	config.Cfg.Mysql.Port = portNumber
	config.Cfg.Mysql.Database = schema
	config.Cfg.Mysql.User = cfg.User
	config.Cfg.Mysql.Password = cfg.Passwd
	config.Cfg.JWT.Secret = "isolated-register-test"
	config.Cfg.JWT.ExpirationDays = 1
	if err := mysql.Init(); err != nil {
		t.Fatal("cannot initialize isolated database")
	}
	req := &dto.RegisterRequest{
		StudentID:  " 202410201016 ",
		Password:   "123456",
		RePassword: "123456",
		DeviceName: " Android 测试机 ",
		AppVersion: " 1.0.0 ",
	}
	data, err := Register(req)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := jwt.ParseToken(data.Token)
	if err != nil || claims.StudentID != "202410201016" || claims.Role != models.RoleStudent {
		t.Fatalf("unusable registration session: %v", err)
	}
	if data.User.StudentID != "202410201016" || data.ProfileCompleted {
		t.Fatalf("unexpected new-user session: %#v", data)
	}
	user, err := mysql.FindUserByStudentID("202410201016")
	if err != nil {
		t.Fatal(err)
	}
	matched, err := passwordutil.Verify(user.Password, req.Password)
	if err != nil || !matched || user.Password == req.Password {
		t.Fatal("registration did not store a password hash")
	}
	// 创建登录需要读取的资料表，在隔离库验证完整注册、登录和改密链路。
	for _, table := range []string{"profiles", "classes"} {
		definition, err := os.ReadFile(filepath.Join("..", "..", "..", "database", "tables", table+".sql"))
		if err != nil {
			t.Fatal(err)
		}
		qualified := strings.Replace(string(definition), "CREATE TABLE `"+table+"`", "CREATE TABLE `"+schema+"`.`"+table+"`", 1)
		if err := server.Exec(qualified).Error; err != nil {
			t.Fatal(err)
		}
	}
	login := &dto.LoginRequest{StudentID: "202410201016", Password: req.Password}
	if _, err := Login(login); err != nil {
		t.Fatalf("registered user cannot log in: %v", err)
	}
	login.Password = "wrong-password"
	if _, err := Login(login); !errors.Is(err, mysql.ErrorInvalidPassword) {
		t.Fatalf("wrong password accepted: %v", err)
	}
	change := &dto.PasswordChangeRequest{Password: "wrong-password", NewPassword: "new-pass", RePassword: "new-pass"}
	if err := ChangePassword(change, login.StudentID); !errors.Is(err, mysql.ErrorNotRightPassword) {
		t.Fatalf("incorrect old password accepted: %v", err)
	}
	change.Password = req.Password
	if err := ChangePassword(change, login.StudentID); err != nil {
		t.Fatal(err)
	}
	login.Password = req.Password
	if _, err := Login(login); !errors.Is(err, mysql.ErrorInvalidPassword) {
		t.Fatalf("old password still works: %v", err)
	}
	login.Password = change.NewPassword
	if _, err := Login(login); err != nil {
		t.Fatalf("new password cannot log in: %v", err)
	}
	updated, err := mysql.FindUserByStudentID(login.StudentID)
	if err != nil {
		t.Fatal(err)
	}
	if matched, err := passwordutil.Verify(updated.Password, change.NewPassword); err != nil || !matched {
		t.Fatal("password change did not store a hash")
	}
	if user.LastSeenAt == nil || user.LastSeenDevice != "Android 测试机" || user.AppVersion != "1.0.0" || user.Status != models.UserStatusActive {
		t.Fatalf("registration usage metadata missing: %#v", user)
	}
	if duplicate, err := Register(req); !errors.Is(err, mysql.ErrorUserExist) || duplicate != nil {
		t.Fatalf("duplicate registration returned a session: data=%#v err=%v", duplicate, err)
	}
	if err := server.Exec("DROP TABLE `" + schema + "`.`users`").Error; err != nil {
		t.Fatal(err)
	}
	if data, err := Register(req); err == nil || data != nil {
		t.Fatalf("failed insert returned a session: data=%#v err=%v", data, err)
	}
}
