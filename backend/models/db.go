/*
 * Database schema
 */
package models

import (
	"database/sql"
	"time"
)

type ExecStatus string

// Execution status
const (
	SUCCESS  ExecStatus = "SUCCESS" // Execute successfully
	FAILED   ExecStatus = "FAILED"  // Execute failed
	CANCELED ExecStatus = "CANCELED"
	PENDING  ExecStatus = "PENDING"
)

// Pipeline's execution report
type Pipeline struct {
	PipelineId int          `json:"pipeline_id" db:"pipeline_id"`
	Repository string       `json:"repository" db:"repository"`
	CommitHash string       `json:"commit_hash" db:"commit_hash"`
	IPAddress  string       `json:"ip_address" db:"ip_address"`
	Name       string       `json:"name" db:"name"`
	StageOrder string       `json:"stage_order" db:"stage_order"`
	Status     ExecStatus   `json:"status" db:"status"`
	StartTime  time.Time    `json:"start_time" db:"start_time"`
	EndTime    sql.NullTime `json:"end_time" db:"end_time"`
}

// Stage's execution report
type Stage struct {
	StageId    int        `json:"stage_id" db:"stage_id"`
	PipelineId int        `json:"pipeline_id" db:"pipeline_id"`
	Name       string     `json:"name" db:"name"`
	Status     ExecStatus `json:"status" db:"status"`
	StartTime  time.Time  `json:"start_time" db:"start_time"`
	EndTime    time.Time  `json:"end_time" db:"end_time"`
}

// Job's execution report
type Job struct {
	JobId       int        `json:"job_id" db:"job_id"`
	StageId     int        `json:"stage_id" db:"stage_id"`
	Name        string     `json:"name" db:"name"`
	Image       string     `json:"image" db:"image"`
	Script      string     `json:"script" db:"script"`
	Status      ExecStatus `json:"status" db:"status"`
	StartTime   time.Time  `json:"start_time" db:"start_time"`
	EndTime     time.Time  `json:"end_time" db:"end_time"`
	ContainerId string     `json:"container_id" db:"container_id"`
}

// Dependencies
type Dependency struct {
	ParentId int `json:"parent_id" db:"parent_id"`
	ChildId  int `json:"child_id" db:"child_id"`
}
