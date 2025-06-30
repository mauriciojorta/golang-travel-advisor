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

func createItinerary(context *gin.Context) {
	log.Debug("Creating itinerary")

	var input struct {
		Title        string                               `json:"title" binding:"required"`
		Description  string                               `json:"description"`
		Notes        *string                              `json:"notes"`
		Destinations *[]models.ItineraryTravelDestination `json:"destinations" binding:"required,dive"`
	}

	userId, exists := context.Get("userId")
	if !exists {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Not authorized."})
		return
	}

	// Bind JSON input to the input struct
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse request data."})
		return
	}

	itineraryService := services.GetItineraryService()

	itinerary := models.NewItinerary(input.Title, input.Description, input.Notes, input.Destinations)

	itinerary.OwnerID = userId.(int64)

	err := itineraryService.ValidateItineraryDestinationsDates(itinerary.TravelDestinations)
	if err != nil {
		log.Errorf("Error validating itinerary destinations dates: %v", err)
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = itineraryService.Create(itinerary)
	if err != nil {
		log.Errorf("Error creating itinerary: %v", err)
		fmt.Print(err)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not create itinerary. Try again later."})
		return
	}

	log.Debugf("Itinerary created successfully for user %d", userId)
	context.JSON(http.StatusCreated, gin.H{"message": "Itinerary created.", "itineraryId": itinerary.ID})
}

func updateItinerary(context *gin.Context) {
	log.Debug("Updating itinerary")

	var input struct {
		ID           int64                                `json:"id" binding:"required"`
		Title        string                               `json:"title" binding:"required"`
		Description  string                               `json:"description"`
		Notes        *string                              `json:"notes"`
		Destinations *[]models.ItineraryTravelDestination `json:"destinations" binding:"required,dive"`
	}

	userId, exists := context.Get("userId")
	if !exists {
		log.Error("User ID not found in context")
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Not authorized."})
		return
	}

	if err := context.ShouldBindJSON(&input); err != nil {
		log.Errorf("Error parsing JSON %v", err)
		context.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse request data."})
		return
	}

	itineraryService := services.GetItineraryService()

	itinerary, err := itineraryService.FindById(input.ID, true)

	if err != nil {
		if strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			log.Warnf("Itinerary with ID %d not found for user %d", input.ID, userId)
			context.JSON(http.StatusNotFound, gin.H{"message": "Itinerary not found."})
		} else {
			log.Errorf("Error retrieving itinerary %v", err)
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not get itinerary. Try again later."})
		}
		return
	}

	if itinerary.OwnerID != userId.(int64) {
		log.Errorf("User %d does not have permission to update itinerary %d", userId, input.ID)
		context.JSON(http.StatusForbidden, gin.H{"message": "You do not have permission to update this itinerary."})
		return
	}

	itinerary.Title = input.Title
	itinerary.Description = input.Description
	itinerary.Notes = input.Notes
	itinerary.TravelDestinations = input.Destinations

	err = itineraryService.ValidateItineraryDestinationsDates(itinerary.TravelDestinations)
	if err != nil {
		log.Errorf("Error validating itinerary destinations dates: %v", err)
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = itineraryService.Update(itinerary)
	if err != nil {
		log.Errorf("Error updating itinerary %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not update itinerary. Try again later."})
		return
	}

	log.Debugf("Itinerary %d updated successfully for user %d", input.ID, userId)
	context.JSON(http.StatusOK, gin.H{"message": "Itinerary updated."})
}

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
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not delete itinerary. Try again later."})
		return
	}

	log.Debugf("Itinerary %d deleted successfully for user %d", itinerary.ID, itinerary.OwnerID)
	context.JSON(http.StatusOK, gin.H{"message": "Itinerary deleted."})
}

func getOwnersItineraries(context *gin.Context) {
	log.Debug("Retrieving owner's itineraries")

	userId, exists := context.Get("userId")
	if !exists {
		log.Error("User ID not found in context")
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Not authorized."})
		return
	}

	itineraryService := services.GetItineraryService()

	itineraries, err := itineraryService.FindByOwnerId(userId.(int64))
	if err != nil {
		log.Errorf("Error retrieving itineraries for user %d: %v", userId, err)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve itineraries. Try again later."})
		return
	}

	log.Debugf("Retrieved itineraries for user %d: %d itineraries found", userId, len(*itineraries))
	context.JSON(http.StatusOK, gin.H{"itineraries": itineraries})
}

