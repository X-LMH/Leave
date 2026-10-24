package common

import (
	"backend/internal/dao/mysql"
	"testing"

	"go.uber.org/zap"
)

func TestApartmentOptionsRejectInvalidGender(t *testing.T) {
	dao := mysql.NewDAO(nil, zap.NewNop())
	for _, enabledOnly := range []bool{true, false} {
		if _, err := NewOptionsService(dao, zap.NewNop()).GetApartmentOptions("unknown", enabledOnly); err != ErrorInvalidApartmentGender {
			t.Fatalf("enabledOnly=%v: %v", enabledOnly, err)
		}
	}
}
