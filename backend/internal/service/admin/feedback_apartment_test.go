package admin

import (
	"backend/internal/dto"
	"backend/internal/models"
	"strings"
	"testing"
)

func TestNormalizeApartment(t *testing.T) {
	order := uint32(0)
	enabled := false
	valid := dto.ApartmentRequest{Name: "  测试楼  ", Gender: models.GenderMale, SortOrder: &order, IsEnabled: &enabled}
	p := valid
	if err := normalizeApartment(&p); err != nil || p.Name != "测试楼" {
		t.Fatalf("valid: %+v %v", p, err)
	}
	for _, change := range []func(*dto.ApartmentRequest){
		func(p *dto.ApartmentRequest) { p.Name = "  " },
		func(p *dto.ApartmentRequest) { p.Name = strings.Repeat("楼", 101) },
		func(p *dto.ApartmentRequest) { p.Gender = "other" },
		func(p *dto.ApartmentRequest) { p.SortOrder = nil },
		func(p *dto.ApartmentRequest) { p.IsEnabled = nil },
	} {
		p = valid
		change(&p)
		if normalizeApartment(&p) != ErrorInvalidAdminInput {
			t.Fatalf("accepted invalid: %+v", p)
		}
	}
	order = ^uint32(0)
	p = valid
	if normalizeApartment(&p) != nil {
		t.Fatal("maximum unsigned order rejected")
	}
}
func TestFeedbackIdentityMask(t *testing.T) {
	id := "20260001"
	row := models.Feedback{StudentID: &id, IsAnonymous: true, Content: "一行\n二行"}
	item := feedbackItem(&row)
	if item.StudentID != nil || item.Content != row.Content {
		t.Fatal("anonymous identity or content changed")
	}
	row.IsAnonymous = false
	if *feedbackItem(&row).StudentID != id {
		t.Fatal("identified identity lost")
	}
	row.StudentID = nil
	if feedbackItem(&row).StudentID != nil {
		t.Fatal("missing identity synthesized")
	}
}
