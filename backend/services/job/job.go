package JobService

import (
	"cicd/pipeci/backend/models"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"
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

// Create a new job report
func (service *JobService) CreateJob(job models.Job) (int, error) {
	// Set the start_time to the current timestamp
	job.StartTime = time.Now()

	result, err := service.db.Exec(
		"INSERT INTO Jobs (stage_id, name, image, script, status, start_time, container_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		job.StageId, job.Name, job.Image, job.Script, job.Status, job.StartTime, job.ContainerId,
	)
	if err != nil {
		return 0, fmt.Errorf("CreateJob: %v", err)
	}

	// Get the ID of the newly inserted job
	jobID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("CreateJob: %v", err)
	}

	return int(jobID), nil
}

// Update job status and end_time
func (service *JobService) UpdateJobStatusAndEndTime(jobID int, containerId string, status models.ExecStatus) error {
	var endTime time.Time = time.Now()
	result, err := service.db.Exec(
		"UPDATE Jobs SET container_id = ?, status = ?, end_time = ? WHERE job_id = ?",
		containerId, status, endTime, jobID,
	)
	if err != nil {
		return fmt.Errorf("UpdateJobStatusAndEndTime: %v", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("UpdateJobStatusAndEndTime: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("UpdateJobStatusAndEndTime: no job found with ID %d", jobID)
	}

	return nil
}
