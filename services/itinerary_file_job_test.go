package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"example.com/travel-advisor/apis"
	"example.com/travel-advisor/models"
	"example.com/travel-advisor/utils"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
)

// --- Mocks ---

func mockItineraryFileJob() *models.ItineraryFileJob {
	ifj := models.NewItineraryFileJob(0)
	ifj.ID = 1
	ifj.ItineraryID = 2
	ifj.FindAliveById = func() error { return nil }
	ifj.FindAliveLightweightById = func() error {
		return nil
	}
	ifj.FindAliveByItineraryId = func() (*[]models.ItineraryFileJob, error) {
		arr := []models.ItineraryFileJob{*ifj}
		return &arr, nil
	}
	ifj.AddAsyncTaskId = func(id string) error { return nil }
	ifj.FailJob = func(desc string) error { return nil }
	ifj.StopJob = func() error { return nil }
	ifj.DeleteJob = func() error { return nil }
	ifj.CompleteJob = func() error { return nil }
	ifj.StartJob = func() error { return nil }
	ifj.PrepareJob = func(it *models.Itinerary) error {
		return nil
	}
	ifj.GetJobsRunningOfUserCount = func(userId int64) (int, error) {
		return 0, nil
	}
	return ifj
}

// --- Tests ---

func TestItineraryFileJobFindById_InvalidID(t *testing.T) {
	svc := &ItineraryFileJobService{}
	job, err := svc.FindAliveById(0)
	assert.Nil(t, job)
	assert.Error(t, err)
}

func TestItineraryFileJobFindById_FailFind(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.FindAliveById = func() error { return errors.New("fail") }

	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob {
		return ifj
	}

	svc := &ItineraryFileJobService{}
	job, err := svc.FindAliveById(1)
	assert.Nil(t, job)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find job by ID")
}

func TestItineraryFileJobFindById_Success(t *testing.T) {
	ifj := mockItineraryFileJob()

	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob {
		return ifj
	}

	svc := &ItineraryFileJobService{}
	job, err := svc.FindAliveById(2)
	assert.NotNil(t, job)
	assert.NoError(t, err)
}

func TestItineraryFileJobFindAliveLightweightById_InvalidID(t *testing.T) {
	svc := &ItineraryFileJobService{}
	job, err := svc.FindAliveLightweightById(0)
	assert.Nil(t, job)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid job ID")
}

func TestItineraryFileJobFindAliveLightweightById_FailFind(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.FindAliveLightweightById = func() error { return errors.New("fail") }

	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob {
		return ifj
	}

	svc := &ItineraryFileJobService{}
	job, err := svc.FindAliveLightweightById(1)
	assert.Nil(t, job)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find job by ID")
}

func TestItineraryFileJobFindAliveLightweightById_Success(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.FindAliveLightweightById = func() error { return nil }

	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob {
		return ifj
	}

	svc := &ItineraryFileJobService{}
	job, err := svc.FindAliveLightweightById(2)
	assert.NotNil(t, job)
	assert.NoError(t, err)
}

func TestItineraryFileJobFindByItineraryId_InvalidID(t *testing.T) {
	svc := &ItineraryFileJobService{}
	jobs, err := svc.FindAliveByItineraryId(0)
	assert.Nil(t, jobs)
	assert.Error(t, err)
}

func TestItineraryFileJobFindByItineraryId_Success(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.FindAliveByItineraryId = func() (*[]models.ItineraryFileJob, error) {
		arr := []models.ItineraryFileJob{*ifj}
		return &arr, nil
	}
	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob {
		return ifj
	}

	svc := &ItineraryFileJobService{}
	jobs, err := svc.FindAliveByItineraryId(1)
	assert.NoError(t, err)
	assert.NotNil(t, jobs)
	assert.Equal(t, int64(1), (*jobs)[0].ID)
}

func TestItineraryFileJobGetJobsRunningOfUserCount_InvalidUser(t *testing.T) {
	svc := &ItineraryFileJobService{}
	count, err := svc.GetJobsRunningOfUserCount(0)
	assert.Equal(t, 0, count)
	assert.Error(t, err)
}

