package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindConfigFile(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  string
	}{
		{name: "local takes precedence", files: []string{"config.local.yaml", "config.test.yaml", "config.server.yaml"}, want: "config/config.local.yaml"},
		{name: "test takes precedence over server", files: []string{"config.test.yaml", "config.server.yaml"}, want: "config/config.test.yaml"},
		{name: "production still supported", files: []string{"config.server.yaml"}, want: "config/config.server.yaml"},
		{name: "templates are not selected", files: []string{"config.test.yaml.example", "other.yaml"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			if err := os.Mkdir("config", 0755); err != nil {
				t.Fatal(err)
			}
			for _, name := range tt.files {
				if err := os.WriteFile(filepath.Join("config", name), nil, 0600); err != nil {
					t.Fatal(err)
				}
			}
			got, err := FindConfigFile()
			if tt.want == "" {
				if err == nil {
					t.Fatal("expected missing config error")
				}
				return
			}
			if err != nil || filepath.Clean(got) != filepath.Clean(tt.want) {
				t.Fatalf("got %q, %v; want %q", got, err, tt.want)
			}
		})
	}
}

func TestLoadLogConfig(t *testing.T) {
	for _, tt := range []struct {
		name    string
		logYAML string
		want    LogConfig
	}{
		{name: "defaults", want: LogConfig{Level: "info", Encoding: "console"}},
		{name: "override", logYAML: "log:\n  level: debug\n  encoding: json\n", want: LogConfig{Level: "debug", Encoding: "json"}},
		{name: "partial", logYAML: "log:\n  encoding: json\n", want: LogConfig{Level: "info", Encoding: "json"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			t.Setenv("MYSQL_PASSWORD", "test-password")
			t.Setenv("JWT_SECRET", "test-secret")
			if err := os.Mkdir("config", 0755); err != nil {
				t.Fatal(err)
			}
			content := "redis:\n  host: localhost\n  port: 6379\njwt:\n  expiration_days: 30\nstorage:\n  root_dir: ./uploads\n  base_url: /uploads\n" + tt.logYAML
			if err := os.WriteFile("config/config.local.yaml", []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			cfg, err := Load()
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Log != tt.want {
				t.Fatalf("got %+v, want %+v", cfg.Log, tt.want)
			}
		})
	}
}

func TestInitReadsTestConfig(t *testing.T) {

	t.Chdir(t.TempDir())
	t.Setenv("MYSQL_PASSWORD", "test-password")
	t.Setenv("JWT_SECRET", "test-secret")
	if err := os.Mkdir("config", 0755); err != nil {
		t.Fatal(err)
	}
	content := "redis:\n  host: localhost\n  port: 6379\napp:\n  port: 9000\nmysql:\n  database: leave_list_test\njwt:\n  expiration_days: 30\nstorage:\n  root_dir: ./uploads\n  base_url: /api/v1/uploads\n"
	if err := os.WriteFile("config/config.test.yaml", []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Mysql.Database != "leave_list_test" || cfg.App.Port != 9000 {
		t.Fatalf("wrong test configuration: database=%q port=%d", cfg.Mysql.Database, cfg.App.Port)
	}
}

func TestInitFromReleaseDirectory(t *testing.T) {
	for _, layout := range []string{"releases/v1", "backend/releases/v1"} {
		t.Run(layout, func(t *testing.T) {

			root := t.TempDir()
			if err := os.Mkdir(filepath.Join(root, "config"), 0755); err != nil {
				t.Fatal(err)
			}
			content := "redis:\n  host: localhost\n  port: 6379\napp:\n  port: 9000\nmysql:\n  database: leave_list_test\njwt:\n  expiration_days: 30\nstorage:\n  root_dir: /test/uploads\n  base_url: /api/v1/uploads\n"
			if err := os.WriteFile(filepath.Join(root, "config/config.test.yaml"), []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, ".env"), []byte("MYSQL_PASSWORD=release-password\nREDIS_PASSWORD=redis-release-password\nJWT_SECRET=release-secret\n"), 0600); err != nil {
				t.Fatal(err)
			}
			for _, key := range []string{"MYSQL_PASSWORD", "REDIS_PASSWORD", "JWT_SECRET"} {
				t.Setenv(key, "")
				if err := os.Unsetenv(key); err != nil {
					t.Fatal(err)
				}
			}
			cwd := filepath.Join(root, filepath.FromSlash(layout))
			if err := os.MkdirAll(cwd, 0755); err != nil {
				t.Fatal(err)
			}
			t.Chdir(cwd)
			cfg, err := Load()
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Mysql.Database != "leave_list_test" || cfg.Mysql.Password != "release-password" || cfg.Redis.Password != "redis-release-password" || cfg.JWT.Secret != "release-secret" {
				t.Fatal("release directory did not load root test YAML and .env")
			}
		})
	}
}

func TestInitValidatesStorage(t *testing.T) {
	t.Setenv("MYSQL_PASSWORD", "test-password")
	t.Setenv("JWT_SECRET", "test-secret")
	tests := []struct {
		name      string
		root      string
		base      string
		wantError string
	}{
		{name: "missing root", base: "/uploads", wantError: "storage.root_dir"},
		{name: "missing base", root: "./uploads", wantError: "storage.base_url"},
		{name: "blank root", root: "  ", base: "/uploads", wantError: "storage.root_dir"},
		{name: "blank base", root: "./uploads", base: "  ", wantError: "storage.base_url"},
		{name: "valid", root: "./uploads", base: "/uploads"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			t.Chdir(t.TempDir())
			if err := os.Mkdir("config", 0755); err != nil {
				t.Fatal(err)
			}
			content := fmt.Sprintf("redis:\n  host: localhost\n  port: 6379\njwt:\n  expiration_days: 30\nstorage:\n  root_dir: %q\n  base_url: %q\n", tt.root, tt.base)
			if err := os.WriteFile("config/config.local.yaml", []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			_, err := Load()
			if tt.wantError == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("expected error for %s, got %v", tt.wantError, err)
			}
		})
	}
}

