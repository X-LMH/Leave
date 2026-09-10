package service

import (
	"backend/internal/dto"
	"strings"
	"testing"
)

func TestValidateAndNormalizeFeedbackRequest(t *testing.T) {
	tests := []struct {
		name  string
		input dto.FeedbackCreateRequest
		valid bool
	}{
		{name: "valid content is trimmed", input: dto.FeedbackCreateRequest{Content: "  页面加载失败  "}, valid: true},
		{name: "blank content", input: dto.FeedbackCreateRequest{Content: " \t "}, valid: false},
		{name: "content exceeds limit", input: dto.FeedbackCreateRequest{Content: strings.Repeat("反", maxFeedbackContentLength+1)}, valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateAndNormalizeFeedbackRequest(&tt.input); got != tt.valid {
				t.Fatalf("validateAndNormalizeFeedbackRequest() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestValidateAndNormalizeFeedbackRequestTrimsContent(t *testing.T) {
	req := &dto.FeedbackCreateRequest{Content: "  建议增加夜间模式  "}
	if !validateAndNormalizeFeedbackRequest(req) {
		t.Fatal("expected feedback request to be valid")
	}
	if req.Content != "建议增加夜间模式" {
		t.Fatalf("content = %q", req.Content)
	}
}
