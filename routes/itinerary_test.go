package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
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
		mockItinerary := models.NewItinerary("", "", nil)
		mockItinerary.Create = func() error {
			return nil // Simulate successful creation
		}
		models.NewItinerary = func(title string, description string, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
			return mockItinerary
		}

		models.ValidateItineraryDestinationsDates = func(destinations *[]models.ItineraryTravelDestination) error {
			return nil // Simulate successful validation
		}

		createItinerary(c)
	})

	input := map[string]interface{}{
		"title":       "Test Itinerary",
		"description": "Test Description",
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

func TestCreateItinerary_DestinationDateValidationError(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/itineraries", func(c *gin.Context) {
		c.Set("userId", int64(1)) // Mock userId

		// Mock the Create method
		mockItinerary := models.NewItinerary("", "", nil)
		mockItinerary.Create = func() error {
			return nil // Simulate successful creation
		}
		models.NewItinerary = func(title string, description string, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
			return mockItinerary
		}

		models.ValidateItineraryDestinationsDates = func(destinations *[]models.ItineraryTravelDestination) error {
			return fmt.Errorf("Arrival date must be before departure date") // Simulate validation error
		}

		createItinerary(c)
	})

	input := map[string]interface{}{
		"title":       "Test Itinerary",
		"description": "Test Description",
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
	assert.Equal(t, http.StatusBadRequest, rec.Code)

}

func TestGetOwnersItineraries_Success(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/itineraries", func(c *gin.Context) {
		c.Set("userId", int64(1)) // Mock userId

		// Mock the FindByOwnerId method
		mockItinerary := models.NewItinerary("", "", nil)
		mockItinerary.FindByOwnerId = func() (*[]models.Itinerary, error) {
			mockItinerary.ID = 1
			mockItinerary.Title = "Mock Itinerary"
			mockItinerary.Description = "Mock Description"
			mockItinerary.OwnerID = 1
			return &[]models.Itinerary{*mockItinerary}, nil // Simulate successful retrieval
		}
		models.NewItinerary = func(title string, description string, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
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

func TestGetItinerary_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/itineraries/:itineraryId", func(c *gin.Context) {
		c.Set("userId", int64(1)) // Mock userId

		getItinerary(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/itineraries/abc", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Invalid itinerary ID.", resp["message"])
}

func TestGetItinerary_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/itineraries/:itineraryId", func(c *gin.Context) {
		// Mock FindById to return error
		mockItinerary := models.NewItinerary("", "", nil)
		mockItinerary.FindById = func() error {
			return fmt.Errorf("db error")
		}
		models.NewItinerary = func(title string, description string, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
			return mockItinerary
		}
		c.Set("userId", int64(1))
		getItinerary(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/itineraries/1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Could not retrieve itinerary. Try again later.", resp["message"])
}

func TestGetItinerary_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/itineraries/:itineraryId", func(c *gin.Context) {
		mockItinerary := models.NewItinerary("", "", nil)
		mockItinerary.FindById = func() error { return nil }
		mockItinerary.OwnerID = 1
		models.NewItinerary = func(title string, description string, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
			return mockItinerary
		}
		getItinerary(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/itineraries/1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Not authorized.", resp["message"])
}

func TestGetItinerary_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/itineraries/:itineraryId", func(c *gin.Context) {
		mockItinerary := models.NewItinerary("", "", nil)
		mockItinerary.FindById = func() error { return nil }
		mockItinerary.OwnerID = 2 // Not matching userId
		models.NewItinerary = func(title string, description string, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
			return mockItinerary
		}
		c.Set("userId", int64(1))
		getItinerary(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/itineraries/1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "You do not have permission to access this itinerary.", resp["message"])
}

func TestGetItinerary_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/itineraries/:itineraryId", func(c *gin.Context) {
		mockItinerary := models.NewItinerary("", "", nil)
		mockItinerary.FindById = func() error { return nil }
		mockItinerary.OwnerID = 1
		mockItinerary.ID = 1
		mockItinerary.Title = "Test"
		models.NewItinerary = func(title string, description string, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
			return mockItinerary
		}
		c.Set("userId", int64(1))
		getItinerary(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/itineraries/1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NotNil(t, resp["itinerary"])
}

func TestRunItineraryFileJob_InvalidItineraryID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/itineraries/:itineraryId/job", func(c *gin.Context) {
		c.Set("userId", int64(1))

		mockItinerary := models.NewItinerary("", "", nil)
		mockItinerary.FindById = func() error {
			return nil
		}
		models.NewItinerary = func(title string, description string, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
			return mockItinerary
		}
		runItineraryFileJob(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/itineraries/abc/job", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Invalid itinerary ID.", resp["message"])
}

