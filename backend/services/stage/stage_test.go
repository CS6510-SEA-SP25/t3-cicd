package StageService

import (
	"cicd/pipeci/backend/models"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetStagesByPipelineId(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewStageService(db)

	// Define the expected rows
	pipelineId := 1
	rows := sqlmock.NewRows([]string{"stage_id", "pipeline_id", "name", "status", "start_time", "end_time"}).
		AddRow(1, pipelineId, "stage1", models.SUCCESS, time.Now(), time.Now()).
		AddRow(2, pipelineId, "stage2", models.SUCCESS, time.Now(), time.Now())

	// Expect the query and return the mock rows
	mock.ExpectQuery("SELECT \\* FROM Stages WHERE pipeline_id = ?").
		WithArgs(pipelineId).
		WillReturnRows(rows)

	// Call the method under test
	stages, err := service.GetStagesByPipelineId(pipelineId)

	// Assert that no error occurred
	assert.NoError(t, err)

	// Assert that the expected stages were returned
	assert.Equal(t, 2, len(stages))
	assert.Equal(t, "stage1", stages[0].Name)
	assert.Equal(t, "stage2", stages[1].Name)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateStage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewStageService(db)

	// Define the stage
	stage := models.Stage{
		PipelineId: 1,
		Name:       "stage1",
		Status:     models.SUCCESS,
		StartTime:  time.Now(),
	}

	// Expect the exec and return a mock result
	mock.ExpectExec("INSERT INTO Stages").
		WithArgs(
			stage.PipelineId,
			stage.Name,
			stage.Status,
			sqlmock.AnyArg(), // Use AnyArg for the start_time argument
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the method under test
	stageId, err := service.CreateStage(stage)

	// Assert that no error occurred
	assert.NoError(t, err)

	// Assert that the expected stage ID was returned
	assert.Equal(t, 1, stageId)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateStageStatusAndEndTime(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewStageService(db)

	// Define the stage ID and status
	stageID := 1
	status := models.SUCCESS

	// Expect the exec and return a mock result
	mock.ExpectExec("UPDATE Stages SET status = \\?, end_time = \\? WHERE stage_id = \\?").
		WithArgs(
			status,
			sqlmock.AnyArg(), // Use AnyArg for the end_time argument
			stageID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Call the method under test
	err = service.UpdateStageStatusAndEndTime(stageID, status)

	// Assert that no error occurred
	assert.NoError(t, err)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateStageStatusAndEndTime_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewStageService(db)

	// Define the stage ID and status
	stageID := 1
	status := models.SUCCESS

	// Expect the exec and return a mock result with no rows affected
	mock.ExpectExec("UPDATE Stages SET status = \\?, end_time = \\? WHERE stage_id = \\?").
		WithArgs(
			status,
			sqlmock.AnyArg(), // Use AnyArg for the end_time argument
			stageID,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Call the method under test
	err = service.UpdateStageStatusAndEndTime(stageID, status)

	// Assert that an error occurred
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("UpdateStageStatusAndEndTime: no stage found with ID %d", stageID), err)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// Test on failed scenarios
func TestGetStagesByPipelineId_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewStageService(db)

	// Define the pipeline ID
	pipelineId := 1

	// Expect the query to return an error
	mock.ExpectQuery("SELECT \\* FROM Stages WHERE pipeline_id = ?").
		WithArgs(pipelineId).
		WillReturnError(fmt.Errorf("database error"))

	// Call the method under test
	stages, err := service.GetStagesByPipelineId(pipelineId)

	// Assert that an error occurred
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GetStagesByPipelineId: database error")

	// Assert that no stages were returned
	assert.Nil(t, stages)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateStage_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewStageService(db)

	// Define the stage
	stage := models.Stage{
		PipelineId: 1,
		Name:       "stage1",
		Status:     models.PENDING,
		StartTime:  time.Now(),
	}

	// Expect the exec to return an error
	mock.ExpectExec("INSERT INTO Stages").
		WithArgs(
			stage.PipelineId,
			stage.Name,
			stage.Status,
			sqlmock.AnyArg(), // Use AnyArg for the start_time argument
		).
		WillReturnError(fmt.Errorf("database error"))

	// Call the method under test
	stageId, err := service.CreateStage(stage)

	// Assert that an error occurred
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CreateStage: database error")

	// Assert that no stage ID was returned
	assert.Equal(t, 0, stageId)

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateStageStatusAndEndTime_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewStageService(db)

	// Define the stage ID and status
	stageID := 1
	status := models.SUCCESS

	// Expect the exec to return an error
	mock.ExpectExec("UPDATE Stages SET status = \\?, end_time = \\? WHERE stage_id = \\?").
		WithArgs(
			status,
			sqlmock.AnyArg(), // Use AnyArg for the end_time argument
			stageID,
		).
		WillReturnError(fmt.Errorf("database error"))

	// Call the method under test
	err = service.UpdateStageStatusAndEndTime(stageID, status)

	// Assert that an error occurred
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UpdateStageStatusAndEndTime: database error")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateStageStatusAndEndTime_NoRowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	service := NewStageService(db)

	// Define the stage ID and status
	stageID := 1
	status := models.SUCCESS

	// Expect the exec to return a result with no rows affected
	mock.ExpectExec("UPDATE Stages SET status = \\?, end_time = \\? WHERE stage_id = \\?").
		WithArgs(
			status,
			sqlmock.AnyArg(), // Use AnyArg for the end_time argument
			stageID,
		).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Call the method under test
	err = service.UpdateStageStatusAndEndTime(stageID, status)

	// Assert that an error occurred
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UpdateStageStatusAndEndTime: no stage found with ID 1")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
