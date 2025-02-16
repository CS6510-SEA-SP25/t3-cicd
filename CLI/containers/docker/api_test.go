package containers

import (
	"cicd/pipeci/schema"
	"fmt"
	"strings"
	"testing"

	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/stretchr/testify/assert"
)

var TEST_PIPELINE_NAME string = "test_pipeline"
var TEST_PIPELINE_PREFIX string = "test_"

// Test initializing the Docker client
func TestInitDockerClient(t *testing.T) {
	dc, err := InitDockerClient()
	assert.NoError(t, err)
	assert.NotNil(t, dc)
	dc.Close()
}

// Test pulling an image
func TestPullImage(t *testing.T) {
	dc, err := InitDockerClient()
	assert.NoError(t, err)
	defer dc.Close()

	// Using a lightweight test image
	err = dc.PullImage("alpine:latest")
	assert.NoError(t, err)
}

// Test listing images to verify pull worked
func TestListImages(t *testing.T) {
	dc, err := InitDockerClient()
	assert.NoError(t, err)
	defer dc.Close()

	images, err := dc.cli.ImageList(dc.ctx, image.ListOptions{})
	assert.NoError(t, err)
	assert.Greater(t, len(images), 0)
}

// Test creating a container
func TestCreateContainer(t *testing.T) {
	dc, err := InitDockerClient()
	assert.NoError(t, err)
	defer dc.Close()

	// Ensure image is available
	err = dc.PullImage("alpine:latest")
	assert.NoError(t, err)

	// Get current working directory
	hostDir, err := os.Getwd()
	assert.NoError(t, err)

	containerDir := "/workspace"
	commands := []string{"echo 'Hello from container'"}

	// Create a container
	containerID, err := dc.CreateContainer(TEST_PIPELINE_PREFIX+"TestCreateContainer", "alpine:latest", commands, hostDir, containerDir)
	assert.NoError(t, err)
	assert.NotEmpty(t, containerID)
}

// Test starting a container
func TestStartContainer(t *testing.T) {
	dc, err := InitDockerClient()
	assert.NoError(t, err)
	defer dc.Close()

	err = dc.PullImage("alpine:latest")
	assert.NoError(t, err)

	hostDir, err := os.Getwd()
	assert.NoError(t, err)

	containerDir := "/workspace"
	commands := []string{"echo 'Container running'"}

	containerID, err := dc.CreateContainer(TEST_PIPELINE_PREFIX+"TestStartContainer", "alpine:latest", commands, hostDir, containerDir)
	assert.NoError(t, err)
	assert.NotEmpty(t, containerID)

	// Start the container
	err = dc.StartContainer(containerID)
	assert.NoError(t, err)
}

// Test waiting for a container to complete execution
func TestWaitContainer(t *testing.T) {
	dc, err := InitDockerClient()
	assert.NoError(t, err)
	defer dc.Close()

	err = dc.PullImage("alpine:latest")
	assert.NoError(t, err)

	hostDir, err := os.Getwd()
	assert.NoError(t, err)

	containerDir := "/workspace"
	commands := []string{"echo 'Wait test successful'"}

	containerID, err := dc.CreateContainer(TEST_PIPELINE_PREFIX+"TestWaitContainer", "alpine:latest", commands, hostDir, containerDir)
	assert.NoError(t, err)

	err = dc.StartContainer(containerID)
	assert.NoError(t, err)

	// Wait for completion
	err = dc.WaitContainer(containerID)
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

	for _, containerID := range removedIds {
		go dc.DeleteContainer(containerID)
	}
	return nil
}

// Test initContainer
func TestExecute(t *testing.T) {
	dc, err := InitDockerClient()
	assert.NoError(t, err)
	defer dc.Close()

	var pipeline schema.PipelineConfiguration = schema.PipelineConfiguration{
		Pipeline: &schema.ConfigurationNode[schema.PipelineInfo]{
			Value: schema.PipelineInfo{Name: &schema.ConfigurationNode[string]{Value: TEST_PIPELINE_NAME}},
		},
		Version:    &schema.ConfigurationNode[string]{Value: "v0"},
		StageOrder: []string{"build"},
		Stages: &schema.ConfigurationNode[map[string]*schema.ConfigurationNode[map[string]*schema.JobConfiguration]]{
			Value: map[string]*schema.ConfigurationNode[map[string]*schema.JobConfiguration]{
				"build": {
					Value: map[string]*schema.JobConfiguration{
						"compile": {
							Name:  &schema.ConfigurationNode[string]{Value: "compile"},
							Stage: &schema.ConfigurationNode[string]{Value: "build"},
							Image: &schema.ConfigurationNode[string]{Value: "maven"},
							Script: &schema.ConfigurationNode[[]string]{
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

	err = Execute(pipeline)
	assert.NoError(t, err)

	// Cleanup
	err = cleanUpAfterTest(dc)
	assert.NoError(t, err) // Expect no error
}

// Test job execution failed
func TestExecuteFailed(t *testing.T) {
	dc, err := InitDockerClient()
	assert.NoError(t, err)
	defer dc.Close()

	var pipeline schema.PipelineConfiguration = schema.PipelineConfiguration{
		Pipeline: &schema.ConfigurationNode[schema.PipelineInfo]{
			Value: schema.PipelineInfo{Name: &schema.ConfigurationNode[string]{Value: TEST_PIPELINE_NAME}},
		},
		Version:    &schema.ConfigurationNode[string]{Value: "v0"},
		StageOrder: []string{"build"},
		Stages: &schema.ConfigurationNode[map[string]*schema.ConfigurationNode[map[string]*schema.JobConfiguration]]{
			Value: map[string]*schema.ConfigurationNode[map[string]*schema.JobConfiguration]{
				"build": {
					Value: map[string]*schema.JobConfiguration{
						"compile": {
							Name:  &schema.ConfigurationNode[string]{Value: "compile"},
							Stage: &schema.ConfigurationNode[string]{Value: "build"},
							Image: &schema.ConfigurationNode[string]{Value: "mavennnnn"},
							Script: &schema.ConfigurationNode[[]string]{
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

	err = Execute(pipeline)
	if err == nil {
		t.Errorf("expected an error but got none")
	} else {
		if !strings.Contains(err.Error(), "terminating pipeline execution, caused by failure in running job") {
			t.Errorf("unexpected error message: %v", err)
		}
	}
	assert.Error(t, err) // Expect no error
}
