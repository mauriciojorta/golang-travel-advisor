package routes

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/travel-advisor/models"
	"example.com/travel-advisor/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// --- Mocks ---

type mockItineraryService struct {
	ValidateErr             error
	CreateErr               error
	UpdateErr               error
	DeleteErr               error
	FindByIdIt              *models.Itinerary
	FindByIdErr             error
	ExistByIdResult         bool
	ExistByIdErr            error
	ValidateOwnershipResult bool
	ValidateOwnershipErr    error
	FindByOwner             *[]models.Itinerary
	FindByOwnerErr          error
}

func (m *mockItineraryService) ValidateItineraryDestinationsDates(_ *[]models.ItineraryTravelDestination) error {
	return m.ValidateErr
}
func (m *mockItineraryService) Create(_ *models.Itinerary) error {
	return m.CreateErr
}
func (m *mockItineraryService) FindById(_ int64, _ bool) (*models.Itinerary, error) {
	return m.FindByIdIt, m.FindByIdErr
}

func (m *mockItineraryService) ExistById(_ int64) (bool, error) {
	return m.ExistByIdResult, m.ExistByIdErr
}

func (m *mockItineraryService) ValidateOwnership(_ int64, _ int64) (bool, error) {
	return m.ValidateOwnershipResult, m.ValidateOwnershipErr
}

func (m *mockItineraryService) FindByOwnerId(_ int64) (*[]models.Itinerary, error) {
	return m.FindByOwner, m.FindByOwnerErr
}

func (m *mockItineraryService) Update(_ *models.Itinerary) error {
	return m.UpdateErr
}

func (m *mockItineraryService) Delete(_ int64) error {
	return m.DeleteErr
}

// For runItineraryFileJob
type mockJobsService struct {
	GetJobsRunningOfUserCountVal int
	GetJobsRunningOfUserCountErr error
	PrepareJobTask               *services.ItineraryFileAsyncTaskPayload
	PrepareJobErr                error
	AddAsyncTaskIdErr            error
	FindByItineraryIdResult      *[]models.ItineraryFileJob
	FindByItineraryIdErr         error
	FindByIdResult               *models.ItineraryFileJob
	FindByIdErr                  error
	SoftDeleteErr                error
	SoftDeleteByItineraryId      error
	DeleteErr                    error
}

func (m *mockJobsService) GetJobsRunningOfUserCount(_ int64) (int, error) {
	return m.GetJobsRunningOfUserCountVal, m.GetJobsRunningOfUserCountErr
}
func (m *mockJobsService) PrepareJob(_ *models.Itinerary) (*services.ItineraryFileAsyncTaskPayload, error) {
	return m.PrepareJobTask, m.PrepareJobErr
}
func (m *mockJobsService) AddAsyncTaskId(_ string, _ *models.ItineraryFileJob) error {
	return m.AddAsyncTaskIdErr
}

func (m *mockJobsService) FindAliveById(_ int64) (*models.ItineraryFileJob, error) {
	return m.FindByIdResult, m.FindByIdErr
}

func (m *mockJobsService) FindAliveLightweightById(_ int64) (*models.ItineraryFileJob, error) {
	return m.FindByIdResult, m.FindByItineraryIdErr
}

func (m *mockJobsService) FindAliveByItineraryId(_ int64) (*[]models.ItineraryFileJob, error) {
	return m.FindByItineraryIdResult, m.FindByItineraryIdErr
}

func (m *mockJobsService) StopJob(_ *models.ItineraryFileJob) error {
	return nil
}

func (m *mockJobsService) FailJob(_ string, _ *models.ItineraryFileJob) error {
	return nil
}

func (m *mockJobsService) SoftDeleteJob(_ *models.ItineraryFileJob) error {
	return m.SoftDeleteErr
}

func (m *mockJobsService) SoftDeleteJobsByItineraryId(_ int64, _ *sql.Tx) error {
	return m.SoftDeleteErr
}

func (m *mockJobsService) DeleteJob(_ *models.ItineraryFileJob) error {
	return m.DeleteErr
}

// For runItineraryFileJob
type mockAsyncqTaskQueue struct {
	EnqueueErr error
	EnqueueId  string
}

func (m *mockAsyncqTaskQueue) EnqueueItineraryFileJob(_ services.ItineraryFileAsyncTaskPayload) (*string, error) {
	if m.EnqueueErr != nil {
		return nil, m.EnqueueErr
	}
	return &m.EnqueueId, nil
}

