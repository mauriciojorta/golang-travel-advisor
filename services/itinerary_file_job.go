package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	FindById(id int64) (*models.ItineraryFileJob, error)
	FindByItineraryId(itineraryId int64) (*[]models.ItineraryFileJob, error)
	GetJobsRunningOfUserCount(userId int64) (int, error)
	PrepareJob(itinerary *models.Itinerary) (*ItineraryFileAsyncTaskPayload, error)
	AddAsyncTaskId(asyncTaskId string, itineraryFileJob *models.ItineraryFileJob) error
	FailJob(errorDescription string, itineraryFileJob *models.ItineraryFileJob) error
	StopJob(itineraryFileJob *models.ItineraryFileJob) error
	DeleteJob(itineraryFileJob *models.ItineraryFileJob) error
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

// FindById retrieves the job by its ID
func (ifjs *ItineraryFileJobService) FindById(id int64) (*models.ItineraryFileJob, error) {
	if id <= 0 {
		return nil, errors.New("invalid job ID")
	}
	job := models.NewItineraryFileJob(0) // Create a new ItineraryFileJob instance
	job.ID = id
	err := job.FindById()
	if err != nil {
		return nil, fmt.Errorf("failed to find job by ID: %w", err)
	}
	return job, nil
}

// FindByItineraryId retrieves jobs by itinerary ID
func (ifjs *ItineraryFileJobService) FindByItineraryId(itineraryId int64) (*[]models.ItineraryFileJob, error) {
	if itineraryId <= 0 {
		return nil, errors.New("invalid itinerary ID")
	}
	job := models.NewItineraryFileJob(itineraryId)
	return job.FindByItineraryId()
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
		return nil, fmt.Errorf("failed to prepare job: %w", err)
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
		return fmt.Errorf("failed to add async task ID: %w", err)
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
		return fmt.Errorf("failed to fail job: %w", err)
	}

	return nil
}

// StopJob stops the job
func (ifjs *ItineraryFileJobService) StopJob(itineraryFileJob *models.ItineraryFileJob) error {
	if itineraryFileJob == nil {
		return errors.New("itinerary file job instance is nil")
	}

	err := itineraryFileJob.StopJob()
	if err != nil {
		return fmt.Errorf("failed to stop job: %w", err)
	}

	return nil
}

// Delete deletes the job
func (ifjs *ItineraryFileJobService) DeleteJob(itineraryFileJob *models.ItineraryFileJob) error {
	if itineraryFileJob == nil {
		return errors.New("itinerary file job instance is nil")
	}

	err := itineraryFileJob.DeleteJob()
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	return nil
}

func HandleItineraryFileJob(ctx context.Context, t *asynq.Task) error {
	var itineraryFileJobTask ItineraryFileAsyncTaskPayload
	if err := json.Unmarshal(t.Payload(), &itineraryFileJobTask); err != nil {
		return fmt.Errorf("could not unmarshal task payload: %w", err)
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
		"notes":              itinerary.Notes,
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
