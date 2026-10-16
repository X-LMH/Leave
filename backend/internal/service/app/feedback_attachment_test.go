package app

import (
	"backend/internal/config"
	"backend/internal/models"
	"bytes"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func feedbackUpload(t *testing.T, data []byte) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "misleading.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(1 << 20); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = req.MultipartForm.RemoveAll() })
	return req.MultipartForm.File["file"][0]
}

func TestReadFeedbackImages(t *testing.T) {
	var small bytes.Buffer
	if err := png.Encode(&small, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	var large bytes.Buffer
	if err := png.Encode(&large, image.NewRGBA(image.Rect(0, 0, 4097, 1))); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		data  []byte
		valid bool
	}{
		{"valid image with misleading filename", small.Bytes(), true},
		{"empty", nil, false},
		{"fake png", []byte("not an image"), false},
		{"truncated png", small.Bytes()[:len(small.Bytes())/2], false},
		{"oversized dimensions", large.Bytes(), false},
		{"oversized file", make([]byte, MaxFeedbackAttachmentSize+1), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			images, err := readFeedbackImages([]*multipart.FileHeader{feedbackUpload(t, tc.data)})
			if tc.valid {
				if err != nil || len(images) != 1 || images[0].extension != ".png" {
					t.Fatalf("images=%v err=%v", images, err)
				}
			} else if !errors.Is(err, ErrorInvalidFeedback) {
				t.Fatalf("expected invalid feedback, got %v", err)
			}
		})
	}
}

func TestSaveFeedbackImagesAndCleanup(t *testing.T) {
	previous := config.Cfg
	t.Cleanup(func() { config.Cfg = previous })
	root := t.TempDir()
	config.Cfg.Storage.RootDir = root
	valid := feedbackImage{data: []byte("validated image bytes"), extension: ".png"}
	for _, count := range []int{0, 1, 3} {
		row := &models.Feedback{}
		err := saveFeedbackWithAttachments(row, makeImages(valid, count), func(row *models.Feedback) error {
			if count == 0 {
				if row.Attachments != nil {
					t.Fatal("no images should be SQL NULL")
				}
				return nil
			}
			var keys []string
			if err := json.Unmarshal([]byte(*row.Attachments), &keys); err != nil || len(keys) != count {
				t.Fatalf("keys=%v err=%v", keys, err)
			}
			for _, key := range keys {
				if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(key))); err != nil {
					t.Fatal(err)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	config.Cfg.Storage.RootDir = t.TempDir()
	failure := errors.New("insert failed")
	if err := saveFeedbackWithAttachments(&models.Feedback{}, []feedbackImage{valid, valid}, func(*models.Feedback) error { return failure }); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(config.Cfg.Storage.RootDir, "feedbacks"))
	if err != nil || len(entries) != 0 {
		t.Fatalf("failed insert left files: %v %v", entries, err)
	}
	invalid := feedbackImage{data: []byte("x"), extension: ".invalid/extension"}
	inserted := false
	if err := saveFeedbackWithAttachments(&models.Feedback{}, []feedbackImage{valid, invalid}, func(*models.Feedback) error { inserted = true; return nil }); err == nil || inserted {
		t.Fatal("saved invalid image or inserted partial feedback")
	}
	entries, err = os.ReadDir(filepath.Join(config.Cfg.Storage.RootDir, "feedbacks"))
	if err != nil || len(entries) != 0 {
		t.Fatalf("failed save left files: %v %v", entries, err)
	}
}
func makeImages(image feedbackImage, count int) []feedbackImage {
	images := make([]feedbackImage, count)
	for index := range images {
		images[index] = image
	}
	return images
}
