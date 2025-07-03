package routes

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"example.com/travel-advisor/models"
	"example.com/travel-advisor/services"
	"github.com/gin-gonic/gin"
)

type CreateItineraryRequest struct {
	Title        string                               `json:"title" binding:"required" example:"Trip to Spain"`
	Description  string                               `json:"description" example:"Summer vacation in Spain"`
	Notes        *string                              `json:"notes" example:"I want to enjoy the nightlife"`
	Destinations *[]models.ItineraryTravelDestination `json:"destinations" binding:"required,dive"`
}

type UpdateItineraryRequest struct {
	ID           int64                                `json:"id" binding:"required" example:"1"`
	Title        string                               `json:"title" binding:"required" example:"Trip to Spain"`
	Description  string                               `json:"description" example:"Summer vacation in Spain"`
	Notes        *string                              `json:"notes" example:"I want to enjoy the nightlife"`
	Destinations *[]models.ItineraryTravelDestination `json:"destinations" binding:"required,dive"`
}

type CreateItineraryResponse struct {
	Message     string `json:"message" example:"Itinerary created."`
	ItineraryID int64  `json:"itineraryId" example:"123"`
}

type GetItineraryResponse struct {
	Itinerary *models.Itinerary `json:"itinerary" example:"{\"id\": 1, \"title\": \"Trip to Spain\", \"description\": \"Summer vacation in Spain\", \"notes\": \"I want to enjoy the nightlife\", \"ownerId\": 42, \"travelDestinations\": [{\"id\": 1, \"country\": \"Spain\", \"city\": \"Madrid\", \"arrivalDate\": \"2024-07-01T00:00:00Z\", \"departureDate\": \"2024-07-05T00:00:00Z\"}]}"` // Example JSON representation
}

type GetItinerariesResponse struct {
	Itineraries *[]models.Itinerary `json:"itineraries" example:"[{\"id\": 1, \"title\": \"Trip to Spain\", \"description\": \"Summer vacation in Spain\", \"notes\": \"I want to enjoy the nightlife\", \"ownerId\": 42, \"travelDestinations\": [{\"id\": 1, \"country\": \"Spain\", \"city\": \"Madrid\", \"arrivalDate\": \"2024-07-01T00:00:00Z\", \"departureDate\": \"2024-07-05T00:00:00Z\"}]}]"` // Example JSON representation
}

type StartItineraryJobResponse struct {
	Message string `json:"message" example:"Job started successfully."`
	JobId   int64  `json:"jobId" example:"123"`
}

type StopItineraryJobResponse struct {
	Message string `json:"message" example:"Itinerary job stopped."`
}

type GetItineraryJobResponse struct {
	Job *models.ItineraryFileJob `json:"job" example:"{\"id\": 1, \"itineraryId\": 1, \"status\": \"running\", \"createdAt\": \"2024-07-01T00:00:00Z\", \"updatedAt\": \"2024-07-01T00:00:00Z\"}"` // Example JSON representation
}

type GetItineraryJobsResponse struct {
	Jobs *[]models.ItineraryFileJob `json:"job" example:"[{\"id\": 1, \"itineraryId\": 1, \"status\": \"running\", \"createdAt\": \"2024-07-01T00:00:00Z\", \"updatedAt\": \"2024-07-01T00:00:00Z\"}]"`
}

type UpdateItineraryResponse struct {
	Message string `json:"message" example:"Itinerary updated."`
}

type DeleteItineraryResponse struct {
	Message string `json:"message" example:"Itinerary deleted."`
}

type DeleteItineraryJobResponse struct {
	Message string `json:"message" example:"Itinerary job deleted."`
}

