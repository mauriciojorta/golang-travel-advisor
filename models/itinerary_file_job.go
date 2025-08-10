package models

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"example.com/travel-advisor/db"
)

type ItineraryFileJob struct {
	ID                int64  `json:"id" example:"1"`
	Status            string `json:"status" example:"completed"`
	StatusDescription string `json:"statusDescription,omitempty" example:"Job completed successfully"`
	// Status can be "running", "completed", "failed", or "stopped"
	CreationDate time.Time `json:"creationDate" example:"2024-06-01T00:00:00Z"` // CreationDate is set when the job is created
	// CreationDate is set when the job is created
	StartDate   time.Time `json:"startDate" example:"2024-06-01T00:00:00Z"`                             // StartDate is set when the job starts
	EndDate     time.Time `json:"endDate,omitempty" example:"2024-06-01T00:01:00Z"`                     // EndDate is set when the job ends
	Filepath    string    `json:"filepath,omitempty" example:"/path/to/file.txt"`                       // Optional, used for file storage
	FileManager string    `json:"fileManager,omitempty" example:"local"`                                // Optional, used for file management
	ItineraryID int64     `json:"itineraryId" example:"123"`                                            // ItineraryID is the ID of the itinerary associated with this job
	AsyncTaskID string    `json:"asyncTaskId,omitempty" example:"e2467dd0-db8a-49db-a5cb-9474f8e63933"` // Optional, async task ID from task manager

	FindAliveById                     func(id int64) (*ItineraryFileJob, error)            `json:"-"`
	FindAliveLightweightById          func(id int64) (*ItineraryFileJob, error)            `json:"-"`
	FindAliveByItineraryId            func(itineraryId int64) ([]*ItineraryFileJob, error) `json:"-"`
	FindDead                          func(fetchLimit int) ([]*ItineraryFileJob, error)    `json:"-"`
	GetInProgressJobsOfUserCount      func(userId int64) (int, error)                      `json:"-"`
	GetInProgressJobsOfItineraryCount func(itineraryId int64) (int, error)                 `json:"-"`
	PrepareJob                        func(itinerary *Itinerary) error                     `json:"-"`
	AddAsyncTaskId                    func(asyncTaskId string) error                       `json:"-"` // Functions for job management
	StartJob                          func() error                                         `json:"-"`
	FailJob                           func(errorDescription string) error                  `json:"-"`
	StopJob                           func() error                                         `json:"-"`
	CompleteJob                       func() error                                         `json:"-"`
	DeleteJob                         func() error                                         `json:"-"`
	SoftDeleteJob                     func() error                                         `json:"-"`
	SoftDeleteJobsByItineraryIdTx     func(itineraryId int64, tx *sql.Tx) error            `json:"-"`
}

var InitItineraryFileJob = func() *ItineraryFileJob {
	return InitItineraryFileJobFunctions(&ItineraryFileJob{})
}

var InitItineraryFileJobFunctions = func(job *ItineraryFileJob) *ItineraryFileJob {
	// Set default SQL implementations for FindById, FindByItineraryId, RunJob, StopJob, and Delete. In the future there could be implementations for
	// other NoSQL DB systems like MongoDB
	job.FindAliveById = job.defaultFindAliveById
	job.FindAliveLightweightById = job.defaultFindAliveLightweightById
	job.FindAliveByItineraryId = job.defaultFindAliveByItineraryId
	job.FindDead = job.defaultFindDead
	job.GetInProgressJobsOfUserCount = job.defaultGetInProgressJobsOfUserCount
	job.GetInProgressJobsOfItineraryCount = job.defaultGetInProgressJobsOfItineraryCount
	job.PrepareJob = job.defaultPrepareJob
	job.StartJob = job.defaultStartJob
	job.AddAsyncTaskId = job.defaultAddAsyncTaskId
	job.FailJob = job.defaultFailJob
	job.StopJob = job.defaultStopJob
	job.CompleteJob = job.defaultCompleteJob
	job.DeleteJob = job.defaultDeleteJob
	job.SoftDeleteJob = job.defaultSoftDeleteJob
	job.SoftDeleteJobsByItineraryIdTx = job.defaultSoftDeleteJobsByItineraryId
	return job
}

