package PipelineService

import (
	"cicd/pipeci/backend/models"
	"database/sql"
	"fmt"
	"sort"
	"strings"
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

// Query pipeline executions by input conditions
func (service *PipelineService) QueryPipelines(filters map[string]interface{}) ([]models.Pipeline, error) {
	var pipelines []models.Pipeline
	query := "SELECT * FROM Pipelines"
	args := make([]interface{}, 0)

	// Sort keys for deterministic query order
	keys := make([]string, 0, len(filters))
	for key := range filters {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Build the conditions
	conditions := make([]string, 0)
	for _, key := range keys {
		value := filters[key]
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
