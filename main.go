package main

import (
	"example.com/travel-advisor/apis"
	"example.com/travel-advisor/db"
	"example.com/travel-advisor/logger"
	"example.com/travel-advisor/routes"
	"example.com/travel-advisor/utils"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file : %v", err)
	}
	logger.SetupLogger()

	log.Info("Environment variables loaded")

	err = utils.InitJwtSecretKey()
	if err != nil {
		log.Fatalf("Error initializing JWT secret key: %v", err)
	}

	err = apis.InitLlmClient()
	if err != nil {
		log.Fatalf("Error initializing LLM client : %v", err)
	}

	log.Info("Starting Travel Advisor API")

	db.InitDB()
	defer db.DB.Close()
	log.Info("Database connection established")

	server := gin.Default()

	routes.RegisterRoutes(server)

	server.Run(":8080") //localhost:8080

}
