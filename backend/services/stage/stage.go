package StageService

import (
	"cicd/pipeci/backend/models"
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

// Get stages in a pipeline
func (service *StageService) GetStagesByPipelineId(pipelineId int) ([]models.Stage, error) {
	var stages []models.Stage

	rows, err := service.db.Query("SELECT * FROM Stages WHERE pipeline_id = ?", pipelineId)
	if err != nil {
		return nil, fmt.Errorf("GetStagesByPipelineId: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stage models.Stage
		if err := rows.Scan(
			&stage.StageId, &stage.PipelineId, &stage.Name,
			&stage.Status, &stage.StartTime, &stage.EndTime,
		); err != nil {
			return nil, fmt.Errorf("GetStagesByPipelineId: %v", err)
		}
		stages = append(stages, stage)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetStagesByPipelineId: %v", err)
	}
	return stages, nil
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