// createItinerary godoc
// @Summary      Create a new itinerary
// @Description  Creates a new itinerary for the authenticated user.
// @Tags         itineraries
// @Accept       json
// @Produce      json
// @Security     Auth
// @Param        itinerary  body  CreateItineraryRequest true  "Itinerary data"
// @Success      201  {object}  CreateItineraryResponse  "Itinerary created."  example({"message": "Itinerary created.", "itineraryId": 123})
// @Failure      400  {object}  ErrorResponse       "Could not parse request data or invalid destinations."
// @Failure      401  {object}  ErrorResponse      "Not authorized."
// @Failure      500  {object}  ErrorResponse      "Could not create itinerary. Try again later."
// @Router       /itineraries [post]
func createItinerary(context *gin.Context) {
	log.Debug("Creating itinerary")

	var input CreateItineraryRequest

	userId, exists := context.Get("userId")
	if !exists {
		context.JSON(http.StatusUnauthorized, &ErrorResponse{Message: "Not authorized."})
		return
	}

	// Bind JSON input to the input struct
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, &ErrorResponse{Message: "Could not parse request data."})
		return
	}

	itineraryService := services.GetItineraryService()

	itinerary := models.NewItinerary(input.Title, input.Description, input.Notes, input.Destinations)

	itinerary.OwnerID = userId.(int64)

	err := itineraryService.ValidateItineraryDestinationsDates(itinerary.TravelDestinations)
	if err != nil {
		log.Errorf("Error validating itinerary destinations dates: %v", err)
		context.JSON(http.StatusBadRequest, &ErrorResponse{Message: err.Error()})
		return
	}

	err = itineraryService.Create(itinerary)
	if err != nil {
		log.Errorf("Error creating itinerary: %v", err)
		fmt.Print(err)
		context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Could not create itinerary. Try again later."})
		return
	}

	log.Debugf("Itinerary created successfully for user %d", userId)
	context.JSON(http.StatusCreated, &CreateItineraryResponse{Message: "Itinerary created.", ItineraryID: itinerary.ID})
}

// updateItinerary godoc
// @Summary      Update an itinerary
// @Description  Updates an existing itinerary for the authenticated user.
// @Tags         itineraries
// @Accept       json
// @Produce      json
// @Security     Auth
// @Param        itinerary  body  UpdateItineraryRequest  true  "Itinerary update data"
// @Success      200  {object}  UpdateItineraryResponse       "Itinerary updated."
// @Failure      400  {object}  ErrorResponse       "Could not parse request data or invalid destinations."
// @Failure      401  {object}  ErrorResponse       "Not authorized."
// @Failure      403  {object}  ErrorResponse       "You do not have permission to update this itinerary."
// @Failure      404  {object}  ErrorResponse       "Itinerary not found."
// @Failure      500  {object}  ErrorResponse       "Could not update itinerary. Try again later."
// @Router       /itineraries [put]
func updateItinerary(context *gin.Context) {
	log.Debug("Updating itinerary")

	var input UpdateItineraryRequest

	userId, exists := context.Get("userId")
	if !exists {
		log.Error("User ID not found in context")
		context.JSON(http.StatusUnauthorized, &ErrorResponse{Message: "Not authorized."})
		return
	}

	if err := context.ShouldBindJSON(&input); err != nil {
		log.Errorf("Error parsing JSON %v", err)
		context.JSON(http.StatusBadRequest, &ErrorResponse{Message: "Could not parse request data."})
		return
	}

	itineraryService := services.GetItineraryService()

	itinerary, err := itineraryService.FindById(input.ID, true)

	if err != nil {
		if strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			log.Warnf("Itinerary with ID %d not found for user %d", input.ID, userId)
			context.JSON(http.StatusNotFound, &ErrorResponse{Message: "Itinerary not found."})
		} else {
			log.Errorf("Error retrieving itinerary %v", err)
			context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Could not get itinerary. Try again later."})
		}
		return
	}

	if itinerary.OwnerID != userId.(int64) {
		log.Errorf("User %d does not have permission to update itinerary %d", userId, input.ID)
		context.JSON(http.StatusForbidden, &ErrorResponse{Message: "You do not have permission to update this itinerary."})
		return
	}

	itinerary.Title = input.Title
	itinerary.Description = input.Description
	itinerary.Notes = input.Notes
	itinerary.TravelDestinations = input.Destinations

	err = itineraryService.ValidateItineraryDestinationsDates(itinerary.TravelDestinations)
	if err != nil {
		log.Errorf("Error validating itinerary destinations dates: %v", err)
		context.JSON(http.StatusBadRequest, &ErrorResponse{Message: err.Error()})
		return
	}

	err = itineraryService.Update(itinerary)
	if err != nil {
		log.Errorf("Error updating itinerary %v", err)
		context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Could not update itinerary. Try again later."})
		return
	}

	log.Debugf("Itinerary %d updated successfully for user %d", input.ID, userId)
	context.JSON(http.StatusOK, &UpdateItineraryResponse{Message: "Itinerary updated."})
}

