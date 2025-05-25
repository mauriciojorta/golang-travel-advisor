package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPasswordAndCheckPasswordHash(t *testing.T) {
	password := "mySecretPassword123!"

	hashed, err := HashPassword(password)
	assert.NoError(t, err, "HashPassword should not return an error")
	assert.NotEmpty(t, hashed, "HashPassword should return a non-empty string")

	// Correct password should return true
	assert.True(t, CheckPasswordHash(hashed, password), "CheckPasswordHash should return true for correct password")

	// Incorrect password should return false
	assert.False(t, CheckPasswordHash(hashed, "wrongPassword"), "CheckPasswordHash should return false for incorrect password")
}

func TestHashPassword_ErrorOnEmpty(t *testing.T) {
	_, err := HashPassword("")
	assert.NoError(t, err, "HashPassword should not return an error for empty password")
}

func TestCheckPasswordHash_InvalidHash(t *testing.T) {
	invalidHash := "notAValidHash"
	password := "anyPassword"
	assert.False(t, CheckPasswordHash(invalidHash, password), "CheckPasswordHash should return false for invalid hash")
}
