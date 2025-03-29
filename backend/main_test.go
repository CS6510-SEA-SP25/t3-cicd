package main

import (
	"cicd/pipeci/backend/cache"
	"cicd/pipeci/backend/db"
	"cicd/pipeci/backend/models"

	// PipelineService "cicd/pipeci/backend/services/pipeline"
	// "cicd/pipeci/backend/storage"
	"cicd/pipeci/backend/types"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var TEST_PIPELINE_NAME string = "test_pipeline"

// Load .env before running tests
func TestMain(m *testing.M) {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Failed to load .env file. Ensure the file exists.")
	}

	log.Println(".env file loaded successfully")

	// Run tests
	os.Exit(m.Run())
}

func TestPingRoute(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestExecuteLocal(t *testing.T) {
	db.Init()
	// storage.Init()
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
	var body = types.ExecuteLocal_RequestBody{
		Pipeline:   pipeline,
		Repository: repository,
	}
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "/execute/local", strings.NewReader(string(jsonBody)))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.NoError(t, err)
}

func TestExecuteLocalFailed_InvalidImage(t *testing.T) {
	db.Init()
	// storage.Init()
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
	var body = types.ExecuteLocal_RequestBody{
		Pipeline:   pipeline,
		Repository: repository,
	}
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "/execute/local", strings.NewReader(string(jsonBody)))
	router.ServeHTTP(w, req)

	// assert.Equal(t, 400, w.Code)
	assert.Equal(t, 200, w.Code) // Async execution returns 200
	assert.NoError(t, err)
}

func TestExecuteLocalFailed(t *testing.T) {
	db.Init()
	// storage.Init()
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
								Value: []string{"ls -la", "invalid command"},
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
	var body = types.ExecuteLocal_RequestBody{
		Pipeline:   pipeline,
		Repository: repository,
	}
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "/execute/local", strings.NewReader(string(jsonBody)))
	router.ServeHTTP(w, req)

	// assert.Equal(t, 400, w.Code)
	assert.Equal(t, 200, w.Code) // Async execution returns 200
	assert.NoError(t, err)
}

func TestReportLocal(t *testing.T) {
	db.Init()
	// storage.Init()
	router := setupRouter()

	w := httptest.NewRecorder()

	var repository models.Repository = models.Repository{
		Url: "https://github.com/CS6510-SEA-SP25/t3-cicd.git",
	}
	var body = types.ReportPastExecutionsLocal_CurrentRepo_RequestBody{
		Repository: repository,
		IPAddress:  "0.0.0.0",
	}
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "/report/local", strings.NewReader(string(jsonBody)))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.NoError(t, err)
}

func TestReportPastExecutionsLocal_ByCondition(t *testing.T) {
	db.Init()
	// storage.Init()
	router := setupRouter()

	w := httptest.NewRecorder()

	var repository models.Repository = models.Repository{
		Url: "https://github.com/CS6510-SEA-SP25/t3-cicd.git",
	}
	var body = types.ReportPastExecutionsLocal_CurrentRepo_RequestBody{
		Repository:   repository,
		IPAddress:    "0.0.0.0",
		PipelineName: "name",
	}
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "/report/local/query", strings.NewReader(string(jsonBody)))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.NoError(t, err)
}

func TestReportStage(t *testing.T) {
	db.Init()
	// storage.Init()
	router := setupRouter()

	w := httptest.NewRecorder()

	var repository models.Repository = models.Repository{
		Url: "https://github.com/CS6510-SEA-SP25/t3-cicd.git",
	}
	var body = types.ReportPastExecutionsLocal_CurrentRepo_RequestBody{
		Repository:   repository,
		IPAddress:    "0.0.0.0",
		PipelineName: "test_pipeline",
		StageName:    "build",
	}
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "/report/local/query", strings.NewReader(string(jsonBody)))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.NoError(t, err)
}

func TestReportJob(t *testing.T) {
	db.Init()

	router := setupRouter()

	w := httptest.NewRecorder()

	var repository models.Repository = models.Repository{
		Url: "https://github.com/CS6510-SEA-SP25/t3-cicd.git",
	}
	var body = types.ReportPastExecutionsLocal_CurrentRepo_RequestBody{
		Repository:   repository,
		IPAddress:    "0.0.0.0",
		PipelineName: "test_pipeline",
		StageName:    "build",
		JobName:      "compile",
	}
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "/report/local/query", strings.NewReader(string(jsonBody)))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.NoError(t, err)
}

func TestRequestExecutionStatus_Success(t *testing.T) {
	db.Init()
	cache.Init()

	router := setupRouter()

	w := httptest.NewRecorder()

	var body = types.RequestExecutionStatus_RequestBody{
		ExecutionId: "id",
	}
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "/status", strings.NewReader(string(jsonBody)))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.NoError(t, err)
}

func TestRequestExecutionStatus_Error(t *testing.T) {
	db.Init()
	cache.Init()

	router := setupRouter()

	w := httptest.NewRecorder()

	var body = ""
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "/status", strings.NewReader(string(jsonBody)))
	router.ServeHTTP(w, req)

	assert.NoError(t, err)
}
