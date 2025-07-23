package utils

import (
	"errors"
	"net/mail"
	"unicode"
)

func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email cannot be a empty string")
	}

	_, err := mail.ParseAddress(email)
	return err
}

func ValidatePassword(s string, minPasswordLength int) bool {
	if s == "" || len(s) < minPasswordLength {
		return false
	}

	hasNumber := false
	hasUpper := false
	hasSpecialChar := false

	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			hasNumber = true
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecialChar = true
		case unicode.IsLetter(c) || c == ' ':
		default:
		}

		if hasNumber && hasUpper && hasSpecialChar {
			return true
		}
	}
	return false
}
