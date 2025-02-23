package JobService

import (
	"cicd/pipeci/backend/models"
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

// Get a job report by id
func (service *JobService) GetJobsByStageId(stageId int) ([]models.Job, error) {
	var jobs []models.Job

	rows, err := service.db.Query("SELECT * FROM Jobs WHERE stage_id = ?", stageId)
	if err != nil {
		return nil, fmt.Errorf("GetJobsByStageId: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var job models.Job
		if err := rows.Scan(
			&job.JobId, &job.StageId, &job.Name, &job.Image, &job.Script,
			&job.Status, &job.StartTime, &job.EndTime,
		); err != nil {
			return nil, fmt.Errorf("GetJobsByStageId: %v", err)
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetJobsByStageId: %v", err)
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
