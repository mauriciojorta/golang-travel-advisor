package services

import (
	"database/sql"
	"errors"
	"testing"

	"example.com/travel-advisor/db"
	"example.com/travel-advisor/models"
	"example.com/travel-advisor/utils"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
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

func TestUserService_GenerateLoginToken_Success(t *testing.T) {
	origGenerateToken := utils.GenerateToken
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	db.DB = dbMock
	defer func() {
		utils.GenerateToken = origGenerateToken
	}()

	utils.GenerateToken = func(email string, id int64) (string, error) {
		return "mocktoken", nil
	}

	mockUser := &models.User{
		Email: "test@example.com",
		ID:    1,
	}
	calledUpdate := false
	mockUser.UpdateLastLoginDate = func(tx *sql.Tx) error {
		calledUpdate = true
		return nil
	}

	origNewAuditEvent := models.NewAuditEvent
	defer func() { models.NewAuditEvent = origNewAuditEvent }()
	mockAuditEvent := &models.AuditEvent{}
	calledCreateAudit := false
	mockAuditEvent.CreateAuditEvent = func(tx *sql.Tx) error {
		calledCreateAudit = true
		return nil
	}
	models.NewAuditEvent = func(userID int64, event string) *models.AuditEvent {
		return mockAuditEvent
	}

	us := &UserService{}
	token, err := us.GenerateLoginToken(mockUser)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token != "mocktoken" {
		t.Errorf("expected token 'mocktoken', got %s", token)
	}
	if !calledUpdate {
		t.Error("expected UpdateLastLoginDate to be called")
	}
	if !calledCreateAudit {
		t.Error("expected CreateAuditEvent to be called")
	}
}

func TestUserService_GenerateLoginToken_ErrorGeneratingToken(t *testing.T) {
	origGenerateToken := utils.GenerateToken
	defer func() { utils.GenerateToken = origGenerateToken }()

	utils.GenerateToken = func(email string, id int64) (string, error) {
		return "", errors.New("token error")
	}

	us := &UserService{}
	mockUser := &models.User{Email: "fail@example.com", ID: 1}
	token, err := us.GenerateLoginToken(mockUser)
	if err == nil || token != "" {
		t.Error("expected error when generating token")
	}
}

func TestUserService_GenerateLoginToken_ErrorBeginTx(t *testing.T) {
	dbMock, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	us := &UserService{}
	mockUser := &models.User{Email: "fail@example.com", ID: 1}
	token, err := us.GenerateLoginToken(mockUser)
	if err == nil || token != "" {
		t.Error("expected error when beginning transaction")
	}
}

func TestUserService_GenerateLoginToken_ErrorUpdateLastLoginDate(t *testing.T) {
	origGenerateToken := utils.GenerateToken
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	defer func() { utils.GenerateToken = origGenerateToken }()
	utils.GenerateToken = func(email string, id int64) (string, error) {
		return "mocktoken", nil
	}

	mock.ExpectBegin()
	mock.ExpectRollback()

	db.DB = dbMock

	mockUser := &models.User{Email: "fail@example.com", ID: 1}
	mockUser.UpdateLastLoginDate = func(tx *sql.Tx) error {
		return errors.New("update error")
	}

	us := &UserService{}
	token, err := us.GenerateLoginToken(mockUser)
	if err == nil || token != "" {
		t.Error("expected error when updating last login date")
	}
}

func TestUserService_GenerateLoginToken_ErrorCreateAuditEvent(t *testing.T) {
	origGenerateToken := utils.GenerateToken
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	defer func() { utils.GenerateToken = origGenerateToken }()
	utils.GenerateToken = func(email string, id int64) (string, error) {
		return "mocktoken", nil
	}

	mock.ExpectBegin()
	mock.ExpectRollback()

	db.DB = dbMock

	mockUser := &models.User{Email: "fail@example.com", ID: 1}
	mockUser.UpdateLastLoginDate = func(tx *sql.Tx) error {
		return nil
	}

	origNewAuditEvent := models.NewAuditEvent
	defer func() { models.NewAuditEvent = origNewAuditEvent }()
	mockAuditEvent := &models.AuditEvent{}
	mockAuditEvent.CreateAuditEvent = func(tx *sql.Tx) error {
		return errors.New("audit error")
	}
	models.NewAuditEvent = func(userID int64, event string) *models.AuditEvent {
		return mockAuditEvent
	}

	us := &UserService{}
	token, err := us.GenerateLoginToken(mockUser)
	if err == nil || token != "" {
		t.Error("expected error when creating audit event")
	}
}
