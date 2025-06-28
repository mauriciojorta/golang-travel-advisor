package services

import (
	"errors"
	"os"
	"testing"

	"example.com/travel-advisor/models"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

type MockAsynqClient struct {
	mock.Mock
}

func (m *MockAsynqClient) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	args := m.Called(task, opts)
	info, _ := args.Get(0).(*asynq.TaskInfo)
	return info, args.Error(1)
}

func (m *MockAsynqClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// --- Helper for injecting mock client ---

func newAsyncqTaskQueueWithMock(client AsyncQueueClientInteface) *AsyncqTaskQueue {
	return &AsyncqTaskQueue{Client: client}
}

// --- Tests ---

func TestNewAsyncqTaskQueue_MissingPassword(t *testing.T) {
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Unsetenv("REDIS_PASSWORD")
	defer os.Unsetenv("REDIS_ADDR")

	queue, err := NewAsyncqTaskQueue()
	assert.Nil(t, queue)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "REDIS_PASSWORD environment variable is not set")
}

func TestNewAsyncqTaskQueue_DefaultAddr(t *testing.T) {
	os.Unsetenv("REDIS_ADDR")
	os.Setenv("REDIS_PASSWORD", "testpass")
	defer os.Unsetenv("REDIS_PASSWORD")

	queue, err := NewAsyncqTaskQueue()
	assert.NoError(t, err)
	assert.NotNil(t, queue)
	queue.Close()
}

func TestAsyncqTaskQueue_Close_Success(t *testing.T) {
	mockClient := new(MockAsynqClient)
	mockClient.On("Close").Return(nil)
	queue := newAsyncqTaskQueueWithMock(mockClient)
	queue.Close()
	mockClient.AssertCalled(t, "Close")
}

func TestAsyncqTaskQueue_Close_Error(t *testing.T) {
	mockClient := new(MockAsynqClient)
	mockClient.On("Close").Return(errors.New("close error"))
	queue := newAsyncqTaskQueueWithMock(mockClient)
	queue.Close()
	mockClient.AssertCalled(t, "Close")
}

func TestAsyncqTaskQueue_EnqueueItineraryFileJob_Success(t *testing.T) {
	mockClient := new(MockAsynqClient)
	payload := ItineraryFileAsyncTaskPayload{&models.Itinerary{}, &models.ItineraryFileJob{}}
	os.Setenv("ASYNC_TASK_TIMEOUT_MINUTES", "1")
	defer os.Unsetenv("ASYNC_TASK_TIMEOUT_MINUTES")

	mockClient.On("Enqueue", mock.AnythingOfType("*asynq.Task"), mock.Anything).Return(&asynq.TaskInfo{ID: "taskid", Queue: "default"}, nil)

	queue := newAsyncqTaskQueueWithMock(mockClient)
	id, err := queue.EnqueueItineraryFileJob(payload)
	assert.NoError(t, err)
	assert.NotNil(t, id)
	assert.Equal(t, "taskid", *id)
}

func TestAsyncqTaskQueue_EnqueueItineraryFileJob_TimeoutParseError(t *testing.T) {
	mockClient := new(MockAsynqClient)
	payload := ItineraryFileAsyncTaskPayload{&models.Itinerary{}, &models.ItineraryFileJob{}}
	os.Setenv("ASYNC_TASK_TIMEOUT_MINUTES", "notanint")
	defer os.Unsetenv("ASYNC_TASK_TIMEOUT_MINUTES")

	queue := newAsyncqTaskQueueWithMock(mockClient)
	_, err := queue.EnqueueItineraryFileJob(payload)
	assert.Error(t, err)
}

func TestAsyncqTaskQueue_EnqueueItineraryFileJob_EnqueueError(t *testing.T) {
	mockClient := new(MockAsynqClient)
	payload := ItineraryFileAsyncTaskPayload{&models.Itinerary{}, &models.ItineraryFileJob{}}
	os.Unsetenv("ASYNC_TASK_TIMEOUT_MINUTES")

	mockClient.On("Enqueue", mock.AnythingOfType("*asynq.Task"), mock.Anything).Return(nil, errors.New("enqueue error"))

	queue := newAsyncqTaskQueueWithMock(mockClient)
	_, err := queue.EnqueueItineraryFileJob(payload)
	assert.Error(t, err)
}

//Note: it is not possible to unit test the json.Marshall error case for EnquqeItineraryFileJob.
// The json.Marshal function will not return an error for the given ItineraryFileAsyncTaskPayload struct, so this case is not testable.
