package file

import (
	"bytes"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"testing"
)

func TestReadImage(t *testing.T) {
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
			data, extension, err := ReadImage(bytes.NewReader(tt.data), 4096)
			if tt.extension == "" {
				if !errors.Is(err, ErrInvalidImage) {
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

func TestReadImagePreservesReadError(t *testing.T) {
	_, _, err := ReadImage(failedAvatarReader{}, 4096)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected read error, got %v", err)
	}
}
