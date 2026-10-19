package mysql

import (
	"backend/internal/models"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	driver "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 仅由已有的 loopback 隔离数据库套件调用，第二个库也随测试销毁。
func testDAOIsolation(t *testing.T, first *DAO, server *gorm.DB, cfg *driver.Config) {
	t.Helper()
	schema := fmt.Sprintf("leave_di_test_%d", time.Now().UnixNano())
	if err := server.Exec("CREATE DATABASE `" + schema + "`").Error; err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := server.Exec("DROP DATABASE `" + schema + "`").Error; err != nil {
			t.Error(err)
		}
	}()
	secondConfig := *cfg
	secondConfig.DBName = schema
	db, err := gorm.Open(gormmysql.Open(secondConfig.FormatDSN()), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatal("open second isolated database failed")
	}
	connection, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	definition, err := os.ReadFile(filepath.Join("..", "..", "..", "database", "tables", "users.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(string(definition)).Error; err != nil {
		t.Fatal(err)
	}
	second := NewDAO(db)
	for _, tc := range []struct {
		dao      *DAO
		password string
	}{
		{dao: first, password: "first-password"},
		{dao: second, password: "second-password"},
	} {
		if err := tc.dao.CreateUser(&models.User{
			StudentID: "di-isolation",
			Password:  tc.password,
			Role:      models.RoleStudent,
			Status:    models.UserStatusActive,
		}); err != nil {
			t.Fatal(err)
		}
	}
	rollback := errors.New("rollback isolated transaction")
	err = first.Transaction(func(tx *gorm.DB) error {
		user, err := first.GetUserForUpdate(tx, "di-isolation")
		if err != nil {
			return err
		}
		if err := first.UpdateUserPassword(tx, user.ID, "rolled-back-password"); err != nil {
			return err
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("transaction returned %v", err)
	}
	for _, tc := range []struct {
		dao      *DAO
		password string
	}{
		{dao: first, password: "first-password"},
		{dao: second, password: "second-password"},
	} {
		user, err := tc.dao.FindUserByStudentID("di-isolation")
		if err != nil {
			t.Fatal(err)
		}
		if user.Password != tc.password {
			t.Fatal("DAO crossed database boundaries or transaction did not roll back")
		}
	}
}
