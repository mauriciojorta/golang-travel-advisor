package routes

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"example.com/travel-advisor/models"
	"github.com/gin-gonic/gin"
)

func createItinerary(context *gin.Context) {
	var input struct {
		Title        string                              `json:"title" binding:"required"`
		Description  string                              `json:"description"`
		Destinations []models.ItineraryTravelDestination `json:"destinations" binding:"required,dive"`
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

	itinerary := models.NewItinerary(input.Title, input.Description, input.Destinations)

	itinerary.OwnerID = userId.(int64)

	err := models.ValidateItineraryDestinationsDates(&itinerary.TravelDestinations)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = itinerary.Create()
	if err != nil {
		fmt.Print(err)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not create itinerary. Try again later."})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "Itinerary created."})
}

func getOwnersItineraries(context *gin.Context) {
	userId, exists := context.Get("userId")
	if !exists {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Not authorized."})
		return
	}

	itinerary := models.NewItinerary("", "", nil)
	itinerary.OwnerID = userId.(int64)

	itineraries, err := itinerary.FindByOwnerId()
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

	itinerary := models.NewItinerary("", "", nil)

	// Convert itineraryId from string to int64
	var id int64
	_, err := fmt.Sscan(itineraryId, &id)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Invalid itinerary ID."})
		return
	}
	itinerary.ID = id

	err = itinerary.FindById()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve itinerary. Try again later."})
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

	itinerary := models.NewItinerary("", "", nil)
	itinerary.ID = id

	err = itinerary.FindById()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve itinerary. Try again later."})
		return
	}

	if itinerary.OwnerID != userId.(int64) {
		context.JSON(http.StatusForbidden, gin.H{"message": "You do not have permission to run this job."})
		return
	}

	job := models.NewItineraryFileJob(itinerary.ID)

	// Check if there is already a job running for this itinerary
	isRunning, err := job.HasItineraryAJobRunning()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not check job status. Try again later."})
		return
	}
	if isRunning {
		context.JSON(http.StatusConflict, gin.H{"message": "A job is already running for this itinerary."})
		return
	}

	// Run the job
	err = job.RunJob(itinerary)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not start job. Try again later."})
		return
	}

	context.JSON(http.StatusAccepted, gin.H{"message": "Job started successfully."})
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

	itinerary := models.NewItinerary("", "", nil)
	itinerary.ID = itineraryId
	err = itinerary.FindById()
	if err != nil {
		log.Error("Error retrieving itinerary: ", err)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve itinerary. Try again later."})
		return
	}

	if itinerary.OwnerID != userId.(int64) {
		context.JSON(http.StatusForbidden, gin.H{"message": "You do not have permission to run this job."})
		return
	}

	job := models.NewItineraryFileJob(itineraryId)

	itineraryFileJobs, err := job.FindByItineraryId()
	if err != nil {
		log.Error("Error retrieving itinerary file jobs: ", err)
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve jobs. Try again later."})
		return
	}

	context.JSON(http.StatusOK, gin.H{"jobs": *itineraryFileJobs})
}
