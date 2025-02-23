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
	assert.NoError(t, err)
	// assert.Error(t, err)
}
