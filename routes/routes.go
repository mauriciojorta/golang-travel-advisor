package routes

import (
	"example.com/travel-advisor/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(server *gin.Engine) {

	server.POST("/signup", signUp)
	server.POST("/login", login)

	authenticated := server.Group("/")
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

}
