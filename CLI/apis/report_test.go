package apis

import (
	"bytes"
	"cicd/pipeci/schema"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReportPastExecutionsLocal_CurrentRepo_General(t *testing.T) {
	err := ReportPastExecutionsLocal_CurrentRepo(schema.Repository{
		Url: "https://github.com/CS6510-SEA-SP25/t3-cicd.git",
	})
	assert.NoError(t, err)

	err = logPipelineExecutionReport("invalid")
	assert.Error(t, err)

	err = logPipelineExecutionReport(map[string]interface{}{
		"commit_hash": "ae47cc929081a0312a54bf85f3f6c232a912e243",
		"end_time":    "2025-02-23T18:36:45Z",
		"ip_address":  "0.0.0.0",
		"name":        "test_pipeline",
		"pipeline_id": 50,
		"repository":  "https://github.com/CS6510-SEA-SP25/t3-cicd.git",
		"stage_order": "build",
		"start_time":  "2025-02-23T18:36:44Z",
		"status":      "FAILED",
	})
	assert.NoError(t, err)
}

func TestReportPastExecutionsLocal_CurrentRepo_NoReport(t *testing.T) {
	// Buffer to capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// Execute
	err := ReportPastExecutionsLocal_CurrentRepo(schema.Repository{
		Url: "repo_not_found",
	})
	assert.NoError(t, err)

	// Assert logs
	expected := "No executions detected."
	if !strings.Contains(buf.String(), expected) {
		t.Errorf("Log output does not contain expected string. Got: %q, Expected: %q", buf.String(), expected)
	}
}

func TestReportPastExecutionsLocal_ByCondition(t *testing.T) {
	err := ReportPastExecutionsLocal_ByCondition(schema.Repository{
		Url: "https://github.com/CS6510-SEA-SP25/t3-cicd.git",
	}, "name")
	assert.NoError(t, err)

	err = logPipelineExecutionReport("invalid")
	assert.Error(t, err)
}

func ReportPastExecutionsLocal_ByCondition_Success(t *testing.T) {
	err := ReportPastExecutionsLocal_ByCondition(schema.Repository{
		Url:        "https://github.com/CS6510-SEA-SP25/t3-cicd.git",
		CommitHash: "5901fe28dc221ed92d5e1ce95afaadc3383f3431",
	}, "name")
	assert.NoError(t, err)
}
