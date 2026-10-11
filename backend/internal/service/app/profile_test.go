package app

import (
	"backend/internal/dto"
	"backend/internal/models"
	"errors"
	"testing"
)

func TestProfileClassReferences(t *testing.T) {
	for _, tc := range []struct {
		name    string
		current uint
		enabled bool
		invalid bool
	}{
		{"enabled", 0, true, false},
		{"retain disabled", 2, false, false},
		{"new disabled", 0, false, true},
		{"different disabled", 1, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateProfileClass(&models.Profile{ClassID: tc.current}, &models.Class{ID: 2, IsEnabled: tc.enabled})
			if errors.Is(err, ErrorInvalidProfileReference) != tc.invalid {
				t.Fatalf("unexpected result: %v", err)
			}
		})
	}
}

func TestToProfileResponse(t *testing.T) {
	profile := &models.Profile{
		StudentID:       "202410201016",
		Name:            "张三",
		Phone:           "13800138000",
		Gender:          models.GenderMale,
		ParentName:      "张父",
		ParentPhone:     "13900139000",
		DormitoryNumber: "302",
		TeacherName:     "王老师",
		ClassID:         2,
		ApartmentID:     4,
	}
	class := &models.Class{
		College:   "计算机学院",
		Major:     "软件工程",
		ClassName: "软件工程 1 班",
	}
	apartment := &models.Apartment{Name: "金川 2 号楼"}

	response := toProfileResponse(profile, class, apartment)
	if response.StudentID != profile.StudentID {
		t.Fatalf("response does not preserve student identity: %#v", response)
	}
	if response.ClassID != profile.ClassID || response.ApartmentID != profile.ApartmentID {
		t.Fatalf("response does not preserve selection IDs: %#v", response)
	}
	if response.ClassInfo.College != class.College || response.ClassInfo.Major != class.Major || response.ClassInfo.ClassName != class.ClassName || response.ApartmentInfo.ApartmentName != apartment.Name || response.ApartmentInfo.DormitoryNumber != profile.DormitoryNumber {
		t.Fatalf("response does not preserve residence data: %#v", response)
	}
}

func TestProfileApartmentReferences(t *testing.T) {
	input := dto.ProfileRequest{ApartmentID: 2, Gender: models.GenderMale}
	for _, tc := range []struct {
		name    string
		current uint
		enabled uint8
		gender  string
		invalid bool
	}{
		{"enabled", 0, 1, models.GenderMale, false},
		{"retain disabled", 2, 0, models.GenderMale, false},
		{"new disabled", 0, 0, models.GenderMale, true},
		{"gender mismatch", 2, 1, models.GenderFemale, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateProfileApartment(&input, &models.Profile{ApartmentID: tc.current}, &models.Apartment{ID: 2, Gender: tc.gender, IsEnabled: tc.enabled})
			if (err != nil) != tc.invalid {
				t.Fatalf("unexpected result: %v", err)
			}
		})
	}
}
