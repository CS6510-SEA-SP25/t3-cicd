package main

import (
	"cicd/cli/schema" // Import the config package
	"fmt"
	"log"
)

func main() {
	// Parse configuration file
	pipeline, err := schema.ParsePipelineConfiguration("/Users/nguyencanhminh/Downloads/code/Northeastern/s25/CS6510-ASD/t3-cicd/.pipeline/sample.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate configuration
	isPipelineValid, validateErr := pipeline.ValidateConfiguration()
	if validateErr != nil {
		log.Fatal(validateErr)
	} else {
		fmt.Println(isPipelineValid)
	}
}
