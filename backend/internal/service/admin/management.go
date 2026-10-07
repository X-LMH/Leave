package admin

import (
	"errors"
	"unicode/utf8"
)

var ErrorInvalidAdminInput = errors.New("invalid management input")

func validAdminText(text string, max int) bool {
	return text != "" && utf8.RuneCountInString(text) <= max
}
