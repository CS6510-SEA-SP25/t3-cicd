package StageService

import (
	"cicd/pipeci/worker/models"
	"database/sql"
	"fmt"
	"time"
)

type StageService struct {
	db *sql.DB
}

func NewStageService(db *sql.DB) *StageService {
	return &StageService{db: db}
}

// Create a stage report
func (service *StageService) CreateStage(stage models.Stage) (int, error) {
	stage.StartTime = time.Now()

	result, err := service.db.Exec(
		"INSERT INTO Stages (pipeline_id, name, status, start_time) VALUES (?, ?, ?, ?)",
		stage.PipelineId, stage.Name, stage.Status, stage.StartTime,
	)
	if err != nil {
		return 0, fmt.Errorf("CreateStage: %v", err)
	}

	// Get the Id of the newly inserted stage
	stageId, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("CreateStage: %v", err)
	}

	return int(stageId), nil
}

// Update stage status and end_time
func (service *StageService) UpdateStageStatusAndEndTime(stageID int, status models.ExecStatus) error {
	var endTime time.Time = time.Now()
	result, err := service.db.Exec(
		"UPDATE Stages SET status = ?, end_time = ? WHERE stage_id = ?",
		status, endTime, stageID,
	)
	if err != nil {
		return fmt.Errorf("UpdateStageStatusAndEndTime: %v", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("UpdateStageStatusAndEndTime: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("UpdateStageStatusAndEndTime: no stage found with ID %d", stageID)
	}

	return nil
}
