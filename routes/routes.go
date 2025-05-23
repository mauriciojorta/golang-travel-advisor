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
	authenticated.GET("/itineraries", getOwnersItineraries)
	authenticated.GET("/itineraries/:id", getItinerary)

}
