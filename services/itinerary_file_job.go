package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"example.com/travel-advisor/apis"
	"example.com/travel-advisor/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	log "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
)

type ItineraryFileJobServiceInterface interface {
	FindAliveById(id int64) (*models.ItineraryFileJob, error)
	FindAliveLightweightById(id int64) (*models.ItineraryFileJob, error)
	FindAliveByItineraryId(itineraryId int64) (*[]models.ItineraryFileJob, error)
	GetJobsRunningOfUserCount(userId int64) (int, error)
	PrepareJob(itinerary *models.Itinerary) (*ItineraryFileAsyncTaskPayload, error)
	AddAsyncTaskId(asyncTaskId string, itineraryFileJob *models.ItineraryFileJob) error
	FailJob(errorDescription string, itineraryFileJob *models.ItineraryFileJob) error
	StopJob(itineraryFileJob *models.ItineraryFileJob) error
	SoftDeleteJob(itineraryFileJob *models.ItineraryFileJob) error
	SoftDeleteJobsByItineraryId(itineraryId int64, tx *sql.Tx) error
	DeleteJob(itineraryFileJob *models.ItineraryFileJob) error
	DeleteDeadJobs(fetchLimit int) error
}

type ItineraryFileJobService struct{}

// singleton instance
var itineraryFileJobServiceInstance = &ItineraryFileJobService{}

// GetItineraryFileJobService returns the singleton instance of ItineraryFileJobService
var GetItineraryFileJobService = func() ItineraryFileJobServiceInterface {
	return itineraryFileJobServiceInstance
}

const (
	TypeItineraryFileGeneration = "itinerary_file_generation"
)

type ItineraryFileAsyncTaskPayload struct {
	Itinerary        *models.Itinerary        `json:"itinerary"`
	ItineraryFileJob *models.ItineraryFileJob `json:"itineraryFileJob"`
}

const itineraryPromptTemplate = `Create a detailed travel itinerary based on the following information:
Title: {{.title}}
Description: {{.description}}
{{if .notes}}Notes: {{.notes}}
{{end}}
Destinations:
{{range .travelDestinations}}
- Country: {{.country}}, City: {{.city}}, Arrival: {{.arrivalDate}}, Departure: {{.departureDate}}
{{end}}

Please provide a day-by-day plan, including recommendations for activities, local attractions, and travel tips for each destination. The plan should provide a schedule for each day, including morning, afternoon, and evening activities. The itinerary should be suitable for a traveler who enjoys cultural experiences, local cuisine, and sightseeing.`

// FindAliveById retrieves the job by its ID
func (ifjs *ItineraryFileJobService) FindAliveById(id int64) (*models.ItineraryFileJob, error) {
	if id <= 0 {
		return nil, errors.New("invalid job ID")
	}
	job := models.NewItineraryFileJob(0) // Create a new ItineraryFileJob instance
	job.ID = id
	err := job.FindAliveById()
	if err != nil {
		log.Errorf("failed to find job by ID: %v", err)
		return nil, errors.New("failed to find job by ID")
	}
	return job, nil
}

// FindAliveLightweightById retrieves the job by its ID as a object containing only the ID and itinerary ID (entity primary and foreign keys)
func (ifjs *ItineraryFileJobService) FindAliveLightweightById(id int64) (*models.ItineraryFileJob, error) {
	if id <= 0 {
		return nil, errors.New("invalid job ID")
	}
	job := models.NewItineraryFileJob(0) // Create a new ItineraryFileJob instance
	job.ID = id
	err := job.FindAliveLightweightById()
	if err != nil {
		log.Errorf("failed to find job by ID: %v", err)
		return nil, errors.New("failed to find job by ID")
	}
	return job, nil
}

// FindAliveByItineraryId retrieves jobs by itinerary ID
func (ifjs *ItineraryFileJobService) FindAliveByItineraryId(itineraryId int64) (*[]models.ItineraryFileJob, error) {
	if itineraryId <= 0 {
		return nil, errors.New("invalid itinerary ID")
	}
	job := models.NewItineraryFileJob(itineraryId)
	return job.FindAliveByItineraryId()
}

// GetJobsRunningOfUserCount retrieves the count of running jobs for a user
func (ifjs *ItineraryFileJobService) GetJobsRunningOfUserCount(userId int64) (int, error) {
	if userId <= 0 {
		return 0, errors.New("invalid user ID")
	}
	job := models.NewItineraryFileJob(0) // Create a new ItineraryFileJob instance
	return job.GetJobsRunningOfUserCount(userId)
}

// PrepareJob prepares the job for execution
func (ifjs *ItineraryFileJobService) PrepareJob(itinerary *models.Itinerary) (*ItineraryFileAsyncTaskPayload, error) {
	if itinerary == nil {
		log.Error("itinerary instance is nil")
		return nil, errors.New("itinerary instance is nil")
	}

	job := models.NewItineraryFileJob(itinerary.ID)
	err := job.PrepareJob(itinerary)
	if err != nil {
		log.Errorf("failed to prepare job: %v", err)
		return nil, errors.New("failed to prepare job")
	}

	payload := &ItineraryFileAsyncTaskPayload{
		Itinerary:        itinerary,
		ItineraryFileJob: job,
	}

	return payload, nil

}

