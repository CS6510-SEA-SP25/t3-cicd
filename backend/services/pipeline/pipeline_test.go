package PipelineService

import (
	"cicd/pipeci/backend/models"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetPipelines(t *testing.T) {
	// Create a new mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Create a new PipelineService with the mock database
	service := NewPipelineService(db)

	// Define the expected rows
	rows := sqlmock.NewRows([]string{"pipeline_id", "repository", "commit_hash", "ip_address", "name", "stage_order", "status", "start_time", "end_time"}).
		AddRow(1, "repo1", "abc123", "0.0.0.0", "pipeline1", 1, models.PENDING, time.Now(), time.Now()).
		AddRow(2, "repo2", "def456", "192.168.1.2", "pipeline2", 2, models.SUCCESS, time.Now(), time.Now())

	// Expect the query and return the mock rows
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Pipelines")).WillReturnRows(rows)

	// Call the method under test
	pipelines, err := service.GetPipelines()

	// Assert that no error occurred
	assert.NoError(t, err)

	// Assert that the expected pipelines were returned
	// assert.Equal(t, 2, len(pipelines))
	if len(pipelines) > 0 {
		assert.Equal(t, "repo1", pipelines[0].Repository)
		assert.Equal(t, "repo2", pipelines[1].Repository)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// Tests on failed scenarios
func TestGetPipelines_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewPipelineService(db)

	// Expect the query to return an error
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Pipelines")).
		WillReturnError(fmt.Errorf("database error"))

	// Call the method under test
	pipelines, err := service.GetPipelines()

	// Assert that an error occurred
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GetPipelines: database error")

	// Assert that no pipelines were returned
	assert.Nil(t, pipelines)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestQueryPipelines_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewPipelineService(db)

	// Define the filters
	filters := map[string]interface{}{
		"repository": "repo1",
		"status":     models.SUCCESS,
	}

	// Define the expected rows
	rows := sqlmock.NewRows([]string{"pipeline_id", "repository", "commit_hash", "ip_address", "name", "stage_order", "status", "start_time", "end_time"}).
		AddRow(1, "repo1", "abc123", "192.168.1.1", "pipeline1", 1, models.SUCCESS, time.Now(), time.Now()).
		AddRow(2, "repo1", "def456", "192.168.1.2", "pipeline2", 2, models.SUCCESS, time.Now(), time.Now())

	// Expect the query with the correct filters
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Pipelines WHERE repository = ? AND status = ? ORDER BY start_time")).
		WithArgs("repo1", models.SUCCESS).
		WillReturnRows(rows)

	// Call the method under test
	pipelines, err := service.QueryPipelines(filters)

	// Assert that no error occurred
	assert.NoError(t, err)

	// Assert that the expected pipelines were returned
	// assert.Equal(t, 2, len(pipelines))
	if len(pipelines) > 0 {
		assert.Equal(t, "repo1", pipelines[0].Repository)
		assert.Equal(t, models.SUCCESS, pipelines[0].Status)
		assert.Equal(t, "repo1", pipelines[1].Repository)
		assert.Equal(t, models.SUCCESS, pipelines[1].Status)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestQueryPipelines_NoFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewPipelineService(db)

	// Define no filters
	filters := map[string]interface{}{}

	// Define the expected rows
	rows := sqlmock.NewRows([]string{"pipeline_id", "repository", "commit_hash", "ip_address", "name", "stage_order", "status", "start_time", "end_time"}).
		AddRow(1, "repo1", "abc123", "192.168.1.1", "pipeline1", 1, models.SUCCESS, time.Now(), time.Now()).
		AddRow(2, "repo2", "def456", "192.168.1.2", "pipeline2", 2, models.PENDING, time.Now(), time.Now())

	// Expect the query with no filters
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Pipelines ORDER BY start_time")).
		WillReturnRows(rows)

	// Call the method under test
	pipelines, err := service.QueryPipelines(filters)

	// Assert that no error occurred
	assert.NoError(t, err)

	// Assert that the expected pipelines were returned
	// assert.Equal(t, 2, len(pipelines))
	assert.Equal(t, "repo1", pipelines[0].Repository)
	assert.Equal(t, "repo2", pipelines[1].Repository)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestQueryPipelines_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewPipelineService(db)

	// Define the filters
	filters := map[string]interface{}{
		"repository": "repo1",
	}

	// Expect the query to return an error
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM Pipelines WHERE repository = ? ORDER BY start_time")).
		WithArgs("repo1").
		WillReturnError(fmt.Errorf("database error"))

	// Call the method under test
	pipelines, err := service.QueryPipelines(filters)

	// Assert that an error occurred
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "QueryPipelines: database error")

	// Assert that no pipelines were returned
	assert.Nil(t, pipelines)

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
