package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/travel-advisor/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticate_NoToken(t *testing.T) {
	// Set up Gin context
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Authenticate)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create a request without an Authorization header
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	resp := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(resp, req)

	// Assert the response
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.JSONEq(t, `{"message": "Not authorized."}`, resp.Body.String())
}

func TestAuthenticate_InvalidToken(t *testing.T) {
	// Mock utils.VerifyToken to return an error
	utils.VerifyToken = func(token string) (int64, error) {
		return 0, assert.AnError
	}

	// Set up Gin context
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Authenticate)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create a request with an invalid Authorization header
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "invalid-token")
	resp := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(resp, req)

	// Assert the response
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.JSONEq(t, `{"message": "Not authorized."}`, resp.Body.String())
}

func TestAuthenticate_ValidToken(t *testing.T) {
	// Mock utils.VerifyToken to return a valid userId
	utils.VerifyToken = func(token string) (int64, error) {
		return 12345, nil
	}

	// Set up Gin context
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Authenticate)
	router.GET("/test", func(c *gin.Context) {
		userId, _ := c.Get("userId")
		c.JSON(http.StatusOK, gin.H{"message": "success", "userId": userId})
	})

	// Create a request with a valid Authorization header
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "valid-token")
	resp := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(resp, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.JSONEq(t, `{"message": "success", "userId": 12345}`, resp.Body.String())
}
