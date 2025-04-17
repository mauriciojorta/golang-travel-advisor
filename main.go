package main

import (
	"example.com/travel-advisor/routes"
	"github.com/gin-gonic/gin"
)

func main() {

	server := gin.Default()

	routes.RegisterRoutes(server)

	server.Run(":8080") //localhost:8080

}
