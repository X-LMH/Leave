package app

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/response"
	appservice "backend/internal/service/app"
	"backend/internal/utils/jwt"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestRegisterRejectsInvalidRequests(t *testing.T) {
	settings := config.Config{}
	dao := mysql.NewDAO(nil, zap.NewNop())
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name string
		edit func(*dto.RegisterRequest)
	}{
		{"invalid student ID", func(p *dto.RegisterRequest) { p.StudentID = "123" }},
		{"short password", func(p *dto.RegisterRequest) {
			p.Password = "123"
			p.RePassword = "123"
		}},
		{"password mismatch", func(p *dto.RegisterRequest) { p.RePassword = "different" }},
		{"long device name", func(p *dto.RegisterRequest) { p.DeviceName = strings.Repeat("机", 256) }},
		{"long version", func(p *dto.RegisterRequest) { p.AppVersion = strings.Repeat("v", 65) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := dto.RegisterRequest{StudentID: "202410201016", Password: "123456", RePassword: "123456"}
			tc.edit(&p)
			body, err := json.Marshal(p)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			NewUserHandler(appservice.NewUserService(dao, jwt.NewManager(settings.JWT.Secret, settings.JWT.ExpirationDays), settings.Storage, zap.NewNop()), zap.NewNop()).RegisterHandler(c)
			var result response.Response
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Code != response.CodeInvalidParam || w.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
}

func TestValidStudentID(t *testing.T) {
	tests := []struct {
		name      string
		studentID string
		want      bool
	}{
		{name: "twelve digits", studentID: "202600010001", want: true},
		{name: "surrounding spaces are ignored", studentID: " 202600010001 ", want: true},
		{name: "too short", studentID: "20260001000", want: false},
		{name: "too long", studentID: "2026000100011", want: false},
		{name: "contains letters", studentID: "20260001000a", want: false},
		{name: "contains punctuation", studentID: "20260001000-", want: false},
		{name: "empty", studentID: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validStudentID(tt.studentID); got != tt.want {
				t.Fatalf("validStudentID(%q) = %v, want %v", tt.studentID, got, tt.want)
			}
		})
	}
}