// AddAsyncTaskId adds an async task ID to the job
func (ifjs *ItineraryFileJobService) AddAsyncTaskId(asyncTaskId string, itineraryFileJob *models.ItineraryFileJob) error {
	if asyncTaskId == "" {
		return errors.New("async task ID cannot be empty")
	}
	if itineraryFileJob == nil {
		return errors.New("itinerary file job instance is nil")
	}

	err := itineraryFileJob.AddAsyncTaskId(asyncTaskId)
	if err != nil {
		log.Errorf("failed to add async task ID: %v", err)
		return errors.New("failed to add async task ID")
	}

	return nil
}

// FailJob marks the job as failed with a description
func (ifjs *ItineraryFileJobService) FailJob(errorDescription string, itineraryFileJob *models.ItineraryFileJob) error {
	if itineraryFileJob == nil {
		return errors.New("itinerary file job instance is nil")
	}
	if errorDescription == "" {
		return errors.New("error description cannot be empty")
	}

	err := itineraryFileJob.FailJob(errorDescription)
	if err != nil {
		log.Errorf("failed to fail job: %v", err)
		return errors.New("failed to fail job")
	}

	return nil
}

// StopJob stops the job
func (ifjs *ItineraryFileJobService) StopJob(itineraryFileJob *models.ItineraryFileJob) error {
	if itineraryFileJob == nil {
		return errors.New("itinerary file job instance is nil")
	}

	if itineraryFileJob.Status == "completed" {
		return errors.New("itinerary file job instance is already completed. It cannot be stopped")
	}

	asyncTaskTimeoutStr := os.Getenv("ASYNC_TASK_TIMEOUT_MINUTES")
	var asyncTaskTimeoutMinutes int
	if asyncTaskTimeoutStr != "" {
		var err error
		asyncTaskTimeoutMinutes, err = strconv.Atoi(asyncTaskTimeoutStr)
		if err != nil {
			return err
		}
	} else {
		asyncTaskTimeoutMinutes = 10 // default timeout in minutes if not set
	}

	if time.Now().Before(itineraryFileJob.CreationDate.Add(time.Duration(asyncTaskTimeoutMinutes) * time.Minute)) {
		return errors.New("itinerary file job instance is still within the timeout period and cannot be stopped")
	}

	err := itineraryFileJob.StopJob()
	if err != nil {
		log.Errorf("failed to stop job: %v", err)
		return errors.New("failed to stop job")
	}

	return nil
}

// SoftDeleteJob marks the job as deleted without removing it from the database
func (ifjs *ItineraryFileJobService) SoftDeleteJob(itineraryFileJob *models.ItineraryFileJob) error {
	if itineraryFileJob == nil {
		return errors.New("itinerary file job instance is nil")
	}
	err := itineraryFileJob.SoftDeleteJob()
	if err != nil {
		return errors.New("failed to soft delete job")
	}
	return nil
}

// SoftDeleteJobByItineraryId soft deletes jobs by itinerary ID
func (ifjs *ItineraryFileJobService) SoftDeleteJobsByItineraryId(itineraryId int64, tx *sql.Tx) error {
	if itineraryId <= 0 {
		return errors.New("invalid itinerary ID")
	}

	if tx == nil {
		return errors.New("transaction instance is nil")
	}

	job := models.NewItineraryFileJob(itineraryId)
	err := job.SoftDeleteJobsByItineraryIdTx(tx)
	if err != nil {
		return errors.New("failed to soft delete job by itinerary ID")
	}
	return nil

}

// Deletes a single job
func (ifjs *ItineraryFileJobService) DeleteJob(itineraryFileJob *models.ItineraryFileJob) error {
	if itineraryFileJob == nil {
		return errors.New("itinerary file job instance is nil")
	}

	if itineraryFileJob.Status != "deleted" {
		return errors.New("itinerary file job cannot be deleted because it is not marked for full deletion")
	}

	// Delete file of job before deleting it from the database
	fileManager := GetFileManager(itineraryFileJob.FileManager)
	err := fileManager.DeleteFile(itineraryFileJob.Filepath)
	if err != nil {
		log.Warnf("Error deleting file for itinerary job %v", itineraryFileJob.ID)
	}

	// Delete job from database
	err = itineraryFileJob.DeleteJob()
	if err != nil {
		return errors.New("failed to delete job")
	}

	return nil
}

