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

func TestGenerateReports(t *testing.T) {
	err := generateReports("invalid")
	assert.Error(t, err)

	err = generateReports([]interface{}{
		map[string]interface{}{
			"end_time":   map[string]interface{}{"Time": "2025-03-16T07:32:00Z", "Valid": true},
			"id":         6.0, // cast float64 -> int
			"name":       "maven_project_1",
			"start_time": "2025-03-16T07:31:02Z",
			"status":     "SUCCESS"},
	})
	assert.NoError(t, err)
}

func TestReportPastExecutionsLocal_CurrentRepo_General(t *testing.T) {
	err := ReportPastExecutionsLocal_CurrentRepo(schema.Repository{
		Url: "https://github.com/CS6510-SEA-SP25/t3-cicd.git",
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
	}, "pipeline_name", "", "", 0)
	assert.NoError(t, err)

	err = generateReports("invalid")
	assert.Error(t, err)
}

func ReportPastExecutionsLocal_ByCondition_Success(t *testing.T) {
	err := ReportPastExecutionsLocal_ByCondition(schema.Repository{
		Url:        "https://github.com/CS6510-SEA-SP25/t3-cicd.git",
		CommitHash: "5901fe28dc221ed92d5e1ce95afaadc3383f3431",
	}, "pipeline_name", "", "", 0)
	assert.NoError(t, err)
}
