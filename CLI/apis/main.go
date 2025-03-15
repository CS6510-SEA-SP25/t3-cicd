package apis

import (
	"cicd/pipeci/schema"
	"fmt"
	"log"
)

// localhost
var BASE_URL string = "http://localhost:8080"

// minikube
// var BASE_URL string = "http://127.0.0.1:63723"

type ExecuteLocal_RequestBody struct {
	Pipeline   schema.PipelineConfiguration `json:"pipeline"`
	Repository schema.Repository            `json:"repository"`
}

/* Execute API on local env */
func ExecuteLocal(pipeline schema.PipelineConfiguration, repository schema.Repository) error {
	var body = &ExecuteLocal_RequestBody{Pipeline: pipeline, Repository: repository}

	res, err := PostRequest(BASE_URL+"/execute/local", body)
	if err != nil {
		return fmt.Errorf("error local pipeline execution: %w", err)
	}

	log.Printf("%v", res)
	return nil
}