// deleteItinerary godoc
// @Summary      Delete an itinerary
// @Description  Deletes an itinerary for the authenticated user.
// @Tags         itineraries
// @Produce      json
// @Security     Auth
// @Param        itineraryId  path  int  true  "Itinerary ID"
// @Success      200  {object}  DeleteItineraryResponse  "Itinerary deleted."
// @Failure      401  {object}  ErrorResponse  "Not authorized."
// @Failure      403  {object}  ErrorResponse "You do not have permission to access this resource."
// @Failure      404  {object}  ErrorResponse "Itinerary not found."
// @Failure      500  {object}  ErrorResponse  "Could not delete itinerary. Try again later."
// @Router       /itineraries/{itineraryId} [delete]
func deleteItinerary(context *gin.Context) {
	log.Debug("Deleting itinerary")

	itinerary := getAndValidateItinerary(context, false)
	if itinerary == nil {
		return
	}

	itineraryService := services.GetItineraryService()

	err := itineraryService.Delete(itinerary.ID)
	if err != nil {
		log.Errorf("Error deleting itinerary %d: %v", itinerary.ID, err)
		context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Could not delete itinerary. Try again later."})
		return
	}

	log.Debugf("Itinerary %d deleted successfully for user %d", itinerary.ID, itinerary.OwnerID)
	context.JSON(http.StatusOK, &DeleteItineraryResponse{Message: "Itinerary deleted."})
}

// getOwnersItineraries godoc
// @Summary      Get all itineraries of the authenticated user
// @Description  Retrieves all itineraries belonging to the authenticated user.
// @Tags         itineraries
// @Produce      json
// @Security     Auth
// @Success      200  {object}  []models.Itinerary  "List of itineraries"  example({"itineraries": [{"id": 1, "title": "Trip to Spain", "description": "Summer vacation", "notes": "Pack sunscreen", "ownerId": 42, "travelDestinations": [{"id": 1, "country": "Spain", "city": "Madrid", "arrivalDate": "2024-07-01T00:00:00Z", "departureDate": "2024-07-05T00:00:00Z"}]}]})
// @Failure      401  {object}  ErrorResponse  "Not authorized."
// @Failure      500  {object}  ErrorResponse  "Could not retrieve itineraries. Try again later."
// @Router       /itineraries [get]
func getOwnersItineraries(context *gin.Context) {
	log.Debug("Retrieving owner's itineraries")

	userId, exists := context.Get("userId")
	if !exists {
		log.Error("User ID not found in context")
		context.JSON(http.StatusUnauthorized, &ErrorResponse{Message: "Not authorized."})
		return
	}

	itineraryService := services.GetItineraryService()

	itineraries, err := itineraryService.FindByOwnerId(userId.(int64))
	if err != nil {
		log.Errorf("Error retrieving itineraries for user %d: %v", userId, err)
		context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Could not retrieve itineraries. Try again later."})
		return
	}

	log.Debugf("Retrieved itineraries for user %d: %d itineraries found", userId, len(*itineraries))
	context.JSON(http.StatusOK, &GetItinerariesResponse{Itineraries: itineraries})
}

// getItinerary godoc
// @Summary      Get an itinerary by ID
// @Description  Retrieves a specific itinerary for the authenticated user.
// @Tags         itineraries
// @Produce      json
// @Security     Auth
// @Param        itineraryId  path  int  true  "Itinerary ID"
// @Success      200  {object}  models.Itinerary  "Itinerary details"
// @Failure      401  {object}  ErrorResponse  "Not authorized."
// @Failure      403  {object}  ErrorResponse  "You do not have permission to access this resource."
// @Failure      404  {object}  ErrorResponse  "Itinerary not found."
// @Failure      500  {object}  ErrorResponse  "Could not get itinerary. Try again later."
// @Router       /itineraries/{itineraryId} [get]
func getItinerary(context *gin.Context) {
	log.Debug("Retrieving itinerary")

	itinerary := getAndValidateItinerary(context, true)
	if itinerary == nil {
		return
	}

	log.Debugf("Retrieved itinerary for user %d: %+v", itinerary.OwnerID, itinerary)
	context.JSON(http.StatusOK, &GetItineraryResponse{Itinerary: itinerary})
}

