package StageService

import (
	"cicd/pipeci/worker/models"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

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
