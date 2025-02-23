package apis

import (
	"cicd/pipeci/schema"
	"fmt"
	"log"
	"strings"
)

type ReportPastExecutionsLocal_RequestBody struct {
	Repository schema.Repository `json:"repository"`
	IPAddress  string            `json:"ip_address"`
}

/*
Report all local pipeline runs for all pipelines configured in the repository located in the working directory
Currently, local execution at ip_address 0.0.0.0
*/
func ReportPastExecutionsLocal(repository schema.Repository) error {
	var body = ReportPastExecutionsLocal_RequestBody{
		Repository: schema.Repository{
			Url: removeTokenFromURL(repository.Url),
		},
		IPAddress: "0.0.0.0",
	}

	res, err := PostRequest(BASE_URL+"/report/local", body)
	if err != nil {
		return fmt.Errorf("error local pipeline report: %w", err)
	}

	if res == nil {
		log.Println("No executions detected.")
		return nil
	}

	// Cast type interface{} -> []interface{}
	pipelines, ok := res.([]interface{})
	if !ok {
		return fmt.Errorf("error local pipeline report: type casting failed for API response")
	}

	// Log each pipeline's details using the function
	for _, pipeline := range pipelines {
		if err = logPipelineExecutionReport(pipeline); err != nil {
			return err
		}
	}
	return nil
}

// Print details of a single pipeline execution report
func logPipelineExecutionReport(input interface{}) error {
	pipeline, ok := input.(map[string]interface{})
	if !ok {
		return fmt.Errorf("error local pipeline report: type casting failed for one pipepline")
	}

	log.Println("Pipeline Details:")
	log.Printf("  Commit Hash: %s\n", pipeline["commit_hash"])
	log.Printf("  Name: %s\n", pipeline["name"])
	log.Printf("  Repository: %s\n", pipeline["repository"])
	log.Printf("  Pipeline ID: %v\n", pipeline["pipeline_id"])
	log.Printf("  Status: %s\n", pipeline["status"])
	log.Printf("  Start Time: %s\n", pipeline["start_time"])
	log.Printf("  End Time: %s\n", pipeline["end_time"])
	log.Printf("  IP Address: %s\n", pipeline["ip_address"])
	log.Printf("  Stage Order: %v\n", pipeline["stage_order"])
	log.Println("----------------------------------------")
	return nil
}

// Remove Personal Access Token from URL if exists
func removeTokenFromURL(url string) string {
	// Split the URL by "@"
	parts := strings.Split(url, "@")
	if len(parts) > 1 {
		// If there's a token, return the part after "@"
		return "https://" + parts[1]
	}
	return url
}
