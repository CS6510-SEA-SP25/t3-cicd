package StageService

import (
	"cicd/pipeci/backend/models"
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

type StageService struct {
	db *sql.DB
}

func NewStageService(db *sql.DB) *StageService {
	return &StageService{db: db}
}

// Query stage executions by input conditions
func (service *StageService) QueryStages(filters map[string]interface{}) ([]models.Stage, error) {
	var stages []models.Stage
	query := "SELECT * FROM Stages"
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
		return nil, fmt.Errorf("QueryStages: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stage models.Stage
		if err := rows.Scan(
			&stage.StageId, &stage.PipelineId, &stage.Name,
			&stage.Status, &stage.StartTime, &stage.EndTime,
		); err != nil {
			return nil, fmt.Errorf("QueryStages: %v", err)
		}
		stages = append(stages, stage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("QueryStages: %v", err)
	}

	return stages, nil
}
