package admin

import (
	"backend/internal/config"
	"backend/internal/dto"
	"backend/internal/models"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestStudentResponse(t *testing.T) {
	previous := config.Cfg.Storage.BaseURL
	config.Cfg.Storage.BaseURL = "/uploads"
	t.Cleanup(func() { config.Cfg.Storage.BaseURL = previous })
	createdAt := time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)
	user := &models.User{
		ID:             7,
		StudentID:      "202600010001",
		Password:       "secret",
		Status:         1,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
		LastSeenAt:     &createdAt,
		LastSeenDevice: "Android",
		AppVersion:     "1.0.0",
	}
	t.Run("without profile", func(t *testing.T) {
		row, err := studentResponse(user, nil)
		if err != nil {
			t.Fatal(err)
		}
		if row.ID != user.ID || row.StudentID != user.StudentID || row.Status != user.Status || row.CreatedAt != createdAt || row.UpdatedAt != createdAt || row.LastSeenAt != user.LastSeenAt || row.LastSeenDevice != user.LastSeenDevice || row.AppVersion != user.AppVersion {
			t.Fatalf("account fields lost: %+v", row)
		}
		if row.ProfileRequest != (dto.ProfileRequest{}) || row.AvatarURL != "/uploads/avatars/default.png" {
			t.Fatalf("incomplete profile defaults: %+v", row)
		}
	})
	t.Run("with profile", func(t *testing.T) {
		profile := &models.Profile{
			ID:              99,
			StudentID:       user.StudentID,
			Name:            "学生",
			Phone:           "13800138000",
			Gender:          models.GenderMale,
			ClassID:         2,
			ParentName:      "家长",
			ParentPhone:     "13900139000",
			TeacherName:     "老师",
			ApartmentID:     3,
			DormitoryNumber: "101",
			AvatarURL:       "avatars/student.png",
			CreatedAt:       createdAt.Add(-time.Hour),
		}
		row, err := studentResponse(user, profile)
		if err != nil {
			t.Fatal(err)
		}
		expected := dto.ProfileRequest{
			Name:            profile.Name,
			Phone:           profile.Phone,
			Gender:          profile.Gender,
			ClassID:         profile.ClassID,
			ParentName:      profile.ParentName,
			ParentPhone:     profile.ParentPhone,
			TeacherName:     profile.TeacherName,
			ApartmentID:     profile.ApartmentID,
			DormitoryNumber: profile.DormitoryNumber,
		}
		if row.ProfileRequest != expected || row.ID != user.ID || row.CreatedAt != user.CreatedAt || row.AvatarURL != "/uploads/avatars/student.png" {
			t.Fatalf("profile or identity mapping: %+v", row)
		}
		body, err := json.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(body), "password") || strings.Contains(string(body), "secret") {
			t.Fatal("credentials exposed")
		}
		profile.AvatarURL = ""
		row, err = studentResponse(user, profile)
		if err != nil || row.AvatarURL != "/uploads/avatars/default.png" {
			t.Fatalf("default avatar: %v %+v", err, row)
		}
	})
}

func TestNormalizeAdminStudent(t *testing.T) {
	valid := dto.ProfileRequest{
		Name:            " 学生 ",
		Phone:           " 13800138000 ",
		Gender:          " male ",
		ParentName:      " 家长 ",
		ParentPhone:     "13900139000",
		TeacherName:     " 老师 ",
		ClassID:         1,
		DormitoryNumber: " 101 ",
	}
	if err := normalizeAdminStudent(&valid); err != nil || valid.Name != "学生" || valid.Phone != "13800138000" || valid.DormitoryNumber != "101" {
		t.Fatalf("normalize: %+v %v", valid, err)
	}
	for _, tc := range []struct {
		name   string
		change func(*dto.ProfileRequest)
	}{
		{"empty name", func(p *dto.ProfileRequest) { p.Name = " " }},
		{"long name", func(p *dto.ProfileRequest) { p.Name = strings.Repeat("字", 65) }},
		{"empty parent", func(p *dto.ProfileRequest) { p.ParentName = "" }},
		{"long parent", func(p *dto.ProfileRequest) { p.ParentName = strings.Repeat("字", 65) }},
		{"empty teacher", func(p *dto.ProfileRequest) { p.TeacherName = "" }},
		{"long teacher", func(p *dto.ProfileRequest) { p.TeacherName = strings.Repeat("字", 65) }},
		{"invalid phone", func(p *dto.ProfileRequest) { p.Phone = "12345678901" }},
		{"invalid parent phone", func(p *dto.ProfileRequest) { p.ParentPhone = "abc" }},
		{"invalid gender", func(p *dto.ProfileRequest) { p.Gender = "other" }},
		{"missing class", func(p *dto.ProfileRequest) { p.ClassID = 0 }},
		{"long dorm", func(p *dto.ProfileRequest) { p.DormitoryNumber = strings.Repeat("字", 33) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := valid
			tc.change(&p)
			if err := normalizeAdminStudent(&p); err != ErrorInvalidAdminInput {
				t.Fatalf("expected invalid input: %v", err)
			}
		})
	}
	valid.Name = strings.Repeat("字", 64)
	valid.ApartmentID = 0
	valid.DormitoryNumber = ""
	if err := normalizeAdminStudent(&valid); err != nil {
		t.Fatalf("boundary/optional apartment: %v", err)
	}
}

func TestValidateStudentReferences(t *testing.T) {
	for _, tc := range []struct {
		name         string
		current      models.Profile
		classEnabled bool
		apartment    *models.Apartment
		wantInvalid  bool
	}{
		{name: "enabled references", classEnabled: true, apartment: &models.Apartment{Gender: models.GenderMale, IsEnabled: 1}},
		{name: "retain disabled references", current: models.Profile{ClassID: 1, ApartmentID: 2}, apartment: &models.Apartment{Gender: models.GenderMale}},
		{name: "new disabled class", apartment: &models.Apartment{Gender: models.GenderMale, IsEnabled: 1}, wantInvalid: true},
		{name: "new disabled apartment", classEnabled: true, apartment: &models.Apartment{Gender: models.GenderMale}, wantInvalid: true},
		{name: "gender mismatch even when retained", current: models.Profile{ClassID: 1, ApartmentID: 2}, apartment: &models.Apartment{Gender: models.GenderFemale}, wantInvalid: true},
		{name: "clear apartment", classEnabled: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := dto.ProfileRequest{ClassID: 1, ApartmentID: 2, Gender: models.GenderMale}
			if tc.apartment == nil {
				input.ApartmentID = 0
			} else {
				tc.apartment.ID = input.ApartmentID
			}
			class := &models.Class{
				ID:        input.ClassID,
				IsEnabled: tc.classEnabled,
			}
			err := validateStudentReferences(&input, &tc.current, class, tc.apartment)
			if (err != nil) != tc.wantInvalid {
				t.Fatalf("invalid = %v, want %v", err, tc.wantInvalid)
			}
		})
	}
}
