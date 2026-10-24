package logger

import (
	"backend/internal/config"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestNewValidatesConfig(t *testing.T) {
	for _, level := range []string{"debug", "info", "warn", "error", "dpanic", "panic", "fatal"} {
		for _, encoding := range []string{"console", "json"} {
			t.Run(level+"/"+encoding, func(t *testing.T) {
				if _, err := New(config.LogConfig{Level: level, Encoding: encoding}); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
	for _, cfg := range []config.LogConfig{
		{Level: "invalid", Encoding: "console"},
		{Level: "info", Encoding: "invalid"},
	} {
		if _, err := New(cfg); err == nil {
			t.Fatalf("expected invalid log configuration to fail: %+v", cfg)
		}
	}
}

func TestOutputAndLevelFiltering(t *testing.T) {
	for _, encoding := range []string{"console", "json"} {
		t.Run(encoding, func(t *testing.T) {
			output, err := os.CreateTemp(t.TempDir(), "stderr-*.log")
			if err != nil {
				t.Fatal(err)
			}
			original := os.Stderr
			os.Stderr = output
			t.Cleanup(func() {
				os.Stderr = original
				if err := output.Close(); err != nil {
					t.Error(err)
				}
			})
			logger, err := New(config.LogConfig{Level: "info", Encoding: encoding})
			if err != nil {
				t.Fatal(err)
			}
			logger.Debug("hidden debug event")
			logger.Info("visible info event")
			if err := logger.Sync(); err != nil {
				t.Fatal(err)
			}
			content, err := os.ReadFile(output.Name())
			if err != nil {
				t.Fatal(err)
			}
			text := strings.TrimSpace(string(content))
			if strings.Contains(text, "hidden debug event") || !strings.Contains(text, "visible info event") {
				t.Fatalf("unexpected output: %s", text)
			}
			if encoding == "json" {
				var entry map[string]any
				if err := json.Unmarshal(content, &entry); err != nil {
					t.Fatal(err)
				}
				if entry["level"] != "info" || entry["time"] == nil || !strings.Contains(entry["caller"].(string), "logger_test.go:") {
					t.Fatalf("missing log fields: %s", text)
				}
			} else if !strings.Contains(text, "logger_test.go:") || !strings.Contains(text, "info") {
				t.Fatalf("missing console fields: %s", text)
			}
		})
	}
}
