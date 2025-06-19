package models

import (
	"testing"
	"time"

	"example.com/travel-advisor/db"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestFileJobDefaultFindAliveByItineraryId_Success(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itineraryID := int64(1)
	asyncTaskId1 := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	asyncTaskId2 := "952057c1-ac50-4014-972e-28ab65242ed6"
	rows := sqlmock.NewRows([]string{"id", "status", "status_description", "creation_date", "start_date", "end_date", "file_path", "file_manager", "itinerary_id", "async_task_id"}).
		AddRow(1, "completed", "Job OK", time.Now(), time.Now().Add(1*time.Minute), time.Now().Add(24*time.Hour), "/path/to/file1", "local", itineraryID, asyncTaskId1).
		AddRow(2, "running", "Job running", time.Now().Add(48*time.Hour), time.Now().Add(49*time.Hour), time.Now().Add(72*time.Hour), "/path/to/file2", "local", itineraryID, asyncTaskId2)

	mock.ExpectQuery("SELECT id, status, status_description, creation_date, start_date, end_date, file_path, file_manager, itinerary_id, async_task_id FROM itinerary_file_jobs WHERE itinerary_id = \\? AND status != 'deleted'").
		WithArgs(itineraryID).
		WillReturnRows(rows)

	job := &ItineraryFileJob{ItineraryID: itineraryID}
	result, err := job.defaultFindAliveByItineraryId()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, *result, 2)

	assert.Equal(t, int64(1), (*result)[0].ID)
	assert.Equal(t, "completed", (*result)[0].Status)
	assert.Equal(t, "Job OK", (*result)[0].StatusDescription)
	assert.Equal(t, "/path/to/file1", (*result)[0].Filepath)
	assert.Equal(t, "local", (*result)[0].FileManager)
	assert.Equal(t, int64(1), (*result)[0].ItineraryID)
	assert.Equal(t, asyncTaskId1, (*result)[0].AsyncTaskID)

	assert.Equal(t, int64(2), (*result)[1].ID)
	assert.Equal(t, "running", (*result)[1].Status)
	assert.Equal(t, "Job running", (*result)[1].StatusDescription)
	assert.Equal(t, "/path/to/file2", (*result)[1].Filepath)
	assert.Equal(t, "local", (*result)[1].FileManager)
	assert.Equal(t, int64(1), (*result)[1].ItineraryID)
	assert.Equal(t, asyncTaskId2, (*result)[1].AsyncTaskID)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestFileJobDefaultFindAliveByItineraryId_Error(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itineraryID := int64(1)

	mock.ExpectQuery("SELECT id, status, status_description, creation_date, start_date, end_date, file_path, file_manager, itinerary_id, async_task_id FROM itinerary_file_jobs WHERE itinerary_id = \\? AND status != 'deleted'").
		WithArgs(itineraryID).
		WillReturnError(sqlmock.ErrCancelled)

	job := &ItineraryFileJob{ItineraryID: itineraryID}
	result, err := job.defaultFindAliveByItineraryId()

	assert.Error(t, err)
	assert.Nil(t, result)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestFileJobDefaultFindAliveById_Success(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	jobID := int64(1)
	asyncTaskId := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	row := sqlmock.NewRows([]string{"id", "status", "status_description", "creation_date", "start_date", "end_date", "file_path", "file_manager", "itinerary_id", "async_task_id"}).
		AddRow(jobID, "completed", "Job OK", time.Now(), time.Now().Add(1*time.Minute), time.Now().Add(24*time.Hour), "/path/to/file", "local", 1, asyncTaskId)

	mock.ExpectQuery("SELECT id, status, status_description, creation_date, start_date, end_date, file_path, file_manager, itinerary_id, async_task_id FROM itinerary_file_jobs WHERE id = \\? AND status != 'deleted'").
		WithArgs(jobID).
		WillReturnRows(row)

	job := &ItineraryFileJob{ID: jobID}
	err = job.defaultFindAliveById()

	assert.NoError(t, err)
	assert.Equal(t, jobID, job.ID)
	assert.Equal(t, "completed", job.Status)
	assert.Equal(t, "/path/to/file", job.Filepath)
	assert.Equal(t, "local", job.FileManager)
	assert.Equal(t, "Job OK", job.StatusDescription)
	assert.Equal(t, int64(1), job.ItineraryID)
	assert.Equal(t, asyncTaskId, job.AsyncTaskID)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestFileJobDefaultFindAliveById_Error(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itineraryID := int64(1)
	mock.ExpectQuery("SELECT id, status, status_description, creation_date, start_date, end_date, file_path, file_manager, itinerary_id, async_task_id FROM itinerary_file_jobs WHERE itinerary_id = \\? AND status != 'deleted'").
		WithArgs(itineraryID).
		WillReturnError(sqlmock.ErrCancelled)

	job := &ItineraryFileJob{ItineraryID: itineraryID}
	result, err := job.defaultFindAliveByItineraryId()

	assert.Error(t, err)
	assert.Nil(t, result)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultGetJobsRunningOfUserCount_Success(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	userId := int64(1)

	mock.ExpectQuery("SELECT COUNT\\(itinerary_file_jobs.id\\) FROM itinerary_file_jobs WHERE status = 'running' AND itinerary_id IN \\(SELECT itineraries.id FROM itineraries WHERE owner_id = \\?\\)").
		WithArgs(userId).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	job := &ItineraryFileJob{}
	count, err := job.defaultGetJobsRunningOfUserCount(userId)

	assert.NoError(t, err)
	assert.Equal(t, 2, count)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultGetJobsRunningOfUserCountZero_Success(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	userId := int64(1)

	mock.ExpectQuery("SELECT COUNT\\(itinerary_file_jobs.id\\) FROM itinerary_file_jobs WHERE status = 'running' AND itinerary_id IN \\(SELECT itineraries.id FROM itineraries WHERE owner_id = \\?\\)").
		WithArgs(userId).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	job := &ItineraryFileJob{}
	count, err := job.defaultGetJobsRunningOfUserCount(userId)

	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultGetJobsRunningOfUserCount_Error(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	userId := int64(1)

	mock.ExpectQuery("SELECT COUNT\\(itinerary_file_jobs.id\\) FROM itinerary_file_jobs WHERE status = 'running' AND itinerary_id IN \\(SELECT itineraries.id FROM itineraries WHERE owner_id = \\?\\)").
		WithArgs(userId).
		WillReturnError(sqlmock.ErrCancelled)

	job := &ItineraryFileJob{}
	count, err := job.defaultGetJobsRunningOfUserCount(userId)

	assert.Error(t, err)
	assert.Equal(t, 0, count)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestFileJobDefaultStopJob_Success(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	job := &ItineraryFileJob{
		ID:     1,
		Status: "running",
	}

	mock.ExpectExec(`UPDATE itinerary_file_jobs SET status = \?, status_description = \?, end_date = \? WHERE id = \?`).
		WithArgs("stopped", "Job stopped by user", sqlmock.AnyArg(), job.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = job.defaultStopJob()

	assert.NoError(t, err)
	assert.Equal(t, "stopped", job.Status)
	assert.Equal(t, "Job stopped by user", job.StatusDescription)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestFileJobDefaultStopJob_Error(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	job := &ItineraryFileJob{
		ID:     1,
		Status: "running",
	}

	mock.ExpectExec(`UPDATE itinerary_file_jobs SET status = \?, status_description = \?, end_date = \? WHERE id = \?`).
		WithArgs("stopped", "Job stopped by user", sqlmock.AnyArg(), job.ID).
		WillReturnError(sqlmock.ErrCancelled)

	err = job.defaultStopJob()

	assert.Error(t, err)
	assert.Equal(t, "running", job.Status)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultDeleteJob_Success(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	job := &ItineraryFileJob{
		ID: 1,
	}

	mock.ExpectExec(`DELETE FROM itinerary_file_jobs WHERE id = \?`).
		WithArgs(job.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = job.defaultDeleteJob()

	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultDeleteJob_Error(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	job := &ItineraryFileJob{
		ID: 1,
	}

	mock.ExpectExec(`DELETE FROM itinerary_file_jobs WHERE id = \?`).
		WithArgs(job.ID).
		WillReturnError(sqlmock.ErrCancelled)

	err = job.defaultDeleteJob()

	assert.Error(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestDefaultPrepareJob_SuccessDefaultFileManager(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{
		ID:          42,
		Title:       "Test Trip",
		Description: "A test trip",
		OwnerID:     7,
	}
	job := &ItineraryFileJob{
		ItineraryID: itinerary.ID,
	}

	mock.ExpectExec(`INSERT INTO itinerary_file_jobs \(status, creation_date, file_manager, itinerary_id\) VALUES \(\?, \?, \?, \?\)`).
		WithArgs("pending", sqlmock.AnyArg(), "local", itinerary.ID).
		WillReturnResult(sqlmock.NewResult(123, 1))

	err = job.defaultPrepareJob(itinerary)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), job.ID)
	assert.Equal(t, "pending", job.Status)
	assert.Equal(t, "local", job.FileManager)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultPrepareJob_SuccessEnvironmentVariableFileManager(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{
		ID:          42,
		Title:       "Test Trip",
		Description: "A test trip",
		OwnerID:     7,
	}
	job := &ItineraryFileJob{
		ItineraryID: itinerary.ID,
	}

	// Set the environment variable for file manager
	t.Setenv("FILE_MANAGER", "s3")

	mock.ExpectExec(`INSERT INTO itinerary_file_jobs \(status, creation_date, file_manager, itinerary_id\) VALUES \(\?, \?, \?, \?\)`).
		WithArgs("pending", sqlmock.AnyArg(), "s3", itinerary.ID).
		WillReturnResult(sqlmock.NewResult(123, 1))

	err = job.defaultPrepareJob(itinerary)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), job.ID)
	assert.Equal(t, "pending", job.Status)
	assert.Equal(t, "s3", job.FileManager)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultPrepareJob_InsertError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{
		ID:          42,
		Title:       "Test Trip",
		Description: "A test trip",
		OwnerID:     7,
	}
	job := &ItineraryFileJob{
		ItineraryID: itinerary.ID,
	}

	mock.ExpectExec(`INSERT INTO itinerary_file_jobs \(status, creation_date, file_manager, itinerary_id\) VALUES \(\?, \?, \?. \?\)`).
		WithArgs("pending", sqlmock.AnyArg(), "local", itinerary.ID).
		WillReturnError(sqlmock.ErrCancelled)

	err = job.defaultPrepareJob(itinerary)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not insert job into database")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultStartJob_Success(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	job := &ItineraryFileJob{
		ID: 1,
	}

	mock.ExpectExec(`UPDATE itinerary_file_jobs SET status = \?, start_date = \? WHERE id = \?`).
		WithArgs("running", sqlmock.AnyArg(), job.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = job.defaultStartJob()
	assert.NoError(t, err)
	assert.Equal(t, "running", job.Status)
	assert.False(t, job.StartDate.IsZero())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultStartJob_Error(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	job := &ItineraryFileJob{
		ID: 1,
	}

	mock.ExpectExec(`UPDATE itinerary_file_jobs SET status = \?, start_date = \? WHERE id = \?`).
		WithArgs("running", sqlmock.AnyArg(), job.ID).
		WillReturnError(sqlmock.ErrCancelled)

	err = job.defaultStartJob()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update job status in database")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultFailJob_Success(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	job := &ItineraryFileJob{
		ID: 1,
	}

	mock.ExpectExec(`UPDATE itinerary_file_jobs SET status = \?, status_description = \?, end_date = \? WHERE id = \?`).
		WithArgs("failed", "some error", sqlmock.AnyArg(), job.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = job.defaultFailJob("some error")
	assert.NoError(t, err)
	assert.Equal(t, "failed", job.Status)
	assert.Equal(t, "some error", job.StatusDescription)
	assert.False(t, job.EndDate.IsZero())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultFailJob_Error(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	job := &ItineraryFileJob{
		ID: 1,
	}

	mock.ExpectExec(`UPDATE itinerary_file_jobs SET status = \?, status_description = \?, end_date = \? WHERE id = \?`).
		WithArgs("failed", "fail reason", sqlmock.AnyArg(), job.ID).
		WillReturnError(sqlmock.ErrCancelled)

	err = job.defaultFailJob("fail reason")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update job status in database")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultAddAsyncTaskId_Success(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	job := &ItineraryFileJob{
		ID: 1,
	}

	asyncTaskId := "task-uuid"
	mock.ExpectExec(`UPDATE itinerary_file_jobs SET async_task_id = \? WHERE id = \?`).
		WithArgs(asyncTaskId, job.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = job.defaultAddAsyncTaskId(asyncTaskId)
	assert.NoError(t, err)
	assert.Equal(t, asyncTaskId, job.AsyncTaskID)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultAddAsyncTaskId_Error(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	job := &ItineraryFileJob{
		ID: 1,
	}

	asyncTaskId := "task-uuid"
	mock.ExpectExec(`UPDATE itinerary_file_jobs SET async_task_id = \? WHERE id = \?`).
		WithArgs(asyncTaskId, job.ID).
		WillReturnError(sqlmock.ErrCancelled)

	err = job.defaultAddAsyncTaskId(asyncTaskId)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update async task ID in database")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultCompleteJob_Success(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	job := &ItineraryFileJob{
		ID: 1,
	}

	mock.ExpectExec(`UPDATE itinerary_file_jobs SET status = \?, status_description = \?, file_path = \?, end_date = \? WHERE id = \?`).
		WithArgs("completed", "Job completed successfully", sqlmock.AnyArg(), sqlmock.AnyArg(), job.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = job.defaultCompleteJob()
	assert.NoError(t, err)
	assert.Equal(t, "completed", job.Status)
	assert.Equal(t, "Job completed successfully", job.StatusDescription)
	assert.False(t, job.EndDate.IsZero())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultCompleteJob_Error(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	job := &ItineraryFileJob{
		ID: 1,
	}

	mock.ExpectExec(`UPDATE itinerary_file_jobs SET status = \?, status_description = \?, file_path = \?, end_date = \? WHERE id = \?`).
		WithArgs("completed", "Job completed successfully", sqlmock.AnyArg(), sqlmock.AnyArg(), job.ID).
		WillReturnError(sqlmock.ErrCancelled)

	err = job.defaultCompleteJob()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update job status in database")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestDefaultFindAliveLightweightById_Success(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	jobID := int64(10)
	itineraryID := int64(20)

	row := sqlmock.NewRows([]string{"id", "itinerary_id"}).
		AddRow(jobID, itineraryID)

	mock.ExpectQuery("SELECT id, itinerary_id FROM itinerary_file_jobs WHERE id = \\? AND status != 'deleted'").
		WithArgs(jobID).
		WillReturnRows(row)

	job := &ItineraryFileJob{ID: jobID}
	err = job.defaultFindAliveLightweightById()

	assert.NoError(t, err)
	assert.Equal(t, jobID, job.ID)
	assert.Equal(t, itineraryID, job.ItineraryID)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultFindAliveLightweightById_Error(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	jobID := int64(10)

	mock.ExpectQuery("SELECT id, itinerary_id FROM itinerary_file_jobs WHERE id = \\? AND status != 'deleted'").
		WithArgs(jobID).
		WillReturnError(sqlmock.ErrCancelled)

	job := &ItineraryFileJob{ID: jobID}
	err = job.defaultFindAliveLightweightById()

	assert.Error(t, err)
	assert.Equal(t, jobID, job.ID) // ID should remain unchanged

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestDefaultSoftDeleteJob_Success(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	job := &ItineraryFileJob{
		ID: 42,
	}

	mock.ExpectExec(`UPDATE itinerary_file_jobs SET status = 'deleted' WHERE id = \?`).
		WithArgs(job.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = job.defaultSoftDeleteJob()
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultSoftDeleteJob_Error(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	job := &ItineraryFileJob{
		ID: 42,
	}

	mock.ExpectExec(`UPDATE itinerary_file_jobs SET status = 'deleted' WHERE id = \?`).
		WithArgs(job.ID).
		WillReturnError(sqlmock.ErrCancelled)

	err = job.defaultSoftDeleteJob()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to soft delete job")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultSoftDeleteJobsByItineraryId_Success(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	itineraryID := int64(99)
	job := &ItineraryFileJob{
		ItineraryID: itineraryID,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE itinerary_file_jobs SET status = 'deleted' WHERE itinerary_id = \?`).
		WithArgs(itineraryID).
		WillReturnResult(sqlmock.NewResult(1, 2))

	tx, err := dbMock.Begin()
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	assert.NoError(t, err)

	err = job.defaultSoftDeleteJobsByItineraryId(tx)
	assert.NoError(t, err)

	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultSoftDeleteJobsByItineraryId_Error(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	itineraryID := int64(99)
	job := &ItineraryFileJob{
		ItineraryID: itineraryID,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE itinerary_file_jobs SET status = 'deleted' WHERE itinerary_id = \?`).
		WithArgs(itineraryID).
		WillReturnError(sqlmock.ErrCancelled)

	tx, err := dbMock.Begin()
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	assert.NoError(t, err)

	err = job.defaultSoftDeleteJobsByItineraryId(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to soft delete jobs by itinerary ID")

	_ = tx.Rollback()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
