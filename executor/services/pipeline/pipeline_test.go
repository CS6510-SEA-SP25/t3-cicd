package PipelineService

import (
	"cicd/pipeci/executor/models"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCreatePipeline(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewPipelineService(db)

	// Define the pipeline
	pipeline := models.Pipeline{
		Repository: "repo1",
		CommitHash: "abc123",
		IPAddress:  "192.168.1.1",
		Name:       "pipeline1",
		StageOrder: "",
		Status:     models.PENDING,
		StartTime:  time.Now(),
	}

	// Expect the exec and return a mock result
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO Pipelines")).
		WithArgs(
			pipeline.Repository,
			pipeline.CommitHash,
			pipeline.IPAddress,
			pipeline.Name,
			pipeline.StageOrder,
			pipeline.Status,
			sqlmock.AnyArg(), // Use AnyArg for the time argument
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the method under test
	pipelineID, err := service.CreatePipeline(pipeline)

	// Assert that no error occurred
	assert.NoError(t, err)

	// Assert that the expected pipeline ID was returned
	assert.Equal(t, 1, pipelineID)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdatePipelineStatusAndEndTime(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewPipelineService(db)

	// Define the pipeline ID and status
	pipelineID := 1
	status := models.SUCCESS

	// Expect the exec and return a mock result
	mock.ExpectExec(regexp.QuoteMeta("UPDATE Pipelines SET status = ?, end_time = ? WHERE pipeline_id = ?")).
		WithArgs(
			status,
			sqlmock.AnyArg(), // Use AnyArg for the end_time argument
			pipelineID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Call the method under test
	err = service.UpdatePipelineStatusAndEndTime(pipelineID, status)

	// Assert that no error occurred
	assert.NoError(t, err)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdatePipelineStatusAndEndTime_NotFound(t *testing.T) {
	// Create a new mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Create a new PipelineService with the mock database
	service := NewPipelineService(db)

	// Define the pipeline ID and status
	pipelineID := 1
	status := models.SUCCESS

	// Expect the exec and return a mock result with no rows affected
	mock.ExpectExec(regexp.QuoteMeta("UPDATE Pipelines SET status = ?, end_time = ? WHERE pipeline_id = ?")).
		WithArgs(status, sqlmock.AnyArg(), pipelineID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Call the method under test
	err = service.UpdatePipelineStatusAndEndTime(pipelineID, status)

	// Assert that an error occurred
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("UpdatePipelineStatusAndEndTime: no pipeline found with ID %d", pipelineID), err)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreatePipeline_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewPipelineService(db)

	// Define the pipeline
	pipeline := models.Pipeline{
		Repository: "repo1",
		CommitHash: "abc123",
		IPAddress:  "192.168.1.1",
		Name:       "pipeline1",
		StageOrder: "",
		Status:     models.PENDING,
		StartTime:  time.Now(),
	}

	// Expect the exec to return an error
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO Pipelines")).
		WithArgs(
			pipeline.Repository,
			pipeline.CommitHash,
			pipeline.IPAddress,
			pipeline.Name,
			pipeline.StageOrder,
			pipeline.Status,
			sqlmock.AnyArg(), // Use AnyArg for the start_time argument
		).
		WillReturnError(fmt.Errorf("database error"))

	// Call the method under test
	pipelineID, err := service.CreatePipeline(pipeline)

	// Assert that an error occurred
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CreatePipeline: database error")

	// Assert that no pipeline ID was returned
	assert.Equal(t, 0, pipelineID)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdatePipelineStatusAndEndTime_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewPipelineService(db)

	// Define the pipeline ID and status
	pipelineID := 1
	status := models.SUCCESS

	// Expect the exec to return an error
	mock.ExpectExec(regexp.QuoteMeta("UPDATE Pipelines SET status = ?, end_time = ? WHERE pipeline_id = ?")).
		WithArgs(
			status,
			sqlmock.AnyArg(), // Use AnyArg for the end_time argument
			pipelineID,
		).
		WillReturnError(fmt.Errorf("database error"))

	// Call the method under test
	err = service.UpdatePipelineStatusAndEndTime(pipelineID, status)

	// Assert that an error occurred
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UpdatePipelineStatusAndEndTime: database error")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdatePipelineStatusAndEndTime_NoRowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewPipelineService(db)

	// Define the pipeline ID and status
	pipelineID := 1
	status := models.SUCCESS

	// Expect the exec to return a result with no rows affected
	mock.ExpectExec(regexp.QuoteMeta("UPDATE Pipelines SET status = ?, end_time = ? WHERE pipeline_id = ?")).
		WithArgs(
			status,
			sqlmock.AnyArg(), // Use AnyArg for the end_time argument
			pipelineID,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Call the method under test
	err = service.UpdatePipelineStatusAndEndTime(pipelineID, status)

	// Assert that an error occurred
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UpdatePipelineStatusAndEndTime: no pipeline found with ID 1")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCleanUpTestPipelines(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	service := NewPipelineService(db)

	// Set up the expectation for the DELETE statement
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM Pipelines WHERE name LIKE ?")).
		WithArgs("test_%").
		WillReturnResult(sqlmock.NewResult(0, 0)) // Rows affected doesn't matter for this test

	// Call the function under test
	err = service.CleanUpTestPipelines()
	if err != nil {
		t.Errorf("expected no error, but got: %v", err)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
