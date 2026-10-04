package file

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveAndDelete(t *testing.T) {
	root := t.TempDir()
	first, err := Save(root, "avatars", ".png", strings.NewReader("image content"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := Save(root, "avatars", ".png", strings.NewReader("another image"))
	if err != nil {
		t.Fatal(err)
	}
	if first == second || !strings.HasPrefix(first, "avatars/") || !strings.HasSuffix(first, ".png") {
		t.Fatalf("unexpected generated paths: %q, %q", first, second)
	}
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(first)))
	if err != nil || string(content) != "image content" {
		t.Fatalf("saved content = %q, error = %v", content, err)
	}
	for range 2 {
		if err := Delete(root, first); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(first))); !os.IsNotExist(err) {
		t.Fatalf("file still exists or stat failed: %v", err)
	}
}

type failingReader struct{}

var errRead = errors.New("read failed")

func (failingReader) Read(buffer []byte) (int, error) {
	return copy(buffer, "partial"), errRead
}

func TestSaveCleansUpFailedWrite(t *testing.T) {
	root := t.TempDir()
	key, err := Save(root, "feedbacks", ".jpg", failingReader{})
	if !errors.Is(err, errRead) || key != "" {
		t.Fatalf("key = %q, error = %v", key, err)
	}
	entries, err := os.ReadDir(filepath.Join(root, "feedbacks"))
	if err != nil || len(entries) != 0 {
		t.Fatalf("partial files remain: %v, error = %v", entries, err)
	}
}

func TestRejectInvalidPaths(t *testing.T) {
	root := t.TempDir()
	for _, key := range []string{"", "../outside", "avatars/../outside", "/absolute", "C:/outside", `avatars\outside`, "avatars//image", "./image"} {
		t.Run(key, func(t *testing.T) {
			if _, err := Save(root, key, ".png", strings.NewReader("content")); !errors.Is(err, ErrInvalidPath) {
				t.Fatalf("Save error = %v", err)
			}
			if err := Delete(root, key); !errors.Is(err, ErrInvalidPath) {
				t.Fatalf("Delete error = %v", err)
			}
			if _, err := AccessURL("/uploads", key); !errors.Is(err, ErrInvalidPath) {
				t.Fatalf("AccessURL error = %v", err)
			}
		})
	}
	for _, extension := range []string{"", "png", ".../jpg", ".png/other"} {
		if _, err := Save(root, "avatars", extension, strings.NewReader("content")); !errors.Is(err, ErrInvalidPath) {
			t.Fatalf("extension %q: error = %v", extension, err)
		}
	}
}

func TestAccessURL(t *testing.T) {
	for _, base := range []string{"/uploads/", "https://example.com/uploads"} {
		got, err := AccessURL(base, "avatars/image.png")
		want := strings.TrimSuffix(base, "/") + "/avatars/image.png"
		if err != nil || got != want {
			t.Fatalf("AccessURL(%q) = %q, %v; want %q", base, got, err, want)
		}
	}
}
