package admin

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestAdminRecordQuery(t *testing.T) {
	for _, values := range []url.Values{
		{},
		{"page": {"2"}, "page_size": {"10"}, "student_id": {" 2026 "}, "name": {" 李 "}, "college": {"学院"}, "major": {"专业"}, "class_name": {"班级"}, "leave_type_id": {"4"}, "is_leave_school": {"false"}, "start_time_from": {"2026-10-10T00:00:00+08:00"}, "start_time_to": {"2026-10-11T00:00:00+08:00"}},
	} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/?"+values.Encode(), nil)
		q, ok := parseAdminRecordQuery(c)
		if !ok {
			t.Fatal("valid query rejected")
		}
		if len(values) == 0 {
			if q.Page != 1 || q.PageSize != 20 || q.IsLeaveSchool != nil {
				t.Fatal("wrong query defaults")
			}
			continue
		}
		if q.Page != 2 || q.PageSize != 10 || q.StudentID != "2026" || q.Name != "李" || q.ClassFilter == nil || q.ClassFilter.College != "学院" || q.ClassFilter.Major != "专业" || q.ClassFilter.ClassName != "班级" || q.LeaveTypeID != 4 || q.IsLeaveSchool == nil || *q.IsLeaveSchool {
			t.Fatalf("wrong filters: %+v", q)
		}
		if q.StartTimeTo.Sub(*q.StartTimeFrom) != 24*time.Hour || q.StartTimeFrom.UTC().Hour() != 16 {
			t.Fatal("time zone or range lost")
		}
	}
}
