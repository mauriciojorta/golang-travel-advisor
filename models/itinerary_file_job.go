package models

import (
	"database/sql"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"example.com/travel-advisor/apis"
	"example.com/travel-advisor/db"
	"example.com/travel-advisor/utils"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
)

type ItineraryFileJob struct {
	ID                int64  `json:"id"`
	Status            string `json:"status"`
	StatusDescription string `json:"statusDescription"`
	// Status can be "running", "completed", "failed", or "stopped"
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
	Filepath    string    `json:"filepath"`
	ItineraryID int64     `json:"itineraryId"`

	FindById                func() error                        `json:"-"`
	FindByItineraryId       func() (*[]ItineraryFileJob, error) `json:"-"`
	HasItineraryAJobRunning func() (bool, error)                `json:"-"`
	RunJob                  func(itinerary *Itinerary) error    `json:"-"`
	StopJob                 func() error                        `json:"-"`
	Delete                  func() error                        `json:"-"`
}

const itineraryPromptTemplate = `Create a detailed travel itinerary based on the following information:
Title: {{.title}}
Description: {{.description}}

Destinations:
{{range .travelDestinations}}
- Country: {{.country}}, City: {{.city}}, Arrival: {{.arrivalDate}}, Departure: {{.departureDate}}
{{end}}

Please provide a day-by-day plan, including recommendations for activities, local attractions, and travel tips for each destination. The plan should provide a schedule for each day, including morning, afternoon, and evening activities. The itinerary should be suitable for a traveler who enjoys cultural experiences, local cuisine, and sightseeing.`

var NewItineraryFileJob = func(itineraryId int64) *ItineraryFileJob {
	job := &ItineraryFileJob{
		ItineraryID: itineraryId,
		Status:      "pending", // Default status
	}
	// Set default implementations for FindById, FindByItineraryId, RunJob, StopJob, and Delete
	job.FindById = job.defaultFindById
	job.FindByItineraryId = job.defaultFindByItineraryId
	job.HasItineraryAJobRunning = job.defaultHasItineraryAJobRunning
	job.RunJob = job.defaultRunJob
	job.StopJob = job.defaultStopJob
	job.Delete = job.defaultDelete
	return job
}

func (ifj *ItineraryFileJob) defaultFindById() error {
	query := `SELECT id, status, status_description, start_date, end_date, file_path, itinerary_id
	FROM itinerary_file_jobs WHERE id = ?`
	row := db.DB.QueryRow(query, ifj.ID)

	var statusDescription sql.NullString
	var endDate sql.NullTime
	var filePath sql.NullString
	err := row.Scan(&ifj.ID, &ifj.Status, &statusDescription, &ifj.StartDate, &endDate, &filePath, &ifj.ItineraryID)
	if err != nil {
		return err
	}

	if statusDescription.Valid {
		ifj.StatusDescription = statusDescription.String
	} else {
		ifj.StatusDescription = ""
	}
	if endDate.Valid {
		ifj.EndDate = endDate.Time
	} else {
		ifj.EndDate = time.Time{}
	}
	if filePath.Valid {
		ifj.Filepath = filePath.String
	} else {
		ifj.Filepath = ""
	}

	return nil
}

