package services

import (
	"errors"
	"testing"

	"example.com/travel-advisor/models"
)

func TestUserService_FindByEmail_Success(t *testing.T) {
	expectedID := int64(42)
	mockUser := &models.User{
		Email: "test@example.com",
	}

	models.NewUser = func(email, password string) *models.User {
		return mockUser
	}
	mockUser.FindByEmail = func(email string) (*models.User, error) {
		mockUser.ID = expectedID
		return mockUser, nil
	}

	origInitUser := models.InitUser

	models.InitUser = func() *models.User {
		return mockUser
	}
	defer func() {
		models.InitUser = origInitUser
	}()

	us := GetUserService()
	user, err := us.FindByEmail("test@example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != expectedID {
		t.Errorf("expected ID %d, got %d", expectedID, user.ID)
	}
}

func TestUserService_FindByEmail_EmptyEmail(t *testing.T) {
	us := &UserService{}
	_, err := us.FindByEmail("")
	if err == nil {
		t.Error("expected error for empty email, got nil")
	}
}

func TestUserService_FindByEmail_Error(t *testing.T) {
	mockUser := &models.User{
		Email: "fail@example.com",
	}
	models.NewUser = func(email, password string) *models.User {
		return mockUser
	}
	mockUser.FindByEmail = func(email string) (*models.User, error) {
		return nil, errors.New("db error")
	}

	origInitUser := models.InitUser

	models.InitUser = func() *models.User {
		return mockUser
	}
	defer func() {
		models.InitUser = origInitUser
	}()

	us := &UserService{}
	// Patch the constructor to return our mock
	origNewUser := models.NewUser
	models.NewUser = func(email, password string) *models.User {
		return mockUser
	}
	defer func() { models.NewUser = origNewUser }()

	_, err := us.FindByEmail("fail@example.com")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestUserService_Create_Success(t *testing.T) {
	mockUser := &models.User{}
	called := false
	mockUser.Create = func() error {
		called = true
		return nil
	}
	us := &UserService{}
	err := us.Create(mockUser)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Error("expected Create to be called")
	}
}

func TestUserService_Create_NilUser(t *testing.T) {
	us := &UserService{}
	err := us.Create(nil)
	if err == nil {
		t.Error("expected error for nil user, got nil")
	}
}

func TestUserService_Create_Error(t *testing.T) {
	mockUser := &models.User{}
	mockUser.Create = func() error {
		return errors.New("insert error")
	}
	us := &UserService{}
	err := us.Create(mockUser)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestUserService_ValidateCredentials_Success(t *testing.T) {
	mockUser := &models.User{Email: "test@example.com"}
	called := false
	mockUser.ValidateCredentials = func(password string) error {
		called = true
		return nil
	}
	us := &UserService{}
	err := us.ValidateCredentials(mockUser, "password123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Error("expected ValidateCredentials to be called")
	}
}

func TestUserService_ValidateCredentials_NilUser(t *testing.T) {
	us := &UserService{}
	err := us.ValidateCredentials(nil, "password123")
	if err == nil {
		t.Error("expected error for nil user, got nil")
	}
}

func TestUserService_ValidateCredentials_EmptyPassword(t *testing.T) {
	mockUser := &models.User{Email: "test@example.com"}
	us := &UserService{}
	err := us.ValidateCredentials(mockUser, "")
	if err == nil {
		t.Error("expected error for nil user, got nil")
	}
}

func TestUserService_ValidateCredentials_Error(t *testing.T) {
	mockUser := &models.User{}
	mockUser.ValidateCredentials = func(password string) error {
		return errors.New("invalid credentials")
	}
	us := &UserService{}
	err := us.ValidateCredentials(mockUser, "password123")
	if err == nil {
		t.Error("expected error, got nil")
	}
}
