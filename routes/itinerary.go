package routes

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"

	"example.com/travel-advisor/models"
	"example.com/travel-advisor/services"
	"github.com/gin-gonic/gin"
)

func createItinerary(context *gin.Context) {
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
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = itineraryService.Create(itinerary)
	if err != nil {
		fmt.Print(err)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not create itinerary. Try again later."})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "Itinerary created."})
}

func updateItinerary(context *gin.Context) {
	var input struct {
		ID           int64                                `json:"id"`
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

	if err := context.ShouldBindJSON(&input); err != nil {
		log.Errorf("Error parsing JSON %v", err)
		context.JSON(http.StatusBadRequest, gin.H{"error": "Could not parse request data."})
		return
	}

	itineraryService := services.GetItineraryService()

	itinerary, err := itineraryService.FindById(input.ID)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"message": "Itinerary not found. Try again later."})
		return
	}

	if itinerary.OwnerID != userId.(int64) {
		context.JSON(http.StatusForbidden, gin.H{"message": "You do not have permission to update this itinerary."})
		return
	}

	itinerary.Title = input.Title
	itinerary.Description = input.Description
	itinerary.Notes = input.Notes
	itinerary.TravelDestinations = input.Destinations

	err = itineraryService.ValidateItineraryDestinationsDates(itinerary.TravelDestinations)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = itineraryService.Update(itinerary)
	if err != nil {
		log.Errorf("Error updating itinerary %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not update itinerary. Try again later."})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Itinerary updated."})
}

func deleteItinerary(context *gin.Context) {
	userId, exists := context.Get("userId")
	if !exists {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Not authorized."})
		return
	}

	itineraryId := context.Param("itineraryId")
	if itineraryId == "" {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Itinerary ID is required."})
		return
	}

	var id int64
	_, err := fmt.Sscan(itineraryId, &id)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid itinerary ID."})
		return
	}

	itineraryService := services.GetItineraryService()

	itinerary, err := itineraryService.FindById(id)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"message": "Itinerary not found."})
		return
	}

	if itinerary.OwnerID != userId.(int64) {
		context.JSON(http.StatusForbidden, gin.H{"message": "You do not have permission to delete this itinerary."})
		return
	}

	err = itineraryService.Delete(itinerary)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not delete itinerary. Try again later."})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Itinerary deleted."})
}

func getOwnersItineraries(context *gin.Context) {
	userId, exists := context.Get("userId")
	if !exists {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Not authorized."})
		return
	}

	itineraryService := services.GetItineraryService()

	itineraries, err := itineraryService.FindByOwnerId(userId.(int64))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve itineraries. Try again later."})
		return
	}

	context.JSON(http.StatusOK, gin.H{"itineraries": itineraries})
}

func getItinerary(context *gin.Context) {
	userId, exists := context.Get("userId")
	if !exists {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Not authorized."})
		return
	}

	itineraryId := context.Param("itineraryId")
	if itineraryId == "" {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Itinerary ID is required."})
		return
	}

	itineraryService := services.GetItineraryService()

	// Convert itineraryId from string to int64
	var id int64
	_, err := fmt.Sscan(itineraryId, &id)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid itinerary ID."})
		return
	}

	itinerary, err := itineraryService.FindById(id)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"message": "Itinerary not found."})
		return
	}

	if itinerary.OwnerID != userId.(int64) {
		context.JSON(http.StatusForbidden, gin.H{"message": "You do not have permission to access this itinerary."})
		return
	}

	context.JSON(http.StatusOK, gin.H{"itinerary": itinerary})
}