// Fully deletes (job file + DB jobs table row removal) a "dead" (in 'deleted' status) jobs from the system.
// It only deletes the first 'n' jobs it finds on the DB based on the fetchLimit argument value (10 by default if the value is equal or less than 0)
func (ifjs *ItineraryFileJobService) DeleteDeadJobs(fetchLimit int) error {
	finalFetchLimit := 10 // Default fetch limit
	if fetchLimit > 0 {
		finalFetchLimit = fetchLimit // If the passed fetchLimit is greater than 0, the finalFetchLimit is updated with its value
	}

	job := models.NewItineraryFileJob(-1)

	deadJobs, err := job.FindDead(finalFetchLimit)
	if err != nil {
		log.Error("failed to find dead jobs", err)
		return errors.New("failed to find dead jobs")
	}
	if len(*deadJobs) == 0 {
		log.Info("No dead jobs found to delete")
		return nil
	}
	for _, deadJob := range *deadJobs {
		log.Infof("Deleting dead job with ID: %d", deadJob.ID)

		if deadJob.Filepath != "" {
			// Delete the file associated with the job
			fileManager := GetFileManager(deadJob.FileManager)
			err = fileManager.DeleteFile(deadJob.Filepath)
			if err != nil {
				log.Warnf("Error deleting file for dead job %v: %v", deadJob.ID, err)
			} else {
				log.Debugf("File %v of job with ID %d deleted successfully", deadJob.Filepath, deadJob.ID)
			}
		}

		// Delete the job from the database
		job.ID = deadJob.ID
		err = job.DeleteJob()
		if err != nil {
			log.Errorf("Error deleting dead job %v from database: %v", deadJob.ID, err)
		} else {
			log.Debugf("Job with ID %d deleted successfully", deadJob.ID)
		}
	}

	return nil

}

func HandleItineraryFileJob(ctx context.Context, t *asynq.Task) error {
	var itineraryFileJobTask ItineraryFileAsyncTaskPayload
	if err := json.Unmarshal(t.Payload(), &itineraryFileJobTask); err != nil {
		log.Errorf("could not unmarshal task payload: %v", err)
		return errors.New("could not unmarshal task payload")
	}

	itinerary := itineraryFileJobTask.Itinerary

	// We regenerate the job from the itinerary ID to have access to the entity methods
	job := models.NewItineraryFileJob(itineraryFileJobTask.ItineraryFileJob.ItineraryID)
	job.ID = itineraryFileJobTask.ItineraryFileJob.ID
	job.Status = itineraryFileJobTask.ItineraryFileJob.Status
	job.StatusDescription = itineraryFileJobTask.ItineraryFileJob.StatusDescription
	job.CreationDate = itineraryFileJobTask.ItineraryFileJob.CreationDate
	job.FileManager = itineraryFileJobTask.ItineraryFileJob.FileManager

	err := job.StartJob()
	if err != nil {
		job.FailJob("Failed to start job: " + err.Error())
		return err
	}

	// Generate the LLM messages for the itinerary
	prompt, err := buildItineraryLlmPrompt(itinerary)
	if err != nil {
		job.FailJob("Failed to build itinerary prompt: " + err.Error())
		return err
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
		job.FailJob("Failed to generate itinerary: " + err.Error())
		return err
	}

	//Get file manager to save the file specified in the configuration settings
	fileManager := GetFileManager(job.FileManager)

	// Generate a unique filename using UUID
	uuidStr := uuid.New().String()
	job.Filepath = "files/users/" + fmt.Sprintf("%d", itinerary.OwnerID) +
		"/itineraries/" + fmt.Sprintf("%d", itinerary.ID) +
		"/" + uuidStr + ".txt"

	// Write the LLM response to the specified file path using the file manager set in configuration
	err = fileManager.SaveContentInFile(job.Filepath, response)
	if err != nil {
		job.FailJob("Failed to write itinerary to file: " + err.Error())
		return err
	}

	defer func(finalError *error, filepath string, fileManager FileManagerInterface) {
		if finalError != nil {
			deleteFileError := fileManager.DeleteFile(filepath)
			if deleteFileError != nil {
				log.Warnf("Error deleting file %v after unexpected failure processing job: %v", filepath, deleteFileError)
			}
		}
	}(&err, job.Filepath, fileManager)

	job.Status = "completed"
	job.StatusDescription = "Itinerary generated successfully"
	job.EndDate = time.Now()

	// Update the job in the database
	err = job.CompleteJob()
	if err != nil {
		job.FailJob("Failed to complete job: " + err.Error())
		return err
	}

	return nil
}

var buildItineraryLlmPrompt = func(itinerary *models.Itinerary) (*string, error) {
	prompt := prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
		prompts.NewHumanMessagePromptTemplate(
			itineraryPromptTemplate,
			[]string{"title", "description", "notes", "travelStartDate", "travelEndDate", "ownerId", "travelDestinations"},
		),
	})

	// Prepare travelDestinations for the template
	var travelDestinations []map[string]any
	for _, dest := range *itinerary.TravelDestinations {
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
		"notes":              itinerary.Notes,
		"ownerId":            itinerary.OwnerID,
		"travelDestinations": travelDestinations,
	}

	message, err := prompt.Format(inputMap)
	if err != nil {
		log.Errorf("failed to format itinerary prompt: %v", err)
		return nil, errors.New("failed to format itinerary prompt")
	}

	log.Debugf("Generated itinerary prompt: %s", message)

	return &message, nil

}
