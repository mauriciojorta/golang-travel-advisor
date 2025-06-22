package main

import (
	"os"
	"strconv"
	"time"

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
	redisPasswr := os.Getenv("REDIS_PASSWORD")
	if redisClientAddr == "" {
		log.Warn("REDIS_ADDR environment variable not set, using default address")
		redisClientAddr = "127.0.0.1:6379"
	}
	if redisPasswr == "" {
		log.Fatal("REDIS_PASSWORD environment variable is not set")
	}
	asyncqSrv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisClientAddr, Password: redisPasswr},
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

	// Start background cleanup process for dead itinerary file jobs
	startDeadItineraryFileJobsCleanup()

	server.Run(":8080") //localhost:8080

}

// This function triggers the full delection of "dead" itinerary file jobs (those marked in status 'deleted') through a periodic timer
func startDeadItineraryFileJobsCleanup() {
	// Initialize a ticker to run every 'n' minutes according to the value defined in the environment variables (10 minutes if absent there)
	intervalInMinutesStr := os.Getenv("DEAD_ITINERARY_FILE_JOBS_TIMER_MINUTES_INTERVAL")
	intervalInMinutes := 10
	var err error
	if intervalInMinutesStr != "" {
		intervalInMinutes, err = strconv.Atoi(intervalInMinutesStr)
		if err != nil {
			log.Errorf("The format of DEAD_ITINERARY_FILE_JOBS_TIMER_MINUTES_INTERVAL environment property is incorrect: %v", err)
			return
		}
	}

	ticker := time.NewTicker(time.Duration(intervalInMinutes) * time.Minute)
	go func() {

		for range ticker.C {
			jobsService := services.GetItineraryFileJobService()

			fetchLimitStr := os.Getenv("DEAD_ITINERARY_FILE_JOBS_FETCH_LIMIT")
			fetchLimit := 10
			if fetchLimitStr != "" {
				fetchLimit, err = strconv.Atoi(fetchLimitStr)
				if err != nil {
					log.Errorf("The format of DEAD_ITINERARY_FILE_JOBS_FETCH_LIMIT environment property is incorrect: %v", err)
					return
				}
			}

			log.Info("Running periodic cleanup of deleted itinerary files")
			err := jobsService.DeleteDeadJobs(fetchLimit)
			if err != nil {
				log.Errorf("Error during periodic cleanup: %v", err)
			}
			log.Info("Periodic cleanup of deleted itinerary files was successful")

		}
	}()
}