func runItineraryFileJob(context *gin.Context) {
	userId, exists := context.Get("userId")
	if !exists {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Not authorized."})
		return
	}

	itineraryId := context.Param("itineraryId")
	if itineraryId == "" {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Itinerary ID is required."})
		return
	}

	// Convert itineraryId from string to int64
	var id int64
	_, err := fmt.Sscan(itineraryId, &id)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid itinerary ID."})
		return
	}

	itineraryService := services.GetItineraryService()

	itinerary, err := itineraryService.FindById(id)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"message": "Itinerary not found."})
		return
	}

	if itinerary.OwnerID != userId.(int64) {
		context.JSON(http.StatusForbidden, gin.H{"message": "You do not have permission to run this job."})
		return
	}

	jobsService := services.GetItineraryFileJobService()

	// Check if there is already a job running for this itinerary
	jobsRunningCount, err := jobsService.GetJobsRunningOfUserCount(userId.(int64))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not check job status. Try again later."})
		return
	}

	jobsRunningLimitStr := os.Getenv("JOBS_RUNNING_PER_USER_LIMIT")
	jobsRunningLimit := 5 // Default limit if not set
	if jobsRunningLimitStr != "" {
		var convErr error
		jobsRunningLimit, convErr = strconv.Atoi(jobsRunningLimitStr)
		if convErr != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid jobs running limit configuration."})
			return
		}
	}

	if jobsRunningCount >= jobsRunningLimit {
		context.JSON(http.StatusConflict, gin.H{"message": "Too many jobs running for your user. Please wait for existing jobs to complete."})
		return
	}

	// Prepare and run the job
	itineraryFileJobTask, err := jobsService.PrepareJob(itinerary)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not create job. Try again later."})
		return
	}

	job := itineraryFileJobTask.ItineraryFileJob

	asyncTaskQueue, err := services.NewAsyncqTaskQueue()
	if err != nil {
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

	context.JSON(http.StatusAccepted, gin.H{"message": "Job started successfully."})
}

func getItineraryJob(context *gin.Context) {
	userId, exists := context.Get("userId")
	if !exists {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Not authorized."})
		return
	}

	itineraryIdStr := context.Param("itineraryId")
	if itineraryIdStr == "" {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Itinerary ID is required."})
		return
	}

	itineraryJobIdStr := context.Param("itineraryJobId")
	if itineraryJobIdStr == "" {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Itinerary Job ID is required."})
		return
	}

	// Convert itineraryId from string to int64
	var itineraryId int64
	_, err := fmt.Sscan(itineraryIdStr, &itineraryId)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid itinerary ID."})
		return
	}

	itineraryService := services.GetItineraryService()

	itinerary, err := itineraryService.FindById(itineraryId)
	if err != nil {
		log.Error("Error retrieving itinerary: ", err)
		context.JSON(http.StatusNotFound, gin.H{"message": "Itinerary not found."})
		return
	}

	if itinerary.OwnerID != userId.(int64) {
		context.JSON(http.StatusForbidden, gin.H{"message": "You do not have permission to run this job."})
		return
	}

	// Convert itineraryJobId from string to int64
	var itineraryJobId int64
	_, err = fmt.Sscan(itineraryJobIdStr, &itineraryJobId)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid itinerary job ID."})
		return
	}

	jobsService := services.GetItineraryFileJobService()

	itineraryJob, err := jobsService.FindById(itineraryJobId)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"message": "Itinerary job not found."})
		return
	}

	context.JSON(http.StatusOK, gin.H{"job": *itineraryJob})

}

func getAllItineraryFileJobs(context *gin.Context) {
	userId, exists := context.Get("userId")
	if !exists {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Not authorized."})
		return
	}

	itineraryIdStr := context.Param("itineraryId")
	if itineraryIdStr == "" {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Itinerary ID is required."})
		return
	}

	// Convert itineraryId from string to int64
	var itineraryId int64
	_, err := fmt.Sscan(itineraryIdStr, &itineraryId)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid itinerary ID."})
		return
	}

	itineraryService := services.GetItineraryService()

	itinerary, err := itineraryService.FindById(itineraryId)
	if err != nil {
		log.Error("Error retrieving itinerary: ", err)
		context.JSON(http.StatusNotFound, gin.H{"message": "Itinerary not found."})
		return
	}

	if itinerary.OwnerID != userId.(int64) {
		context.JSON(http.StatusForbidden, gin.H{"message": "You do not have permission to run this job."})
		return
	}

	jobsService := services.GetItineraryFileJobService()

	itineraryFileJobs, err := jobsService.FindByItineraryId(itineraryId)
	if err != nil {
		log.Error("Error retrieving itinerary file jobs: ", err)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve jobs. Try again later."})
		return
	}

	context.JSON(http.StatusOK, gin.H{"jobs": *itineraryFileJobs})
}
