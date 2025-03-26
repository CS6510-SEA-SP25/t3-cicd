package main

import (
	"cicd/pipeci/worker/db"
	"cicd/pipeci/worker/storage"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

// Load .env before running tests
func TestMain(m *testing.M) {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Failed to load .env file. Ensure the file exists.")
	}

	log.Println(".env file loaded successfully")

	// Run tests
	os.Exit(m.Run())
}

// Test the `processTask` function with JSON input
func TestProcessTaskWithJsonInput(t *testing.T) {

	// Init database
	db.Init()

	// Init artifact storage
	storage.Init()

	// Set the JSON input as if it was passed via the --json flag
	jsonInput := `{"id":"id","message":{"pipeline":{"Version":{"Value":"v0","Location":{"Line":1,"Column":1}},"Pipeline":{"Value":{"Name":{"Value":"test_pipeline","Location":{"Line":5,"Column":3}}},"Location":{"Line":4,"Column":1}},"Stages":{"Value":{"build":{"Value":{"compile":{"Name":{"Value":"compile","Location":{"Line":13,"Column":5}},"Stage":{"Value":"build","Location":{"Line":14,"Column":5}},"Image":{"Value":"maven","Location":{"Line":15,"Column":5}},"Script":{"Value":["ls -la","mvn -v"],"Location":{"Line":16,"Column":5}},"Dependencies":null}},"Location":{"Line":9,"Column":5}}},"Location":{"Line":8,"Column":1}},"StageOrder":["build"],"ExecOrder":{"build":[["compile"]]}},"repository":{"Url":"https://github.com/CS6510-SEA-SP25/t3-cicd.git","CommitHash":"729d2fbb87d47fc770cdbe287e28bbe4ab2faa42"}}}`

	// Call the processTask function directly (this is testable now)
	err := processTask(jsonInput)

	// Assertions to verify that the task processing was done correctly
	assert.NoError(t, err, "Task should be processed without errors")
}

func TestProcessTask_Failed(t *testing.T) {
	// Init database
	db.Init()

	// Init artifact storage
	storage.Init()

	// Call the processTask function directly (this is testable now)
	err := processTask("")

	// Assertions to verify that the task processing was done correctly
	assert.Error(t, err, "Task should be processed without errors")
}
