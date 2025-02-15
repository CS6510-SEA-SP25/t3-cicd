package containers

import (
	"cicd/pipeci/schema"
	"testing"

	"os"

	"github.com/docker/docker/api/types/image"
	"github.com/stretchr/testify/assert"
)

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
	containerID, err := dc.CreateContainer("alpine:latest", commands, hostDir, containerDir)
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

	containerID, err := dc.CreateContainer("alpine:latest", commands, hostDir, containerDir)
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

	containerID, err := dc.CreateContainer("alpine:latest", commands, hostDir, containerDir)
	assert.NoError(t, err)

	err = dc.StartContainer(containerID)
	assert.NoError(t, err)

	// Wait for completion
	err = dc.WaitContainer(containerID)
	assert.NoError(t, err)
}

// Test initContainer
func TestExecute(t *testing.T) {
	var pipeline schema.PipelineConfiguration = schema.PipelineConfiguration{
		Version:    &schema.ConfigurationNode[string]{Value: "v0"},
		StageOrder: []string{"build"},
		Stages: &schema.ConfigurationNode[map[string]*schema.ConfigurationNode[map[string]*schema.JobConfiguration]]{
			Value: map[string]*schema.ConfigurationNode[map[string]*schema.JobConfiguration]{
				"build": {
					Value: map[string]*schema.JobConfiguration{
						"compile": {
							Name:  &schema.ConfigurationNode[string]{Value: "compile"},
							Stage: &schema.ConfigurationNode[string]{Value: "build"},
							Image: &schema.ConfigurationNode[string]{Value: "maven:3.8.6"},
							Script: &schema.ConfigurationNode[[]string]{
								Value: []string{"ps -u", "mvn -v"},
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

	err := Execute(pipeline)
	assert.NoError(t, err)
}
