package router

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/response"
	adminservice "backend/internal/service/admin"
	appservice "backend/internal/service/app"
	"backend/internal/utils/jwt"
	passwordutil "backend/internal/utils/password"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	driver "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAdminPasswordIntegration(t *testing.T) {
	settings := config.Config{}
	dao := mysql.NewDAO(nil)
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
	server, err := gorm.Open(gormmysql.Open(cfg.FormatDSN()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("cannot connect to isolated MySQL server")
	}
	serverSQL, err := server.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer serverSQL.Close()
	schema := fmt.Sprintf("leave_password_test_%d", time.Now().UnixNano())
	if err := server.Exec("CREATE DATABASE `" + schema + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci").Error; err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := server.Exec("DROP DATABASE `" + schema + "`").Error; err != nil {
			t.Errorf("drop test schema: %v", err)
		}
	}()
	cfg.DBName = schema
	testDB, err := gorm.Open(gormmysql.Open(cfg.FormatDSN()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("cannot open isolated test schema")
	}
	connection, err := testDB.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	definition, err := os.ReadFile(filepath.Join("..", "..", "database", "tables", "users.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if err := testDB.Exec(string(definition)).Error; err != nil {
		t.Fatal(err)
	}

	settings.JWT.Secret = "isolated-password-test"
	settings.JWT.ExpirationDays = 1
	dao = mysql.NewDAO(testDB)
	initialHash, err := passwordutil.Hash("old-pass")
	if err != nil {
		t.Fatal(err)
	}
	for _, user := range []models.User{
		{StudentID: "admin", Password: initialHash, Role: models.RoleAdmin, Status: models.UserStatusActive},
		{StudentID: "student", Password: initialHash, Role: models.RoleStudent, Status: models.UserStatusActive},
		{StudentID: "disabled", Password: initialHash, Role: models.RoleAdmin, Status: models.UserStatusDisabled},
	} {
		if err := testDB.Create(&user).Error; err != nil {
			t.Fatal(err)
		}
	}
	token, err := jwt.NewManager(settings.JWT.Secret, settings.JWT.ExpirationDays).GenerateToken("admin", models.RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	router := SetupRouter(testDependencies(dao, settings))
	call := func(auth, body string) response.Response {
		t.Helper()
		req := httptest.NewRequest("POST", "/api/v1/admin/auth/password", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if auth != "" {
			req.Header.Set("Authorization", "Bearer "+auth)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var result response.Response
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if w.Code != result.Code.HTTPStatus() || result.Data != nil {
			t.Fatal("unexpected HTTP status or credential data in response")
		}
		return result
	}
	validBody := `{"password":"old-pass","new_password":"new-pass","re_password":"new-pass"}`
	if got := call("", validBody); got.Code != response.CodeNeedLogin {
		t.Fatal("anonymous password change allowed")
	}
	for _, account := range []struct{ id, role string }{
		{"student", models.RoleStudent},
		{"student", models.RoleAdmin}, // Stale or incorrect admin claim must not grant access.
		{"disabled", models.RoleAdmin},
		{"missing", models.RoleAdmin},
	} {
		auth, err := jwt.NewManager(settings.JWT.Secret, settings.JWT.ExpirationDays).GenerateToken(account.id, account.role)
		if err != nil {
			t.Fatal(err)
		}
		if got := call(auth, validBody); got.Code != response.CodeNeedLogin {
			t.Fatalf("account %s accepted", account.id)
		}
	}
	for _, body := range []string{
		`{`,
		`{}`,
		`{"password":"old-pass","new_password":"short","re_password":"short"}`,
		`{"password":"old-pass","new_password":"old-pass","re_password":"old-pass"}`,
		`{"password":"old-pass","new_password":"new-pass","re_password":"different"}`,
	} {
		if got := call(token, body); got.Code != response.CodeInvalidParam {
			t.Fatalf("invalid request accepted: %d", got.Code)
		}
	}
	if got := call(token, strings.Replace(validBody, "old-pass", "wrong-pass", 1)); got.Code != response.CodeNotRightPassword {
		t.Fatal("incorrect original password accepted")
	}
	user, err := dao.FindUserByStudentID("admin")
	if err != nil || user.Password != initialHash {
		t.Fatal("failed request modified password")
	}
	// No App version headers are sent, and the same token remains valid after success.
	if got := call(token, validBody); got.Code != response.CodeSuccess {
		t.Fatalf("change failed: %d", got.Code)
	}
	authService := adminservice.NewAuthService(dao, jwt.NewManager(settings.JWT.Secret, settings.JWT.ExpirationDays))
	if _, err := authService.AdminLogin(&dto.AdminLoginRequest{UserName: "admin", Password: "old-pass"}); !errors.Is(err, mysql.ErrorInvalidPassword) {
		t.Fatal("old password still accepted for login")
	}
	if _, err := authService.AdminLogin(&dto.AdminLoginRequest{UserName: "admin", Password: "new-pass"}); err != nil {
		t.Fatal("new password cannot log in")
	}
	if got := call(token, `{"password":"new-pass","new_password":" pass ","re_password":" pass "}`); got.Code != response.CodeSuccess {
		t.Fatal("current session or password spaces were not preserved")
	}
	// Both requests start with the same original password; exactly one can commit.
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, password := range []string{"winner-one", "winner-two"} {
		go func(next string) {
			<-start
			results <- authService.ChangePassword("admin", &dto.PasswordChangeRequest{
				Password:    " pass ",
				NewPassword: next,
				RePassword:  next,
			})
		}(password)
	}
	close(start)
	successes := 0
	for range 2 {
		if err := <-results; err == nil {
			successes++
		} else if !errors.Is(err, mysql.ErrorNotRightPassword) {
			t.Fatalf("unexpected concurrent result: %v", err)
		}
	}
	if successes != 1 {
		t.Fatalf("concurrent successes: %d", successes)
	}
	user, err = dao.FindUserByStudentID("admin")
	if err != nil {
		t.Fatal(err)
	}
	first, firstErr := passwordutil.Verify(user.Password, "winner-one")
	second, secondErr := passwordutil.Verify(user.Password, "winner-two")
	if firstErr != nil || secondErr != nil || first == second || user.Role != models.RoleAdmin || user.Status != models.UserStatusActive {
		t.Fatal("unexpected persisted account state")
	}
	userService := appservice.NewUserService(dao, jwt.NewManager(settings.JWT.Secret, settings.JWT.ExpirationDays), settings.Storage)
	t.Run("student password reset", func(t *testing.T) {
		definition, err := os.ReadFile(filepath.Join("..", "..", "database", "tables", "profiles.sql"))
		if err != nil {
			t.Fatal(err)
		}
		if err := testDB.Exec(string(definition)).Error; err != nil {
			t.Fatal(err)
		}

		callReset := func(auth, id string) response.Response {
			t.Helper()
			req := httptest.NewRequest("POST", "/api/v1/admin/students/"+id+"/reset-password", nil)
			if auth != "" {
				req.Header.Set("Authorization", "Bearer "+auth)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			var result response.Response
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if w.Code != result.Code.HTTPStatus() || result.Data != nil {
				t.Fatal("unexpected status or credential response")
			}
			return result
		}
		var student models.User
		if err := testDB.Where("student_id = ?", "student").First(&student).Error; err != nil {
			t.Fatal(err)
		}
		id := strconv.FormatUint(uint64(student.ID), 10)
		if got := callReset("", id); got.Code != response.CodeNeedLogin {
			t.Fatal("anonymous reset allowed")
		}
		for _, account := range []struct{ id, role string }{
			{"student", models.RoleStudent},
			{"student", models.RoleAdmin}, // 数据库角色必须覆盖令牌中的过期角色声明。
			{"disabled", models.RoleAdmin},
			{"missing", models.RoleAdmin},
		} {
			auth, err := jwt.NewManager(settings.JWT.Secret, settings.JWT.ExpirationDays).GenerateToken(account.id, account.role)
			if err != nil {
				t.Fatal(err)
			}
			if got := callReset(auth, id); got.Code != response.CodeNeedLogin {
				t.Fatalf("account %s reset allowed", account.id)
			}
		}
		var unchangedStudent models.User
		if err := testDB.First(&unchangedStudent, student.ID).Error; err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(student, unchangedStudent) {
			t.Fatal("rejected reset changed student account")
		}
		for _, invalid := range []string{"0", "-1", "abc", "4294967296"} {
			if got := callReset(token, invalid); got.Code != response.CodeInvalidParam {
				t.Fatal("invalid ID accepted")
			}
		}
		var admin models.User
		if err := testDB.Where("student_id = ?", "admin").First(&admin).Error; err != nil {
			t.Fatal(err)
		}
		if got := callReset(token, strconv.FormatUint(uint64(admin.ID), 10)); got.Code != response.CodeManagementNotFound {
			t.Fatal("admin target accepted")
		}
		if got := callReset(token, "4294967295"); got.Code != response.CodeManagementNotFound {
			t.Fatal("missing target accepted")
		}
		for _, target := range []*models.User{
			&student,
			{StudentID: "disabled-student", Password: initialHash, Role: models.RoleStudent, Status: models.UserStatusDisabled},
			{StudentID: "deleted-student", Password: initialHash, Role: models.RoleStudent, Status: models.UserStatusActive},
		} {
			if target.ID == 0 {
				if err := testDB.Create(target).Error; err != nil {
					t.Fatal(err)
				}
			}
			deleted := target.StudentID == "deleted-student"
			if deleted {
				if err := testDB.Delete(target).Error; err != nil {
					t.Fatal(err)
				}
			}
			if err := testDB.Unscoped().First(target, target.ID).Error; err != nil {
				t.Fatal(err)
			}
			before := *target
			got := callReset(token, strconv.FormatUint(uint64(target.ID), 10))
			expected := response.CodeSuccess
			if deleted {
				expected = response.CodeManagementNotFound
			}
			if got.Code != expected {
				t.Fatalf("reset %s: %d", target.StudentID, got.Code)
			}
			var after models.User
			if err := testDB.Unscoped().First(&after, target.ID).Error; err != nil {
				t.Fatal(err)
			}
			if deleted {
				if after.Password != before.Password {
					t.Fatal("deleted account changed")
				}
			} else {
				matched, err := passwordutil.Verify(after.Password, "12345678")
				if err != nil || !matched || after.Password == "12345678" {
					t.Fatal("default password hash invalid")
				}
				matched, err = passwordutil.Verify(after.Password, "old-pass")
				if err != nil || matched {
					t.Fatal("old password still valid")
				}
			}
			after.Password, after.UpdatedAt = before.Password, before.UpdatedAt
			if !reflect.DeepEqual(after, before) {
				t.Fatal("reset changed unrelated account fields")
			}
		}

		if _, err := userService.Login(&dto.LoginRequest{StudentID: "student", Password: "old-pass"}); !errors.Is(err, mysql.ErrorInvalidPassword) {
			t.Fatal("old student password can still log in")
		}
		if _, err := userService.Login(&dto.LoginRequest{StudentID: "student", Password: "12345678"}); err != nil {
			t.Fatalf("default password cannot log in: %v", err)
		}
		var unchangedAdmin models.User
		if err := testDB.First(&unchangedAdmin, admin.ID).Error; err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(admin, unchangedAdmin) {
			t.Fatal("reset changed administrator")
		}
		var untouched models.User
		if err := testDB.Where("student_id = ?", "disabled").First(&untouched).Error; err != nil {
			t.Fatal(err)
		}
		if untouched.Password != initialHash {
			t.Fatal("unrelated password changed")
		}
	})
}
