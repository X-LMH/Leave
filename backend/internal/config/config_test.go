package config

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestInitValidatesStorage(t *testing.T) {
	previous := Cfg
	t.Cleanup(func() {
		Cfg = previous
		viper.Reset()
	})
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
			viper.Reset()
			t.Chdir(t.TempDir())
			if err := os.Mkdir("config", 0755); err != nil {
				t.Fatal(err)
			}
			content := fmt.Sprintf("jwt:\n  expiration_days: 30\nstorage:\n  root_dir: %q\n  base_url: %q\n", tt.root, tt.base)
			if err := os.WriteFile("config/config.local.yaml", []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			err := Init()
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