// runItineraryFileJob godoc
// @Summary      Start itinerary file generation job
// @Description  Starts an asynchronous job to generate a file for the specified itinerary. The user must be authenticated and the itinerary must belong to them.
// @Tags         itineraries
// @Produce      json
// @Security     Auth
// @Param        itineraryId  path  int  true  "Itinerary ID"
// @Success      202  {object}  map[string]interface{}  "Job started successfully."
// @Failure      401  {object}  ErrorResponse       "Not authorized."
// @Failure      403  {object}  ErrorResponse       "You do not have permission to access this resource."
// @Failure      404  {object}  ErrorResponse       "Itinerary not found."
// @Failure      409  {object}  ErrorResponse       "Too many jobs running for your user. Please wait for existing jobs to complete."
// @Failure      500  {object}  ErrorResponse       "Could not create job. Try again later."
// @Router       /itineraries/{itineraryId}/jobs [post]
func runItineraryFileJob(context *gin.Context) {
	log.Debug("Running itinerary file job")

	itinerary := getAndValidateItinerary(context, true)
	if itinerary == nil {
		return
	}

	jobsService := services.GetItineraryFileJobService()

	// Check if there is already a job running for this user
	userId, _ := context.Get("userId")

	jobsRunningCount, err := jobsService.GetJobsRunningOfUserCount(userId.(int64))
	if err != nil {
		log.Errorf("Error checking running jobs for user %d: %v", userId, err)
		context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Could not check job status. Try again later."})
		return
	}

	jobsRunningLimitStr := os.Getenv("JOBS_RUNNING_PER_USER_LIMIT")
	jobsRunningLimit := 5 // Default limit if not set
	if jobsRunningLimitStr != "" {
		var convErr error
		jobsRunningLimit, convErr = strconv.Atoi(jobsRunningLimitStr)
		if convErr != nil {
			log.Errorf("Invalid JOBS_RUNNING_PER_USER_LIMIT environment variable: %v", convErr)
			context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Invalid jobs running limit configuration."})
			return
		}
	}

	if jobsRunningCount >= jobsRunningLimit {
		log.Errorf("User %d has too many jobs running: %d", userId, jobsRunningCount)
		context.JSON(http.StatusConflict, &ErrorResponse{Message: "Too many jobs running for your user. Please wait for existing jobs to complete."})
		return
	}

	// Prepare and run the job
	itineraryFileJobTask, err := jobsService.PrepareJob(itinerary)
	if err != nil {
		log.Errorf("Error preparing itinerary file job: %v", err)
		context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Could not create job. Try again later."})
		return
	}

	job := itineraryFileJobTask.ItineraryFileJob

	asyncTaskQueue, err := services.NewAsyncqTaskQueue()
	if err != nil {
		log.Errorf("Error initializing async task queue: %v", err)
		context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Could not create job. Try again later."})
		return
	}
	defer asyncTaskQueue.Close()

	if asyncTaskQueue == nil {
		log.Error("Async task queue is not initialized.")
		context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Internal server error. Try again later."})
		return
	}

	asyncTaskId, err := asyncTaskQueue.EnqueueItineraryFileJob(*itineraryFileJobTask)
	if err != nil {
		log.Error("Error enqueuing itinerary file job: ", err)
		jobsService.FailJob("Could not enqueue job", job)
		context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Could not enqueue job. Try again later."})
		return
	}

	err = jobsService.AddAsyncTaskId(*asyncTaskId, job)
	if err != nil {
		log.Error("Error adding async task ID to job: ", err)
		jobsService.FailJob("Could not add async task ID to job", job)
		context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Could not add async task ID to job. Try again later."})
		return
	}

	log.Debugf("Itinerary file job started successfully with ID %d for user %d", job.ID, userId)
	context.JSON(http.StatusAccepted, &StartItineraryJobResponse{Message: "Job started successfully.", JobId: job.ID})
}

