package routes

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"example.com/travel-advisor/models"
	"example.com/travel-advisor/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// --- Mocks ---

type mockItineraryService struct {
	ValidateErr            error
	CreateErr              error
	UpdateErr              error
	DeleteErr              error
	FindByIdIt             *models.Itinerary
	FindByIdErr            error
	FindLightweightByIdIt  *models.Itinerary
	FindLightweightByIdErr error
	FindByOwner            *[]models.Itinerary
	FindByOwnerErr         error
}

func (m *mockItineraryService) ValidateItineraryDestinationsDates(_ []*models.ItineraryTravelDestination) error {
	return m.ValidateErr
}
func (m *mockItineraryService) Create(_ *models.Itinerary) error {
	return m.CreateErr
}
func (m *mockItineraryService) FindById(_ int64, _ bool) (*models.Itinerary, error) {
	return m.FindByIdIt, m.FindByIdErr
}

func (m *mockItineraryService) FindLightweightById(_ int64) (*models.Itinerary, error) {
	return m.FindLightweightByIdIt, m.FindLightweightByIdErr
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
	GetJobsRunningOfUserCountVal   int
	GetJobsRunningOfUserCountErr   error
	PrepareJobTask                 *services.ItineraryFileAsyncTaskPayload
	PrepareJobErr                  error
	StopJobErr                     error
	AddAsyncTaskIdErr              error
	FindByItineraryIdResult        *[]models.ItineraryFileJob
	FindByItineraryIdErr           error
	FindByIdResult                 *models.ItineraryFileJob
	FindByIdErr                    error
	FindAliveLightweightByIdResult *models.ItineraryFileJob
	FindAliveLightweightByIdErr    error
	OpenItineraryJobFileResult     io.ReadSeekCloser
	OpenItineraryJobFileErr        error
	SoftDeleteErr                  error
	SoftDeleteByItineraryId        error
	DeleteErr                      error
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
	return m.FindAliveLightweightByIdResult, m.FindAliveLightweightByIdErr //unused in routes
}

func (m *mockJobsService) FindAliveByItineraryId(_ int64) (*[]models.ItineraryFileJob, error) {
	return m.FindByItineraryIdResult, m.FindByItineraryIdErr
}

func (m *mockJobsService) StopJob(_ *models.ItineraryFileJob) error {
	return m.StopJobErr
}