func getItinerary(context *gin.Context) {
	log.Debug("Retrieving itinerary")

	itinerary := getAndValidateItinerary(context, true)
	if itinerary == nil {
		return
	}

	log.Debugf("Retrieved itinerary for user %d: %+v", itinerary.OwnerID, itinerary)
	context.JSON(http.StatusOK, gin.H{"itinerary": itinerary})
}

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
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not check job status. Try again later."})
		return
	}

	jobsRunningLimitStr := os.Getenv("JOBS_RUNNING_PER_USER_LIMIT")
	jobsRunningLimit := 5 // Default limit if not set
	if jobsRunningLimitStr != "" {
		var convErr error
		jobsRunningLimit, convErr = strconv.Atoi(jobsRunningLimitStr)
		if convErr != nil {
			log.Errorf("Invalid JOBS_RUNNING_PER_USER_LIMIT environment variable: %v", convErr)
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid jobs running limit configuration."})
			return
		}
	}

	if jobsRunningCount >= jobsRunningLimit {
		log.Errorf("User %d has too many jobs running: %d", userId, jobsRunningCount)
		context.JSON(http.StatusConflict, gin.H{"message": "Too many jobs running for your user. Please wait for existing jobs to complete."})
		return
	}

	// Prepare and run the job
	itineraryFileJobTask, err := jobsService.PrepareJob(itinerary)
	if err != nil {
		log.Errorf("Error preparing itinerary file job: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not create job. Try again later."})
		return
	}

	job := itineraryFileJobTask.ItineraryFileJob

	asyncTaskQueue, err := services.NewAsyncqTaskQueue()
	if err != nil {
		log.Errorf("Error initializing async task queue: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not create job. Try again later."})
		return
	}
	defer asyncTaskQueue.Close()

	if asyncTaskQueue == nil {
		log.Error("Async task queue is not initialized.")
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error. Try again later."})
		return
	}

	asyncTaskId, err := asyncTaskQueue.EnqueueItineraryFileJob(*itineraryFileJobTask)
	if err != nil {
		log.Error("Error enqueuing itinerary file job: ", err)
		jobsService.FailJob("Could not enqueue job", job)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not enqueue job. Try again later."})
		return
	}

	err = jobsService.AddAsyncTaskId(*asyncTaskId, job)
	if err != nil {
		log.Error("Error adding async task ID to job: ", err)
		jobsService.FailJob("Could not add async task ID to job", job)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not add async task ID to job. Try again later."})
		return
	}

	log.Debugf("Itinerary file job started successfully with ID %d for user %d", job.ID, userId)
	context.JSON(http.StatusAccepted, gin.H{"message": "Job started successfully.", "jobId": job.ID})
}

func getItineraryJob(context *gin.Context) {
	log.Debug("Retrieving itinerary job")

	itinerary := getAndValidateItinerary(context, false)
	if itinerary == nil {
		return
	}

	itineraryJobIdStr := context.Param("itineraryJobId")
	if itineraryJobIdStr == "" {
		log.Error("Itinerary Job ID is required but not provided.")
		context.JSON(http.StatusBadRequest, gin.H{"message": "Itinerary Job ID is required."})
		return
	}

	// Convert itineraryJobId from string to int64
	var itineraryJobId int64
	_, err := fmt.Sscan(itineraryJobIdStr, &itineraryJobId)
	if err != nil {
		log.Error("Invalid itinerary job ID format: ", err)
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid itinerary job ID."})
		return
	}

	jobsService := services.GetItineraryFileJobService()

	itineraryJob, err := jobsService.FindAliveById(itineraryJobId)
	if err != nil {
		if strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			log.Error("Itinerary job not found: ", err)
			context.JSON(http.StatusNotFound, gin.H{"message": "Itinerary job not found."})
		} else {
			log.Error("Error retrieving itinerary job: ", err)
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not get itinerary job. Try again later."})
		}
		return
	}

	itineraryJob = validateItineraryJobOwnership(itinerary.ID, itineraryJob, context)
	if itineraryJob == nil {
		return
	}

	log.Debugf("Retrieved itinerary job: %+v", itineraryJob)
	context.JSON(http.StatusOK, gin.H{"job": *itineraryJob})

}

