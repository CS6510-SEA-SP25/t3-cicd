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
