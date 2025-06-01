package main

import (
	"os"

	"example.com/travel-advisor/apis"
	"example.com/travel-advisor/db"
	"example.com/travel-advisor/logger"
	"example.com/travel-advisor/routes"
	"example.com/travel-advisor/services"
	"example.com/travel-advisor/utils"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
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

	// Initialize Asyncq Server
	redisClientAddr := os.Getenv("REDIS_ADDR")
	if redisClientAddr == "" {
		log.Warn("REDIS_ADDR environment variable not set, using default address")
		redisClientAddr = "127.0.0.1:6379"
	}
	asyncqSrv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisClientAddr},
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
		},
	)

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(services.TypeItineraryFileGeneration, services.HandleItineraryFileJob)

	go func() {
		if err := asyncqSrv.Run(mux); err != nil {
			log.Fatalf("could not run asyncq server: %v", err)
		}
	}()

	// Initialize Gin server
	server := gin.Default()

	routes.RegisterRoutes(server)

	server.Run(":8080") //localhost:8080

}