func (m *mockJobsService) FailJob(_ string, _ *models.ItineraryFileJob) error {
	return nil // unused in routes
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

func (m *mockJobsService) DeleteDeadJobs(_ int) error {
	return nil // unused in routes
}

func (m *mockJobsService) OpenItineraryJobFile(itineraryFileJob *models.ItineraryFileJob) (io.ReadSeekCloser, error) {
	return m.OpenItineraryJobFileResult, m.OpenItineraryJobFileErr
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

func Test_getItinerary_NotFound(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindByIdErr: sql.ErrNoRows}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getItinerary(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_getItinerary_UnexpectedError(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindByIdErr: errors.New("unexpected error")}
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
		return &mockItineraryService{FindByIdIt: &models.Itinerary{ID: 1, OwnerID: 2}}
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

func Test_runItineraryFileJob_ItineraryNotFound(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindByIdErr: sql.ErrNoRows}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	runItineraryFileJob(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_runItineraryFileJob_ItineraryUnexpectedError(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindByIdErr: errors.New("unexpected error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	runItineraryFileJob(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_runItineraryFileJob_ItineraryForbidden(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindByIdIt: &models.Itinerary{ID: 1, OwnerID: 2}}
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

func Test_runItineraryFileJob_InitAsyncTaskQueueClient_Error(t *testing.T) {
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
	services.NewAsyncqTaskQueue = func() (services.AsyncTaskQueueInterface, error) {
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
	services.NewAsyncqTaskQueue = func() (services.AsyncTaskQueueInterface, error) {
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
	services.NewAsyncqTaskQueue = func() (services.AsyncTaskQueueInterface, error) {
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
	services.NewAsyncqTaskQueue = func() (services.AsyncTaskQueueInterface, error) {
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

func Test_getAllItineraryFileJobs_ItineraryNotFound(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdErr: sql.ErrNoRows}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getAllItineraryFileJobs(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_getAllItineraryFileJobs_ItineraryUnexpectedError(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdErr: errors.New("unexpected error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getAllItineraryFileJobs(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_getAllItineraryFileJobs_ItineraryForbidden(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 2}}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	getAllItineraryFileJobs(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_getAllItineraryFileJobs_ItineraryFileJobsUnexpectedError(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{FindByItineraryIdResult: nil, FindByItineraryIdErr: errors.New("unexpected jobs error")}
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
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{OwnerID: 1}}
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

func Test_updateItinerary_ItineraryNotFound(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindByIdErr: sql.ErrNoRows}
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

func Test_updateItinerary_ItineraryUnexpectedError(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindByIdErr: errors.New("unexpected error")}
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
		return &mockItineraryService{FindByIdIt: &models.Itinerary{ID: 1, OwnerID: 2}}
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
			FindByIdIt:  &models.Itinerary{OwnerID: 1},
			ValidateErr: errors.New("validation error"),
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
		FindByIdIt: &models.Itinerary{OwnerID: 1},
		UpdateErr:  errors.New("update error"),
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
		FindByIdIt: &models.Itinerary{OwnerID: 1},
		UpdateErr:  nil,
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

func Test_deleteItinerary_ItineraryNotFound(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdErr: sql.ErrNoRows}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	deleteItinerary(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_deleteItinerary_FindItineraryUnexpectedError(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdErr: errors.New("unexpected error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	deleteItinerary(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_deleteItinerary_Forbidden(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 2}}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	deleteItinerary(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_deleteItinerary_Delete_Error(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	mock := &mockItineraryService{
		FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1},
		DeleteErr:             errors.New("delete error"),
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
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
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
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
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

func Test_getItineraryJob_ItineraryNotFound(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdErr: sql.ErrNoRows}
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

func Test_getItineraryJob_ItineraryUnexpectedError(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdErr: errors.New("unexpected error")}
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

func Test_getItineraryJob_ItineraryForbidden(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 2}}
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

func Test_getItineraryJob_ItineraryJobForbidden(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
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
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{OwnerID: 1}}
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

func Test_getItineraryJob_ItineraryJobNotFound(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{
			FindByIdResult: nil,
			FindByIdErr:    sql.ErrNoRows,
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

func Test_getItineraryJob_ItineraryJobUnexpectedError(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{
			FindByIdResult: nil,
			FindByIdErr:    errors.New("unexpected error"),
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
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_getItineraryJob_Success(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
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

func Test_deleteItineraryJob_ItineraryNotFound(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdErr: sql.ErrNoRows}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	deleteItineraryJob(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_deleteItineraryJob_ItineraryUnexpectedError(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdErr: errors.New("unexpected error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	deleteItineraryJob(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_deleteItineraryJob_ItineraryForbidden(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 2}}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	deleteItineraryJob(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_deleteItineraryJob_BadRequest_NoItineraryJobId(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
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
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
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

func Test_deleteItineraryJob_ItineraryJobNotFound(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{FindAliveLightweightByIdErr: sql.ErrNoRows}
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

func Test_deleteItineraryJob_ItineraryJobUnexpectedError(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{FindAliveLightweightByIdErr: errors.New("unexpected error")}
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

func Test_deleteItineraryJob_Forbidden_JobOwnership(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{FindAliveLightweightByIdResult: &models.ItineraryFileJob{ID: 2, ItineraryID: 99}}
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
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{
			FindAliveLightweightByIdResult: &models.ItineraryFileJob{ID: 2, ItineraryID: 1},
			SoftDeleteErr:                  errors.New("delete error"),
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
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{
			FindAliveLightweightByIdResult: &models.ItineraryFileJob{ID: 2, ItineraryID: 1},
			SoftDeleteErr:                  nil,
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
func Test_stopItineraryJob_Unauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	stopItineraryJob(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_stopItineraryJob_BadRequest_NoItineraryId(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	stopItineraryJob(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_stopItineraryJob_BadRequest_InvalidItineraryId(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "abc"}}
	stopItineraryJob(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_stopItineraryJob_ItineraryNotFound(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdErr: sql.ErrNoRows}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	stopItineraryJob(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_stopItineraryJob_ItineraryUnexpectedError(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdErr: errors.New("unexpected error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	stopItineraryJob(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_stopItineraryJob_ItineraryForbidden(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 2}}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	stopItineraryJob(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_stopItineraryJob_BadRequest_NoItineraryJobId(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	stopItineraryJob(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_stopItineraryJob_BadRequest_InvalidItineraryJobId(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "abc"},
	}
	stopItineraryJob(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_stopItineraryJob_ItineraryJobNotFound(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{FindByIdErr: sql.ErrNoRows}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "2"},
	}
	stopItineraryJob(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_stopItineraryJob_ItineraryJobUnexpectedError(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{FindByIdErr: errors.New("unexpected error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "2"},
	}
	stopItineraryJob(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_stopItineraryJob_ItineraryJobForbidden(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
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
	stopItineraryJob(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_stopItineraryJob_StopJob_Error(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{
			FindByIdResult:       &models.ItineraryFileJob{ID: 2, ItineraryID: 1},
			FindByItineraryIdErr: nil,
			StopJobErr:           errors.New("stop error"),
		}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "2"},
	}
	stopItineraryJob(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_stopItineraryJob_Success(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{
			FindByIdResult:       &models.ItineraryFileJob{ID: 2, ItineraryID: 1},
			FindByItineraryIdErr: nil,
			StopJobErr:           nil,
		}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "2"},
	}
	stopItineraryJob(c)
	assert.Equal(t, http.StatusOK, w.Code)
}
func Test_downloadItineraryJobFile_Unauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	downloadItineraryJobFile(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_downloadItineraryJobFile_BadRequest_NoItineraryId(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	downloadItineraryJobFile(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_downloadItineraryJobFile_BadRequest_InvalidItineraryId(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "abc"}}
	downloadItineraryJobFile(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_downloadItineraryJobFile_ItineraryNotFound(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdErr: sql.ErrNoRows}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	downloadItineraryJobFile(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_downloadItineraryJobFile_ItineraryUnexpectedError(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdErr: errors.New("unexpected error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	downloadItineraryJobFile(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_downloadItineraryJobFile_ItineraryForbidden(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 2}}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	downloadItineraryJobFile(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_downloadItineraryJobFile_BadRequest_NoItineraryJobId(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{{Key: "itineraryId", Value: "1"}}
	downloadItineraryJobFile(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_downloadItineraryJobFile_BadRequest_InvalidItineraryJobId(t *testing.T) {
	orig := services.GetItineraryService
	defer func() { services.GetItineraryService = orig }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "abc"},
	}
	downloadItineraryJobFile(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_downloadItineraryJobFile_ItineraryJobNotFound(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{FindByIdErr: sql.ErrNoRows}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "2"},
	}
	downloadItineraryJobFile(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_downloadItineraryJobFile_ItineraryJobUnexpectedError(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{FindByIdErr: errors.New("unexpected error")}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "2"},
	}
	downloadItineraryJobFile(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_downloadItineraryJobFile_Forbidden_JobOwnership(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
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
	downloadItineraryJobFile(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_downloadItineraryJobFile_JobFilePathEmpty(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{FindByIdResult: &models.ItineraryFileJob{ID: 2, ItineraryID: 1, Filepath: ""}}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "2"},
	}
	downloadItineraryJobFile(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

type fakeReadSeekCloser struct {
	io.Reader
}

func (f *fakeReadSeekCloser) Close() error                                 { return nil }
func (f *fakeReadSeekCloser) Seek(offset int64, whence int) (int64, error) { return 0, nil }

func Test_downloadItineraryJobFile_OpenFile_Error(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{
			FindByIdResult:          &models.ItineraryFileJob{ID: 2, ItineraryID: 1, Filepath: "file.txt"},
			OpenItineraryJobFileErr: errors.New("open error"),
		}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "2"},
	}
	downloadItineraryJobFile(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_downloadItineraryJobFile_FileNotOsFile(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{
			FindByIdResult:             &models.ItineraryFileJob{ID: 2, ItineraryID: 1, Filepath: "file.txt"},
			OpenItineraryJobFileResult: &fakeReadSeekCloser{Reader: bytes.NewReader([]byte("data"))},
		}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "2"},
	}
	downloadItineraryJobFile(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

type statErrorFile struct {
	fakeReadSeekCloser
}

func (f *statErrorFile) Stat() (os.FileInfo, error) {
	return nil, errors.New("stat error")
}

func Test_downloadItineraryJobFile_FileStat_Error(t *testing.T) {
	origIt := services.GetItineraryService
	defer func() { services.GetItineraryService = origIt }()
	services.GetItineraryService = func() services.ItineraryServiceInterface {
		return &mockItineraryService{FindLightweightByIdIt: &models.Itinerary{ID: 1, OwnerID: 1}}
	}
	origJobs := services.GetItineraryFileJobService
	defer func() { services.GetItineraryFileJobService = origJobs }()
	badFile := &statErrorFile{fakeReadSeekCloser{Reader: bytes.NewReader([]byte("data"))}}
	services.GetItineraryFileJobService = func() services.ItineraryFileJobServiceInterface {
		return &mockJobsService{
			FindByIdResult:             &models.ItineraryFileJob{ID: 2, ItineraryID: 1, Filepath: "file.txt"},
			OpenItineraryJobFileResult: badFile,
		}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	setUserId(c, 1)
	c.Params = gin.Params{
		{Key: "itineraryId", Value: "1"},
		{Key: "itineraryJobId", Value: "2"},
	}
	downloadItineraryJobFile(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// Note: It is not possible to unit test the success case for downloadItineraryJobFile
// in a pure unit test, because http.ServeContent writes directly to the http.ResponseWriter
// and expects an *os.File for Stat(). Mocking *os.File is not feasible in Go.
// Integration tests with a real file are required for a true success case.
