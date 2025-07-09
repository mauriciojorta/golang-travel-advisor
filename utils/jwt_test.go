package utils

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitJwtSecretKey_Success(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", "testsecret")
	defer os.Unsetenv("JWT_SECRET_KEY")

	// Reset sync.Once for testing
	secretKeyOnce = *new(sync.Once)

	err := InitJwtSecretKey()
	assert.NoError(t, err, "expected no error, got %v", err)
}

func TestInitJwtSecretKey_MissingEnv(t *testing.T) {
	os.Unsetenv("JWT_SECRET_KEY")

	// Reset sync.Once for testing
	secretKeyOnce = *new(sync.Once)

	err := InitJwtSecretKey()
	assert.Error(t, err, "expected error for missing JWT_SECRET_KEY, got nil")
}

func TestGenerateAndVerifyToken(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", "testsecret")
	defer os.Unsetenv("JWT_SECRET_KEY")

	// Reset sync.Once for testing
	secretKeyOnce = *new(sync.Once)
	secretKey = ""
	if err := InitJwtSecretKey(); err != nil {
		t.Fatalf("InitJwtSecretKey failed: %v", err)
	}

	email := "test@example.com"
	userId := int64(12345)

	token, err := GenerateToken(email, userId)

	assert.NoError(t, err, "GenerateToken failed: %v", err)
	assert.NotEmpty(t, token, "expected token to be non-empty")

	gotUserId, err := VerifyToken(token)

	assert.NoError(t, err, "VerifyToken failed: %v", err)
	assert.Equal(t, userId, gotUserId, "expected userId to match")
}

func TestVerifyToken_InvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", "testsecret")
	defer os.Unsetenv("JWT_SECRET_KEY")

	// Reset sync.Once for testing
	secretKeyOnce = *new(sync.Once)
	err := InitJwtSecretKey()
	assert.NoError(t, err, "InitJwtSecretKey failed: %v", err)

	invalidToken := "invalid.token.value"
	_, err = VerifyToken(invalidToken)
	assert.Error(t, err, "expected error for invalid token, got nil")
}