// getItineraryJob godoc
// @Summary      Get an itinerary file job by ID
// @Description  Retrieves a specific itinerary file job for the authenticated user.
// @Tags         itineraries
// @Produce      json
// @Security     Auth
// @Param        itineraryId     path  int  true  "Itinerary ID"
// @Param        itineraryJobId  path  int  true  "Itinerary Job ID"
// @Success      200  {object}  models.ItineraryFileJob  "Itinerary file job details"  example({"job": {"id": 1, "status": "completed", "statusDescription": "Job OK", "creationDate": "2024-07-01T00:00:00Z", "startDate": "2024-07-01T01:00:00Z", "endDate": "2024-07-01T02:00:00Z", "filePath": "/path/to/file", "fileManager": "local", "itineraryId": 1, "asyncTaskId": "a1b2c3"}})
// @Failure      400  {object}  ErrorResponse  "Bad request."
// @Failure      401  {object}  ErrorResponse  "Not authorized."
// @Failure      403  {object}  ErrorResponse  "You do not have permission to access this resource."
// @Failure      404  {object}  ErrorResponse  "Itinerary job not found."
// @Failure      500  {object}  ErrorResponse  "Could not get itinerary job. Try again later."
// @Router       /itineraries/{itineraryId}/jobs/{itineraryJobId} [get]
func getItineraryJob(context *gin.Context) {
	log.Debug("Retrieving itinerary job")

	itinerary := getAndValidateItinerary(context, false)
	if itinerary == nil {
		return
	}

	itineraryJobIdStr := context.Param("itineraryJobId")
	if itineraryJobIdStr == "" {
		log.Error("Itinerary Job ID is required but not provided.")
		context.JSON(http.StatusBadRequest, &ErrorResponse{Message: "Itinerary Job ID is required."})
		return
	}

	// Convert itineraryJobId from string to int64
	var itineraryJobId int64
	_, err := fmt.Sscan(itineraryJobIdStr, &itineraryJobId)
	if err != nil {
		log.Error("Invalid itinerary job ID format: ", err)
		context.JSON(http.StatusBadRequest, &ErrorResponse{Message: "Invalid itinerary job ID."})
		return
	}

	jobsService := services.GetItineraryFileJobService()

	itineraryJob, err := jobsService.FindAliveById(itineraryJobId)
	if err != nil {
		if strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			log.Error("Itinerary job not found: ", err)
			context.JSON(http.StatusNotFound, &ErrorResponse{Message: "Itinerary job not found."})
		} else {
			log.Error("Error retrieving itinerary job: ", err)
			context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Could not get itinerary job. Try again later."})
		}
		return
	}

	itineraryJob = validateItineraryJobOwnership(itinerary.ID, itineraryJob, context)
	if itineraryJob == nil {
		return
	}

	log.Debugf("Retrieved itinerary job: %+v", itineraryJob)
	context.JSON(http.StatusOK, &GetItineraryJobResponse{Job: itineraryJob})

}

