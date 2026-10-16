package file

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveAndDelete(t *testing.T) {
	root := t.TempDir()
	first, err := saveOne(root, "avatars", ".png", strings.NewReader("image content"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := saveOne(root, "avatars", ".png", strings.NewReader("another image"))
	if err != nil {
		t.Fatal(err)
	}
	if first == second || !strings.HasPrefix(first, "avatars/") || !strings.HasSuffix(first, ".png") {
		t.Fatalf("unexpected generated paths: %q, %q", first, second)
	}
	saved, err := Open(root, first)
	if err != nil {
		t.Fatal(err)
	}
	content, readErr := io.ReadAll(saved)
	err = errors.Join(readErr, saved.Close())
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
	key, err := saveOne(root, "feedbacks", ".jpg", failingReader{})
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
			if _, err := Save(root, key, Input{Extension: ".png", Content: strings.NewReader("content")}); !errors.Is(err, ErrInvalidPath) {
				t.Fatalf("Save error = %v", err)
			}
			if _, err := Open(root, key); !errors.Is(err, ErrInvalidPath) {
				t.Fatalf("Open error = %v", err)
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
		if _, err := Save(root, "avatars", Input{Extension: extension, Content: strings.NewReader("content")}); !errors.Is(err, ErrInvalidPath) {
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

func TestSaveBatch(t *testing.T) {
	for _, count := range []int{0, 1, 3} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			root := t.TempDir()
			files := make([]Input, count)
			for i := range files {
				files[i] = Input{Extension: ".txt", Content: strings.NewReader(fmt.Sprint(i))}
			}
			keys, err := Save(root, "files", files...)
			if err != nil || len(keys) != count {
				t.Fatalf("keys=%v err=%v", keys, err)
			}
			for i, key := range keys {
				data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(key)))
				if err != nil || string(data) != fmt.Sprint(i) {
					t.Fatalf("data=%q err=%v", data, err)
				}
			}
		})
	}
}

func TestSaveBatchRollback(t *testing.T) {
	for _, last := range []Input{
		{Extension: ".txt", Content: failingReader{}},
		{Extension: ".invalid/extension", Content: strings.NewReader("invalid")},
	} {
		root := t.TempDir()
		existing, err := Save(root, "files", Input{Extension: ".txt", Content: strings.NewReader("keep")})
		if err != nil {
			t.Fatal(err)
		}
		keys, err := Save(root, "files", Input{Extension: ".txt", Content: strings.NewReader("first")}, last)
		if err == nil || keys != nil {
			t.Fatalf("keys=%v err=%v", keys, err)
		}
		if last.Extension == ".txt" && !errors.Is(err, errRead) {
			t.Fatalf("lost read error: %v", err)
		}
		entries, err := os.ReadDir(filepath.Join(root, "files"))
		if err != nil || len(entries) != 1 {
			t.Fatalf("entries=%v err=%v", entries, err)
		}
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(existing[0])))
		if err != nil || string(data) != "keep" {
			t.Fatalf("existing file changed: %q %v", data, err)
		}
	}
}