func (ifj *ItineraryFileJob) defaultFindByItineraryId() (*[]ItineraryFileJob, error) {
	query := `SELECT id, status, status_description, start_date, end_date, file_path, itinerary_id
	FROM itinerary_file_jobs WHERE itinerary_id = ?`
	rows, err := db.DB.Query(query, ifj.ItineraryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobs []ItineraryFileJob
	for rows.Next() {
		var job ItineraryFileJob
		var statusDescription sql.NullString
		var endDate sql.NullTime
		var filePath sql.NullString
		err := rows.Scan(&job.ID, &job.Status, &statusDescription, &job.StartDate, &endDate, &filePath, &job.ItineraryID)

		if err != nil {
			return nil, err
		}
		if statusDescription.Valid {
			job.StatusDescription = statusDescription.String
		} else {
			job.StatusDescription = ""
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

		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &jobs, nil
}

func (ifj *ItineraryFileJob) defaultHasItineraryAJobRunning() (bool, error) {
	query := `SELECT COUNT(id) FROM itinerary_file_jobs WHERE itinerary_id = ? AND status = 'running'`
	row := db.DB.QueryRow(query, ifj.ItineraryID)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (ifj *ItineraryFileJob) defaultRunJob(itinerary *Itinerary) error {
	ifj.Status = "running"
	ifj.StartDate = time.Now()

	// Insert the job into the database
	query := `INSERT INTO itinerary_file_jobs (status, start_date, itinerary_id) VALUES (?, ?, ?)`
	res, err := db.DB.Exec(query, ifj.Status, ifj.StartDate, ifj.ItineraryID)
	if err == nil {
		id, err := res.LastInsertId()
		if err == nil {
			ifj.ID = id
		}
	} else {
		log.Error("Error inserting job into database:", err)
		return err
	}

	// Start job processing asynchronously
	go func(job *ItineraryFileJob) {

		// Generate the LLM messages for the itinerary
		prompt, err := buildItineraryLlmPrompt(itinerary)
		if err != nil {
			job.Status = "failed"
			job.EndDate = time.Now()
			job.StatusDescription = "Failed to build itinerary prompt: " + err.Error()
			// Update the job in the database
			updateQuery := `UPDATE itinerary_file_jobs SET status = ?, status_description = ?, end_date = ? WHERE id = ?`
			_, err = db.DB.Exec(updateQuery, job.Status, job.StatusDescription, job.EndDate, job.ID)
			if err != nil {
				log.Error("Error updating job status in database:", err)
			}
			return
		}

		// Call the LLM to generate the itinerary
		messages := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeSystem,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "You are a helpful expert and guide of international travel."},
				},
			},
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: *prompt},
				},
			},
		}

		response, err := apis.CallLlm(messages)
		if err != nil {
			job.Status = "failed"
			job.EndDate = time.Now()
			job.StatusDescription = "Failed to generate itinerary: " + err.Error()
			// Update the job in the database
			updateQuery := `UPDATE itinerary_file_jobs SET status = ?, status_description = ?, end_date = ? WHERE id = ?`
			_, err = db.DB.Exec(updateQuery, job.Status, job.StatusDescription, job.EndDate, job.ID)
			if err != nil {
				fmt.Println("Error updating job status in database:", err)
			}
			return
		}

		// Generate a unique filename using UUID
		uuidStr := uuid.New().String()
		job.Filepath = "files/users/" + fmt.Sprintf("%d", itinerary.OwnerID) +
			"/itineraries/" + fmt.Sprintf("%d", itinerary.ID) +
			"/" + uuidStr + ".txt"

		// Save the generated itinerary to the specified file
		// Write the LLM response to the specified file
		err = utils.WriteFile(job.Filepath, []byte(*response), 0644)
		if err != nil {
			job.Status = "failed"
			job.EndDate = time.Now()
			job.StatusDescription = "Failed to write itinerary to file: " + err.Error()
			// Update the job in the database
			updateQuery := `UPDATE itinerary_file_jobs SET status = ?, status_description = ?, end_date = ?, file_path = ? WHERE id = ?`
			db.DB.Exec(updateQuery, job.Status, job.StatusDescription, job.EndDate, job.Filepath, job.ID)
			return
		}

		job.Status = "completed"
		job.StatusDescription = "Itinerary generated successfully"
		job.EndDate = time.Now()

		// Update the job in the database
		updateQuery := `UPDATE itinerary_file_jobs SET status = ?, status_description = ?, end_date = ?, file_path = ? WHERE id = ?`
		_, err = db.DB.Exec(updateQuery, job.Status, job.StatusDescription, job.EndDate, job.Filepath, job.ID)
		if err != nil {
			fmt.Println("Error updating job status in database:", err)
		}
	}(ifj)

	return nil
}

func (ifj *ItineraryFileJob) defaultStopJob() error {
	// Simulate stopping the job
	prevJobStatus := ifj.Status
	if ifj.Status == "running" {
		ifj.Status = "stopped"
		ifj.StatusDescription = "Job stopped by user"
		ifj.EndDate = time.Now()
		// Update the row in the database
		query := `UPDATE itinerary_file_jobs SET status = ?, status_description = ?, end_date = ? WHERE id = ?`
		_, err := db.DB.Exec(query, ifj.Status, ifj.StatusDescription, ifj.EndDate, ifj.ID)
		if err != nil {
			ifj.Status = prevJobStatus // Revert status if update fails
		}
		return err
	}
	return nil // If the job is not running, nothing to stop
}

func (ifj *ItineraryFileJob) defaultDelete() error {
	query := `DELETE FROM itinerary_file_jobs WHERE id = ?`
	_, err := db.DB.Exec(query, ifj.ID)
	return err
}

var buildItineraryLlmPrompt = func(itinerary *Itinerary) (*string, error) {
	prompt := prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
		prompts.NewHumanMessagePromptTemplate(
			itineraryPromptTemplate,
			[]string{"title", "description", "travelStartDate", "travelEndDate", "ownerId", "travelDestinations"},
		),
	})

	// Prepare travelDestinations for the template
	var travelDestinations []map[string]any
	for _, dest := range itinerary.TravelDestinations {
		travelDestinations = append(travelDestinations, map[string]any{
			"country":       dest.Country,
			"city":          dest.City,
			"arrivalDate":   dest.ArrivalDate,
			"departureDate": dest.DepartureDate,
		})
	}

	// Build the input map
	inputMap := map[string]any{
		"title":              itinerary.Title,
		"description":        itinerary.Description,
		"ownerId":            itinerary.OwnerID,
		"travelDestinations": travelDestinations,
	}

	message, err := prompt.Format(inputMap)
	if err != nil {
		return nil, fmt.Errorf("failed to format itinerary prompt: %w", err)
	}

	log.Debugf("Generated itinerary prompt: %s", message)

	return &message, nil

}
