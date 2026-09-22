package app

import (
	"backend/internal/dto"
	"backend/internal/models"
	"testing"
)

func TestValidateAndNormalizeProfile(t *testing.T) {
	tests := []struct {
		name  string
		input dto.ProfileRequest
		valid bool
	}{
		{
			name: "valid profile trims optional fields",
			input: dto.ProfileRequest{
				Name: " 张三 ", Phone: "13800138000", Gender: " male ",
				ParentName: " 李四 ", ParentPhone: "13900139000",
				ClassID: 1, ApartmentID: 3, DormitoryNumber: " 302 ", TeacherName: " 王老师 ",
			},
			valid: true,
		},
		{
			name: "missing required name",
			input: dto.ProfileRequest{
				Phone: "13800138000", Gender: models.GenderMale, ClassID: 1,
				ParentName: "李四", ParentPhone: "13900139000", TeacherName: "王老师",
			},
			valid: false,
		},
		{
			name: "missing required class ID",
			input: dto.ProfileRequest{
				Name: "张三", Phone: "13800138000", Gender: models.GenderMale,
				ParentName: "李四", ParentPhone: "13900139000", TeacherName: "王老师",
			},
			valid: false,
		},
		{
			name: "invalid gender",
			input: dto.ProfileRequest{
				Name: "张三", Phone: "13800138000", Gender: "未知", ClassID: 1,
				ParentName: "李四", ParentPhone: "13900139000", TeacherName: "王老师",
			},
			valid: false,
		},
		{
			name: "invalid phone",
			input: dto.ProfileRequest{
				Name: "张三", Phone: "12800138000", Gender: models.GenderMale, ClassID: 1,
				ParentName: "李四", ParentPhone: "13900139000", TeacherName: "王老师",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateAndNormalizeProfile(&tt.input); got != tt.valid {
				t.Fatalf("validateAndNormalizeProfile() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestValidateAndNormalizeProfileTrimsFields(t *testing.T) {
	p := &dto.ProfileRequest{
		Name: " 张三 ", Phone: " 13800138000 ", Gender: " female ",
		ParentName: " 李四 ", ParentPhone: " 13900139000 ",
		ClassID: 1, ApartmentID: 3, DormitoryNumber: " 302 ", TeacherName: " 王老师 ",
	}

	if !validateAndNormalizeProfile(p) {
		t.Fatal("expected profile to be valid")
	}
	if p.Name != "张三" || p.ClassID != 1 || p.ApartmentID != 3 || p.DormitoryNumber != "302" {
		t.Fatalf("profile fields were not normalized: %#v", p)
	}
}
