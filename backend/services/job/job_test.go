package JobService

import (
	"cicd/pipeci/backend/models"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestQueryJobs_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewJobService(db)

	// Define filters
	filters := map[string]interface{}{
		"status": models.SUCCESS,
	}

	// Define expected rows
	rows := sqlmock.NewRows([]string{"job_id", "stage_id", "name", "image", "script", "status", "start_time", "end_time", "container_id"}).
		AddRow(1, 1, "job1", "", "", models.SUCCESS, time.Now(), time.Now(), "").
		AddRow(2, 2, "job2", "", "", models.SUCCESS, time.Now(), time.Now(), "")

	// Expect query with correct filters
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Jobs WHERE status = ? ORDER BY start_time")).
		WithArgs(models.SUCCESS).
		WillReturnRows(rows)

	// Call method under test
	jobs, err := service.QueryJobs(filters)

	// Assert no error occurred
	assert.NoError(t, err)

	// Assert expected jobs were returned
	assert.Equal(t, 2, len(jobs))
	assert.Equal(t, "job1", jobs[0].Name)
	assert.Equal(t, "job2", jobs[1].Name)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestQueryJobs_NoFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewJobService(db)

	// Define no filters
	filters := map[string]interface{}{}

	// Define expected rows
	rows := sqlmock.NewRows([]string{"job_id", "stage_id", "name", "image", "script", "status", "start_time", "end_time", "container_id"}).
		AddRow(1, 1, "job1", "", "", models.SUCCESS, time.Now(), time.Now(), "").
		AddRow(2, 2, "job2", "", "", models.SUCCESS, time.Now(), time.Now(), "")

	// Expect query with no filters
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Jobs ORDER BY start_time")).
		WillReturnRows(rows)

	// Call method under test
	jobs, err := service.QueryJobs(filters)

	// Assert no error occurred
	assert.NoError(t, err)

	// Assert expected jobs were returned
	assert.Equal(t, "job1", jobs[0].Name)
	assert.Equal(t, "job2", jobs[1].Name)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestQueryJobs_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewJobService(db)

	// Define filters
	filters := map[string]interface{}{
		"name": "job1",
	}

	// Expect query to return an error
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Jobs WHERE name = ? ORDER BY start_time")).
		WithArgs("job1").
		WillReturnError(fmt.Errorf("database error"))

	// Call method under test
	jobs, err := service.QueryJobs(filters)

	// Assert an error occurred
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "QueryJobs: database error")

	// Assert no jobs were returned
	assert.Nil(t, jobs)

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
