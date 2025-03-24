package apis

import (
	"cicd/pipeci/schema"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteLocal(t *testing.T) {
	// Override BASE_URL for testing
	BASE_URL = "http://localhost:8080"

	pipeline := schema.PipelineConfiguration{ /* mock data */ }
	repository := schema.Repository{ /* mock data */ }

	err := ExecuteLocal(pipeline, repository)
	assert.NoError(t, err)
}
