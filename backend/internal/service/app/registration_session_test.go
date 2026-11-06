package app

import (
	"context"
	"errors"
	"testing"

	"backend/internal/config"
	"backend/internal/dao/mysql"
	redisdao "backend/internal/dao/redis"
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/session"
	"backend/internal/testutil/sessiontest"

	"go.uber.org/zap"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestRegistrationSessionFailurePreservesCreatedAccount(t *testing.T) {
	db, err := gorm.Open(gormmysql.New(gormmysql.Config{
		DSN:                       "tcp(127.0.0.1:1)/isolated_register_test",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
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
	var created *models.User
	var insertErr error
	if err := db.Callback().Create().Replace("gorm:create", func(tx *gorm.DB) {
		if insertErr != nil {
			tx.AddError(insertErr)
			return
		}
		user := tx.Statement.Dest.(*models.User)
		user.ID = 1
		copy := *user
		created = &copy
		tx.RowsAffected = 1
	}); err != nil {
		t.Fatal(err)
	}
	r := sessiontest.NewSessionRedis()
	r.Err = errors.New("isolated Redis unavailable")
	s := session.NewStore(redisdao.NewDAO(r, "default"), config.SessionConfig{})
	service := NewUserService(mysql.NewDAO(db, zap.NewNop()), s, config.StorageConfig{BaseURL: "/uploads"}, zap.NewNop())
	input := &dto.RegisterRequest{
		StudentID:  "202410201016",
		Password:   "123456",
		RePassword: "123456",
	}
	data, err := service.Register(context.Background(), input)
	if data != nil || !errors.Is(err, session.ErrRegistrationSession) || created == nil || created.AuthVersion != 1 {
		t.Fatal("registration partial success was not distinguished", err)
	}
	if len(r.Values) != 0 {
		t.Fatal("failed session returned a credential")
	}
	// A failed database insert must never establish a Redis session.
	r.Err = nil
	insertErr = errors.New("isolated insert failed")
	data, err = service.Register(context.Background(), input)
	if data != nil || err == nil || errors.Is(err, session.ErrRegistrationSession) || len(r.Values) != 0 {
		t.Fatal("insert failure created session", err)
	}
	insertErr = nil
	data, err = service.Register(context.Background(), input)
	if err != nil || data == nil || data.Token == "" {
		t.Fatal("registration success", err)
	}
	record, err := s.Get(context.Background(), data.Token)
	if err != nil || record.UserID != created.ID || record.StudentID != input.StudentID {
		t.Fatal("session not bound to inserted account", err)
	}
}
