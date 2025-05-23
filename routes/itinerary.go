package routes

import (
	"fmt"
	"net/http"
	"time"

	"example.com/travel-advisor/models"
	"github.com/gin-gonic/gin"
)

func createItinerary(context *gin.Context) {
	var input struct {
		Title           string                              `json:"title" binding:"required"`
		Description     string                              `json:"description"`
		TravelStartDate time.Time                           `json:"travelStartDate"`
		TravelEndDate   time.Time                           `json:"travelEndDate"`
		Destinations    []models.ItineraryTravelDestination `json:"destinations" binding:"required,dive"`
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

	itinerary := models.NewItinerary(input.Title, input.Description, input.TravelStartDate, input.TravelEndDate, input.Destinations)

	itinerary.OwnerID = userId.(int64)

	err := itinerary.Create()
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

	itinerary := models.NewItinerary("", "", time.Time{}, time.Time{}, nil)
	itinerary.OwnerID = userId.(int64)

	itineraries, err := itinerary.FindByOwnerId()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not retrieve itineraries. Try again later."})
		return
	}

	context.JSON(http.StatusOK, gin.H{"itineraries": itineraries})
}

func getItinerary(context *gin.Context) {
	itineraryId := context.Param("itineraryId")
	if itineraryId == "" {
		context.JSON(http.StatusBadRequest, gin.H{"message": "Itinerary ID is required."})
		return
	}

	itinerary := models.NewItinerary("", "", time.Time{}, time.Time{}, nil)

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

	userId, exists := context.Get("userId")
	if !exists {
		context.JSON(http.StatusUnauthorized, gin.H{"message": "Not authorized."})
		return
	}

	if itinerary.OwnerID != userId.(int64) {
		context.JSON(http.StatusForbidden, gin.H{"message": "You do not have permission to access this itinerary."})
		return
	}

	context.JSON(http.StatusOK, gin.H{"itinerary": itinerary})
}
