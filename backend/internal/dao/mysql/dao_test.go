package mysql

import (
	"backend/internal/models"
	"testing"

	"go.uber.org/zap"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestDAOsUseTheirOwnConnections(t *testing.T) {
	newDAO := func(calls *int) *DAO {
		t.Helper()
		db, err := gorm.Open(gormmysql.New(gormmysql.Config{
			DSN:                       "root:test-only@tcp(localhost:3306)/leave",
			SkipInitializeWithVersion: true,
		}), &gorm.Config{
			DryRun:                 true,
			DisableAutomaticPing:   true,
			SkipDefaultTransaction: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		connection, err := db.DB()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := connection.Close(); err != nil {
				t.Error(err)
			}
		})
		if err := db.Callback().Create().After("gorm:create").Register("count_instance_creates", func(*gorm.DB) {
			*calls++
		}); err != nil {
			t.Fatal(err)
		}
		return NewDAO(db, zap.NewNop())
	}
	var firstCalls, secondCalls int
	first := newDAO(&firstCalls)
	second := newDAO(&secondCalls)
	for _, dao := range []*DAO{first, second, first} {
		if err := dao.CreateUser(&models.User{StudentID: "isolated"}); err != nil {
			t.Fatal(err)
		}
	}
	if firstCalls != 2 || secondCalls != 1 {
		t.Fatalf("queries used another DAO's connection: first=%d second=%d", firstCalls, secondCalls)
	}
}
