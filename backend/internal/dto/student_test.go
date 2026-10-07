package dto

import (
	"encoding/json"
	"testing"
)

func TestStudentListResponseFields(t *testing.T) {
	body, err := json.Marshal(StudentListResponse{
		Items:    []*StudentListItem{{ID: 7, StudentID: "202600010001"}},
		Total:    1,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	var response struct {
		Items []map[string]json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	allowed := []string{"id", "student_id", "name", "class_id", "gender", "phone", "status", "app_version", "app_version_status", "created_at", "last_seen_at"}
	if len(response.Items) != 1 || len(response.Items[0]) != len(allowed) {
		t.Fatalf("unexpected list fields: %s", body)
	}
	for _, key := range allowed {
		if _, ok := response.Items[0][key]; !ok {
			t.Errorf("missing list field %s", key)
		}
	}
}
