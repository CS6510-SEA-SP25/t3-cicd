package PipelineService

import (
	"cicd/pipeci/backend/models"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type PipelineService struct {
	db *sql.DB
}

func NewPipelineService(db *sql.DB) *PipelineService {
	return &PipelineService{db: db}
}

// Get all pipelines
func (service *PipelineService) GetPipelines() ([]models.Pipeline, error) {
	var pipelines []models.Pipeline

	rows, err := service.db.Query("SELECT * FROM Pipelines")
	if err != nil {
		return nil, fmt.Errorf("GetPipelines: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var pipeline models.Pipeline
		if err := rows.Scan(
			&pipeline.PipelineId, &pipeline.Repository, &pipeline.CommitHash, &pipeline.IPAddress,
			&pipeline.Name, &pipeline.StageOrder,
			&pipeline.Status, &pipeline.StartTime, &pipeline.EndTime,
		); err != nil {
			return nil, fmt.Errorf("GetPipelines: %v", err)
		}
		pipelines = append(pipelines, pipeline)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetPipelines: %v", err)
	}
	return pipelines, nil
}

// Create a new pipeline report
func (service *PipelineService) CreatePipeline(pipeline models.Pipeline) (int, error) {
	pipeline.StartTime = time.Now()

	result, err := service.db.Exec(
		"INSERT INTO Pipelines (repository, commit_hash, ip_address, name, stage_order, status, start_time) VALUES (?, ?, ?, ?, ?, ?, ?)",
		pipeline.Repository, pipeline.CommitHash, pipeline.IPAddress, pipeline.Name, pipeline.StageOrder, pipeline.Status, pipeline.StartTime,
	)
	if err != nil {
		return 0, fmt.Errorf("CreatePipeline: %v", err)
	}

	// Get the Id of the newly inserted pipeline
	pipelineId, err := result.LastInsertId()

	if err != nil {
		return 0, fmt.Errorf("CreatePipeline: %v", err)
	}

	return int(pipelineId), nil
}

// Update pipeline status and end_time
func (service *PipelineService) UpdatePipelineStatusAndEndTime(pipelineID int, status models.ExecStatus) error {
	var endTime time.Time = time.Now()
	result, err := service.db.Exec(
		"UPDATE Pipelines SET status = ?, end_time = ? WHERE pipeline_id = ?",
		status, endTime, pipelineID,
	)
	if err != nil {
		return fmt.Errorf("UpdatePipelineStatusAndEndTime: %v", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("UpdatePipelineStatusAndEndTime: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("UpdatePipelineStatusAndEndTime: no pipeline found with ID %d", pipelineID)
	}

	return nil
}

// Query pipeline executions by input conditions
func (service *PipelineService) QueryPipelines(filters map[string]interface{}) ([]models.Pipeline, error) {
	var pipelines []models.Pipeline
	query := "SELECT * FROM Pipelines"
	args := make([]interface{}, 0)
	conditions := make([]string, 0)

	// Build the conditions
	for key, value := range filters {
		conditions = append(conditions, fmt.Sprintf("%s = ?", key))
		args = append(args, value)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Execute
	rows, err := service.db.Query(query+" ORDER BY start_time", args...)
	if err != nil {
		return nil, fmt.Errorf("QueryPipelines: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pipeline models.Pipeline
		if err := rows.Scan(
			&pipeline.PipelineId, &pipeline.Repository, &pipeline.CommitHash, &pipeline.IPAddress,
			&pipeline.Name, &pipeline.StageOrder,
			&pipeline.Status, &pipeline.StartTime, &pipeline.EndTime,
		); err != nil {
			return nil, fmt.Errorf("QueryPipelines: %v", err)
		}
		pipelines = append(pipelines, pipeline)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("QueryPipelines: %v", err)
	}

	return pipelines, nil
}

// Deletes all pipelines with names starting with "test_"
func (service *PipelineService) CleanUpTestPipelines() error {
	result, err := service.db.Exec("DELETE FROM Pipelines WHERE name LIKE ?", "test_%")
	if err != nil {
		return fmt.Errorf("CleanUpTestPipelines: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("CleanUpTestPipelines: %v", err)
	}
	fmt.Printf("Clean %#v rows after tests.", rowsAffected)
	return nil
}