func (m *mockAsyncqTaskQueue) Close() {
	// No-op for mock
}

// --- Helpers ---

func setUserId(ctx *gin.Context, id int64) {
	ctx.Set("userId", id)
}

// --- Tests ---

func Test_createItinerary_Unauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	createItinerary(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_createItinerary_BadRequest_BindJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte("{bad json")))
	createItinerary(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_createItinerary_ValidateItineraryDestinationsDates_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ValidateErr: errors.New("validation error")}
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	body := map[string]interface{}{
		"title":        "Test",
		"description":  "Desc",
		"destinations": []models.ItineraryTravelDestination{},
	}
	b, _ := json.Marshal(body)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(b))
	c.Request.Header.Set("Content-Type", "application/json")
	createItinerary(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_createItinerary_Create_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{CreateErr: errors.New("create error")}
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	body := map[string]interface{}{
		"title":        "Test",
		"description":  "Desc",
		"destinations": []models.ItineraryTravelDestination{},
	}
	b, _ := json.Marshal(body)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(b))
	c.Request.Header.Set("Content-Type", "application/json")
	createItinerary(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_createItinerary_Success(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{}
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	body := map[string]interface{}{
		"title":        "Test",
		"description":  "Desc",
		"notes":        "Test notes",
		"destinations": []models.ItineraryTravelDestination{},
	}
	b, _ := json.Marshal(body)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(b))
	c.Request.Header.Set("Content-Type", "application/json")
	createItinerary(c)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func Test_createItinerary_SuccessNullNotes(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{}
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	body := map[string]interface{}{
		"title":        "Test",
		"description":  "Desc",
		"destinations": []models.ItineraryTravelDestination{},
	}
	b, _ := json.Marshal(body)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(b))
	c.Request.Header.Set("Content-Type", "application/json")
	createItinerary(c)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func Test_getOwnersItineraries_Unauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	getOwnersItineraries(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_getOwnersItineraries_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindByOwnerErr: errors.New("find error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	getOwnersItineraries(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_getOwnersItineraries_Success(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		itineraries := []models.Itinerary{{ID: 1}}
		return &mockItineraryService{FindByOwner: &itineraries}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	getOwnersItineraries(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_getItinerary_Unauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	getItinerary(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_getItinerary_BadRequest_NoParam(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	getItinerary(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_getItinerary_BadRequest_InvalidId(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "abc"}}
	getItinerary(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_getItinerary_ExistById_NotFound(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: false}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getItinerary(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_getItinerary_ExistById_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdErr: errors.New("exist error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getItinerary(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_getItinerary_ValidateOwnership_Forbidden(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: false}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getItinerary(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_getItinerary_ValidateOwnership_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipErr: errors.New("validate ownership error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getItinerary(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_getItinerary_FindById_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true, FindByIdErr: errors.New("find error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getItinerary(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_getItinerary_Success(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true, FindByIdIt: &models.Itinerary{OwnerID: 1}}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getItinerary(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_runItineraryFileJob_Unauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	runItineraryFileJob(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_runItineraryFileJob_BadRequest_NoParam(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	runItineraryFileJob(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_runItineraryFileJob_BadRequest_InvalidId(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "abc"}}
	runItineraryFileJob(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_runItineraryFileJob_ExistById_NotFound(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: false}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	runItineraryFileJob(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_runItineraryFileJob_ExistById_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdErr: errors.New("exist error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	runItineraryFileJob(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_runItineraryFileJob_ValidateOwnership_Forbidden(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: false}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	runItineraryFileJob(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_runItineraryFileJob_ValidateOwnership_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipErr: errors.New("validate ownership error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	runItineraryFileJob(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_runItineraryFileJob_FindById_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true, FindByIdErr: errors.New("find error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	runItineraryFileJob(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_runItineraryFileJob_GetJobsRunningOfUserCount_Error(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true, FindByIdIt: &models.Itinerary{OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{GetJobsRunningOfUserCountErr: errors.New("count error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	runItineraryFileJob(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_runItineraryFileJob_TooManyJobsRunning(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true, FindByIdIt: &models.Itinerary{OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{GetJobsRunningOfUserCountVal: 5}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	runItineraryFileJob(c)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func Test_runItineraryFileJob_PrepareJob_Error(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true, FindByIdIt: &models.Itinerary{OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{
			GetJobsRunningOfUserCountVal: 0,
			PrepareJobErr:                errors.New("prepare error"),
		}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	runItineraryFileJob(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_runItineraryFileJob_InitAsyncTaskQueueClient_Error(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true, FindByIdIt: &models.Itinerary{OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{
			GetJobsRunningOfUserCountVal: 0,
			PrepareJobTask: &services.ItineraryFileAsyncTaskPayload{
				Itinerary:        &models.Itinerary{OwnerID: 1},
				ItineraryFileJob: &models.ItineraryFileJob{},
			},
		}
	}
	origQueue := services.NewAsyncqTaskQueue
	defer func() { services.NewAsyncqTaskQueue = origQueue }()
	services.NewAsyncqTaskQueue = func() (services.AsyncqTaskQueueInterface, error) {
		return nil, errors.New("REDIS_PASSWORD environment variable is not set")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	runItineraryFileJob(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_runItineraryFileJob_Enqueue_Error(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true, FindByIdIt: &models.Itinerary{OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{
			GetJobsRunningOfUserCountVal: 0,
			PrepareJobTask: &services.ItineraryFileAsyncTaskPayload{
				Itinerary:        &models.Itinerary{OwnerID: 1},
				ItineraryFileJob: &models.ItineraryFileJob{},
			},
		}
	}
	origQueue := services.NewAsyncqTaskQueue
	defer func() { services.NewAsyncqTaskQueue = origQueue }()
	services.NewAsyncqTaskQueue = func() (services.AsyncqTaskQueueInterface, error) {
		return &mockAsyncqTaskQueue{EnqueueErr: errors.New("enqueue error")}, nil
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	runItineraryFileJob(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_runItineraryFileJob_AddAsyncTaskId_Error(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true, FindByIdIt: &models.Itinerary{OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{
			GetJobsRunningOfUserCountVal: 0,
			PrepareJobTask: &services.ItineraryFileAsyncTaskPayload{
				Itinerary:        &models.Itinerary{OwnerID: 1},
				ItineraryFileJob: &models.ItineraryFileJob{},
			},
			AddAsyncTaskIdErr: errors.New("add async id error"),
		}
	}
	origQueue := services.NewAsyncqTaskQueue
	defer func() { services.NewAsyncqTaskQueue = origQueue }()
	services.NewAsyncqTaskQueue = func() (services.AsyncqTaskQueueInterface, error) {
		return &mockAsyncqTaskQueue{EnqueueId: "taskid"}, nil
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	runItineraryFileJob(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_runItineraryFileJob_Success(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true, FindByIdIt: &models.Itinerary{OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{
			GetJobsRunningOfUserCountVal: 0,
			PrepareJobTask: &services.ItineraryFileAsyncTaskPayload{
				Itinerary:        &models.Itinerary{OwnerID: 1},
				ItineraryFileJob: &models.ItineraryFileJob{},
			},
		}
	}
	origQueue := services.NewAsyncqTaskQueue
	defer func() { services.NewAsyncqTaskQueue = origQueue }()
	services.NewAsyncqTaskQueue = func() (services.AsyncqTaskQueueInterface, error) {
		return &mockAsyncqTaskQueue{EnqueueId: "taskid"}, nil
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	runItineraryFileJob(c)
	assert.Equal(t, http.StatusAccepted, w.Code)
}

func Test_getAllItineraryFileJobs_Unauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	getAllItineraryFileJobs(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_getAllItineraryFileJobs_BadRequest_NoParam(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	getAllItineraryFileJobs(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_getAllItineraryFileJobs_BadRequest_InvalidId(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "abc"}}
	getAllItineraryFileJobs(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_getAllItineraryFileJobs_ExistById_NotFound(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: false}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getAllItineraryFileJobs(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_getAllItineraryFileJobs_ExistById_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdErr: errors.New("exist error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getAllItineraryFileJobs(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_getAllItineraryFileJobs_ValidateOwnership_Forbidden(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: false}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getAllItineraryFileJobs(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_getAllItineraryFileJobs_ValidateOwnership_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipErr: errors.New("validate ownership error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getAllItineraryFileJobs(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_getAllItineraryFileJobs_FindByItineraryId_Error(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true, FindByIdIt: &models.Itinerary{OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{FindByItineraryIdResult: nil, FindByItineraryIdErr: errors.New("find jobs error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getAllItineraryFileJobs(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_getAllItineraryFileJobs_Success(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true, FindByIdIt: &models.Itinerary{OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	jobs := []models.ItineraryFileJob{{ID: 1}, {ID: 2}}
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{FindByItineraryIdResult: &jobs}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getAllItineraryFileJobs(c)
	assert.Equal(t, http.StatusOK, w.Code)
}
func Test_updateItinerary_Unauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	updateItinerary(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_updateItinerary_BadRequest_NoParam(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	updateItinerary(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_updateItinerary_BadRequest_InvalidIdBindJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	body := map[string]interface{}{
		"id":           "1",
		"title":        "Test",
		"description":  "Desc",
		"destinations": []models.ItineraryTravelDestination{},
	}
	b, _ := json.Marshal(body)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(b))
	updateItinerary(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_updateItinerary_BadRequest_BindJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer([]byte("{bad json")))
	updateItinerary(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_updateItinerary_ExistById_NotFound(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: false}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	body := map[string]interface{}{
		"id":           1,
		"title":        "Test",
		"description":  "Desc",
		"destinations": []models.ItineraryTravelDestination{},
	}
	b, _ := json.Marshal(body)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(b))
	c.Request.Header.Set("Content-Type", "application/json")
	updateItinerary(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_updateItinerary_ExistById_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdErr: errors.New("exist error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	body := map[string]interface{}{
		"id":           1,
		"title":        "Test",
		"description":  "Desc",
		"destinations": []models.ItineraryTravelDestination{},
	}
	b, _ := json.Marshal(body)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(b))
	c.Request.Header.Set("Content-Type", "application/json")
	updateItinerary(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_updateItinerary_FindById_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, FindByIdErr: errors.New("find error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	body := map[string]interface{}{
		"id":           1,
		"title":        "Test",
		"description":  "Desc",
		"destinations": []models.ItineraryTravelDestination{},
	}
	b, _ := json.Marshal(body)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(b))
	c.Request.Header.Set("Content-Type", "application/json")
	updateItinerary(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_updateItinerary_Forbidden(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, FindByIdIt: &models.Itinerary{OwnerID: 2}}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	body := map[string]interface{}{
		"id":           1,
		"title":        "Test",
		"description":  "Desc",
		"destinations": []models.ItineraryTravelDestination{},
	}
	b, _ := json.Marshal(body)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(b))
	c.Request.Header.Set("Content-Type", "application/json")
	updateItinerary(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_updateItinerary_ValidateItineraryDestinationsDates_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{
			ExistByIdResult: true,
			FindByIdIt:      &models.Itinerary{OwnerID: 1},
			ValidateErr:     errors.New("validation error"),
		}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	body := map[string]interface{}{
		"id":           1,
		"title":        "Test",
		"description":  "Desc",
		"destinations": []models.ItineraryTravelDestination{},
	}
	b, _ := json.Marshal(body)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(b))
	c.Request.Header.Set("Content-Type", "application/json")
	updateItinerary(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_updateItinerary_Update_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	mock := &mockItineraryService{
		ExistByIdResult: true,
		FindByIdIt:      &models.Itinerary{OwnerID: 1},
		UpdateErr:       errors.New("update error"),
	}
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return mock
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	body := map[string]interface{}{
		"id":           1,
		"title":        "Test",
		"description":  "Desc",
		"destinations": []models.ItineraryTravelDestination{},
	}
	b, _ := json.Marshal(body)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(b))
	c.Request.Header.Set("Content-Type", "application/json")
	updateItinerary(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_updateItinerary_Success(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	mock := &mockItineraryService{
		ExistByIdResult: true,
		FindByIdIt:      &models.Itinerary{OwnerID: 1},
		UpdateErr:       nil,
	}
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return mock
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	body := map[string]interface{}{
		"id":           1,
		"title":        "Test",
		"description":  "Desc",
		"notes":        "Test notes",
		"destinations": []models.ItineraryTravelDestination{},
	}
	b, _ := json.Marshal(body)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(b))
	c.Request.Header.Set("Content-Type", "application/json")
	updateItinerary(c)
	assert.Equal(t, http.StatusOK, w.Code)
}
func Test_deleteItinerary_Unauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	deleteItinerary(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_deleteItinerary_BadRequest_NoParam(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	deleteItinerary(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_deleteItinerary_BadRequest_InvalidId(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "abc"}}
	deleteItinerary(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_deleteItinerary_ExistById_NotFound(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: false}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	deleteItinerary(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_deleteItinerary_ExistById_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdErr: errors.New("exist error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	deleteItinerary(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_deleteItinerary_ValidateOwnership_Forbidden(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: false}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	deleteItinerary(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_deleteItinerary_ValidateOwnership_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipErr: errors.New("validate ownership error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	deleteItinerary(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_deleteItinerary_Delete_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	mock := &mockItineraryService{
		ExistByIdResult:         true,
		ValidateOwnershipResult: true,
		DeleteErr:               errors.New("delete error"),
	}
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return mock
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	deleteItinerary(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_deleteItinerary_Success(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	deleteItinerary(c)
	assert.Equal(t, http.StatusOK, w.Code)
}
func Test_getItineraryJob_Unauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	getItineraryJob(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_getItineraryJob_BadRequest_NoItineraryId(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	getItineraryJob(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_getItineraryJob_BadRequest_NoItineraryJobId(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true}
	}
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getItineraryJob(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_getItineraryJob_BadRequest_InvalidItineraryId(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "abc"},
		{Key: "itineraryJobId", Value: "2"},
	}
	getItineraryJob(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_getItineraryJob_ExistItineraryById_NotFound(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: false}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "1"},
	}
	getItineraryJob(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_getItineraryJob_ExistItineraryById_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdErr: errors.New("exist error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "1"},
	}
	getItineraryJob(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_getItineraryJob_ValidateItineraryOwnership_Forbidden(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: false}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "1"},
	}
	getItineraryJob(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_getItineraryJob_ValidateItineraryOwnership_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipErr: errors.New("validate ownership error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "1"},
	}
	getItineraryJob(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_getItineraryJob_ForbiddenItineraryJob(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true, FindByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{
			FindByIdResult:       &models.ItineraryFileJob{ID: 1, ItineraryID: 2},
			FindByItineraryIdErr: nil,
		}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "1"},
	}
	getItineraryJob(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_getItineraryJob_BadRequest_InvalidItineraryJobId(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true, FindByIdIt: &models.Itinerary{OwnerID: 1}}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "abc"},
	}
	getItineraryJob(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_getItineraryJob_FindByJobId_Error(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true, FindByIdIt: &models.Itinerary{OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{
			FindByIdResult: nil,
			FindByIdErr:    errors.New("Itinerary not found"),
		}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "1"},
	}

	getItineraryJob(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_getItineraryJob_Success(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true, FindByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{
			FindByIdResult:       &models.ItineraryFileJob{ID: 1, ItineraryID: 1},
			FindByItineraryIdErr: nil,
		}
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "1"},
	}
	getItineraryJob(c)
	assert.Equal(t, http.StatusOK, w.Code)
}
func Test_deleteItineraryJob_Unauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	deleteItineraryJob(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_deleteItineraryJob_BadRequest_NoItineraryId(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	deleteItineraryJob(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_deleteItineraryJob_BadRequest_InvalidItineraryId(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "abc"}}
	deleteItineraryJob(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_deleteItineraryJob_ExistById_NotFound(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: false}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	deleteItineraryJob(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_deleteItineraryJob_ExistById_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdErr: errors.New("exist error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	deleteItineraryJob(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_deleteItineraryJob_ValidateOwnership_Forbidden(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: false}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	deleteItineraryJob(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_deleteItineraryJob_ValidateOwnership_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipErr: errors.New("validate ownership error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	deleteItineraryJob(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_deleteItineraryJob_BadRequest_NoItineraryJobId(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	deleteItineraryJob(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_deleteItineraryJob_BadRequest_InvalidItineraryJobId(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "abc"},
	}
	deleteItineraryJob(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_deleteItineraryJob_FindAliveLightweightById_NotFound(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{FindByItineraryIdErr: sql.ErrNoRows}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "2"},
	}
	deleteItineraryJob(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_deleteItineraryJob_FindAliveLightweightById_Error(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{FindByItineraryIdErr: errors.New("find error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "2"},
	}
	deleteItineraryJob(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_deleteItineraryJob_Forbidden_JobOwnership(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{FindByIdResult: &models.ItineraryFileJob{ID: 2, ItineraryID: 99}}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "2"},
	}
	deleteItineraryJob(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_deleteItineraryJob_SoftDeleteJob_Error(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true, FindByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{
			FindByIdResult: &models.ItineraryFileJob{ID: 2, ItineraryID: 1},
			SoftDeleteErr:  errors.New("delete error"),
		}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "2"},
	}
	deleteItineraryJob(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_deleteItineraryJob_Success(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{ExistByIdResult: true, ValidateOwnershipResult: true}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{
			FindByIdResult: &models.ItineraryFileJob{ID: 2, ItineraryID: 1},
			SoftDeleteErr:  nil,
		}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "2"},
	}
	deleteItineraryJob(c)
	assert.Equal(t, http.StatusOK, w.Code)
}
