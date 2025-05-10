package routes

import (
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
		context.JSON(http.StatusInternalServerError, gin.H{"message": "Could not create itinerary. Try again later."})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "Itinerary created.", "itinerary": itinerary})
}
