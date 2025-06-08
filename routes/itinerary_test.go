package routes

import (
	"bytes"
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
	ValidateErr    error
	CreateErr      error
	FindByIdIt     *models.Itinerary
	FindByIdErr    error
	FindByOwner    *[]models.Itinerary
	FindByOwnerErr error
}

func (m *mockItineraryService) ValidateItineraryDestinationsDates(_ *[]models.ItineraryTravelDestination) error {
	return m.ValidateErr
}
func (m *mockItineraryService) Create(_ *models.Itinerary) error {
	return m.CreateErr
}
func (m *mockItineraryService) FindById(_ int64) (*models.Itinerary, error) {
	return m.FindByIdIt, m.FindByIdErr
}
func (m *mockItineraryService) FindByOwnerId(_ int64) (*[]models.Itinerary, error) {
	return m.FindByOwner, m.FindByOwnerErr
}

func (m *mockItineraryService) Update(_ *models.Itinerary) error {
	return nil // Not used in tests
}

func (m *mockItineraryService) Delete(_ *models.Itinerary) error {
	return nil // Not used in tests
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

func (m *mockJobsService) FindById(_ int64) (*models.ItineraryFileJob, error) {
	return nil, nil // Not used in tests
}

func (m *mockJobsService) FindByItineraryId(_ int64) (*[]models.ItineraryFileJob, error) {
	return m.FindByItineraryIdResult, m.FindByItineraryIdErr
}

func (m *mockJobsService) StopJob(_ *models.ItineraryFileJob) error {
	return nil // Not used in tests
}

func (m *mockJobsService) FailJob(_ string, _ *models.ItineraryFileJob) error {
	return nil // Not used in tests
}

func (m *mockJobsService) DeleteJob(_ *models.ItineraryFileJob) error {
	return nil // Not used in tests
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

func Test_getItinerary_FindById_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindByIdErr: errors.New("find error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getItinerary(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_getItinerary_Forbidden(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindByIdIt: &models.Itinerary{OwnerID: 2}}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getItinerary(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_getItinerary_Success(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindByIdIt: &models.Itinerary{OwnerID: 1}}
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

func Test_runItineraryFileJob_FindById_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindByIdErr: errors.New("find error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	runItineraryFileJob(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_runItineraryFileJob_Forbidden(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindByIdIt: &models.Itinerary{OwnerID: 2}}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	runItineraryFileJob(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_runItineraryFileJob_GetJobsRunningOfUserCount_Error(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindByIdIt: &models.Itinerary{OwnerID: 1}}
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
		return &mockItineraryService{FindByIdIt: &models.Itinerary{OwnerID: 1}}
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
		return &mockItineraryService{FindByIdIt: &models.Itinerary{OwnerID: 1}}
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

func Test_runItineraryFileJob_Enqueue_Error(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindByIdIt: &models.Itinerary{OwnerID: 1}}
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
	services.NewAsyncqTaskQueue = func() services.AsyncqTaskQueueInterface {
		return &mockAsyncqTaskQueue{EnqueueErr: errors.New("enqueue error")}
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
		return &mockItineraryService{FindByIdIt: &models.Itinerary{OwnerID: 1}}
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
	services.NewAsyncqTaskQueue = func() services.AsyncqTaskQueueInterface {
		return &mockAsyncqTaskQueue{EnqueueId: "taskid"}
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
		return &mockItineraryService{FindByIdIt: &models.Itinerary{OwnerID: 1}}
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
	services.NewAsyncqTaskQueue = func() services.AsyncqTaskQueueInterface {
		return &mockAsyncqTaskQueue{EnqueueId: "taskid"}
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

func Test_getAllItineraryFileJobs_FindById_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindByIdErr: errors.New("find error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getAllItineraryFileJobs(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_getAllItineraryFileJobs_Forbidden(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindByIdIt: &models.Itinerary{OwnerID: 2}}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getAllItineraryFileJobs(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_getAllItineraryFileJobs_FindByItineraryId_Error(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindByIdIt: &models.Itinerary{OwnerID: 1}}
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
		return &mockItineraryService{FindByIdIt: &models.Itinerary{OwnerID: 1}}
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
