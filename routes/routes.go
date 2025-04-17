package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(server *gin.Engine) {

	server.GET("/hello", helloWorld)

}

func helloWorld(context *gin.Context) {

	context.JSON(http.StatusCreated, gin.H{"message": "Hello World."})

}
