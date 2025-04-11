package types

import (
	"cicd/pipeci/executor/models"
	"database/sql"
	"time"
)

type JobExecutor_RequestBody struct {
	Job        models.JobConfiguration `json:"job"`
	Repository models.Repository       `json:"repository"`
}

type ReportPastExecutionsLocal_CurrentRepo_RequestBody struct {
	Repository   models.Repository `json:"repository"`
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
