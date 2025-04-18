package routes

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/travel-advisor/models"
	"example.com/travel-advisor/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSignUp_InvalidRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/signup", signUp)

	body := `{"email": "invalid-email"}`
	req, _ := http.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Could not parse request data.")
}

func TestSignUp_UserAlreadyExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/signup", func(c *gin.Context) {

		models.NewUser = func(email, password string) *models.User {
			user := &models.User{
				Email:    email,
				Password: password,
			}

			user.FindUser = func() error {
				return nil // Simulate user already exists
			}

			return user
		}

		signUp(c)
	})

	body := `{"email": "test@example.com", "password": "password123"}`
	req, _ := http.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Could not create user. It already exists.")
}

func TestSignUp_CreateUserError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/signup", func(c *gin.Context) {
		models.NewUser = func(email, password string) *models.User {
			user := &models.User{
				Email:    email,
				Password: password,
			}

			// Set default implementations for FindUser and Create
			mockFindUser := func() error {
				return errors.New("user not found") // Simulate user does not exist
			}
			user.FindUser = mockFindUser

			mockCreate := func() error {
				return errors.New("create error") // Simulate create user error
			}
			user.Create = mockCreate

			return user
		}

		signUp(c)
	})

	body := `{"email": "test@example.com", "password": "password123"}`
	req, _ := http.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "Could not create user. Try again later.")
}

func TestSignUp_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/signup", func(c *gin.Context) {
		models.NewUser = func(email, password string) *models.User {
			user := &models.User{
				Email:    email,
				Password: password,
			}

			// Set default implementations for FindUser and Create
			user.FindUser = func() error {
				return errors.New("user not found") // Simulate user does not exist
			}
			user.Create = func() error {
				return nil // Simulate successful user creation
			}

			return user
		}

		signUp(c)
	})

	body := `{"email": "test@example.com", "password": "password123"}`
	req, _ := http.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "User created.")
}

func TestLogin_InvalidRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/login", login)

	body := `{"email": "invalid-email"}`
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Could not parse request data.")
}

func TestLogin_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/login", func(c *gin.Context) {
		models.NewUser = func(email, password string) *models.User {
			user := &models.User{
				Email:    email,
				Password: password,
			}

			user.ValidateCredentials = func() error {
				return errors.New("invalid credentials")
			}

			return user
		}

		login(c)
	})

	body := `{"email": "test@example.com", "password": "wrongpassword"}`
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Wrong user credentials.")
}

func TestLogin_TokenGenerationFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/login", func(c *gin.Context) {
		models.NewUser = func(email, password string) *models.User {
			user := &models.User{
				Email:    email,
				Password: password,
			}

			user.ValidateCredentials = func() error {
				return nil // Simulate valid credentials
			}

			return user
		}
		utils.GenerateToken = func(email string, userID int64) (string, error) {
			return "", errors.New("token generation failed")
		}

		login(c)
	})

	body := `{"email": "test@example.com", "password": "password123"}`
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Wrong user credentials.")
}

func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/login", func(c *gin.Context) {

		models.NewUser = func(email, password string) *models.User {
			user := &models.User{
				Email:    email,
				Password: password,
			}

			user.ValidateCredentials = func() error {
				return nil // Simulate valid credentials
			}

			return user
		}

		utils.GenerateToken = func(email string, userID int64) (string, error) {
			return "mocked-token", nil // Simulate successful token generation
		}
		login(c)
	})

	body := `{"email": "test@example.com", "password": "password123"}`
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Login successful!")
	assert.Contains(t, rec.Body.String(), "mocked-token")
}