func TestItineraryFileJobGetJobsRunningOfUserCount_Success(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.GetJobsRunningOfUserCount = func(userId int64) (int, error) {
		if userId == 5 {
			return 3, nil // Simulate 3 running jobs for user ID 5
		}
		return 0, nil // Default case
	}
	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob {
		return ifj
	}

	svc := &ItineraryFileJobService{}
	count, err := svc.GetJobsRunningOfUserCount(5)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestItineraryFileJobPrepareJob_NilItinerary(t *testing.T) {
	svc := &ItineraryFileJobService{}
	payload, err := svc.PrepareJob(nil)
	assert.Nil(t, payload)
	assert.Error(t, err)
}

func TestItineraryFileJobPrepareJob_PrepareJobFails(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.PrepareJob = func(it *models.Itinerary) error {
		return errors.New("prepare job failed")
	}
	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob {
		return ifj
	}

	svc := &ItineraryFileJobService{}
	it := &models.Itinerary{ID: 1}
	payload, err := svc.PrepareJob(it)
	assert.Nil(t, payload)
	assert.Error(t, err)
}

func TestItineraryFileJobPrepareJob_Success(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.PrepareJob = func(it *models.Itinerary) error {
		return nil // Simulate successful preparation
	}
	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob {
		return ifj
	}

	svc := &ItineraryFileJobService{}
	it := &models.Itinerary{ID: 2}
	payload, err := svc.PrepareJob(it)
	assert.NoError(t, err)
	assert.NotNil(t, payload)
	assert.Equal(t, it, payload.Itinerary)
}

func TestItineraryFileJobAddAsyncTaskId_EmptyTaskId(t *testing.T) {
	svc := &ItineraryFileJobService{}
	err := svc.AddAsyncTaskId("", &models.ItineraryFileJob{})
	assert.Error(t, err)
}

func TestItineraryFileJobAddAsyncTaskId_NilJob(t *testing.T) {
	svc := &ItineraryFileJobService{}
	err := svc.AddAsyncTaskId("taskid", nil)
	assert.Error(t, err)
}

func TestItineraryFileJobAddAsyncTaskId_Fail(t *testing.T) {
	ifj := mockItineraryFileJob()

	ifj.AddAsyncTaskId = func(id string) error {
		return errors.New("fail")
	}
	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob {
		return ifj
	}

	err := (&ItineraryFileJobService{}).AddAsyncTaskId("taskid", ifj)
	assert.Error(t, err)
}

func TestItineraryFileJobAddAsyncTaskId_Success(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.AddAsyncTaskId = func(id string) error {
		return nil // Simulate successful addition of async task ID
	}
	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob {
		return ifj
	}

	err := (&ItineraryFileJobService{}).AddAsyncTaskId("taskid", ifj)
	assert.NoError(t, err)
}

func TestItineraryFileJobFailJob_NilJob(t *testing.T) {
	svc := &ItineraryFileJobService{}
	err := svc.FailJob("desc", nil)
	assert.Error(t, err)
}

func TestItineraryFileJobFailJob_EmptyDesc(t *testing.T) {
	svc := &ItineraryFileJobService{}
	err := svc.FailJob("", &models.ItineraryFileJob{})
	assert.Error(t, err)
}

func TestItineraryFileJobFailJob_Fail(t *testing.T) {

	ifj := mockItineraryFileJob()
	ifj.FailJob = func(desc string) error {
		return errors.New("fail")
	}

	err := (&ItineraryFileJobService{}).FailJob("desc", ifj)
	assert.Error(t, err)
}

func TestItineraryFileJobFailJob_Success(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.FailJob = func(desc string) error {
		return nil // Simulate successful failure of job
	}

	err := (&ItineraryFileJobService{}).FailJob("desc", ifj)
	assert.NoError(t, err)
}

func TestItineraryFileJobStopJob_NilJob(t *testing.T) {
	svc := &ItineraryFileJobService{}
	err := svc.StopJob(nil)
	assert.Error(t, err)
}

func TestItineraryFileJobStopJob_JobCompleted(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.Status = "completed"
	defer func() {
		ifj.Status = ""
	}()

	err := (&ItineraryFileJobService{}).StopJob(ifj)
	assert.Error(t, err)
}

func TestItineraryFileJobStopJob_TimeoutNotReached(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.CreationDate = time.Now().Add(4 * -time.Minute) // Simulate job started 4 minutes ago (below default timeout of 10 minutes)
	defer func() {
		ifj.CreationDate = time.Now()
	}()

	err := (&ItineraryFileJobService{}).StopJob(ifj)
	assert.Error(t, err)
}

func TestItineraryFileJobStopJob_Fail(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.StopJob = func() error {
		return errors.New("fail")
	}

	err := (&ItineraryFileJobService{}).StopJob(ifj)
	assert.Error(t, err)
}

func TestItineraryFileJobStopJob_Success(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.CreationDate = time.Now().Add(11 * -time.Minute) // Simulate job started 11 minutes ago (above default timeout of 10 minutes)
	defer func() {
		ifj.CreationDate = time.Now()
	}()

	ifj.StopJob = func() error {
		return nil // Simulate successful stopping of job
	}

	err := (&ItineraryFileJobService{}).StopJob(ifj)
	assert.NoError(t, err)
}

func TestItineraryFileJobStopJob_Success_CustomTimeout(t *testing.T) {
	os.Setenv("ASYNC_TASK_TIMEOUT_MINUTES", "2")

	ifj := mockItineraryFileJob()
	ifj.CreationDate = time.Now().Add(4 * -time.Minute) // Simulate job started 4 minutes ago (above custom timeout of 2 minutes)
	defer func() {
		ifj.CreationDate = time.Now()
		defer os.Clearenv()
	}()

	ifj.StopJob = func() error {
		return nil // Simulate successful stopping of job
	}

	err := (&ItineraryFileJobService{}).StopJob(ifj)
	assert.NoError(t, err)
}

func TestHandleItineraryFileJob_UnmarshalError(t *testing.T) {
	// Invalid JSON payload
	task := asynq.NewTask("ItineraryFileJob", []byte("{invalid-json}"))

	err := HandleItineraryFileJob(nil, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not unmarshal task payload")
}

func TestHandleItineraryFileJob_StartJobFails(t *testing.T) {
	// Removed mocking of uuid.New as it cannot be reassigned.

	it := &models.Itinerary{ID: 1, OwnerID: 2}
	job := mockItineraryFileJob()
	job.StartJob = func() error { return errors.New("start job fail") }
	job.FailJob = func(desc string) error {
		assert.Contains(t, desc, "Failed to start job")
		return nil
	}

	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob {
		return job
	}

	payload := ItineraryFileAsyncTaskPayload{
		Itinerary:        it,
		ItineraryFileJob: job,
	}
	payloadBytes, _ := json.Marshal(payload)
	task := asynq.NewTask("ItineraryFileJob", payloadBytes)

	err := HandleItineraryFileJob(nil, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "start job fail")
}

func TestHandleItineraryFileJob_BuildPromptFails(t *testing.T) {
	it := &models.Itinerary{ID: 1, OwnerID: 2}
	job := mockItineraryFileJob()
	job.StartJob = func() error { return nil }
	job.FailJob = func(desc string) error {
		assert.Contains(t, desc, "Failed to build itinerary prompt")
		return nil
	}

	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob {
		return job
	}

	origBuildPrompt := buildItineraryLlmPrompt
	defer func() { buildItineraryLlmPrompt = origBuildPrompt }()
	buildItineraryLlmPrompt = func(it *models.Itinerary) (*string, error) {
		return nil, errors.New("prompt fail")
	}

	payload := ItineraryFileAsyncTaskPayload{
		Itinerary:        it,
		ItineraryFileJob: job,
	}
	payloadBytes, _ := json.Marshal(payload)
	task := asynq.NewTask("ItineraryFileJob", payloadBytes)

	err := HandleItineraryFileJob(nil, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prompt fail")
}

func TestHandleItineraryFileJob_LlmCallFails(t *testing.T) {
	it := &models.Itinerary{ID: 1, OwnerID: 2}
	job := mockItineraryFileJob()
	job.StartJob = func() error { return nil }
	job.FailJob = func(desc string) error {
		assert.Contains(t, desc, "Failed to generate itinerary")
		return nil
	}

	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob {
		return job
	}

	origBuildPrompt := buildItineraryLlmPrompt
	defer func() { buildItineraryLlmPrompt = origBuildPrompt }()
	prompt := "prompt"
	buildItineraryLlmPrompt = func(it *models.Itinerary) (*string, error) {
		return &prompt, nil
	}

	origCallLlm := apis.CallLlm
	defer func() { apis.CallLlm = origCallLlm }()
	apis.CallLlm = func(msgs []llms.MessageContent) (*string, error) {
		return nil, errors.New("llm fail")
	}

	payload := ItineraryFileAsyncTaskPayload{
		Itinerary:        it,
		ItineraryFileJob: job,
	}
	payloadBytes, _ := json.Marshal(payload)
	task := asynq.NewTask("ItineraryFileJob", payloadBytes)

	err := HandleItineraryFileJob(nil, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "llm fail")
}

func TestHandleItineraryFileJob_WriteFileFails(t *testing.T) {
	it := &models.Itinerary{ID: 1, OwnerID: 2}
	job := mockItineraryFileJob()
	job.StartJob = func() error { return nil }
	job.FailJob = func(desc string) error {
		assert.Contains(t, desc, "Failed to write itinerary to file")
		return nil
	}

	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob {
		return job
	}

	origBuildPrompt := buildItineraryLlmPrompt
	defer func() { buildItineraryLlmPrompt = origBuildPrompt }()
	prompt := "prompt"
	buildItineraryLlmPrompt = func(it *models.Itinerary) (*string, error) {
		return &prompt, nil
	}

	origCallLlm := apis.CallLlm
	defer func() { apis.CallLlm = origCallLlm }()
	resp := "llm response"
	apis.CallLlm = func(msgs []llms.MessageContent) (*string, error) {
		return &resp, nil
	}

	origWriteLocalFile := utils.WriteLocalFile
	defer func() { utils.WriteLocalFile = origWriteLocalFile }()
	utils.WriteLocalFile = func(path string, data []byte, perm os.FileMode) error {
		return errors.New("write fail")
	}

	payload := ItineraryFileAsyncTaskPayload{
		Itinerary:        it,
		ItineraryFileJob: job,
	}
	payloadBytes, _ := json.Marshal(payload)
	task := asynq.NewTask("ItineraryFileJob", payloadBytes)

	err := HandleItineraryFileJob(nil, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "write fail")
}

func TestHandleItineraryFileJob_CompleteJobFails(t *testing.T) {
	it := &models.Itinerary{ID: 1, OwnerID: 2}
	job := mockItineraryFileJob()
	job.StartJob = func() error { return nil }
	job.FailJob = func(desc string) error {
		assert.Contains(t, desc, "Failed to complete job")
		return nil
	}
	job.CompleteJob = func() error { return errors.New("complete fail") }

	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob {
		return job
	}

	origBuildPrompt := buildItineraryLlmPrompt
	defer func() { buildItineraryLlmPrompt = origBuildPrompt }()
	prompt := "prompt"
	buildItineraryLlmPrompt = func(it *models.Itinerary) (*string, error) {
		return &prompt, nil
	}

	origCallLlm := apis.CallLlm
	defer func() { apis.CallLlm = origCallLlm }()
	resp := "llm response"
	apis.CallLlm = func(msgs []llms.MessageContent) (*string, error) {
		return &resp, nil
	}

	origWriteLocalFile := utils.WriteLocalFile
	defer func() { utils.WriteLocalFile = origWriteLocalFile }()
	utils.WriteLocalFile = func(path string, data []byte, perm os.FileMode) error {
		return nil
	}
	utils.DeleteLocalFile = func(p string) error {
		return nil
	}

	payload := ItineraryFileAsyncTaskPayload{
		Itinerary:        it,
		ItineraryFileJob: job,
	}
	payloadBytes, _ := json.Marshal(payload)
	task := asynq.NewTask("ItineraryFileJob", payloadBytes)

	err := HandleItineraryFileJob(nil, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "complete fail")
}

func TestHandleItineraryFileJob_Success(t *testing.T) {
	it := &models.Itinerary{ID: 1, OwnerID: 2}
	job := mockItineraryFileJob()
	job.StartJob = func() error { return nil }
	job.FailJob = func(desc string) error {
		t.Errorf("FailJob should not be called on success")
		return nil
	}
	job.CompleteJob = func() error { return nil }

	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob {
		return job
	}

	origBuildPrompt := buildItineraryLlmPrompt
	defer func() { buildItineraryLlmPrompt = origBuildPrompt }()
	prompt := "prompt"
	buildItineraryLlmPrompt = func(it *models.Itinerary) (*string, error) {
		return &prompt, nil
	}

	origCallLlm := apis.CallLlm
	defer func() { apis.CallLlm = origCallLlm }()
	resp := "llm response"
	apis.CallLlm = func(msgs []llms.MessageContent) (*string, error) {
		return &resp, nil
	}

	origWriteLocalFile := utils.WriteLocalFile
	defer func() { utils.WriteLocalFile = origWriteLocalFile }()
	utils.WriteLocalFile = func(path string, data []byte, perm os.FileMode) error {
		return nil
	}

	payload := ItineraryFileAsyncTaskPayload{
		Itinerary:        it,
		ItineraryFileJob: job,
	}
	payloadBytes, _ := json.Marshal(payload)
	task := asynq.NewTask("ItineraryFileJob", payloadBytes)

	err := HandleItineraryFileJob(nil, task)
	assert.NoError(t, err)
}
func TestItineraryFileJobService_SoftDeleteJob_NilJob(t *testing.T) {
	svc := &ItineraryFileJobService{}
	err := svc.SoftDeleteJob(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "itinerary file job instance is nil")
}

func TestItineraryFileJobService_SoftDeleteJob_Fail(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.SoftDeleteJob = func() error { return errors.New("fail soft delete") }
	err := (&ItineraryFileJobService{}).SoftDeleteJob(ifj)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to soft delete job")
}

func TestItineraryFileJobService_SoftDeleteJob_Success(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.SoftDeleteJob = func() error { return nil }
	err := (&ItineraryFileJobService{}).SoftDeleteJob(ifj)
	assert.NoError(t, err)
}

func TestItineraryFileJobService_SoftDeleteJobsByItineraryId_InvalidID(t *testing.T) {
	svc := &ItineraryFileJobService{}
	err := svc.SoftDeleteJobsByItineraryId(0, &sql.Tx{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid itinerary ID")
}

func TestItineraryFileJobService_SoftDeleteJobsByItineraryId_NilTx(t *testing.T) {
	svc := &ItineraryFileJobService{}
	err := svc.SoftDeleteJobsByItineraryId(1, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction instance is nil")
}

func TestItineraryFileJobService_SoftDeleteJobsByItineraryId_Fail(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.SoftDeleteJobsByItineraryIdTx = func(tx *sql.Tx) error { return errors.New("fail soft delete by itinerary id") }
	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob { return ifj }
	err := (&ItineraryFileJobService{}).SoftDeleteJobsByItineraryId(1, &sql.Tx{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to soft delete job by itinerary ID")
}

func TestItineraryFileJobService_SoftDeleteJobsByItineraryId_Success(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.SoftDeleteJobsByItineraryIdTx = func(tx *sql.Tx) error { return nil }
	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob { return ifj }
	err := (&ItineraryFileJobService{}).SoftDeleteJobsByItineraryId(1, &sql.Tx{})
	assert.NoError(t, err)
}

type mockFileManager struct {
	deleteFileCalled bool
	deleteFileErr    error
	returnReader     io.ReadSeekCloser
	openFileErr      error
}

func (m *mockFileManager) DeleteFile(path string) error {
	m.deleteFileCalled = true
	return m.deleteFileErr
}

func (m *mockFileManager) SaveContentInFile(path string, content *string) error { return nil }

func (m *mockFileManager) OpenFile(path string) (io.ReadSeekCloser, error) {
	return m.returnReader, m.openFileErr
}
func TestItineraryFileJobService_DeleteJob_NilJob(t *testing.T) {
	svc := &ItineraryFileJobService{}
	err := svc.DeleteJob(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "itinerary file job instance is nil")
}

func TestItineraryFileJobService_DeleteJob_NotMarkedDeleted(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.Status = "completed"
	err := (&ItineraryFileJobService{}).DeleteJob(ifj)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "itinerary file job cannot be deleted because it is not marked for full deletion")
}

func TestItineraryFileJobService_DeleteJob_FileDeleteFails(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.Status = "deleted"
	ifj.Filepath = "some/path"
	mgr := &mockFileManager{deleteFileErr: errors.New("file delete fail")}
	GetFileManager = func(name string) FileManagerInterface { return mgr }
	ifj.DeleteJob = func() error { return nil }
	err := (&ItineraryFileJobService{}).DeleteJob(ifj)
	assert.NoError(t, err)
	assert.True(t, mgr.deleteFileCalled)
}

func TestItineraryFileJobService_DeleteJob_DeleteJobFails(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.Status = "deleted"
	ifj.Filepath = "some/path"
	mgr := &mockFileManager{}
	GetFileManager = func(name string) FileManagerInterface { return mgr }
	ifj.DeleteJob = func() error { return errors.New("delete job fail") }
	err := (&ItineraryFileJobService{}).DeleteJob(ifj)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete job")
	assert.True(t, mgr.deleteFileCalled)
}

func TestItineraryFileJobService_DeleteJob_Success(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.Status = "deleted"
	ifj.Filepath = "some/path"
	mgr := &mockFileManager{}
	GetFileManager = func(name string) FileManagerInterface { return mgr }
	ifj.DeleteJob = func() error { return nil }
	err := (&ItineraryFileJobService{}).DeleteJob(ifj)
	assert.NoError(t, err)
	assert.True(t, mgr.deleteFileCalled)
}

func TestItineraryFileJobService_DeleteDeadJobs_FindDeadFails(t *testing.T) {
	origNewItineraryFileJob := models.NewItineraryFileJob
	defer func() { models.NewItineraryFileJob = origNewItineraryFileJob }()

	mockJob := mockItineraryFileJob()
	mockJob.FindDead = func(limit int) (*[]models.ItineraryFileJob, error) {
		return nil, errors.New("find dead fail")
	}
	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob { return mockJob }

	svc := &ItineraryFileJobService{}
	err := svc.DeleteDeadJobs(5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find dead jobs")
}

func TestItineraryFileJobService_DeleteDeadJobs_NoDeadJobs(t *testing.T) {
	origNewItineraryFileJob := models.NewItineraryFileJob
	defer func() { models.NewItineraryFileJob = origNewItineraryFileJob }()

	mockJob := mockItineraryFileJob()
	mockJob.FindDead = func(limit int) (*[]models.ItineraryFileJob, error) {
		arr := []models.ItineraryFileJob{}
		return &arr, nil
	}
	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob { return mockJob }

	svc := &ItineraryFileJobService{}
	err := svc.DeleteDeadJobs(0)
	assert.NoError(t, err)
}

func TestItineraryFileJobService_DeleteDeadJobs_FileDeleteFails(t *testing.T) {
	origNewItineraryFileJob := models.NewItineraryFileJob
	defer func() { models.NewItineraryFileJob = origNewItineraryFileJob }()

	mockDeadJob := mockItineraryFileJob()
	mockDeadJob.ID = 123
	mockDeadJob.FileManager = "mock"
	mockDeadJob.Filepath = "dead/path"
	mockDeadJob.DeleteJob = func() error { return nil }

	mockJob := mockItineraryFileJob()
	mockJob.FindDead = func(limit int) (*[]models.ItineraryFileJob, error) {
		arr := []models.ItineraryFileJob{*mockDeadJob}
		return &arr, nil
	}
	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob { return mockJob }

	mgr := &mockFileManager{deleteFileErr: errors.New("file delete fail")}
	GetFileManager = func(name string) FileManagerInterface { return mgr }

	svc := &ItineraryFileJobService{}
	err := svc.DeleteDeadJobs(1)
	assert.NoError(t, err)
	assert.True(t, mgr.deleteFileCalled)
}

func TestItineraryFileJobService_DeleteDeadJobs_DeleteJobFails(t *testing.T) {
	origNewItineraryFileJob := models.NewItineraryFileJob
	defer func() { models.NewItineraryFileJob = origNewItineraryFileJob }()

	mockDeadJob := mockItineraryFileJob()
	mockDeadJob.ID = 456
	mockDeadJob.FileManager = "mock"
	mockDeadJob.Filepath = "dead/path"
	mockDeadJob.DeleteJob = func() error { return errors.New("delete job fail") }

	mockJob := mockItineraryFileJob()
	mockJob.FindDead = func(limit int) (*[]models.ItineraryFileJob, error) {
		arr := []models.ItineraryFileJob{*mockDeadJob}
		return &arr, nil
	}
	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob { return mockJob }

	mgr := &mockFileManager{}
	GetFileManager = func(name string) FileManagerInterface { return mgr }

	svc := &ItineraryFileJobService{}
	err := svc.DeleteDeadJobs(1)
	assert.NoError(t, err)
	assert.True(t, mgr.deleteFileCalled)
}

func TestItineraryFileJobService_DeleteDeadJobs_Success(t *testing.T) {
	origNewItineraryFileJob := models.NewItineraryFileJob
	defer func() { models.NewItineraryFileJob = origNewItineraryFileJob }()

	mockDeadJob := mockItineraryFileJob()
	mockDeadJob.ID = 789
	mockDeadJob.FileManager = "mock"
	mockDeadJob.Filepath = "dead/path"
	mockDeadJob.DeleteJob = func() error { return nil }

	mockJob := mockItineraryFileJob()
	mockJob.FindDead = func(limit int) (*[]models.ItineraryFileJob, error) {
		arr := []models.ItineraryFileJob{*mockDeadJob}
		return &arr, nil
	}
	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob { return mockJob }

	mgr := &mockFileManager{}
	GetFileManager = func(name string) FileManagerInterface { return mgr }

	svc := &ItineraryFileJobService{}
	err := svc.DeleteDeadJobs(1)
	assert.NoError(t, err)
	assert.True(t, mgr.deleteFileCalled)
}

func TestItineraryFileJobService_DeleteDeadJobs_AbsentFileSuccess(t *testing.T) {
	origNewItineraryFileJob := models.NewItineraryFileJob
	defer func() { models.NewItineraryFileJob = origNewItineraryFileJob }()

	mockDeadJob := mockItineraryFileJob()
	mockDeadJob.ID = 789
	mockDeadJob.FileManager = "mock"
	mockDeadJob.Filepath = ""
	mockDeadJob.DeleteJob = func() error { return nil }

	mockJob := mockItineraryFileJob()
	mockJob.FindDead = func(limit int) (*[]models.ItineraryFileJob, error) {
		arr := []models.ItineraryFileJob{*mockDeadJob}
		return &arr, nil
	}
	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob { return mockJob }

	mgr := &mockFileManager{}
	GetFileManager = func(name string) FileManagerInterface { return mgr }

	svc := &ItineraryFileJobService{}
	err := svc.DeleteDeadJobs(1)
	assert.NoError(t, err)
	assert.False(t, mgr.deleteFileCalled)
}
func TestItineraryFileJobService_OpenItineraryJobFile_NilJob(t *testing.T) {
	svc := &ItineraryFileJobService{}
	file, err := svc.OpenItineraryJobFile(nil)
	assert.Nil(t, file)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "itinerary file job instance is nil")
}

func TestItineraryFileJobService_OpenItineraryJobFile_EmptyFilepath(t *testing.T) {
	svc := &ItineraryFileJobService{}
	job := &models.ItineraryFileJob{Filepath: ""}
	file, err := svc.OpenItineraryJobFile(job)
	assert.Nil(t, file)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "itinerary file job filepath is empty")
}

type mockReadSeeker struct{}

func (m *mockReadSeeker) Read(p []byte) (n int, err error)             { return 0, io.EOF }
func (m *mockReadSeeker) Close() error                                 { return nil }
func (m *mockReadSeeker) Seek(offset int64, whence int) (int64, error) { return 0, nil }

func TestItineraryFileJobService_OpenItineraryJobFile_OpenFileFails(t *testing.T) {
	svc := &ItineraryFileJobService{}
	job := &models.ItineraryFileJob{Filepath: "some/path", FileManager: "mock"}
	mgr := &mockFileManager{openFileErr: errors.New("fail open")}
	GetFileManager = func(name string) FileManagerInterface { return mgr }

	file, err := svc.OpenItineraryJobFile(job)
	assert.Nil(t, file)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open itinerary job file")
}

func TestItineraryFileJobService_OpenItineraryJobFile_Success(t *testing.T) {
	svc := &ItineraryFileJobService{}
	job := &models.ItineraryFileJob{Filepath: "some/path", FileManager: "mock"}
	reader := &mockReadSeeker{}
	mgr := &mockFileManager{returnReader: reader}
	GetFileManager = func(name string) FileManagerInterface { return mgr }

	file, err := svc.OpenItineraryJobFile(job)
	assert.NoError(t, err)
	assert.Equal(t, reader, file)
}
