package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoadRedisConfig(t *testing.T) {
	for _, tt := range []struct {
		name      string
		yaml      string
		password  string
		want      RedisConfig
		wantError string
	}{
		{name: "defaults", want: RedisConfig{Host: "127.0.0.1", Port: 6379}},
		{name: "yaml", yaml: "redis:\n  host: localhost\n  port: 6380\n  username: app\n  password: yaml-secret\n  db: 2\n", want: RedisConfig{Host: "localhost", Port: 6380, Username: "app", Password: "yaml-secret", DB: 2}},
		{name: "environment", yaml: "redis:\n  password: yaml-secret\n", password: "env-secret", want: RedisConfig{Host: "127.0.0.1", Port: 6379, Password: "env-secret"}},
		{name: "empty host", yaml: "redis:\n  host: ' '\n", wantError: "redis.host"},
		{name: "invalid port", yaml: "redis:\n  port: 65536\n", wantError: "redis.port"},
		{name: "zero port", yaml: "redis:\n  port: 0\n", wantError: "redis.port"},
		{name: "negative db", yaml: "redis:\n  db: -1\n", wantError: "redis.db"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			t.Setenv("MYSQL_PASSWORD", "test-password")
			t.Setenv("JWT_SECRET", "test-secret")
			t.Setenv("REDIS_PASSWORD", tt.password)
			if tt.password == "" {
				if err := os.Unsetenv("REDIS_PASSWORD"); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Mkdir("config", 0755); err != nil {
				t.Fatal(err)
			}
			content := "jwt:\n  expiration_days: 30\nstorage:\n  root_dir: ./uploads\n  base_url: /uploads\n" + tt.yaml
			if err := os.WriteFile("config/config.local.yaml", []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			cfg, err := Load()
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("got %v, want %s error", err, tt.wantError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Redis != tt.want {
				t.Fatal("unexpected Redis configuration")
			}
		})
	}
}