func TestLoadKeepsIndependentConfiguration(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("MYSQL_PASSWORD", "environment-password")
	t.Setenv("JWT_SECRET", "environment-secret")
	if err := os.Mkdir("config", 0755); err != nil {
		t.Fatal(err)
	}
	path := "config/config.local.yaml"
	firstYAML := "redis:\n  host: localhost\n  port: 6379\napp:\n  port: 9001\n  package_dir: ./first-apks\nmysql:\n  password: yaml-password\njwt:\n  secret: yaml-secret\n  expiration_days: 1\nstorage:\n  root_dir: ./first\n  base_url: /first\n"
	if err := os.WriteFile(path, []byte(firstYAML), 0600); err != nil {
		t.Fatal(err)
	}
	first, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	secondYAML := "redis:\n  host: localhost\n  port: 6379\njwt:\n  expiration_days: 7\nstorage:\n  root_dir: ./second\n  base_url: /second\n"
	if err := os.WriteFile(path, []byte(secondYAML), 0600); err != nil {
		t.Fatal(err)
	}
	second, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if first.App.Port != 9001 || first.Storage.RootDir != "./first" || first.JWT.ExpirationDays != 1 {
		t.Fatal("loading another configuration changed the first instance")
	}
	if second.App.Port != 10000 || second.App.PackageDir != "" || second.Storage.RootDir != "./second" || second.JWT.ExpirationDays != 7 {
		t.Fatal("second load retained settings from the first load")
	}
	for _, cfg := range []Config{first, second} {
		if cfg.Mysql.Password != "environment-password" || cfg.JWT.Secret != "environment-secret" {
			t.Fatal("environment overrides were not preserved")
		}
	}
}
