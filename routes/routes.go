package routes

import (
	"example.com/travel-advisor/middlewares"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
)

type ErrorResponse struct {
	Message string `json:"message" example:"An error occurred."`
}

func RegisterRoutes(server *gin.Engine) {
	api := server.Group("/api/v1")

	api.POST("/signup", signUp)
	api.POST("/login", login)

	authenticated := api.Group("/")
	authenticated.Use(middlewares.Authenticate)
	authenticated.POST("/itineraries", createItinerary)
	authenticated.PUT("/itineraries", updateItinerary)
	authenticated.GET("/itineraries", getOwnersItineraries)
	authenticated.GET("/itineraries/:itineraryId", getItinerary)
	authenticated.DELETE("/itineraries/:itineraryId", deleteItinerary)
	authenticated.POST("/itineraries/:itineraryId/jobs", runItineraryFileJob)
	authenticated.GET("/itineraries/:itineraryId/jobs", getAllItineraryFileJobs)
	authenticated.GET("/itineraries/:itineraryId/jobs/:itineraryJobId", getItineraryJob)
	authenticated.GET("/itineraries/:itineraryId/jobs/:itineraryJobId/file", downloadItineraryJobFile)
	authenticated.PUT("/itineraries/:itineraryId/jobs/:itineraryJobId/stop", stopItineraryJob)
	authenticated.DELETE("/itineraries/:itineraryId/jobs/:itineraryJobId", deleteItineraryJob)

	api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
