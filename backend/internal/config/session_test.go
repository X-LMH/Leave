package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSessionConfiguration(t *testing.T) {
	const appYAML = "app:\n  namespace: isolated\n  port: 8080\n"
	const sessionYAML = "session:\n  student_ttl: 24h\n  admin_ttl: 2h\n"
	for _, tc := range []struct {
		name      string
		yaml      string
		wantError string
	}{
		{name: "explicit without JWT", yaml: appYAML + sessionYAML},
		{name: "missing namespace", yaml: "app:\n  port: 8080\n" + sessionYAML, wantError: "app.namespace"},
		{name: "empty namespace", yaml: "app:\n  namespace: ''\n  port: 8080\n" + sessionYAML, wantError: "app.namespace"},
		{name: "ambiguous namespace", yaml: "app:\n  namespace: 'test:auth'\n  port: 8080\n" + sessionYAML, wantError: "app.namespace"},
		{name: "legacy namespace does not replace app namespace", yaml: "app:\n  port: 8080\nsession:\n  namespace: legacy\n  student_ttl: 24h\n  admin_ttl: 2h\n", wantError: "app.namespace"},
		{name: "missing port", yaml: "app:\n  namespace: isolated\n" + sessionYAML, wantError: "app.port"},
		{name: "missing session", yaml: appYAML, wantError: "session TTLs"},
		{name: "missing admin TTL", yaml: appYAML + "session:\n  student_ttl: 24h\n", wantError: "session TTLs"},
		{name: "missing student TTL", yaml: appYAML + "session:\n  admin_ttl: 2h\n", wantError: "session TTLs"},
		{name: "invalid duration", yaml: appYAML + "session:\n  student_ttl: invalid\n  admin_ttl: 2h\n", wantError: "session.student_ttl"},
		{name: "zero duration", yaml: appYAML + "session:\n  student_ttl: 24h\n  admin_ttl: 0s\n", wantError: "session TTLs"},
		{name: "negative duration", yaml: appYAML + "session:\n  student_ttl: 24h\n  admin_ttl: -1h\n", wantError: "session TTLs"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			t.Setenv("MYSQL_PASSWORD", "isolated-test-password")
			t.Setenv("JWT_SECRET", "")
			if err := os.Mkdir("config", 0755); err != nil {
				t.Fatal(err)
			}
			content := "redis:\n  host: localhost\n  port: 6379\nstorage:\n  root_dir: ./uploads\n  base_url: /uploads\n" + tc.yaml
			if err := os.WriteFile(filepath.Join("config", "config.local.yaml"), []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			cfg, err := Load()
			if tc.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantError) {
					t.Fatalf("got %v, want %s error", err, tc.wantError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if cfg.App.Namespace != "isolated" || cfg.App.Port != 8080 || cfg.Session.StudentTTL != 24*time.Hour || cfg.Session.AdminTTL != 2*time.Hour {
				t.Fatal("explicit configuration was not decoded correctly")
			}
		})
	}
}
