package PipelineService

import (
	"cicd/pipeci/executor/models"
	"database/sql"
	"fmt"
	"time"
)

type PipelineService struct {
	db *sql.DB
}

func NewPipelineService(db *sql.DB) *PipelineService {
	return &PipelineService{db: db}
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
