package app

import (
	"backend/internal/models"
	"testing"
)

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
