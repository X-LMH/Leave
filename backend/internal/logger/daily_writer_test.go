package logger

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDailyWriterAppendRolloverAndPreserveHistory(t *testing.T) {
	dir := t.TempDir()
	old := filepath.Join(dir, "2000-01-01.jsonl")
	if err := os.WriteFile(old, []byte("history\n"), 0600); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 26, 23, 59, 59, 0, time.FixedZone("CST", 8*3600))
	for _, message := range []string{"first\n", "restart\n"} {
		w, err := newDailyWriter(dir)
		if err != nil {
			t.Fatal(err)
		}
		w.now = func() time.Time { return now }
		if _, err := w.Write([]byte(message)); err != nil {
			t.Fatal(err)
		}
		if message == "restart\n" {
			now = now.Add(time.Second)
			if _, err := w.Write([]byte("next day\n")); err != nil {
				t.Fatal(err)
			}
		}
	}
	for name, want := range map[string]string{
		"2000-01-01.jsonl": "history\n",
		"2026-10-26.jsonl": "first\nrestart\n",
		"2026-10-27.jsonl": "next day\n",
	} {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || string(got) != want {
			t.Fatalf("%s: %q, %v", name, got, err)
		}
	}
}

func TestDailyWriterConcurrentLines(t *testing.T) {
	w, err := newDailyWriter(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	w.now = func() time.Time { return now }
	var workers sync.WaitGroup
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 50 {
				if _, err := w.Write([]byte("{\"ok\":true}\n")); err != nil {
					t.Error(err)
					return
				}
			}
		}()
	}
	workers.Wait()
	content, err := os.ReadFile(filepath.Join(w.dir, now.Format(time.DateOnly)+".jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 400 {
		t.Fatalf("got %d lines", len(lines))
	}
	for _, line := range lines {
		if line != `{"ok":true}` {
			t.Fatalf("broken line: %s", line)
		}
	}
}

func TestDailyWriterOpenFailures(t *testing.T) {
	dir := t.TempDir()
	blocked := filepath.Join(dir, "file")
	if err := os.WriteFile(blocked, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := newDailyWriter(filepath.Join(blocked, "logs")); err == nil {
		t.Fatal("expected directory failure")
	}
	w, err := newDailyWriter(dir)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().AddDate(0, 0, 1)
	path := filepath.Join(dir, now.Format(time.DateOnly)+".jsonl")
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	w.now = func() time.Time { return now }
	if _, err := w.Write([]byte("failed\n")); err == nil {
		t.Fatal("expected open failure")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("recovered\n")); err != nil {
		t.Fatal(err)
	}
}