// downloadItineraryJobFile godoc
// @Summary      Download itinerary job file
// @Description  Downloads the generated file for the specified itinerary job. The user must be authenticated and the itinerary must belong to them.
// @Tags         itineraries
// @Produce      application/octet-stream
// @Security     Auth
// @Param        itineraryId     path  int  true  "Itinerary ID"
// @Param        itineraryJobId  path  int  true  "Itinerary Job ID"
// @Success      200  {file}  file  "File downloaded successfully."
// @Failure      400  {object}  ErrorResponse  "Bad request."
// @Failure      401  {object}  ErrorResponse  "Not authorized."
// @Failure      403  {object}  ErrorResponse  "You do not have permission to access this resource."
// @Failure      404  {object}  ErrorResponse  "Itinerary job or file not found."
// @Failure      500  {object}  ErrorResponse  "Could not download file. Try again later."
// @Router       /itineraries/{itineraryId}/jobs/{itineraryJobId}/file [get]
func downloadItineraryJobFile(context *gin.Context) {
	log.Debug("Downloading itinerary job file")

	itinerary := getAndValidateItinerary(context, false)
	if itinerary == nil {
		return
	}

	itineraryJobIdStr := context.Param("itineraryJobId")
	if itineraryJobIdStr == "" {
		log.Error("Itinerary Job ID is required but not provided.")
		context.JSON(http.StatusBadRequest, &ErrorResponse{Message: "Itinerary Job ID is required."})
		return
	}

	// Convert itineraryJobId from string to int64
	var itineraryJobId int64
	_, err := fmt.Sscan(itineraryJobIdStr, &itineraryJobId)
	if err != nil {
		log.Error("Invalid itinerary job ID format: ", err)
		context.JSON(http.StatusBadRequest, &ErrorResponse{Message: "Invalid itinerary job ID."})
		return
	}

	jobsService := services.GetItineraryFileJobService()

	itineraryJob, err := jobsService.FindAliveById(itineraryJobId)
	if err != nil {
		if strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			log.Error("Itinerary job not found: ", err)
			context.JSON(http.StatusNotFound, &ErrorResponse{Message: "Itinerary job not found."})
		} else {
			log.Error("Error retrieving itinerary job: ", err)
			context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Could not get itinerary job. Try again later."})
		}
		return
	}

	itineraryJob = validateItineraryJobOwnership(itinerary.ID, itineraryJob, context)
	if itineraryJob == nil {
		return
	}

	filePath := itineraryJob.Filepath
	if filePath == "" {
		context.JSON(http.StatusNotFound, &ErrorResponse{Message: "Itinerary job file not found."})
		return
	}

	file, err := jobsService.OpenItineraryJobFile(itineraryJob)
	if err != nil {
		log.Error("Error opening itinerary job file: ", err)
		context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Could not open itinerary job file. Try again later."})
		return
	}
	defer file.Close()

	// Assert file to *os.File to access Stat()
	osFile, ok := file.(*os.File)
	if !ok {
		log.Error("File is not an *os.File, cannot get file info")
		// TODO in the future, instead of returning an error, we could handle different file types (like a S3 file)
		context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Internal server error. Try again later."})
		return
	}

	fileInfo, err := osFile.Stat()
	if err != nil {
		log.Error("Error getting file info: ", err)
		context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Could not get file info. Try again later."})
		return
	}
	context.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileInfo.Name()))
	context.Header("Content-Type", "application/octet-stream")
	context.Header("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	http.ServeContent(context.Writer, context.Request, fileInfo.Name(), fileInfo.ModTime(), file)

	log.Debugf("File %s served successfully for itinerary job ID %d", fileInfo.Name(), itineraryJobId)
	context.Status(http.StatusOK)
	// Note: The file will be served directly to the client, so no further action is needed here.

}

// stopItineraryJob godoc
// @Summary      Stop an itinerary file job
// @Description  Stops an active itinerary file job for the authenticated user. The user must own the itinerary.
// @Tags         itineraries
// @Produce      json
// @Security     Auth
// @Param        itineraryId     path  int  true  "Itinerary ID"
// @Param        itineraryJobId  path  int  true  "Itinerary Job ID"
// @Success      200  {object}  StopItineraryJobResponse  "Itinerary job stopped."  example({"message": "Itinerary job stopped."})
// @Failure      400  {object}  ErrorResponse  "Bad request."
// @Failure      401  {object}  ErrorResponse  "Not authorized."
// @Failure      403  {object}  ErrorResponse  "You do not have permission to access this resource."
// @Failure      404  {object}  ErrorResponse  "Itinerary job not found."
// @Failure      500  {object}  ErrorResponse  "Could not get itinerary job. Try again later."
// @Router       /itineraries/{itineraryId}/jobs/{itineraryJobId}/stop [post]
func stopItineraryJob(context *gin.Context) {
	log.Debug("Stopping itinerary job")
	itinerary := getAndValidateItinerary(context, false)
	if itinerary == nil {
		return
	}

	itineraryJobIdStr := context.Param("itineraryJobId")
	if itineraryJobIdStr == "" {
		context.JSON(http.StatusBadRequest, &ErrorResponse{Message: "Itinerary Job ID is required."})
		return
	}

	// Convert itineraryJobId from string to int64
	var itineraryJobId int64
	_, err := fmt.Sscan(itineraryJobIdStr, &itineraryJobId)
	if err != nil {
		context.JSON(http.StatusBadRequest, &ErrorResponse{Message: "Invalid itinerary job ID."})
		return
	}

	jobsService := services.GetItineraryFileJobService()

	itineraryJob, err := jobsService.FindAliveById(itineraryJobId)
	if err != nil {
		if strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			log.Error("Itinerary job not found: ", err)
			context.JSON(http.StatusNotFound, &ErrorResponse{Message: "Itinerary job not found."})
		} else {
			log.Error("Error retrieving itinerary job: ", err)
			context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Could not get itinerary job. Try again later."})
		}
		return
	}

	itineraryJob = validateItineraryJobOwnership(itinerary.ID, itineraryJob, context)
	if itineraryJob == nil {
		return
	}

	err = jobsService.StopJob(itineraryJob)
	if err != nil {
		log.Error("Error stopping itinerary job: ", err)
		context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: fmt.Sprintf("Could not stop job: %v", err)})
		return
	}

	log.Debugf("Itinerary job %d stopped successfully for itinerary ID %d", itineraryJobId, itinerary.ID)
	context.JSON(http.StatusOK, &StopItineraryJobResponse{Message: "Itinerary job stopped."})
}

