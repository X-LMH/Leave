package common

import "testing"

func TestApartmentOptionsRejectInvalidGender(t *testing.T) {
	for _, enabledOnly := range []bool{true, false} {
		if _, err := GetApartmentOptions("unknown", enabledOnly); err != ErrorInvalidApartmentGender {
			t.Fatalf("enabledOnly=%v: %v", enabledOnly, err)
		}
	}
}
