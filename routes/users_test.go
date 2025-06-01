package routes

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/travel-advisor/models"
	"example.com/travel-advisor/services"
	"example.com/travel-advisor/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// --- Mocks ---

type mockUserService struct {
	findByEmailFunc         func(email string) (*models.User, error)
	createFunc              func(user *models.User) error
	validateCredentialsFunc func(user *models.User) error
}

func (m *mockUserService) FindByEmail(email string) (*models.User, error) {
	return m.findByEmailFunc(email)
}
func (m *mockUserService) Create(user *models.User) error {
	return m.createFunc(user)
}
func (m *mockUserService) ValidateCredentials(user *models.User) error {
	return m.validateCredentialsFunc(user)
}

// Patch services.GetUserService to return our mock
func setMockUserService(mock services.UserServiceInterface) func() {
	orig := services.GetUserService
	services.GetUserService = func() services.UserServiceInterface {
		return mock
	}
	return func() { services.GetUserService = orig }
}

// Patch utils.GenerateToken
func setMockGenerateToken(token string, err error) func() {
	orig := utils.GenerateToken
	utils.GenerateToken = func(email string, userId int64) (string, error) {
		return token, err
	}
	return func() { utils.GenerateToken = orig }
}

// --- Tests ---

func TestSignUp_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockUserService{
		findByEmailFunc: func(email string) (*models.User, error) { return nil, errors.New("not found") },
		createFunc:      func(user *models.User) error { return nil },
	}
	restore := setMockUserService(mockSvc)
	defer restore()

	body := []byte(`{"email":"test@example.com","password":"pass123"}`)
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	signUp(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "User created.")
}

func TestSignUp_UserExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockUserService{
		findByEmailFunc: func(email string) (*models.User, error) { return &models.User{}, nil },
	}
	restore := setMockUserService(mockSvc)
	defer restore()

	body := []byte(`{"email":"test@example.com","password":"pass123"}`)
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	signUp(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "It already exists")
}

func TestSignUp_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer([]byte(`{bad json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	signUp(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Could not parse request data")
}

func TestSignUp_CreateError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockUserService{
		findByEmailFunc: func(email string) (*models.User, error) { return nil, errors.New("not found") },
		createFunc:      func(user *models.User) error { return errors.New("db error") },
	}
	restore := setMockUserService(mockSvc)
	defer restore()

	body := []byte(`{"email":"test@example.com","password":"pass123"}`)
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	signUp(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Could not create user")
}

func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockUserService{
		validateCredentialsFunc: func(user *models.User) error { return nil },
	}
	restoreSvc := setMockUserService(mockSvc)
	defer restoreSvc()
	restoreToken := setMockGenerateToken("mocktoken", nil)
	defer restoreToken()

	body := []byte(`{"email":"test@example.com","password":"pass123"}`)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	login(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Login successful")
	assert.Contains(t, w.Body.String(), "mocktoken")
}

func TestLogin_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte(`{bad json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	login(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Could not parse request data")
}

func TestLogin_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockUserService{
		validateCredentialsFunc: func(user *models.User) error { return errors.New("invalid") },
	}
	restore := setMockUserService(mockSvc)
	defer restore()

	body := []byte(`{"email":"test@example.com","password":"wrongpass"}`)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Wrong user credentials")
}

func TestLogin_TokenError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := &mockUserService{
		validateCredentialsFunc: func(user *models.User) error { return nil },
	}
	restoreSvc := setMockUserService(mockSvc)
	defer restoreSvc()
	restoreToken := setMockGenerateToken("", errors.New("token error"))
	defer restoreToken()

	body := []byte(`{"email":"test@example.com","password":"pass123"}`)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Wrong user credentials")
}