// deleteItineraryJob godoc
// @Summary      Delete an itinerary file job
// @Description  Soft deletes an itinerary file job for the authenticated user. The user must own the itinerary.
// @Tags         itineraries
// @Produce      json
// @Security     Auth
// @Param        itineraryId     path  int  true  "Itinerary ID"
// @Param        itineraryJobId  path  int  true  "Itinerary Job ID"
// @Success      200  {object}  DeleteItineraryJobResponse "Itinerary job deleted."  example({"message": "Itinerary job deleted."})
// @Failure      400  {object}  ErrorResponse  "Bad request."
// @Failure      401  {object}  ErrorResponse  "Not authorized."
// @Failure      403  {object}  ErrorResponse  "You do not have permission to access this resource."
// @Failure      404  {object}  ErrorResponse  "Itinerary job not found."
// @Failure      500  {object}  ErrorResponse  "Could not delete job. Try again later."
// @Router       /itineraries/{itineraryId}/jobs/{itineraryJobId} [delete]
func deleteItineraryJob(context *gin.Context) {
	log.Debug("Deleting itinerary job")

	itinerary := getAndValidateItinerary(context, false)
	if itinerary == nil {
		return
	}

	itineraryJobIdStr := context.Param("itineraryJobId")
	if itineraryJobIdStr == "" {
		log.Error("Itinerary Job ID is required but not provided.")
		context.JSON(http.StatusBadRequest, &ErrorResponse{Message: "Itinerary Job ID is required."})
		return
	}

	// Convert itineraryJobId from string to int64
	var itineraryJobId int64
	_, err := fmt.Sscan(itineraryJobIdStr, &itineraryJobId)
	if err != nil {
		log.Error("Invalid itinerary job ID format: ", err)
		context.JSON(http.StatusBadRequest, &ErrorResponse{Message: "Invalid itinerary job ID."})
		return
	}

	jobsService := services.GetItineraryFileJobService()

	itineraryJob, err := jobsService.FindAliveLightweightById(itineraryJobId)
	if err != nil {
		if strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			log.Error("Itinerary job not found: ", err)
			context.JSON(http.StatusNotFound, &ErrorResponse{Message: "Itinerary job not found."})
		} else {
			log.Error("Error retrieving itinerary job: ", err)
			context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Could not get itinerary job. Try again later."})
		}
		return
	}

	itineraryJob = validateItineraryJobOwnership(itinerary.ID, itineraryJob, context)
	if itineraryJob == nil {
		return
	}

	// We soft delete the job instead of hard deleting it to safely delete files later in a background task
	err = jobsService.SoftDeleteJob(itineraryJob)
	if err != nil {
		log.Error("Error deleting itinerary job: ", err)
		context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Could not delete job. Try again later."})
		return
	}

	log.Debugf("Itinerary job %d deleted successfully for itinerary ID %d", itineraryJobId, itinerary.ID)
	context.JSON(http.StatusOK, &DeleteItineraryJobResponse{Message: "Itinerary job deleted."})

}

