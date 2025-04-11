package main

import (
	"cicd/pipeci/executor/cache"
	DockerService "cicd/pipeci/executor/containers/docker"
	"cicd/pipeci/executor/db"
	"cicd/pipeci/executor/storage"
	"cicd/pipeci/executor/types"
	"encoding/json"
	"flag"
	"log"

	"github.com/joho/godotenv"
)

// RabbitMQ Item
type QueueItem struct {
	Id         string                        `json:"id"`
	JobId      int                           `json:"jobId"`
	StageId    int                           `json:"stageId"`
	PipelineId int                           `json:"pipelineId"`
	Message    types.JobExecutor_RequestBody `json:"message"`
}

/* Process QueueItem received from message queue */
func processQueueItem(jsonInput string) error {
	// Parse the JSON input
	var queueItem QueueItem
	if err := json.Unmarshal([]byte(jsonInput), &queueItem); err != nil {
		log.Printf("[Executor] Error JSON Unmarshalling: %v", err)
		return err
	}

	log.Printf("[Executor] JSON parsed done.")

	// Execute the Docker service
	err := DockerService.Execute(
		queueItem.PipelineId,
		queueItem.StageId,
		queueItem.JobId,
		queueItem.Id,
		queueItem.Message.Job,
		queueItem.Message.Repository)
	if err != nil {
		log.Printf("[Executor] Error executing job: %v", err)
		return err
	}

	log.Printf("[Executor] QueueItem %s for Job %d completed successfully", queueItem.Id, queueItem.JobId)
	return nil
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("No .env file, use default env variables.")
	} else {
		log.Printf("Loading .env file.")
	}

	// Init database
	db.Init()

	// Init artifact storage
	storage.Init()

	// Init cache
	cache.Init()

	// Parse command-line flags
	jsonInput := flag.String("job", "", "Job JSON input")
	flag.Parse()

	// JSON passed via CLI flag
	if *jsonInput != "" {
		if err := processQueueItem(*jsonInput); err != nil {
			log.Fatalf("[Executor] Job failed: %v", err)
		}
	}

	log.Println("[Executor] Job execution complete.")
}
