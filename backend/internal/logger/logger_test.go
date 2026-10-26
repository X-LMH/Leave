package logger

import (
	"backend/internal/config"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewValidatesConfig(t *testing.T) {
	for _, level := range []string{"debug", "info", "warn", "error", "dpanic", "panic", "fatal"} {
		for _, encoding := range []string{encodingConsole, encodingJSON} {
			t.Run(level+"/"+encoding, func(t *testing.T) {
				if _, err := NewLogger(config.LogConfig{Level: level, Encoding: encoding}); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
	for _, cfg := range []config.LogConfig{
		{Level: "invalid", Encoding: encodingConsole},
		{Level: "info", Encoding: "invalid"},
	} {
		if _, err := NewLogger(cfg); err == nil {
			t.Fatalf("expected invalid log configuration to fail: %+v", cfg)
		}
	}
}

func TestOutputAndLevelFiltering(t *testing.T) {
	for _, encoding := range []string{encodingConsole, encodingJSON} {
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
			logger, err := NewLogger(config.LogConfig{Level: "info", Encoding: encoding})
			if err != nil {
				t.Fatal(err)
			}
			logger.Debug("hidden debug event")
			logger.Info("visible info event", zap.Duration("elapsed", 150*time.Millisecond))
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
			if encoding == encodingJSON {
				var entry map[string]any
				if err := json.Unmarshal(content, &entry); err != nil {
					t.Fatal(err)
				}
				if entry["level"] != "info" || entry["time"] == nil || !strings.Contains(entry["caller"].(string), "logger_test.go:") {
					t.Fatalf("missing log fields: %s", text)
				}
				if _, err := time.Parse("2006-01-02 15:04:05", entry["time"].(string)); err != nil {
					t.Fatalf("unexpected JSON time: %s", text)
				}
				if entry["elapsed"] != "150ms" || strings.Contains(text, `\u001b`) {
					t.Fatalf("unexpected JSON style: %s", text)
				}
			} else {
				if !strings.Contains(text, "INFO") || !strings.Contains(text, "\x1b[") || !strings.Contains(text, "150ms") {
					t.Fatalf("unexpected console style: %s", text)
				}
				if _, err := time.Parse("2006-01-02 15:04:05", strings.SplitN(text, "\t", 2)[0]); err != nil {
					t.Fatalf("unexpected console time: %s", text)
				}
			}
			if !strings.Contains(text, "backend/internal/logger/logger_test.go:") {
				t.Fatalf("missing relative caller: %s", text)
			}
		})
	}
}