// getAllItineraryFileJobs godoc
// @Summary      Get all file jobs for an itinerary
// @Description  Retrieves all file jobs associated with the specified itinerary. The user must be authenticated and own the itinerary.
// @Tags         itineraries
// @Produce      json
// @Security     Auth
// @Param        itineraryId  path  int  true  "Itinerary ID"
// @Success      200  {object}  []models.ItineraryFileJob  "List of itinerary file jobs"  example({"jobs": [{"id": 1, "status": "completed", "statusDescription": "Job OK", "creationDate": "2024-07-01T00:00:00Z", "startDate": "2024-07-01T01:00:00Z", "endDate": "2024-07-01T02:00:00Z", "filePath": "/path/to/file", "fileManager": "local", "itineraryId": 1, "asyncTaskId": "a1b2c3"}]})
// @Failure      401  {object}  ErrorResponse  "Not authorized."
// @Failure      403  {object}  ErrorResponse  "You do not have permission to access this resource."
// @Failure      404  {object}  ErrorResponse  "Itinerary not found."
// @Failure      500  {object}  ErrorResponse  "Could not retrieve jobs. Try again later."
// @Router       /itineraries/{itineraryId}/jobs [get]
func getAllItineraryFileJobs(context *gin.Context) {
	log.Debug("Retrieving all itinerary file jobs for an itinerary")

	itinerary := getAndValidateItinerary(context, false)
	if itinerary == nil {
		return
	}

	jobsService := services.GetItineraryFileJobService()

	itineraryFileJobs, err := jobsService.FindAliveByItineraryId(itinerary.ID)
	if err != nil {
		log.Error("Error retrieving itinerary file jobs: ", err)
		context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Could not retrieve jobs. Try again later."})
		return
	}

	log.Debugf("Retrieved %d itinerary file jobs for itinerary ID %d", len(*itineraryFileJobs), itinerary.ID)
	context.JSON(http.StatusOK, &GetItineraryJobsResponse{Jobs: itineraryFileJobs})
}

func validateAuthenticatedUser(context *gin.Context) *int64 {
	userId, exists := context.Get("userId")
	if !exists {
		log.Error("User ID not found in context")
		context.JSON(http.StatusUnauthorized, &ErrorResponse{Message: "Not authorized."})
		return nil
	}
	uid := userId.(int64)
	return &uid
}

func getAndValidateItinerary(context *gin.Context, fullItinerary bool) *models.Itinerary {
	userId := validateAuthenticatedUser(context)
	if userId == nil {
		return nil
	}

	itineraryIdStr := context.Param("itineraryId")
	if itineraryIdStr == "" {
		log.Error("Itinerary ID is required but not provided.")
		context.JSON(http.StatusBadRequest, &ErrorResponse{Message: "Itinerary ID is required."})
		return nil
	}

	// Convert itineraryId from string to int64
	var itineraryId int64
	_, err := fmt.Sscan(itineraryIdStr, &itineraryId)
	if err != nil {
		log.Error("Invalid itinerary ID format: ", err)
		context.JSON(http.StatusBadRequest, &ErrorResponse{Message: "Invalid itinerary ID."})
		return nil
	}

	itineraryService := services.GetItineraryService()

	var itinerary *models.Itinerary

	if fullItinerary {
		itinerary, err = itineraryService.FindById(itineraryId, true)
	} else {
		itinerary, err = itineraryService.FindLightweightById(itineraryId)
	}

	if err != nil {
		if strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			log.Errorf("Itinerary with ID %d not found for user %d", itineraryId, *userId)
			context.JSON(http.StatusNotFound, &ErrorResponse{Message: "Itinerary not found."})
		} else {
			log.Errorf("Error retrieving itinerary %v", err)
			context.JSON(http.StatusInternalServerError, &ErrorResponse{Message: "Could not get itinerary. Try again later."})
		}
		return nil
	}

	if itinerary.OwnerID != *userId {
		log.Errorf("User %d does not have permission to access itinerary %d", *userId, itineraryId)
		context.JSON(http.StatusForbidden, &ErrorResponse{Message: "You do not have permission to access this resource."})
		return nil
	}

	return itinerary
}

func validateItineraryJobOwnership(itineraryId int64, itineraryFileJob *models.ItineraryFileJob, context *gin.Context) *models.ItineraryFileJob {
	if itineraryId != itineraryFileJob.ItineraryID {
		log.Errorf("Itinerary ID %d does not match job's itinerary ID %d", itineraryId, itineraryFileJob.ItineraryID)
		context.JSON(http.StatusForbidden, &ErrorResponse{Message: "You do not have permission to access this resource."})
		return nil
	}

	return itineraryFileJob
}
