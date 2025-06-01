package services

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

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
	ifj.FindById = func() error { return nil }
	ifj.FindByItineraryId = func() (*[]models.ItineraryFileJob, error) {
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
	job, err := svc.FindById(0)
	assert.Nil(t, job)
	assert.Error(t, err)
}

func TestItineraryFileJobFindById_FailFind(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.FindById = func() error { return errors.New("fail") }

	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob {
		return ifj
	}

	svc := &ItineraryFileJobService{}
	job, err := svc.FindById(1)
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
	job, err := svc.FindById(2)
	assert.NotNil(t, job)
	assert.NoError(t, err)
}

func TestItineraryFileJobFindByItineraryId_InvalidID(t *testing.T) {
	svc := &ItineraryFileJobService{}
	jobs, err := svc.FindByItineraryId(0)
	assert.Nil(t, jobs)
	assert.Error(t, err)
}

func TestItineraryFileJobFindByItineraryId_Success(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.FindByItineraryId = func() (*[]models.ItineraryFileJob, error) {
		arr := []models.ItineraryFileJob{*ifj}
		return &arr, nil
	}
	models.NewItineraryFileJob = func(itineraryId int64) *models.ItineraryFileJob {
		return ifj
	}

	svc := &ItineraryFileJobService{}
	jobs, err := svc.FindByItineraryId(1)
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

	ifj.StopJob = func() error {
		return nil // Simulate successful stopping of job
	}

	err := (&ItineraryFileJobService{}).StopJob(ifj)
	assert.NoError(t, err)
}

func TestItineraryFileJobDeleteJob_NilJob(t *testing.T) {
	svc := &ItineraryFileJobService{}
	err := svc.DeleteJob(nil)
	assert.Error(t, err)
}

func TestItineraryFileJobDeleteJob_Fail(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.DeleteJob = func() error {
		return errors.New("fail")
	}
	err := (&ItineraryFileJobService{}).DeleteJob(ifj)
	assert.Error(t, err)
}

func TestItineraryFileJobDeleteJob_Success(t *testing.T) {
	ifj := mockItineraryFileJob()
	ifj.DeleteJob = func() error {
		return nil // Simulate successful deletion of job
	}

	err := (&ItineraryFileJobService{}).DeleteJob(ifj)
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
