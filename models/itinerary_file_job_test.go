package models

import (
	"os"
	"testing"
	"time"

	"example.com/travel-advisor/apis"
	"example.com/travel-advisor/db"
	"example.com/travel-advisor/utils"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
)

func TestFileJobDefaultFindByItineraryId(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itineraryID := int64(1)
	rows := sqlmock.NewRows([]string{"id", "status", "status_description", "start_date", "end_date", "file_path", "itinerary_id"}).
		AddRow(1, "completed", "Job OK", time.Now(), time.Now().Add(24*time.Hour), "/path/to/file1", itineraryID).
		AddRow(2, "running", "Job running", time.Now().Add(48*time.Hour), time.Now().Add(72*time.Hour), "/path/to/file2", itineraryID)

	mock.ExpectQuery("SELECT id, status, status_description, start_date, end_date, file_path, itinerary_id FROM itinerary_file_jobs WHERE itinerary_id = ?").
		WithArgs(itineraryID).
		WillReturnRows(rows)

	job := &ItineraryFileJob{ItineraryID: itineraryID}
	result, err := job.defaultFindByItineraryId()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, *result, 2)

	assert.Equal(t, int64(1), (*result)[0].ID)
	assert.Equal(t, "completed", (*result)[0].Status)
	assert.Equal(t, "Job OK", (*result)[0].StatusDescription)
	assert.Equal(t, "/path/to/file1", (*result)[0].Filepath)

	assert.Equal(t, int64(2), (*result)[1].ID)
	assert.Equal(t, "running", (*result)[1].Status)
	assert.Equal(t, "Job running", (*result)[1].StatusDescription)
	assert.Equal(t, "/path/to/file2", (*result)[1].Filepath)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestFileJobDefaultFindByItineraryIdError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itineraryID := int64(1)

	mock.ExpectQuery("SELECT id, status, status_description, start_date, end_date, file_path, itinerary_id FROM itinerary_file_jobs WHERE itinerary_id = ?").
		WithArgs(itineraryID).
		WillReturnError(sqlmock.ErrCancelled)

	job := &ItineraryFileJob{ItineraryID: itineraryID}
	result, err := job.defaultFindByItineraryId()

	assert.Error(t, err)
	assert.Nil(t, result)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestFileJobDefaultFindById(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	jobID := int64(1)
	row := sqlmock.NewRows([]string{"id", "status", "status_description", "start_date", "end_date", "file_path", "itinerary_id"}).
		AddRow(jobID, "completed", "Job OK", time.Now(), time.Now().Add(24*time.Hour), "/path/to/file", 1)

	mock.ExpectQuery("SELECT id, status, status_description, start_date, end_date, file_path, itinerary_id FROM itinerary_file_jobs WHERE id = ?").
		WithArgs(jobID).
		WillReturnRows(row)

	job := &ItineraryFileJob{ID: jobID}
	err = job.defaultFindById()

	assert.NoError(t, err)
	assert.Equal(t, jobID, job.ID)
	assert.Equal(t, "completed", job.Status)
	assert.Equal(t, "/path/to/file", job.Filepath)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestFileJobDefaultIdError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itineraryID := int64(1)

	mock.ExpectQuery("SELECT id, status, status_description, start_date, end_date, file_path, itinerary_id FROM itinerary_file_jobs WHERE itinerary_id = ?").
		WithArgs(itineraryID).
		WillReturnError(sqlmock.ErrCancelled)

	job := &ItineraryFileJob{ItineraryID: itineraryID}
	result, err := job.defaultFindByItineraryId()

	assert.Error(t, err)
	assert.Nil(t, result)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultHasItineraryAJobRunning_True(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itineraryID := int64(123)
	mock.ExpectQuery("SELECT COUNT\\(id\\) FROM itinerary_file_jobs WHERE itinerary_id = \\? AND status = 'running'").
		WithArgs(itineraryID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	job := &ItineraryFileJob{ItineraryID: itineraryID}
	running, err := job.defaultHasItineraryAJobRunning()

	assert.NoError(t, err)
	assert.True(t, running)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultHasItineraryAJobRunning_False(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itineraryID := int64(456)
	mock.ExpectQuery("SELECT COUNT\\(id\\) FROM itinerary_file_jobs WHERE itinerary_id = \\? AND status = 'running'").
		WithArgs(itineraryID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	job := &ItineraryFileJob{ItineraryID: itineraryID}
	running, err := job.defaultHasItineraryAJobRunning()

	assert.NoError(t, err)
	assert.False(t, running)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultHasItineraryAJobRunning_Error(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itineraryID := int64(789)
	mock.ExpectQuery("SELECT COUNT\\(id\\) FROM itinerary_file_jobs WHERE itinerary_id = \\? AND status = 'running'").
		WithArgs(itineraryID).
		WillReturnError(sqlmock.ErrCancelled)

	job := &ItineraryFileJob{ItineraryID: itineraryID}
	running, err := job.defaultHasItineraryAJobRunning()

	assert.Error(t, err)
	assert.False(t, running)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestFileJobDefaultStopJob(t *testing.T) {
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

func TestFileJobDefaultStopJobError(t *testing.T) {
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

func TestDefaultDelete(t *testing.T) {
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

	err = job.defaultDelete()

	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultDeleteError(t *testing.T) {
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

	err = job.defaultDelete()

	assert.Error(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestDefaultRunJob_Success(t *testing.T) {
	// Patch dependencies
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	// Patch apis.CallLlm and os.WriteFile
	origCallLlm := apis.CallLlm
	origWriteFile := utils.WriteFile
	defer func() {
		apis.CallLlm = origCallLlm
		utils.WriteFile = origWriteFile
	}()

	apis.CallLlm = func(_ []llms.MessageContent) (*string, error) {
		resp := "LLM itinerary"
		return &resp, nil
	}
	utils.WriteFile = func(name string, data []byte, perm os.FileMode) error {
		return nil
	}

	itinerary := &Itinerary{
		ID:          10,
		OwnerID:     20,
		Title:       "Trip",
		Description: "Desc",
		TravelDestinations: []ItineraryTravelDestination{
			{Country: "Country", City: "City", ArrivalDate: time.Now(), DepartureDate: time.Now().Add(2 * 24 * time.Hour)},
		},
	}

	job := &ItineraryFileJob{ItineraryID: itinerary.ID}

	// Expect INSERT
	mock.ExpectExec("INSERT INTO itinerary_file_jobs").
		WithArgs("running", sqlmock.AnyArg(), itinerary.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect UPDATEs (could be more, but at least the last one)
	mock.ExpectExec("UPDATE itinerary_file_jobs SET status = \\?, status_description = \\?, end_date = \\?, file_path = \\? WHERE id = \\?").
		WithArgs("completed", "Itinerary generated successfully", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Run
	job.defaultRunJob(itinerary)

	// Wait for goroutine to finish (not ideal, but works for this async test)
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, "completed", job.Status)
	assert.Equal(t, "Itinerary generated successfully", job.StatusDescription)
	assert.NotEmpty(t, job.Filepath)
	assert.Equal(t, int64(1), job.ID)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultRunJob_InsertError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{ID: 1, OwnerID: 2}
	job := &ItineraryFileJob{ItineraryID: itinerary.ID}

	mock.ExpectExec("INSERT INTO itinerary_file_jobs").
		WithArgs("running", sqlmock.AnyArg(), itinerary.ID).
		WillReturnError(sqlmock.ErrCancelled)

	err = job.defaultRunJob(itinerary)

	assert.Error(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultRunJob_PromptError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	origBuildPrompt := buildItineraryLlmPrompt
	defer func() { buildItineraryLlmPrompt = origBuildPrompt }()
	buildItineraryLlmPrompt = func(_ *Itinerary) (*string, error) {
		return nil, assert.AnError
	}

	itinerary := &Itinerary{ID: 1, OwnerID: 2}
	job := &ItineraryFileJob{ItineraryID: itinerary.ID}

	mock.ExpectExec("INSERT INTO itinerary_file_jobs").
		WithArgs("running", sqlmock.AnyArg(), itinerary.ID).
		WillReturnResult(sqlmock.NewResult(2, 1))

	mock.ExpectExec("UPDATE itinerary_file_jobs SET status = \\?, status_description = \\?, end_date = \\? WHERE id = \\?").
		WithArgs("failed", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(2)).
		WillReturnResult(sqlmock.NewResult(2, 1))

	err = job.defaultRunJob(itinerary)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, "failed", job.Status)
	assert.Contains(t, job.StatusDescription, "Failed to build itinerary prompt")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultRunJob_LlmError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	origCallLlm := apis.CallLlm
	defer func() { apis.CallLlm = origCallLlm }()
	apis.CallLlm = func(_ []llms.MessageContent) (*string, error) {
		return nil, assert.AnError
	}

	itinerary := &Itinerary{
		ID:          3,
		OwnerID:     4,
		Title:       "Trip",
		Description: "Desc",
		TravelDestinations: []ItineraryTravelDestination{
			{Country: "Country", City: "City", ArrivalDate: time.Now(), DepartureDate: time.Now().Add(2 * 24 * time.Hour)},
		},
	}
	job := &ItineraryFileJob{ItineraryID: itinerary.ID}

	mock.ExpectExec("INSERT INTO itinerary_file_jobs").
		WithArgs("running", sqlmock.AnyArg(), itinerary.ID).
		WillReturnResult(sqlmock.NewResult(3, 1))

	mock.ExpectExec("UPDATE itinerary_file_jobs SET status = \\?, status_description = \\?, end_date = \\? WHERE id = \\?").
		WithArgs("failed", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(3)).
		WillReturnResult(sqlmock.NewResult(3, 1))

	err = job.defaultRunJob(itinerary)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, "failed", job.Status)
	assert.Contains(t, job.StatusDescription, "Failed to generate itinerary")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDefaultRunJob_WriteFileError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	origCallLlm := apis.CallLlm
	origWriteFile := utils.WriteFile
	defer func() {
		apis.CallLlm = origCallLlm
		utils.WriteFile = origWriteFile
	}()

	apis.CallLlm = func(_ []llms.MessageContent) (*string, error) {
		resp := "LLM itinerary"
		return &resp, nil
	}
	utils.WriteFile = func(name string, data []byte, perm os.FileMode) error {
		return assert.AnError
	}

	itinerary := &Itinerary{
		ID:          5,
		OwnerID:     6,
		Title:       "Trip",
		Description: "Desc",
		TravelDestinations: []ItineraryTravelDestination{
			{Country: "Country", City: "City", ArrivalDate: time.Now(), DepartureDate: time.Now().Add(2 * 24 * time.Hour)},
		},
	}
	job := &ItineraryFileJob{ItineraryID: itinerary.ID}

	mock.ExpectExec("INSERT INTO itinerary_file_jobs").
		WithArgs("running", sqlmock.AnyArg(), itinerary.ID).
		WillReturnResult(sqlmock.NewResult(4, 1))

	mock.ExpectExec("UPDATE itinerary_file_jobs SET status = \\?, status_description = \\?, end_date = \\?, file_path = \\? WHERE id = \\?").
		WithArgs("failed", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), int64(4)).
		WillReturnResult(sqlmock.NewResult(4, 1))

	err = job.defaultRunJob(itinerary)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, "failed", job.Status)
	assert.Contains(t, job.StatusDescription, "Failed to write itinerary to file")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
