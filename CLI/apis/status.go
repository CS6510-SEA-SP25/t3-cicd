package apis

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type RequestExecutionStatus_RequestBody struct {
	ExecutionId string `json:"execution_id"`
}

type RequestExecutionStatus_ResponseBody struct {
	Pipeline PipelineExecutionStatus         `json:"pipeline"`
	Stages   map[string]StageExecutionStatus `json:"stages"`
}

type PipelineExecutionStatus struct {
	PipelineId int    `json:"pipeline_id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	StageOrder string `json:"stage_order"`
}

type StageExecutionStatus struct {
	StageId int                  `json:"stage_id"`
	Name    string               `json:"name"`
	Status  string               `json:"status"`
	Jobs    []JobExecutionStatus `json:"jobs"`
}

type JobExecutionStatus struct { // * put OrderBy created_at when query status
	JobId  int    `json:"job_id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// Convert interface{} to RequestExecutionStatus_ResponseBody
func convertToPipelineExecStatus(rawData interface{}) (RequestExecutionStatus_ResponseBody, error) {
	var response = RequestExecutionStatus_ResponseBody{}
	jsonBytes, err := json.Marshal(rawData)
	if err != nil {
		return response, fmt.Errorf("error converting map to Report: %v", err)
	}
	json.Unmarshal(jsonBytes, &response)
	return response, nil
}

// Print pipeline execution status
func printExecutionStatus(response RequestExecutionStatus_ResponseBody) {
	// Print pipeline header
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("🚀 PIPELINE: %s (ID: %d)\n", response.Pipeline.Name, response.Pipeline.PipelineId)
	fmt.Printf("   Status: %s\t\tStage Order: %s\n", colorStatus(response.Pipeline.Status), response.Pipeline.StageOrder)
	fmt.Println(strings.Repeat("─", 60))

	// Print each stage
	for _, stage := range response.Stages {
		fmt.Printf("\n📦 STAGE: %s (ID: %d)\n", stage.Name, stage.StageId)
		fmt.Printf("   Status: %s\n", colorStatus(stage.Status))

		// Print jobs table
		fmt.Println("   ┌──────────┬──────────────┬────────────┐")
		fmt.Println("   │   Job ID │ Name         │ Status     │")
		fmt.Println("   ├──────────┼──────────────┼────────────┤")
		for _, job := range stage.Jobs {
			fmt.Printf("   │ %8d │ %-12s │ %-19s │ \n",
				job.JobId,
				job.Name,
				colorStatus(job.Status))
		}
		fmt.Println("   └──────────┴──────────────┴────────────┘")
	}
	fmt.Println(strings.Repeat("═", 60))
}

// Helper function to add color to status
func colorStatus(status string) string {
	switch strings.ToUpper(status) {
	case "SUCCESS":
		return "\033[32m" + status + "\033[0m" // Green
	case "FAILED":
		return "\033[31m" + status + "\033[0m" // Red
	case "PENDING":
		return "\033[33m" + status + "\033[0m" // Yellow
	default:
		return "\033[34m" + status + "\033[0m" // Blue
	}
}

/* Generate pipeline execution status report */
func generateStatusReport(rawData interface{}) error {
	response, err := convertToPipelineExecStatus(rawData)
	if err != nil {
		return fmt.Errorf("generateStatusReport %w", err)
	}

	printExecutionStatus(response)
	return nil
}

/* Get execution status of a pipeline */
func GetExecutionStatus(execId string) error {
	var body = RequestExecutionStatus_RequestBody{
		ExecutionId: execId,
	}

	rawData, err := PostRequest(BASE_URL+"/status", body)
	if err != nil {
		return fmt.Errorf("error local pipeline report: %w", err)
	}
	if rawData == nil {
		log.Println("No executions detected.")
		return nil
	}

	return generateStatusReport(rawData)
}
