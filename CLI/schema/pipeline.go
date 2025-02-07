package schema

import (
	"errors"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

// Job configuration.
type JobConfiguration struct {
	Name         string   `yaml:"name"`   // (required) Name of job within a stage.
	Stage        string   `yaml:"stage"`  // (required) Stage name.
	Image        string   `yaml:"image"`  // (required) Docker image to be used.
	Script       []string `yaml:"script"` // (required) List of scripts to be executed sequentially.
	Dependencies []string `yaml:"needs"`  // (optional) Jobs within the same statge that must complete successfully before this job can start executing.
}

// Pipeline identifier info.
type PipelineInfo struct {
	Name *string `yaml:"name"` // (required) Name of pipeline.
}

// Pipeline configuration
type PipelineConfiguration struct {
	Version *string `yaml:"version"` // (required) API version. Currently set at v0.

	Pipeline *PipelineInfo `yaml:"pipeline"`

	Stages []string `yaml:"stages"` // (required) List of stages. Also represent the order of execution.

	Jobs []JobConfiguration `yaml:"jobs"` // (required) List of jobs' configuration. Each stage has at least one job.
}

// Reads YAML file then parse to Pipeline
func ParsePipelineConfiguration(filename string) (*PipelineConfiguration, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var pipeline PipelineConfiguration
	if err := yaml.Unmarshal(data, &pipeline); err != nil {
		return nil, err
	}

	return &pipeline, nil
}

// Validate Pipeline configuration
func (pipeline *PipelineConfiguration) ValidateConfiguration() (bool, error) {
	log.SetPrefix("Greetings: ")
	log.SetFlags(0)
	// Validate version
	if pipeline.Version == nil || *pipeline.Version != "v0" {
		return false, errors.New("version syntax error")
	}

	// Validate pipeline info
	if pipeline.Pipeline == nil || *pipeline.Pipeline.Name == "" {
		return false, errors.New("pipeline syntax error")
	}

	// // Validate stages and jobs
	// stages := pipeline.Stages
	// jobsMap := make(map[string]map[string]JobConfiguration)

	return true, nil
}