func TestRunItineraryFileJob_ItineraryNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/itineraries/:itineraryId/job", func(c *gin.Context) {
		c.Set("userId", int64(1))

		mockItinerary := models.NewItinerary("", "", nil)
		mockItinerary.FindById = func() error {
			return fmt.Errorf("not found")
		}
		models.NewItinerary = func(title string, description string, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
			return mockItinerary
		}
		runItineraryFileJob(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/itineraries/1/job", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Could not retrieve itinerary. Try again later.", resp["message"])
}

func TestRunItineraryFileJob_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/itineraries/:itineraryId/job", runItineraryFileJob)

	req, _ := http.NewRequest(http.MethodPost, "/itineraries/1/job", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Not authorized.", resp["message"])
}

func TestRunItineraryFileJob_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/itineraries/:itineraryId/job", func(c *gin.Context) {
		mockItinerary := models.NewItinerary("", "", nil)
		mockItinerary.FindById = func() error { return nil }
		mockItinerary.OwnerID = 2 // Not matching userId
		models.NewItinerary = func(title string, description string, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
			return mockItinerary
		}
		c.Set("userId", int64(1))
		runItineraryFileJob(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/itineraries/1/job", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "You do not have permission to run this job.", resp["message"])
}

func TestRunItineraryFileJob_JobAlreadyRunning(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/itineraries/:itineraryId/job", func(c *gin.Context) {
		mockItinerary := models.NewItinerary("", "", nil)
		mockItinerary.FindById = func() error { return nil }
		mockItinerary.OwnerID = 1
		models.NewItinerary = func(title string, description string, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
			return mockItinerary
		}
		mockJob := models.NewItineraryFileJob(1)
		mockJob.HasItineraryAJobRunning = func() (bool, error) {
			return true, nil // Simulate job already running
		}
		models.NewItineraryFileJob = func(itineraryID int64) *models.ItineraryFileJob {
			return mockJob
		}
		c.Set("userId", int64(1))
		runItineraryFileJob(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/itineraries/1/job", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "A job is already running for this itinerary.", resp["message"])
}

func TestRunItineraryFileJob_RunJobError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/itineraries/:itineraryId/job", func(c *gin.Context) {
		mockItinerary := models.NewItinerary("", "", nil)
		mockItinerary.FindById = func() error { return nil }
		mockItinerary.OwnerID = 1
		models.NewItinerary = func(title string, description string, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
			return mockItinerary
		}
		mockJob := models.NewItineraryFileJob(1)
		mockJob.HasItineraryAJobRunning = func() (bool, error) {
			return false, nil // No job running
		}
		mockJob.RunJob = func(itinerary *models.Itinerary) error {
			return fmt.Errorf("job run error") // Simulate job run error
		}
		models.NewItineraryFileJob = func(itineraryID int64) *models.ItineraryFileJob {
			return mockJob
		}
		c.Set("userId", int64(1))
		runItineraryFileJob(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/itineraries/1/job", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Could not start job. Try again later.", resp["message"])
}

func TestRunItineraryFileJob_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/itineraries/:itineraryId/job", func(c *gin.Context) {
		mockItinerary := models.NewItinerary("", "", nil)
		mockItinerary.FindById = func() error { return nil }
		mockItinerary.OwnerID = 1
		models.NewItinerary = func(title string, description string, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
			return mockItinerary
		}
		mockJob := models.NewItineraryFileJob(1)
		mockJob.HasItineraryAJobRunning = func() (bool, error) {
			return false, nil // No job running
		}
		mockJob.RunJob = func(itinerary *models.Itinerary) error {
			return nil
		}
		models.NewItineraryFileJob = func(itineraryID int64) *models.ItineraryFileJob {
			return mockJob
		}
		c.Set("userId", int64(1))
		runItineraryFileJob(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/itineraries/1/job", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusAccepted, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Job started successfully.", resp["message"])
}

func TestGetAllItineraryFileJobs_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/itineraries/:itineraryId/jobs", getAllItineraryFileJobs)

	req, _ := http.NewRequest(http.MethodGet, "/itineraries/1/jobs", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Not authorized.", resp["message"])
}

