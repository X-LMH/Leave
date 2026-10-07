package app

import (
	"backend/internal/service"
	"bytes"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestReadAvatar(t *testing.T) {
	var pngData, jpegData, gifData, wideData bytes.Buffer
	picture := image.NewRGBA(image.Rect(0, 0, 2, 2))
	if err := png.Encode(&pngData, picture); err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(&jpegData, picture, nil); err != nil {
		t.Fatal(err)
	}
	if err := gif.Encode(&gifData, picture, nil); err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(&wideData, image.NewRGBA(image.Rect(0, 0, 4097, 1))); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		data      []byte
		extension string
	}{
		{name: "PNG", data: pngData.Bytes(), extension: ".png"},
		{name: "JPEG", data: jpegData.Bytes(), extension: ".jpg"},
		{name: "empty"},
		{name: "text disguised as image", data: []byte("not an image")},
		{name: "unsupported GIF", data: gifData.Bytes()},
		{name: "truncated PNG", data: pngData.Bytes()[:len(pngData.Bytes())-20]},
		{name: "oversized dimensions", data: wideData.Bytes()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, extension, err := readAvatar(bytes.NewReader(tt.data))
			if tt.extension == "" {
				if !errors.Is(err, ErrorInvalidAvatar) {
					t.Fatalf("expected invalid avatar, got %v", err)
				}
				return
			}
			if err != nil || extension != tt.extension || !bytes.Equal(data, tt.data) {
				t.Fatalf("readAvatar returned extension=%q err=%v", extension, err)
			}
		})
	}
}

type failedAvatarReader struct{}

func (failedAvatarReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestReadAvatarPreservesReadError(t *testing.T) {
	_, _, err := readAvatar(failedAvatarReader{})
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected read error, got %v", err)
	}
}

func TestDeleteOldAvatarPreservesDefault(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "avatars"), 0755); err != nil {
		t.Fatal(err)
	}
	keys := []string{service.DefaultAvatarPath, "avatars/old.jpg", "other.jpg"}
	for _, key := range keys {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(key)), []byte("test"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, key := range append(keys, "", "avatars/missing.jpg") {
		if err := deleteOldAvatar(root, key); err != nil {
			t.Fatalf("deleteOldAvatar(%q): %v", key, err)
		}
	}
	for _, key := range []string{service.DefaultAvatarPath, "other.jpg"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(key))); err != nil {
			t.Fatalf("protected file %q should remain: %v", key, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "avatars", "old.jpg")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("old avatar should be deleted, got %v", err)
	}
}
