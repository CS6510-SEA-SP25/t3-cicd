package main

import (
	DockerService "cicd/pipeci/worker/containers/docker"
	"cicd/pipeci/worker/db"
	"cicd/pipeci/worker/storage"
	"cicd/pipeci/worker/types"
	"encoding/json"
	"flag"
	"log"

	"github.com/joho/godotenv"
)

// Task struct
type Task struct {
	Id      string                         `json:"id"`
	Message types.ExecuteLocal_RequestBody `json:"message"`
}

/* Process task received from message queue */
func processTask(jsonInput string) error {
	// Parse the JSON input
	var task Task
	if err := json.Unmarshal([]byte(jsonInput), &task); err != nil {
		log.Printf("[Worker] Error JSON Unmarshalling: %v", err)
		return err
	}

	log.Printf("[Worker] JSON parsed done.")

	// Execute the Docker service
	err := DockerService.Execute(task.Message.Pipeline, task.Message.Repository)
	if err != nil {
		log.Printf("[Worker] Error executing pipeline: %v", err)
		return err
	}

	log.Printf("[Worker] Task %s completed successfully", task.Id)
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

	// Parse command-line flags
	jsonInput := flag.String("task", "", "Task JSON input")
	flag.Parse()

	// JSON passed via CLI flag
	if *jsonInput != "" {
		if err := processTask(*jsonInput); err != nil {
			log.Fatalf("[Worker] Task failed: %v", err)
		}
	}

	log.Println("[Worker] Pipeline execution complete.")
}
