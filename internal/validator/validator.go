package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var EmailRX = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type Validator struct {
	FieldErrors map[string]string
}

func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0
}

func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}
	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

func NotBlank(s string) bool {
	return strings.TrimSpace(s) != ""
}

func MaxChars(s string, n int) bool {
	return utf8.RuneCountInString(s) <= n
}

func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

func Matches(rx *regexp.Regexp, s string) bool {
	return rx.MatchString(s)
}
