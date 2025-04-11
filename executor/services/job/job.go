package JobService

import (
	"cicd/pipeci/executor/models"
	"database/sql"
	"fmt"
	"time"
)

type JobService struct {
	db *sql.DB
}

func NewJobService(db *sql.DB) *JobService {
	return &JobService{db: db}
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