var NewItineraryFileJob = func(itineraryId int64) *ItineraryFileJob {
	job := &ItineraryFileJob{
		ItineraryID: itineraryId,
	}

	return InitItineraryFileJobFunctions(job)

}

func (ifj *ItineraryFileJob) defaultFindAliveById(id int64) (*ItineraryFileJob, error) {
	query := `SELECT id, status, status_description, creation_date, start_date, end_date, file_path, file_manager, itinerary_id, async_task_id
	FROM itinerary_file_jobs WHERE id = ? AND status != 'deleted'`
	row := db.DB.QueryRow(query, id)

	itineraryFileJob := &ItineraryFileJob{}

	var statusDescription sql.NullString
	var startDate sql.NullTime
	var endDate sql.NullTime
	var filePath sql.NullString
	var fileManager sql.NullString
	var asyncTaskId sql.NullString
	err := row.Scan(&itineraryFileJob.ID, &itineraryFileJob.Status, &statusDescription, &itineraryFileJob.CreationDate, &startDate, &endDate, &filePath, &fileManager, &itineraryFileJob.ItineraryID, &asyncTaskId)
	if err != nil {
		return nil, err
	}

	if statusDescription.Valid {
		itineraryFileJob.StatusDescription = statusDescription.String
	} else {
		itineraryFileJob.StatusDescription = ""
	}
	if startDate.Valid {
		itineraryFileJob.StartDate = startDate.Time
	} else {
		itineraryFileJob.StartDate = time.Time{}
	}
	if endDate.Valid {
		itineraryFileJob.EndDate = endDate.Time
	} else {
		itineraryFileJob.EndDate = time.Time{}
	}
	if filePath.Valid {
		itineraryFileJob.Filepath = filePath.String
	} else {
		itineraryFileJob.Filepath = ""
	}
	if fileManager.Valid {
		itineraryFileJob.FileManager = fileManager.String
	} else {
		itineraryFileJob.Filepath = ""
	}
	if asyncTaskId.Valid {
		itineraryFileJob.AsyncTaskID = asyncTaskId.String
	} else {
		itineraryFileJob.AsyncTaskID = ""
	}

	return itineraryFileJob, nil
}

func (ifj *ItineraryFileJob) defaultFindAliveLightweightById(id int64) (*ItineraryFileJob, error) {
	query := `SELECT id, itinerary_id
	FROM itinerary_file_jobs WHERE id = ? AND status != 'deleted'`
	row := db.DB.QueryRow(query, id)

	itineraryFileJob := &ItineraryFileJob{}

	err := row.Scan(&itineraryFileJob.ID, &itineraryFileJob.ItineraryID)
	if err != nil {
		return nil, err
	}

	return itineraryFileJob, nil
}

