package utils

import (
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email       string
		expectError bool
	}{
		{"", true},
		{"plainaddress", true},
		{"@missingusername.com", true},
		{"username@.com", true},
		{"username@domain.com", false},
		{"user.name+tag@sub.domain.co.uk", false},
	}

	for _, tt := range tests {
		err := ValidateEmail(tt.email)
		if (err != nil) != tt.expectError {
			t.Errorf("ValidateEmail(%q) error = %v, expectError = %v", tt.email, err, tt.expectError)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		password    string
		minLength   int
		expectValid bool
	}{
		{"", 8, false},
		{"sho12A!", 8, false},
		{"longpassword", 8, false},
		{"Longpassword1", 8, false},
		{"Longpassword!", 8, false},
		{"Password1!", 8, true},
		{"Pass1!", 6, true},
		{"PASSWORD1!", 8, true},
		{"password1!", 8, false},
		{"Password!", 8, false},
		{"Password1", 8, false},
		{"P@ssw0rd", 8, true},
	}

	for _, tt := range tests {
		valid := ValidatePassword(tt.password, tt.minLength)
		if valid != tt.expectValid {
			t.Errorf("ValidatePassword(%q, %d) = %v, expectValid = %v", tt.password, tt.minLength, valid, tt.expectValid)
		}
	}
}
