package apis

import (
	"cicd/pipeci/schema"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

type ReportPastExecutionsLocal_CurrentRepo_RequestBody struct {
	Repository   schema.Repository `json:"repository"`
	IPAddress    string            `json:"ip_address"`
	PipelineName string            `json:"pipeline_name"`
	StageName    string            `json:"stage_name"`
	JobName      string            `json:"job_name"`
	RunCounter   int               `json:"run_counter"`
}

type Report_ResponseBody struct {
	Id        int          `json:"id"`
	Name      string       `json:"name"`
	StartTime time.Time    `json:"start_time"`
	EndTime   sql.NullTime `json:"end_time"`
	Status    string       `json:"status"`
	// RunCounter int          `json:"run_counter"`
}

/*
Report ALL local pipeline runs for all pipelines configured in the repository located in the working directory
Currently, local execution at ip_address 0.0.0.0
*/
func ReportPastExecutionsLocal_CurrentRepo(repository schema.Repository) error {
	var body = ReportPastExecutionsLocal_CurrentRepo_RequestBody{
		Repository: schema.Repository{
			Url: removeTokenFromURL(repository.Url),
		},
		IPAddress: "0.0.0.0",
	}

	rawData, err := PostRequest(BASE_URL+"/report/local", body)

	if err != nil {
		return fmt.Errorf("error local pipeline report: %w", err)
	}

	if rawData == nil {
		log.Println("No executions detected.")
		return nil
	}

	return generateReports(rawData)
}

/*
Returns the list of all executions by query conditions.
Currently,
- Local execution at ip_address 0.0.0.0
- only matching pipelineName, more to come
*/
func ReportPastExecutionsLocal_ByCondition(repository schema.Repository, pipelineName, stageName, jobName string, runCounter int) error {
	var body = ReportPastExecutionsLocal_CurrentRepo_RequestBody{
		Repository: schema.Repository{
			Url:        removeTokenFromURL(repository.Url),
			CommitHash: repository.CommitHash,
		},
		IPAddress:    "0.0.0.0",
		PipelineName: strings.TrimSpace(pipelineName),
		StageName:    strings.TrimSpace(stageName),
		JobName:      strings.TrimSpace(jobName),
		RunCounter:   runCounter,
	}

	rawData, err := PostRequest(BASE_URL+"/report/local/query", body)
	if err != nil {
		return fmt.Errorf("error local executions report: %w", err)
	}

	if rawData == nil {
		log.Println("No executions detected.")
		return nil
	}

	return generateReports(rawData)
}

/* Generate execution reports */
func generateReports(rawData interface{}) error {
	reports, err := convertToReports(rawData)

	if err != nil {
		return fmt.Errorf("error local pipeline report: type casting failed for API response, %w", err)
	}

	log.Println("CI/CD Execution details:")
	// Log each pipeline's details using the function
	for _, report := range reports {
		if err = logExecutionReport(report); err != nil {
			return err
		}
	}
	return nil
}

// Print details of a single pipeline execution report
func logExecutionReport(input Report_ResponseBody) error {
	// Calculate duration if both times exist
	duration := ""
	if !input.StartTime.IsZero() && !input.EndTime.Time.IsZero() {
		duration = fmt.Sprintf("(Duration: %v)", input.EndTime.Time.Sub(input.StartTime).Round(time.Second))
	}

	// Colorize status
	status := fmtStatus(input.Status)

	// Create the formatted output
	output := fmt.Sprintf(`
╔═══════════════════════════════════════════════════
║ 🚀 PIPELINE EXECUTION REPORT
╟───────────────────────────────────────────────────
║ 📛 Name:       %s
║ 🆔 ID:         %v
║ 🏷️ Status:     %s
║ 🕒 Start Time: %s
║ 🕓 End Time:   %s %s
╚═══════════════════════════════════════════════════`,
		input.Name,
		input.Id,
		status,
		input.StartTime.Format("2006-01-02 15:04:05 MST"),
		input.EndTime.Time.Format("2006-01-02 15:04:05 MST"),
		duration,
	)

	log.Println(output)
	return nil
}

// Helper function to colorize status
func fmtStatus(status string) string {
	switch strings.ToUpper(status) {
	case "SUCCESS":
		return "\033[32m✔ " + status + "\033[0m" // Green check
	case "FAILED":
		return "\033[31m✘ " + status + "\033[0m" // Red X
	case "RUNNING":
		return "\033[33m↻ " + status + "\033[0m" // Yellow arrow
	default:
		return "➔ " + status // Default arrow
	}
}

// Convert interface{} to []Report_ResponseBody
func convertToReports(data interface{}) ([]Report_ResponseBody, error) {
	var reports = make([]Report_ResponseBody, 0)
	// Type assert to []interface{}
	dataList, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to cast to []interface{}")
	}

	for _, item := range dataList {
		// Type assert to map[string]interface{}
		record, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to cast item to map[string]interface{}")
		}

		// Convert map to Report struct
		report, err := mapToReport(record)
		if err != nil {
			return nil, fmt.Errorf("error converting map to Report: %v", err)
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// Convert map[string]interface{} to Report_ResponseBody
func mapToReport(data map[string]interface{}) (Report_ResponseBody, error) {
	var report Report_ResponseBody
	var err error

	// Extract Id
	if id, ok := data["id"].(float64); ok {
		report.Id = int(id)
	} else {
		return report, fmt.Errorf("invalid id field")
	}

	// Extract Name
	if name, ok := data["name"].(string); ok {
		report.Name = name
	} else {
		return report, fmt.Errorf("invalid name field")
	}

	// Extract StartTime
	if startTimeStr, ok := data["start_time"].(string); ok {
		report.StartTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return report, fmt.Errorf("invalid start_time format: %v", err)
		}
	} else {
		return report, fmt.Errorf("invalid start_time field")
	}

	// Extract EndTime (handling sql.NullTime)
	if endTimeMap, ok := data["end_time"].(map[string]interface{}); ok {
		if endTimeStr, exists := endTimeMap["Time"].(string); exists {
			t, err := time.Parse(time.RFC3339, endTimeStr)
			if err != nil {
				return report, fmt.Errorf("invalid end_time format: %v", err)
			}
			report.EndTime = sql.NullTime{Time: t, Valid: true}
		}
	}

	// Extract Status
	if status, ok := data["status"].(string); ok {
		report.Status = status
	} else {
		return report, fmt.Errorf("invalid status field")
	}

	return report, nil
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
