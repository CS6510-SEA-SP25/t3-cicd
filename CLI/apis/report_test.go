package apis

import (
	"cicd/pipeci/schema"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReportPastExecutionsLocal(t *testing.T) {
	err := ReportPastExecutionsLocal(schema.Repository{
		Url: "https://github.com/CS6510-SEA-SP25/t3-cicd.git",
	})
	// assert.NoError(t, err)
	// assert.Error(t, err)

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

func TestQueryPastExectionsLocal(t *testing.T) {
	err := QueryPastExectionsLocal(schema.Repository{
		Url: "https://github.com/CS6510-SEA-SP25/t3-cicd.git",
	}, "name")
	// assert.NoError(t, err)
	// assert.Error(t, err)

	err = logPipelineExecutionReport("invalid")
	assert.Error(t, err)
}
