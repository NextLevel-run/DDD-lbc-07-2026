package domain

import (
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// Email is a value object representing a validated, normalized email address.
type Email struct {
	value string
}

// NewEmail validates and builds an Email from a raw string, lowercasing it.
func NewEmail(s string) (Email, error) {
	normalized := strings.ToLower(strings.TrimSpace(s))
	if !emailRegex.MatchString(normalized) {
		return Email{}, ErrInvalidEmail
	}
	return Email{value: normalized}, nil
}

// String returns the normalized email address.
func (e Email) String() string {
	return e.value
}
