package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/travel-advisor/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateItinerary_Success(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/itineraries", func(c *gin.Context) {
		c.Set("userId", int64(1)) // Mock userId

		// Mock the Create method
		mockItinerary := models.NewItinerary("", "", time.Time{}, time.Time{}, nil)
		mockItinerary.Create = func() error {
			return nil // Simulate successful creation
		}
		models.NewItinerary = func(title string, description string, travelStartDate time.Time, travelEndDate time.Time, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
			return mockItinerary
		}

		createItinerary(c)
	})

	input := map[string]interface{}{
		"title":           "Test Itinerary",
		"description":     "Test Description",
		"travelStartDate": time.Now(),
		"travelEndDate":   time.Now().Add(48 * time.Hour),
		"destinations": []models.ItineraryTravelDestination{
			{Country: "Country 1", City: "City 1", ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
		},
	}

	body, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, "/itineraries", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusCreated, rec.Code)
	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Itinerary created.", response["message"])
}

func TestCreateItinerary_Unauthorized(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/itineraries", createItinerary)

	req, _ := http.NewRequest(http.MethodPost, "/itineraries", nil)
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCreateItinerary_BadRequest(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/itineraries", func(c *gin.Context) {
		c.Set("userId", int64(1)) // Mock userId
		createItinerary(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/itineraries", bytes.NewBuffer([]byte(`invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetOwnersItineraries_Success(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/itineraries", func(c *gin.Context) {
		c.Set("userId", int64(1)) // Mock userId

		// Mock the FindByOwnerId method
		mockItinerary := models.NewItinerary("", "", time.Time{}, time.Time{}, nil)
		mockItinerary.FindByOwnerId = func() error {
			mockItinerary.ID = 1
			mockItinerary.Title = "Mock Itinerary"
			mockItinerary.Description = "Mock Description"
			mockItinerary.TravelStartDate = time.Now()
			mockItinerary.TravelEndDate = time.Now().Add(48 * time.Hour)
			mockItinerary.OwnerID = 1
			return nil // Simulate successful retrieval
		}
		models.NewItinerary = func(title string, description string, travelStartDate time.Time, travelEndDate time.Time, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
			return mockItinerary
		}

		getOwnersItineraries(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/itineraries", nil)
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["itineraries"])
}

func TestGetOwnersItineraries_Unauthorized(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/itineraries", getOwnersItineraries)

	req, _ := http.NewRequest(http.MethodGet, "/itineraries", nil)
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
