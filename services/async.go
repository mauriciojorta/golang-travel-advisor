package services

import (
	"encoding/json"
	"os"
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

var NewAsyncqTaskQueue = func() AsyncqTaskQueueInterface {
	redisClientAddr := os.Getenv("REDIS_ADDR")
	if redisClientAddr == "" {
		log.Warn("REDIS_ADDR environment variable not set, using default address")
		redisClientAddr = "127.0.0.1:6379"
	}
	queueClient := asynq.NewClient(asynq.RedisClientOpt{Addr: redisClientAddr})
	return &AsyncqTaskQueue{
		Client: queueClient,
	}
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

	asyncTask := asynq.NewTask(TypeItineraryFileGeneration, asyncTaskPayloadJson, asynq.Timeout(10*time.Minute))

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: "127.0.0.1:6379"})
	defer client.Close()

	info, err := client.Enqueue(asyncTask)
	if err != nil {
		log.Errorf("could not enqueue itinerary file job task: %v", err)
		return nil, err
	}
	log.Debugf("enqueued itinerary file job task: id=%s queue=%s", info.ID, info.Queue)

	return &info.ID, nil

}
