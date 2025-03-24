package JobService

import (
	"cicd/pipeci/backend/models"
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

type JobService struct {
	db *sql.DB
}

func NewJobService(db *sql.DB) *JobService {
	return &JobService{db: db}
}

// Query job executions by input conditions
func (service *JobService) QueryJobs(filters map[string]interface{}) ([]models.Job, error) {
	var jobs []models.Job
	query := "SELECT * FROM Jobs"
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
		return nil, fmt.Errorf("QueryJobs: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var job models.Job
		if err := rows.Scan(
			&job.JobId, &job.StageId, &job.Name,
			&job.Image, &job.Script, &job.Status,
			&job.StartTime, &job.EndTime, &job.ContainerId,
		); err != nil {
			return nil, fmt.Errorf("QueryJobs: %v", err)
		}
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("QueryJobs: %v", err)
	}

	return jobs, nil
}
