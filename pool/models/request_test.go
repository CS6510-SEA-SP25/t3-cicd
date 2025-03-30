package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYAMLFileLocation(t *testing.T) {
	loc := &YAMLFileLocation{
		Line:   10,
		Column: 5,
	}

	assert.Equal(t, 10, loc.Line)
	assert.Equal(t, 5, loc.Column)
}

func TestConfigurationNode(t *testing.T) {
	loc := &YAMLFileLocation{Line: 1, Column: 1}
	node := &ConfigurationNode[string]{
		Value:    "test-value",
		Location: loc,
	}

	assert.Equal(t, "test-value", node.Value)
	assert.Equal(t, 1, node.Location.Line)
	assert.Equal(t, 1, node.Location.Column)
}

func TestJobConfiguration(t *testing.T) {
	loc := &YAMLFileLocation{Line: 1, Column: 1}
	job := &JobConfiguration{
		Name:         &ConfigurationNode[string]{Value: "job1", Location: loc},
		Stage:        &ConfigurationNode[string]{Value: "build", Location: loc},
		Image:        &ConfigurationNode[string]{Value: "golang:1.19", Location: loc},
		Script:       &ConfigurationNode[[]string]{Value: []string{"echo 'Hello, World!'"}, Location: loc},
		Dependencies: &ConfigurationNode[[]string]{Value: []string{"job2", "job3"}, Location: loc},
	}

	assert.Equal(t, "job1", job.Name.Value)
	assert.Equal(t, "build", job.Stage.Value)
	assert.Equal(t, "golang:1.19", job.Image.Value)
	assert.Len(t, job.Script.Value, 1)
	assert.Equal(t, "echo 'Hello, World!'", job.Script.Value[0])
	assert.Len(t, job.Dependencies.Value, 2)
	assert.Equal(t, "job2", job.Dependencies.Value[0])
	assert.Equal(t, "job3", job.Dependencies.Value[1])
}

func TestPipelineInfo(t *testing.T) {
	loc := &YAMLFileLocation{Line: 1, Column: 1}
	pipelineInfo := &PipelineInfo{
		Name: &ConfigurationNode[string]{Value: "pipeline1", Location: loc},
	}

	assert.Equal(t, "pipeline1", pipelineInfo.Name.Value)
}

func TestPipelineConfiguration(t *testing.T) {
	loc := &YAMLFileLocation{Line: 1, Column: 1}
	pipelineConfig := &PipelineConfiguration{
		Version:  &ConfigurationNode[string]{Value: "v0", Location: loc},
		Pipeline: &ConfigurationNode[PipelineInfo]{Value: PipelineInfo{Name: &ConfigurationNode[string]{Value: "pipeline1", Location: loc}}, Location: loc},
		Stages: &ConfigurationNode[map[string]*ConfigurationNode[map[string]*JobConfiguration]]{
			Value: map[string]*ConfigurationNode[map[string]*JobConfiguration]{
				"build": {
					Value: map[string]*JobConfiguration{
						"job1": {
							Name:   &ConfigurationNode[string]{Value: "job1", Location: loc},
							Stage:  &ConfigurationNode[string]{Value: "build", Location: loc},
							Image:  &ConfigurationNode[string]{Value: "golang:1.19", Location: loc},
							Script: &ConfigurationNode[[]string]{Value: []string{"echo 'Hello, World!'"}, Location: loc},
						},
					},
					Location: loc,
				},
			},
			Location: loc,
		},
		StageOrder: []string{"build"},
		ExecOrder:  map[string][][]string{"build": {{"job1"}}},
	}

	assert.Equal(t, "v0", pipelineConfig.Version.Value)
	assert.Equal(t, "pipeline1", pipelineConfig.Pipeline.Value.Name.Value)
	assert.Len(t, pipelineConfig.Stages.Value, 1)
	assert.Equal(t, []string{"build"}, pipelineConfig.StageOrder)
	assert.Len(t, pipelineConfig.ExecOrder["build"], 1)
	assert.Equal(t, []string{"job1"}, pipelineConfig.ExecOrder["build"][0])
}

func TestRepository(t *testing.T) {
	repo := &Repository{
		Url:        "https://github.com/example/repo.git",
		CommitHash: "abc123",
	}

	assert.Equal(t, "https://github.com/example/repo.git", repo.Url)
	assert.Equal(t, "abc123", repo.CommitHash)
}
