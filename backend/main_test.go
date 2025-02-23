package main

import (
	"cicd/pipeci/backend/db"
	"cicd/pipeci/backend/models"
	"cicd/pipeci/backend/routes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var TEST_PIPELINE_NAME string = "test_pipeline"

func TestPingRoute(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestExecuteLocal(t *testing.T) {
	db.Init()
	router := setupRouter()

	w := httptest.NewRecorder()

	var pipeline models.PipelineConfiguration = models.PipelineConfiguration{
		Pipeline: &models.ConfigurationNode[models.PipelineInfo]{
			Value: models.PipelineInfo{Name: &models.ConfigurationNode[string]{Value: TEST_PIPELINE_NAME}},
		},
		Version:    &models.ConfigurationNode[string]{Value: "v0"},
		StageOrder: []string{"build"},
		Stages: &models.ConfigurationNode[map[string]*models.ConfigurationNode[map[string]*models.JobConfiguration]]{
			Value: map[string]*models.ConfigurationNode[map[string]*models.JobConfiguration]{
				"build": {
					Value: map[string]*models.JobConfiguration{
						"compile": {
							Name:  &models.ConfigurationNode[string]{Value: "compile"},
							Stage: &models.ConfigurationNode[string]{Value: "build"},
							Image: &models.ConfigurationNode[string]{Value: "maven"},
							Script: &models.ConfigurationNode[[]string]{
								Value: []string{"ls -la", "mvn -v"},
							},
						},
					},
				},
			},
		},
		ExecOrder: map[string][][]string{
			"build": {{"compile"}},
		},
	}

	var repository models.Repository = models.Repository{
		Url: "https://github.com/CS6510-SEA-SP25/t3-cicd.git", CommitHash: "ae47cc929081a0312a54bf85f3f6c232a912e243",
	}
	// // Create an example user for testing
	var body = routes.ExecuteLocal_RequestBody{
		Pipeline:   pipeline,
		Repository: repository,
	}
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "/execute/local", strings.NewReader(string(jsonBody)))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.NoError(t, err)
}

func TestExecuteLocalFailed(t *testing.T) {
	db.Init()
	router := setupRouter()

	w := httptest.NewRecorder()

	var pipeline models.PipelineConfiguration = models.PipelineConfiguration{
		Pipeline: &models.ConfigurationNode[models.PipelineInfo]{
			Value: models.PipelineInfo{Name: &models.ConfigurationNode[string]{Value: TEST_PIPELINE_NAME}},
		},
		Version:    &models.ConfigurationNode[string]{Value: "v0"},
		StageOrder: []string{"build"},
		Stages: &models.ConfigurationNode[map[string]*models.ConfigurationNode[map[string]*models.JobConfiguration]]{
			Value: map[string]*models.ConfigurationNode[map[string]*models.JobConfiguration]{
				"build": {
					Value: map[string]*models.JobConfiguration{
						"compile": {
							Name:  &models.ConfigurationNode[string]{Value: "compile"},
							Stage: &models.ConfigurationNode[string]{Value: "build"},
							Image: &models.ConfigurationNode[string]{Value: "mavennnn"},
							Script: &models.ConfigurationNode[[]string]{
								Value: []string{"ls -la", "mvn -v"},
							},
						},
					},
				},
			},
		},
		ExecOrder: map[string][][]string{
			"build": {{"compile"}},
		},
	}

	var repository models.Repository = models.Repository{
		Url: "https://github.com/CS6510-SEA-SP25/t3-cicd.git", CommitHash: "ae47cc929081a0312a54bf85f3f6c232a912e243",
	}
	// // Create an example user for testing
	var body = routes.ExecuteLocal_RequestBody{
		Pipeline:   pipeline,
		Repository: repository,
	}
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "/execute/local", strings.NewReader(string(jsonBody)))
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	assert.NoError(t, err)
}
