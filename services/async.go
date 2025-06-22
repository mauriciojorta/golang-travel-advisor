package services

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/hibiken/asynq"
)

type AsyncqTaskQueueInterface interface {
	Close()
	EnqueueItineraryFileJob(itineraryTaskPayload ItineraryFileAsyncTaskPayload) (*string, error)
}

type AsyncqTaskQueue struct {
	Client *asynq.Client
}

type ItineraryFileTaskQueue interface {
	EnqueueItineraryFileJob(itineraryTaskPayload ItineraryFileAsyncTaskPayload) error
}

var NewAsyncqTaskQueue = func() (AsyncqTaskQueueInterface, error) {
	redisClientAddr := os.Getenv("REDIS_ADDR")
	redisPasswr := os.Getenv("REDIS_PASSWORD")
	if redisClientAddr == "" {
		log.Warn("REDIS_ADDR environment variable not set, using default address")
		redisClientAddr = "127.0.0.1:6379"
	}
	if redisPasswr == "" {
		errorMsg := "REDIS_PASSWORD environment variable is not set"
		log.Error(errorMsg)
		return nil, errors.New(errorMsg)
	}
	queueClient := asynq.NewClient(asynq.RedisClientOpt{Addr: redisClientAddr, Password: redisPasswr})
	return &AsyncqTaskQueue{
		Client: queueClient,
	}, nil
}

func (q *AsyncqTaskQueue) Close() {
	if q.Client != nil {
		err := q.Client.Close()
		if err != nil {
			log.Errorf("could not close asyncq client: %v", err)
		} else {
			log.Info("asyncq client closed successfully")
		}
	}
}

func (q *AsyncqTaskQueue) EnqueueItineraryFileJob(itineraryTaskPayload ItineraryFileAsyncTaskPayload) (*string, error) {

	asyncTaskPayloadJson, err := json.Marshal(itineraryTaskPayload)
	if err != nil {
		return nil, err
	}

	asyncTaskTimeoutStr := os.Getenv("ASYNC_TASK_TIMEOUT_MINUTES")
	var asyncTaskTimeoutMinutes int
	if asyncTaskTimeoutStr != "" {
		asyncTaskTimeoutMinutes, err = strconv.Atoi(asyncTaskTimeoutStr)
		if err != nil {
			return nil, err
		}
	} else {
		asyncTaskTimeoutMinutes = 10 // default timeout in minutes if not set
	}

	asyncTask := asynq.NewTask(TypeItineraryFileGeneration, asyncTaskPayloadJson, asynq.MaxRetry(0), asynq.Timeout(time.Duration(asyncTaskTimeoutMinutes)*time.Minute))

	info, err := q.Client.Enqueue(asyncTask)
	if err != nil {
		log.Errorf("could not enqueue itinerary file job task: %v", err)
		return nil, err
	}
	log.Debugf("enqueued itinerary file job task: id=%s queue=%s", info.ID, info.Queue)

	return &info.ID, nil

}
