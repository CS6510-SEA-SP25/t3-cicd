package JobService

import (
	"cicd/pipeci/backend/models"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetJobsByStageId(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewJobService(db)

	// Define the expected rows
	stageId := 1
	rows := sqlmock.NewRows([]string{"job_id", "stage_id", "name", "image", "script", "status", "start_time", "end_time"}).
		AddRow(1, stageId, "job1", "image1", "script1", models.PENDING, time.Now(), time.Now()).
		AddRow(2, stageId, "job2", "image2", "script2", models.SUCCESS, time.Now(), time.Now())

	// Expect the query and return the mock rows
	mock.ExpectQuery("SELECT \\* FROM Jobs WHERE stage_id = ?").
		WithArgs(stageId).
		WillReturnRows(rows)

	// Call the method under test
	jobs, err := service.GetJobsByStageId(stageId)

	// Assert that no error occurred
	assert.NoError(t, err)

	// Assert that the expected jobs were returned
	assert.Equal(t, 2, len(jobs))
	assert.Equal(t, "job1", jobs[0].Name)
	assert.Equal(t, "job2", jobs[1].Name)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateJob(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewJobService(db)

	// Define the job
	job := models.Job{
		StageId:     1,
		Name:        "job1",
		Image:       "image1",
		Script:      "script1",
		Status:      models.PENDING,
		StartTime:   time.Now(),
		ContainerId: "container1",
	}

	// Expect the exec and return a mock result
	mock.ExpectExec("INSERT INTO Jobs").
		WithArgs(
			job.StageId,
			job.Name,
			job.Image,
			job.Script,
			job.Status,
			sqlmock.AnyArg(), // Use AnyArg for the start_time argument
			job.ContainerId,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the method under test
	jobID, err := service.CreateJob(job)

	// Assert that no error occurred
	assert.NoError(t, err)

	// Assert that the expected job ID was returned
	assert.Equal(t, 1, jobID)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateJobStatusAndEndTime(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewJobService(db)

	// Define the job ID, container ID, and status
	jobID := 1
	containerId := "container1"
	status := models.SUCCESS

	// Expect the exec and return a mock result
	mock.ExpectExec("UPDATE Jobs SET container_id = \\?, status = \\?, end_time = \\? WHERE job_id = \\?").
		WithArgs(
			containerId,
			status,
			sqlmock.AnyArg(), // Use AnyArg for the end_time argument
			jobID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Call the method under test
	err = service.UpdateJobStatusAndEndTime(jobID, containerId, status)

	// Assert that no error occurred
	assert.NoError(t, err)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateJobStatusAndEndTime_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewJobService(db)

	// Define the job ID, container ID, and status
	jobID := 1
	containerId := "container1"
	status := models.SUCCESS

	// Expect the exec and return a mock result with no rows affected
	mock.ExpectExec("UPDATE Jobs SET container_id = \\?, status = \\?, end_time = \\? WHERE job_id = \\?").
		WithArgs(
			containerId,
			status,
			sqlmock.AnyArg(), // Use AnyArg for the end_time argument
			jobID,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Call the method under test
	err = service.UpdateJobStatusAndEndTime(jobID, containerId, status)

	// Assert that an error occurred
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("UpdateJobStatusAndEndTime: no job found with ID %d", jobID), err)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// Test on failed scenarios
func TestGetJobsByStageId_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewJobService(db)

	// Define the stage ID
	stageId := 1

	// Expect the query to return an error
	mock.ExpectQuery("SELECT \\* FROM Jobs WHERE stage_id = ?").
		WithArgs(stageId).
		WillReturnError(fmt.Errorf("database error"))

	// Call the method under test
	jobs, err := service.GetJobsByStageId(stageId)

	// Assert that an error occurred
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GetJobsByStageId: database error")

	// Assert that no jobs were returned
	assert.Nil(t, jobs)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateJob_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewJobService(db)

	// Define the job
	job := models.Job{
		StageId:     1,
		Name:        "job1",
		Image:       "image1",
		Script:      "script1",
		Status:      models.PENDING,
		StartTime:   time.Now(),
		ContainerId: "container1",
	}

	// Expect the exec to return an error
	mock.ExpectExec("INSERT INTO Jobs").
		WithArgs(
			job.StageId,
			job.Name,
			job.Image,
			job.Script,
			job.Status,
			sqlmock.AnyArg(), // Use AnyArg for the start_time argument
			job.ContainerId,
		).
		WillReturnError(fmt.Errorf("database error"))

	// Call the method under test
	jobID, err := service.CreateJob(job)

	// Assert that an error occurred
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CreateJob: database error")

	// Assert that no job ID was returned
	assert.Equal(t, 0, jobID)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateJobStatusAndEndTime_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewJobService(db)

	// Define the job ID, container ID, and status
	jobID := 1
	containerId := "container1"
	status := models.PENDING

	// Expect the exec to return an error
	mock.ExpectExec("UPDATE Jobs SET container_id = \\?, status = \\?, end_time = \\? WHERE job_id = \\?").
		WithArgs(
			containerId,
			status,
			sqlmock.AnyArg(), // Use AnyArg for the end_time argument
			jobID,
		).
		WillReturnError(fmt.Errorf("database error"))

	// Call the method under test
	err = service.UpdateJobStatusAndEndTime(jobID, containerId, status)

	// Assert that an error occurred
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UpdateJobStatusAndEndTime: database error")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateJobStatusAndEndTime_NoRowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewJobService(db)

	// Define the job ID, container ID, and status
	jobID := 1
	containerId := "container1"
	status := models.SUCCESS

	// Expect the exec to return a result with no rows affected
	mock.ExpectExec("UPDATE Jobs SET container_id = \\?, status = \\?, end_time = \\? WHERE job_id = \\?").
		WithArgs(
			containerId,
			status,
			sqlmock.AnyArg(), // Use AnyArg for the end_time argument
			jobID,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Call the method under test
	err = service.UpdateJobStatusAndEndTime(jobID, containerId, status)

	// Assert that an error occurred
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UpdateJobStatusAndEndTime: no job found with ID 1")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