func TestGetAllItineraryFileJobs_InvalidItineraryID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/itineraries/:itineraryId/jobs", func(c *gin.Context) {
		mockItinerary := models.NewItinerary("", "", nil)
		mockItinerary.FindById = func() error { return nil }
		mockItinerary.OwnerID = 1
		models.NewItinerary = func(title string, description string, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
			return mockItinerary
		}
		mockJob := models.NewItineraryFileJob(1)
		mockJob.FindByItineraryId = func() (*[]models.ItineraryFileJob, error) {
			jobs := []models.ItineraryFileJob{
				{ID: 1, Status: "Completed"},
				{ID: 2, Status: "Running"},
			}
			return &jobs, nil
		}
		models.NewItineraryFileJob = func(itineraryID int64) *models.ItineraryFileJob {
			return mockJob
		}
		c.Set("userId", int64(1))
		getAllItineraryFileJobs(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/itineraries/abc/jobs", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Invalid itinerary ID.", resp["message"])
}

func TestGetAllItineraryFileJobs_MissingItineraryID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/itineraries/:itineraryId/jobs", func(c *gin.Context) {
		mockItinerary := models.NewItinerary("", "", nil)
		mockItinerary.FindById = func() error { return nil }
		mockItinerary.OwnerID = 1
		models.NewItinerary = func(title string, description string, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
			return mockItinerary
		}
		mockJob := models.NewItineraryFileJob(1)
		mockJob.FindByItineraryId = func() (*[]models.ItineraryFileJob, error) {
			jobs := []models.ItineraryFileJob{
				{ID: 1, Status: "Completed"},
				{ID: 2, Status: "Running"},
			}
			return &jobs, nil
		}
		models.NewItineraryFileJob = func(itineraryID int64) *models.ItineraryFileJob {
			return mockJob
		}
		c.Set("userId", int64(1))
		getAllItineraryFileJobs(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/itineraries//jobs", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Itinerary ID is required.", resp["message"])
}

func TestGetAllItineraryFileJobs_ItineraryNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/itineraries/:itineraryId/jobs", func(c *gin.Context) {
		c.Set("userId", int64(1))

		mockItinerary := models.NewItinerary("", "", nil)
		mockItinerary.FindById = func() error {
			return fmt.Errorf("not found")
		}
		models.NewItinerary = func(title string, description string, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
			return mockItinerary
		}
		getAllItineraryFileJobs(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/itineraries/1/jobs", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Could not retrieve itinerary. Try again later.", resp["message"])
}

func TestGetAllItineraryFileJobs_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/itineraries/:itineraryId/jobs", func(c *gin.Context) {
		mockItinerary := models.NewItinerary("", "", nil)
		mockItinerary.FindById = func() error { return nil }
		mockItinerary.OwnerID = 2 // Not matching userId
		models.NewItinerary = func(title string, description string, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
			return mockItinerary
		}
		c.Set("userId", int64(1))
		getAllItineraryFileJobs(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/itineraries/1/jobs", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "You do not have permission to run this job.", resp["message"])
}

func TestGetAllItineraryFileJobs_RetrieveJobsError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/itineraries/:itineraryId/jobs", func(c *gin.Context) {
		mockItinerary := models.NewItinerary("", "", nil)
		mockItinerary.FindById = func() error { return nil }
		mockItinerary.OwnerID = 1
		models.NewItinerary = func(title string, description string, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
			return mockItinerary
		}
		mockJob := models.NewItineraryFileJob(1)
		mockJob.FindByItineraryId = func() (*[]models.ItineraryFileJob, error) {
			return nil, fmt.Errorf("db error")
		}
		models.NewItineraryFileJob = func(itineraryID int64) *models.ItineraryFileJob {
			return mockJob
		}
		c.Set("userId", int64(1))
		getAllItineraryFileJobs(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/itineraries/1/jobs", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Could not retrieve jobs. Try again later.", resp["message"])
}

func TestGetAllItineraryFileJobs_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/itineraries/:itineraryId/jobs", func(c *gin.Context) {
		mockItinerary := models.NewItinerary("", "", nil)
		mockItinerary.FindById = func() error { return nil }
		mockItinerary.OwnerID = 1
		models.NewItinerary = func(title string, description string, travelDestinations []models.ItineraryTravelDestination) *models.Itinerary {
			return mockItinerary
		}
		mockJob := models.NewItineraryFileJob(1)
		mockJob.FindByItineraryId = func() (*[]models.ItineraryFileJob, error) {
			jobs := []models.ItineraryFileJob{
				{ID: 1, Status: "Completed"},
				{ID: 2, Status: "Running"},
			}
			return &jobs, nil
		}
		models.NewItineraryFileJob = func(itineraryID int64) *models.ItineraryFileJob {
			return mockJob
		}
		c.Set("userId", int64(1))
		getAllItineraryFileJobs(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/itineraries/1/jobs", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NotNil(t, resp["jobs"])
	assert.Len(t, resp["jobs"], 2)
}
