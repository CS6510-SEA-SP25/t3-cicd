package schema

import (
	"errors"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Job configuration.
type JobConfiguration struct {
	Name         *string  `yaml:"name"`   // (required) Name of job within a stage.
	Stage        *string  `yaml:"stage"`  // (required) Stage name.
	Image        *string  `yaml:"image"`  // (required) Docker image to be used.
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
		return false, errors.New("syntax error: version")
	}

	// Validate pipeline info
	if pipeline.Pipeline == nil || *pipeline.Pipeline.Name == "" {
		return false, errors.New("syntax error: pipeline")
	}

	// Validate stages and jobs
	stages := make(map[string]map[string]*JobConfiguration)

	// Check stages
	if len(pipeline.Stages) == 0 {
		return false, errors.New("syntax error: empty stages")
	}
	for _, stage := range pipeline.Stages {
		if stage == "" {
			return false, errors.New("syntax error: stage name must be a non-empty string")
		} else if _, ok := stages[stage]; ok {
			return false, errors.New("syntax error: duplicated stages")
		} else {
			stages[stage] = make(map[string]*JobConfiguration)
		}
	}

	if len(pipeline.Jobs) == 0 {
		return false, errors.New("syntax error: empty jobs")
	}

	for _, job := range pipeline.Jobs {
		// Check format
		// Name
		if job.Name == nil {
			return false, errors.New("syntax error: empty job name")
		}
		// Stage
		if job.Stage == nil {
			return false, errors.New("syntax error: empty job stage")
		}
		if stages[*job.Stage] == nil {
			return false, errors.New("syntax error: job stage must be defined in stages")
		}
		if stages[*job.Stage][*job.Name] != nil {
			return false, errors.New("syntax error: duplicated job name within a stage")
		}
		// Image
		if job.Image == nil {
			return false, errors.New("syntax error: empty job image")
		}
		// Script
		if len(job.Script) == 0 {
			return false, errors.New("syntax error: empty job script")
		}

		// TODO: Check logic
		stages[*job.Stage][*job.Name] = &job
	}

	fmt.Printf("Jobs mapping by stages: %#v", stages)

	return true, nil
}
