package DockerService

import (
	"cicd/pipeci/backend/db"
	"cicd/pipeci/backend/models"
	"cicd/pipeci/backend/storage"
	"fmt"

	"log"
	"os"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var TEST_PIPELINE_NAME string = "test_pipeline"
var TEST_PIPELINE_PREFIX string = "test_"

// Load .env before running tests
func TestMain(m *testing.M) {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("Failed to load .env file. Ensure the file exists.")
	}

	log.Println(".env file loaded successfully")

	// Run tests
	os.Exit(m.Run())
}

// Test initializing the Docker client
func TestInitDockerClient(t *testing.T) {
	dc, err := initDockerClient()
	assert.NoError(t, err)
	assert.NotNil(t, dc)
	dc.Close()
}

// Test pulling an image
func TestPullImage(t *testing.T) {
	dc, err := initDockerClient()
	assert.NoError(t, err)
	defer dc.Close()

	// Using a lightweight test image
	err = dc.pullImage("alpine:latest")
	assert.NoError(t, err)
}

// Test listing images to verify pull worked
func TestListImages(t *testing.T) {
	dc, err := initDockerClient()
	assert.NoError(t, err)
	defer dc.Close()

	images, err := dc.cli.ImageList(dc.ctx, image.ListOptions{})
	assert.NoError(t, err)
	assert.Greater(t, len(images), 0)
}

// Test creating a container
func TestCreateContainer(t *testing.T) {
	dc, err := initDockerClient()
	assert.NoError(t, err)
	defer dc.Close()

	// Ensure image is available
	err = dc.pullImage("alpine:latest")
	assert.NoError(t, err)

	commands := []string{"echo 'Hello from container'"}

	// Create a container
	containerId, err := dc.createContainer(TEST_PIPELINE_PREFIX+"TestCreateContainer", "alpine:latest", commands)
	assert.NoError(t, err)
	assert.NotEmpty(t, containerId)

	err = dc.deleteContainer(containerId)
	assert.NoError(t, err)
}

// Test starting a container
func TestStartContainer(t *testing.T) {
	dc, err := initDockerClient()
	assert.NoError(t, err)
	defer dc.Close()

	err = dc.pullImage("alpine:latest")
	assert.NoError(t, err)

	commands := []string{"echo 'Container running'"}

	containerId, err := dc.createContainer(TEST_PIPELINE_PREFIX+"TestStartContainer", "alpine:latest", commands)
	assert.NoError(t, err)
	assert.NotEmpty(t, containerId)

	// Start the container
	err = dc.startContainer(containerId)
	assert.NoError(t, err)

	err = dc.deleteContainer(containerId)
	assert.NoError(t, err)
}

// Test waiting for a container to complete execution
func TestWaitContainer(t *testing.T) {
	dc, err := initDockerClient()
	assert.NoError(t, err)
	defer dc.Close()

	err = dc.pullImage("alpine:latest")
	assert.NoError(t, err)

	commands := []string{"echo 'Wait test successful'"}

	containerId, err := dc.createContainer(TEST_PIPELINE_PREFIX+"TestWaitContainer", "alpine:latest", commands)
	assert.NoError(t, err)

	err = dc.startContainer(containerId)
	assert.NoError(t, err)

	// Wait for completion
	err = dc.WaitContainer(containerId)
	assert.NoError(t, err)

	err = dc.deleteContainer(containerId)
	assert.NoError(t, err)
}

/* Delete all containers with prefix "test_" after test */
func cleanUpAfterTest(dc *DockerClient) error {
	// List all containers
	containers, err := dc.cli.ContainerList(dc.ctx, container.ListOptions{
		All: true, // Include stopped containers
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	var removedIds []string
	for _, ctn := range containers {
		for _, name := range ctn.Names {
			trimmedName := strings.TrimPrefix(name, "/")
			if strings.Contains(trimmedName, TEST_PIPELINE_NAME) || strings.HasPrefix(trimmedName, TEST_PIPELINE_PREFIX) {
				removedIds = append(removedIds, ctn.ID)
			}
		}
	}

	// Gather container ids, currently doing nothing
	// TODO: clean up artifacts
	for _, containerId := range removedIds {
		// go dc.deleteContainer(containerId)
		fmt.Printf("clean up after test container id: %v", containerId)
	}
	return nil
}

// Test initContainer
func TestExecute(t *testing.T) {
	db.Init()
	storage.Init()
	dc, err := initDockerClient()
	assert.NoError(t, err)
	defer dc.Close()

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

	err = Execute(pipeline, models.Repository{
		Url: "https://github.com/CS6510-SEA-SP25/t3-cicd.git", CommitHash: "ae47cc929081a0312a54bf85f3f6c232a912e243",
	})
	assert.NoError(t, err)

	// Cleanup
	err = cleanUpAfterTest(dc)
	assert.NoError(t, err) // Expect no error
}

// Test job execution failed
func TestExecuteFailed(t *testing.T) {
	db.Init()
	storage.Init()
	dc, err := initDockerClient()
	assert.NoError(t, err)
	defer dc.Close()

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
							Image: &models.ConfigurationNode[string]{Value: "mavennnnn"},
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

	err = Execute(pipeline, models.Repository{})
	if err == nil {
		t.Errorf("expected an error but got none")
	} else {
		if !strings.Contains(err.Error(), "terminating pipeline execution, caused by failure in running job") {
			t.Errorf("unexpected error message: %v", err)
		}
	}
	assert.Error(t, err) // Expect no error
}

func TestExecuteFailed_InvalidCommand(t *testing.T) {
	db.Init()
	storage.Init()
	dc, err := initDockerClient()
	assert.NoError(t, err)
	defer dc.Close()

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

	err = Execute(pipeline, models.Repository{})
	if err == nil {
		t.Errorf("expected an error but got none")
	} else {
		if !strings.Contains(err.Error(), "terminating pipeline execution, caused by failure in running job") {
			t.Errorf("unexpected error message: %v", err)
		}
	}
	assert.Error(t, err) // Expect no error
}

func TestExecuteFailed_TerminatedJobs(t *testing.T) {
	db.Init()
	storage.Init()
	dc, err := initDockerClient()
	assert.NoError(t, err)
	defer dc.Close()

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
						"depends_on_compile": {
							Name:  &models.ConfigurationNode[string]{Value: "compile"},
							Stage: &models.ConfigurationNode[string]{Value: "build"},
							Image: &models.ConfigurationNode[string]{Value: "maven"},
							Script: &models.ConfigurationNode[[]string]{
								Value: []string{"ls -la"},
							},
							Dependencies: &models.ConfigurationNode[[]string]{
								Value: []string{"compile"},
							},
						},
						"not_depends_on_compile": {
							Name:  &models.ConfigurationNode[string]{Value: "compile"},
							Stage: &models.ConfigurationNode[string]{Value: "build"},
							Image: &models.ConfigurationNode[string]{Value: "maven"},
							Script: &models.ConfigurationNode[[]string]{
								Value: []string{"ls -la"},
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

	err = Execute(pipeline, models.Repository{})
	if err == nil {
		t.Errorf("expected an error but got none")
	} else {
		if !strings.Contains(err.Error(), "terminating pipeline execution, caused by failure in running job") {
			t.Errorf("unexpected error message: %v", err)
		}
	}
	assert.Error(t, err) // Expect no error
}
