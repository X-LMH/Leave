package app

import (
	"backend/internal/config"
	"backend/internal/dao/mysql"
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
	settings := config.Config{}
	dao := mysql.NewDAO(nil)

	root := t.TempDir()
	settings.Storage.RootDir = root
	valid := feedbackImage{data: []byte("validated image bytes"), extension: ".png"}
	for _, count := range []int{0, 1, 3} {
		row := &models.Feedback{}
		err := NewFeedbackService(dao, settings.Storage).saveFeedbackWithAttachments(row, makeImages(valid, count), func(row *models.Feedback) error {
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
	settings.Storage.RootDir = t.TempDir()
	failure := errors.New("insert failed")
	if err := NewFeedbackService(dao, settings.Storage).saveFeedbackWithAttachments(&models.Feedback{}, []feedbackImage{valid, valid}, func(*models.Feedback) error { return failure }); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(settings.Storage.RootDir, "feedbacks"))
	if err != nil || len(entries) != 0 {
		t.Fatalf("failed insert left files: %v %v", entries, err)
	}
	invalid := feedbackImage{data: []byte("x"), extension: ".invalid/extension"}
	inserted := false
	if err := NewFeedbackService(dao, settings.Storage).saveFeedbackWithAttachments(&models.Feedback{}, []feedbackImage{valid, invalid}, func(*models.Feedback) error { inserted = true; return nil }); err == nil || inserted {
		t.Fatal("saved invalid image or inserted partial feedback")
	}
	entries, err = os.ReadDir(filepath.Join(settings.Storage.RootDir, "feedbacks"))
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

func TestFeedbackServicesKeepIndependentStorage(t *testing.T) {
	firstRoot := t.TempDir()
	secondRoot := t.TempDir()
	settings := config.StorageConfig{RootDir: firstRoot}
	first := NewFeedbackService(nil, settings)
	settings.RootDir = secondRoot
	second := NewFeedbackService(nil, settings)
	image := feedbackImage{data: []byte("validated bytes"), extension: ".png"}
	for _, tc := range []struct {
		service *FeedbackService
		root    string
		other   string
	}{
		{service: first, root: firstRoot, other: secondRoot},
		{service: second, root: secondRoot, other: firstRoot},
	} {
		row := &models.Feedback{}
		if err := tc.service.saveFeedbackWithAttachments(row, []feedbackImage{image}, func(*models.Feedback) error { return nil }); err != nil {
			t.Fatal(err)
		}
		var keys []string
		if err := json.Unmarshal([]byte(*row.Attachments), &keys); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(tc.root, filepath.FromSlash(keys[0]))); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(tc.other, filepath.FromSlash(keys[0]))); !os.IsNotExist(err) {
			t.Fatalf("attachment leaked into another service's storage: %v", err)
		}
	}
}