func (ifj *ItineraryFileJob) defaultFindAliveByItineraryId(itineraryId int64) ([]*ItineraryFileJob, error) {
	query := `SELECT id, status, status_description, creation_date, start_date, end_date, file_path, file_manager, itinerary_id, async_task_id
	FROM itinerary_file_jobs WHERE itinerary_id = ? AND status != 'deleted'`
	rows, err := db.DB.Query(query, itineraryId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobs []*ItineraryFileJob
	for rows.Next() {
		var job ItineraryFileJob
		var statusDescription sql.NullString
		var startDate sql.NullTime
		var endDate sql.NullTime
		var filePath sql.NullString
		var fileManager sql.NullString
		var asyncTaskId sql.NullString
		err := rows.Scan(&job.ID, &job.Status, &statusDescription, &job.CreationDate, &startDate, &endDate, &filePath, &fileManager, &job.ItineraryID, &asyncTaskId)

		if err != nil {
			return nil, err
		}
		if statusDescription.Valid {
			job.StatusDescription = statusDescription.String
		} else {
			job.StatusDescription = ""
		}
		if startDate.Valid {
			job.StartDate = startDate.Time
		} else {
			job.StartDate = time.Time{}
		}
		if endDate.Valid {
			job.EndDate = endDate.Time
		} else {
			job.EndDate = time.Time{}
		}
		if filePath.Valid {
			job.Filepath = filePath.String
		} else {
			job.Filepath = ""
		}
		if fileManager.Valid {
			job.FileManager = fileManager.String
		} else {
			job.FileManager = ""
		}
		if asyncTaskId.Valid {
			job.AsyncTaskID = asyncTaskId.String
		} else {
			job.AsyncTaskID = ""
		}

		jobs = append(jobs, &job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (ifj *ItineraryFileJob) defaultFindDead(fetchLimit int) ([]*ItineraryFileJob, error) {
	query := `SELECT id, status, status_description, creation_date, start_date, end_date, file_path, file_manager, itinerary_id, async_task_id
	FROM itinerary_file_jobs WHERE status = 'deleted' ORDER BY creation_date ASC LIMIT ?`
	rows, err := db.DB.Query(query, fetchLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobs []*ItineraryFileJob
	for rows.Next() {
		var job ItineraryFileJob
		var statusDescription sql.NullString
		var startDate sql.NullTime
		var endDate sql.NullTime
		var filePath sql.NullString
		var fileManager sql.NullString
		var asyncTaskId sql.NullString
		err := rows.Scan(&job.ID, &job.Status, &statusDescription, &job.CreationDate, &startDate, &endDate, &filePath, &fileManager, &job.ItineraryID, &asyncTaskId)

		if err != nil {
			return nil, err
		}
		if statusDescription.Valid {
			job.StatusDescription = statusDescription.String
		} else {
			job.StatusDescription = ""
		}
		if startDate.Valid {
			job.StartDate = startDate.Time
		} else {
			job.StartDate = time.Time{}
		}
		if endDate.Valid {
			job.EndDate = endDate.Time
		} else {
			job.EndDate = time.Time{}
		}
		if filePath.Valid {
			job.Filepath = filePath.String
		} else {
			job.Filepath = ""
		}
		if fileManager.Valid {
			job.FileManager = fileManager.String
		} else {
			job.FileManager = ""
		}
		if asyncTaskId.Valid {
			job.AsyncTaskID = asyncTaskId.String
		} else {
			job.AsyncTaskID = ""
		}

		jobs = append(jobs, &job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (ifj *ItineraryFileJob) defaultGetInProgressJobsOfUserCount(userId int64) (int, error) {
	query := `SELECT COUNT(itinerary_file_jobs.id) FROM itinerary_file_jobs WHERE status IN ('pending','running') AND itinerary_id IN (SELECT itineraries.id FROM itineraries WHERE owner_id = ?)`
	row := db.DB.QueryRow(query, userId)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (ifj *ItineraryFileJob) defaultGetInProgressJobsOfItineraryCount(itineraryId int64) (int, error) {
	query := `SELECT COUNT(itinerary_file_jobs.id) FROM itinerary_file_jobs WHERE status IN ('pending','running') AND itinerary_id = ?`
	row := db.DB.QueryRow(query, itineraryId)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (ifj *ItineraryFileJob) defaultPrepareJob(itinerary *Itinerary) error {
	ifj.Status = "pending"
	ifj.StartDate = time.Now()

	filemanager := os.Getenv("FILE_MANAGER")
	if filemanager == "" {
		filemanager = "local" // Local file manager if not set
	}

	ifj.FileManager = filemanager

	// Insert the job into the database
	query := `INSERT INTO itinerary_file_jobs (status, creation_date, file_manager, itinerary_id) VALUES (?, ?, ?, ?)`
	res, err := db.DB.Exec(query, ifj.Status, time.Now(), ifj.FileManager, itinerary.ID)
	if err == nil {
		id, err := res.LastInsertId()
		if err == nil {
			ifj.ID = id
		}
	} else {
		return fmt.Errorf("could not insert job into database: %w", err)
	}

	return nil
}

func (ifj *ItineraryFileJob) defaultStartJob() error {
	ifj.Status = "running"
	ifj.StartDate = time.Now()
	// Update the job in the database
	query := `UPDATE itinerary_file_jobs SET status = ?, start_date = ? WHERE id = ?`
	_, err := db.DB.Exec(query, ifj.Status, ifj.StartDate, ifj.ID)
	if err != nil {
		log.Errorf("Error updating job status to 'running' in database: %v", err)
		return fmt.Errorf("failed to update job status to 'running' in database: %w", err)
	}
	return nil
}

func (ifj *ItineraryFileJob) defaultFailJob(errorDescription string) error {
	ifj.Status = "failed"
	ifj.StatusDescription = errorDescription
	ifj.EndDate = time.Now()

	// Update the job in the database
	query := `UPDATE itinerary_file_jobs SET status = ?, status_description = ?, end_date = ? WHERE id = ?`
	_, err := db.DB.Exec(query, ifj.Status, ifj.StatusDescription, ifj.EndDate, ifj.ID)
	if err != nil {
		log.Warnf("Error updating job status to 'failed' in database: %v", err)
		return fmt.Errorf("failed to update job status to 'failed' in database: %w", err)
	}
	return nil
}

func (ifj *ItineraryFileJob) defaultAddAsyncTaskId(asyncTaskId string) error {
	ifj.AsyncTaskID = asyncTaskId
	query := `UPDATE itinerary_file_jobs SET async_task_id = ? WHERE id = ?`
	_, err := db.DB.Exec(query, ifj.AsyncTaskID, ifj.ID)
	if err != nil {
		log.Errorf("Error updating async task ID in database: %v", err)
		return fmt.Errorf("failed to update async task ID in database: %w", err)
	}
	return nil
}

func (ifj *ItineraryFileJob) defaultStopJob() error {
	// Simulate stopping the job
	prevJobStatus := ifj.Status
	if ifj.Status == "running" || ifj.Status == "pending" {
		ifj.Status = "stopped"
		ifj.StatusDescription = "Job stopped by user"
		ifj.EndDate = time.Now()
		// Update the row in the database
		query := `UPDATE itinerary_file_jobs SET status = ?, status_description = ?, end_date = ? WHERE id = ?`
		_, err := db.DB.Exec(query, ifj.Status, ifj.StatusDescription, ifj.EndDate, ifj.ID)
		if err != nil {
			log.Errorf("Error updating job status to 'stopped' in database: %v", err)
			ifj.Status = prevJobStatus // Revert status if update fails
			return fmt.Errorf("failed to update job status to 'stopped' in database: %w", err)
		}
	}
	return nil // If the job is not running, nothing to stop
}

func (ifj *ItineraryFileJob) defaultCompleteJob() error {
	ifj.Status = "completed"
	ifj.StatusDescription = "Job completed successfully"
	ifj.EndDate = time.Now()

	// Update the job in the database
	query := `UPDATE itinerary_file_jobs SET status = ?, status_description = ?, file_path = ?, end_date = ? WHERE id = ?`
	_, err := db.DB.Exec(query, ifj.Status, ifj.StatusDescription, ifj.Filepath, ifj.EndDate, ifj.ID)
	if err != nil {
		log.Errorf("Error updating job status to 'completed' in database: %v", err)
		return fmt.Errorf("failed to update job status to 'completed' in database: %w", err)
	}
	return nil
}

func (ifj *ItineraryFileJob) defaultDeleteJob() error {
	query := `DELETE FROM itinerary_file_jobs WHERE id = ?`
	_, err := db.DB.Exec(query, ifj.ID)
	if err != nil {
		log.Errorf("Error deleting job from database: %v", err)
		return fmt.Errorf("failed to delete job from database: %w", err)
	}
	return err
}

func (ifj *ItineraryFileJob) defaultSoftDeleteJob() error {
	query := `UPDATE itinerary_file_jobs SET status = 'deleted' WHERE id = ?`
	_, err := db.DB.Exec(query, ifj.ID)
	if err != nil {
		log.Errorf("Error soft deleting job: %v", err)
		return fmt.Errorf("failed to soft delete job: %w", err)
	}
	return nil
}

func (ifj *ItineraryFileJob) defaultSoftDeleteJobsByItineraryId(itineraryId int64, tx *sql.Tx) error {
	query := `UPDATE itinerary_file_jobs SET status = 'deleted' WHERE itinerary_id = ?`
	_, err := tx.Exec(query, itineraryId)
	if err != nil {
		log.Errorf("Error soft deleting jobs by itinerary ID: %v", err)
		return fmt.Errorf("failed to soft delete jobs by itinerary ID: %w", err)
	}
	return nil
}
