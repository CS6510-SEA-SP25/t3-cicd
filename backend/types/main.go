package types

import (
	"cicd/pipeci/backend/models"
	"database/sql"
	"time"
)

// run
type ExecuteLocal_RequestBody struct {
	Pipeline   models.PipelineConfiguration `json:"pipeline"`
	Repository models.Repository            `json:"repository"`
}

// report
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

// status
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
