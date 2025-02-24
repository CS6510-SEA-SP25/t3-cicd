package models

import (
	"testing"
)

func TestYAMLFileLocation(t *testing.T) {
	loc := &YAMLFileLocation{
		Line:   10,
		Column: 5,
	}

	if loc.Line != 10 {
		t.Errorf("Expected Line to be 10, got %d", loc.Line)
	}

	if loc.Column != 5 {
		t.Errorf("Expected Column to be 5, got %d", loc.Column)
	}
}

func TestConfigurationNode(t *testing.T) {
	loc := &YAMLFileLocation{Line: 1, Column: 1}
	node := &ConfigurationNode[string]{
		Value:    "test-value",
		Location: loc,
	}

	if node.Value != "test-value" {
		t.Errorf("Expected Value to be 'test-value', got '%s'", node.Value)
	}

	if node.Location.Line != 1 || node.Location.Column != 1 {
		t.Errorf("Expected Location to be {1, 1}, got {%d, %d}", node.Location.Line, node.Location.Column)
	}
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

	if job.Name.Value != "job1" {
		t.Errorf("Expected Name to be 'job1', got '%s'", job.Name.Value)
	}

	if job.Stage.Value != "build" {
		t.Errorf("Expected Stage to be 'build', got '%s'", job.Stage.Value)
	}

	if job.Image.Value != "golang:1.19" {
		t.Errorf("Expected Image to be 'golang:1.19', got '%s'", job.Image.Value)
	}

	if len(job.Script.Value) != 1 || job.Script.Value[0] != "echo 'Hello, World!'" {
		t.Errorf("Expected Script to be ['echo \"Hello, World!\"'], got %v", job.Script.Value)
	}

	if len(job.Dependencies.Value) != 2 || job.Dependencies.Value[0] != "job2" || job.Dependencies.Value[1] != "job3" {
		t.Errorf("Expected Dependencies to be ['job2', 'job3'], got %v", job.Dependencies.Value)
	}
}

func TestPipelineInfo(t *testing.T) {
	loc := &YAMLFileLocation{Line: 1, Column: 1}
	pipelineInfo := &PipelineInfo{
		Name: &ConfigurationNode[string]{Value: "pipeline1", Location: loc},
	}

	if pipelineInfo.Name.Value != "pipeline1" {
		t.Errorf("Expected Name to be 'pipeline1', got '%s'", pipelineInfo.Name.Value)
	}
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

	if pipelineConfig.Version.Value != "v0" {
		t.Errorf("Expected Version to be 'v0', got '%s'", pipelineConfig.Version.Value)
	}

	if pipelineConfig.Pipeline.Value.Name.Value != "pipeline1" {
		t.Errorf("Expected Pipeline Name to be 'pipeline1', got '%s'", pipelineConfig.Pipeline.Value.Name.Value)
	}

	if len(pipelineConfig.Stages.Value) != 1 {
		t.Errorf("Expected 1 stage, got %d", len(pipelineConfig.Stages.Value))
	}

	if pipelineConfig.StageOrder[0] != "build" {
		t.Errorf("Expected StageOrder to be ['build'], got %v", pipelineConfig.StageOrder)
	}

	if len(pipelineConfig.ExecOrder["build"]) != 1 || pipelineConfig.ExecOrder["build"][0][0] != "job1" {
		t.Errorf("Expected ExecOrder to be map[build:[[job1]]], got %v", pipelineConfig.ExecOrder)
	}
}

func TestRepository(t *testing.T) {
	repo := &Repository{
		Url:        "https://github.com/example/repo.git",
		CommitHash: "abc123",
	}

	if repo.Url != "https://github.com/example/repo.git" {
		t.Errorf("Expected Url to be 'https://github.com/example/repo.git', got '%s'", repo.Url)
	}

	if repo.CommitHash != "abc123" {
		t.Errorf("Expected CommitHash to be 'abc123', got '%s'", repo.CommitHash)
	}
}
