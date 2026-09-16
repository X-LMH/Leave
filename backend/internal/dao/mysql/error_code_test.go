package mysql

import (
	"errors"
	"testing"

	"gorm.io/gorm"
)

func TestIsDuplicateKeyError(t *testing.T) {
	if !IsDuplicateKeyError(gorm.ErrDuplicatedKey) {
		t.Fatal("expected duplicated key error to be recognized")
	}
	if !IsDuplicateKeyError(errors.Join(errors.New("write failed"), gorm.ErrDuplicatedKey)) {
		t.Fatal("expected wrapped duplicated key error to be recognized")
	}
	if IsDuplicateKeyError(errors.New("write failed")) {
		t.Fatal("unexpected duplicated key error")
	}
}