func downloadItineraryJobFile(context *gin.Context) {
	log.Debug("Downloading itinerary job file")

	itinerary := getAndValidateItinerary(context, false)
	if itinerary == nil {
		return
	}

	itineraryJobIdStr := context.Param("itineraryJobId")
	if itineraryJobIdStr == "" {
		log.Error("Itinerary Job ID is required but not provided.")
		context.JSON(http.StatusBadRequest, gin.H{"message": "Itinerary Job ID is required."})
		return
	}

	// Convert itineraryJobId from string to int64
	var itineraryJobId int64
	_, err := fmt.Sscan(itineraryJobIdStr, &itineraryJobId)
	if err != nil {
		log.Error("Invalid itinerary job ID format: ", err)
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid itinerary job ID."})
		return
	}

	jobsService := services.GetItineraryFileJobService()

	itineraryJob, err := jobsService.FindAliveById(itineraryJobId)
	if err != nil {
		if strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			log.Error("Itinerary job not found: ", err)
			context.JSON(http.StatusNotFound, gin.H{"message": "Itinerary job not found."})
		} else {
			log.Error("Error retrieving itinerary job: ", err)
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not get itinerary job. Try again later."})
		}
		return
	}

	itineraryJob = validateItineraryJobOwnership(itinerary.ID, itineraryJob, context)
	if itineraryJob == nil {
		return
	}

	filePath := itineraryJob.Filepath
	if filePath == "" {
		context.JSON(http.StatusNotFound, gin.H{"message": "Itinerary job file not found."})
		return
	}

	file, err := jobsService.OpenItineraryJobFile(itineraryJob)
	if err != nil {
		log.Error("Error opening itinerary job file: ", err)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not open itinerary job file. Try again later."})
		return
	}
	defer file.Close()

	// Assert file to *os.File to access Stat()
	osFile, ok := file.(*os.File)
	if !ok {
		log.Error("File is not an *os.File, cannot get file info")
		// TODO in the future, instead of returning an error, we could handle different file types (like a S3 file)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error. Try again later."})
		return
	}

	fileInfo, err := osFile.Stat()
	if err != nil {
		log.Error("Error getting file info: ", err)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not get file info. Try again later."})
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

func stopItineraryJob(context *gin.Context) {
	log.Debug("Stopping itinerary job")
	itinerary := getAndValidateItinerary(context, false)
	if itinerary == nil {
		return
	}

	itineraryJobIdStr := context.Param("itineraryJobId")
	if itineraryJobIdStr == "" {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Itinerary Job ID is required."})
		return
	}

	// Convert itineraryJobId from string to int64
	var itineraryJobId int64
	_, err := fmt.Sscan(itineraryJobIdStr, &itineraryJobId)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid itinerary job ID."})
		return
	}

	jobsService := services.GetItineraryFileJobService()

	itineraryJob, err := jobsService.FindAliveById(itineraryJobId)
	if err != nil {
		if strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			log.Error("Itinerary job not found: ", err)
			context.JSON(http.StatusNotFound, gin.H{"message": "Itinerary job not found."})
		} else {
			log.Error("Error retrieving itinerary job: ", err)
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not get itinerary job. Try again later."})
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
		context.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Could not stop job: %v", err)})
		return
	}

	log.Debugf("Itinerary job %d stopped successfully for itinerary ID %d", itineraryJobId, itinerary.ID)
	context.JSON(http.StatusOK, gin.H{"message": "Itinerary job stopped."})
}

func deleteItineraryJob(context *gin.Context) {
	log.Debug("Deleting itinerary job")

	itinerary := getAndValidateItinerary(context, false)
	if itinerary == nil {
		return
	}

	itineraryJobIdStr := context.Param("itineraryJobId")
	if itineraryJobIdStr == "" {
		log.Error("Itinerary Job ID is required but not provided.")
		context.JSON(http.StatusBadRequest, gin.H{"message": "Itinerary Job ID is required."})
		return
	}

	// Convert itineraryJobId from string to int64
	var itineraryJobId int64
	_, err := fmt.Sscan(itineraryJobIdStr, &itineraryJobId)
	if err != nil {
		log.Error("Invalid itinerary job ID format: ", err)
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid itinerary job ID."})
		return
	}

	jobsService := services.GetItineraryFileJobService()

	itineraryJob, err := jobsService.FindAliveLightweightById(itineraryJobId)
	if err != nil {
		if strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
			log.Error("Itinerary job not found: ", err)
			context.JSON(http.StatusNotFound, gin.H{"message": "Itinerary job not found."})
		} else {
			log.Error("Error retrieving itinerary job: ", err)
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not get itinerary job. Try again later."})
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
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not delete job. Try again later."})
		return
	}

	log.Debugf("Itinerary job %d deleted successfully for itinerary ID %d", itineraryJobId, itinerary.ID)
	context.JSON(http.StatusOK, gin.H{"message": "Itinerary job deleted."})

}

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
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve jobs. Try again later."})
		return
	}

	log.Debugf("Retrieved %d itinerary file jobs for itinerary ID %d", len(*itineraryFileJobs), itinerary.ID)
	context.JSON(http.StatusOK, gin.H{"jobs": *itineraryFileJobs})
}

func validateAuthenticatedUser(context *gin.Context) *int64 {
	userId, exists := context.Get("userId")
	if !exists {
		log.Error("User ID not found in context")
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Not authorized."})
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
		context.JSON(http.StatusBadRequest, gin.H{"message": "Itinerary ID is required."})
		return nil
	}

	// Convert itineraryId from string to int64
	var itineraryId int64
	_, err := fmt.Sscan(itineraryIdStr, &itineraryId)
	if err != nil {
		log.Error("Invalid itinerary ID format: ", err)
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid itinerary ID."})
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
			context.JSON(http.StatusNotFound, gin.H{"message": "Itinerary not found."})
		} else {
			log.Errorf("Error retrieving itinerary %v", err)
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not get itinerary. Try again later."})
		}
		return nil
	}

	if itinerary.OwnerID != *userId {
		log.Errorf("User %d does not have permission to access itinerary %d", *userId, itineraryId)
		context.JSON(http.StatusForbidden, gin.H{"message": "You do not have permission to access this resource."})
		return nil
	}

	return itinerary
}

func validateItineraryJobOwnership(itineraryId int64, itineraryFileJob *models.ItineraryFileJob, context *gin.Context) *models.ItineraryFileJob {
	if itineraryId != itineraryFileJob.ItineraryID {
		log.Errorf("Itinerary ID %d does not match job's itinerary ID %d", itineraryId, itineraryFileJob.ItineraryID)
		context.JSON(http.StatusForbidden, gin.H{"message": "You do not have permission to access this resource."})
		return nil
	}

	return itineraryFileJob
}
