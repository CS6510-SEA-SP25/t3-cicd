package schema

import (
	"errors"
	"log"
	"os"

	"fmt"

	"gopkg.in/yaml.v3"
)

// Job configuration.
type JobConfiguration struct {
	Name         *string  `yaml:"name"`   // (required) Name of job within a stage.
	Stage        *string  `yaml:"stage"`  // (required) Stage name.
	Image        *string  `yaml:"image"`  // (required) Docker image to be used.
	Script       []string `yaml:"script"` // (required) List of scripts to be executed sequentially.
	Dependencies []string `yaml:"needs"`  // (optional) Jobs within the same statge that must complete successfully before this job can start executing.

	// YAML Node line number
	NameLine         int
	StageLine        int
	ImageLine        int
	ScriptLine       int
	DependenciesLine int

	// YAML Node column number
	NameColumn         int
	StageColumn        int
	ImageColumn        int
	ScriptColumn       int
	DependenciesColumn int
}

// Pipeline identifier info.
type PipelineInfo struct {
	Name       *string `yaml:"name"` // (required) Name of pipeline.
	NameLine   int
	NameColumn int
}

// Pipeline configuration
type PipelineConfiguration struct {
	Version  *string            `yaml:"version"` // (required) API version. Currently set at v0.
	Pipeline *PipelineInfo      `yaml:"pipeline"`
	Stages   []string           `yaml:"stages"` // (required) List of stages. Also represent the order of execution.
	Jobs     []JobConfiguration `yaml:"jobs"`   // (required) List of jobs' configuration. Each stage has at least one job.

	// YAML Node line number
	PipelineLine int
	VersionLine  int
	StagesLine   int
	JobsLine     int

	// YAML Node column number
	PipelineColumn int
	VersionColumn  int
	StagesColumn   int
	JobsColumn     int

	/*
		Jobs execution order for each stage (topological)
		e.g. "build" -> [[job1, job2], [job3], [job4]]
	*/
	ExecOrder map[string][][]string
}

// parsePipelineConfig extracts values and line numbers from yaml.Node
func parsePipelineConfig(root *yaml.Node, config *PipelineConfiguration) {
	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		fmt.Println("Invalid YAML structure")
		return
	}

	// Access the mapping node
	mapping := root.Content[0]
	if mapping.Kind != yaml.MappingNode {
		fmt.Println("Expected a mapping node")
		return
	}

	for i := 0; i < len(mapping.Content); i += 2 {
		keyNode := mapping.Content[i]
		valueNode := mapping.Content[i+1]

		switch keyNode.Value {
		case "version":
			config.Version = &valueNode.Value
			config.VersionLine = keyNode.Line
			config.VersionColumn = keyNode.Column
		case "pipeline":
			config.Pipeline = &PipelineInfo{}
			config.PipelineLine = keyNode.Line
			config.PipelineColumn = keyNode.Column
			parsePipelineInfo(valueNode, config.Pipeline)
		case "stages":
			config.StagesLine = keyNode.Line
			config.StagesColumn = keyNode.Column
			if valueNode.Kind == yaml.SequenceNode {
				for _, item := range valueNode.Content {
					config.Stages = append(config.Stages, item.Value)
				}
			}
		case "jobs":
			config.JobsLine = keyNode.Line
			config.JobsColumn = keyNode.Column
			if valueNode.Kind == yaml.SequenceNode {
				for _, jobNode := range valueNode.Content {
					var job JobConfiguration
					parseJobConfig(jobNode, &job)
					config.Jobs = append(config.Jobs, job)
				}
			}
		}
	}
}

// parsePipelineInfo extracts pipeline details
func parsePipelineInfo(node *yaml.Node, pipeline *PipelineInfo) {
	if node.Kind != yaml.MappingNode {
		fmt.Println("Expected a mapping node for pipeline")
		return
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		switch keyNode.Value {
		case "name":
			pipeline.Name = &valueNode.Value
			pipeline.NameLine = keyNode.Line
			pipeline.NameColumn = keyNode.Column
		}
	}
}

// parseJobConfig extracts job details and line numbers
func parseJobConfig(node *yaml.Node, job *JobConfiguration) {
	if node.Kind != yaml.MappingNode {
		fmt.Println("Expected a mapping node for job")
		return
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		switch keyNode.Value {
		case "name":
			job.Name = &valueNode.Value
			job.NameLine = keyNode.Line
			job.NameColumn = keyNode.Column
		case "stage":
			job.Stage = &valueNode.Value
			job.StageLine = keyNode.Line
			job.StageColumn = keyNode.Column
		case "image":
			job.Image = &valueNode.Value
			job.ImageLine = keyNode.Line
			job.ImageColumn = keyNode.Column
		case "script":
			job.ScriptLine = keyNode.Line
			job.ScriptColumn = keyNode.Column
			if valueNode.Kind == yaml.SequenceNode {
				for _, item := range valueNode.Content {
					job.Script = append(job.Script, item.Value)
				}
			}
		case "needs":
			job.DependenciesLine = keyNode.Line
			job.DependenciesColumn = keyNode.Column
			if valueNode.Kind == yaml.SequenceNode {
				for _, item := range valueNode.Content {
					job.Dependencies = append(job.Dependencies, item.Value)
				}
			}
		}
	}
}

// Reads YAML file then parse to Pipeline
func ParseYAMLFile(filename string) (*PipelineConfiguration, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var pipeline PipelineConfiguration
	var root yaml.Node

	if err := yaml.Unmarshal(data, &root); err != nil {
		log.Fatalf("%#v\n", err)
	}

	parsePipelineConfig(&root, &pipeline)

	return &pipeline, nil
}

// Validate Pipeline configuration
func (pipeline *PipelineConfiguration) ValidateConfiguration() (bool, int, int, error) {
	// Validate version
	if pipeline.Version == nil || *pipeline.Version != "v0" {
		return false, pipeline.VersionLine, pipeline.PipelineColumn, errors.New("syntax error: version")
	}

	// Validate pipeline info
	if pipeline.Pipeline == nil {
		return false, pipeline.PipelineLine, pipeline.PipelineColumn, errors.New("syntax error: missing key `pipeline`")
	}
	if pipeline.Pipeline.Name == nil || *pipeline.Pipeline.Name == "" {
		return false, pipeline.Pipeline.NameLine, pipeline.Pipeline.NameColumn, errors.New("syntax error: pipeline name is required")
	}

	// Validate stages and jobs
	stages := make(map[string]map[string]*JobConfiguration)

	// Check stages
	if pipeline.Stages == nil {
		return false, pipeline.StagesLine, pipeline.StagesColumn, errors.New("syntax error: missing key `stages`")
	}
	if len(pipeline.Stages) == 0 {
		return false, pipeline.StagesLine, pipeline.StagesColumn, errors.New("syntax error: stages must have at least one item")
	}

	for _, stage := range pipeline.Stages {
		if stage == "" {
			return false, pipeline.StagesLine, pipeline.StagesColumn, errors.New("syntax error: stage name must be a non-empty string")
		} else if _, ok := stages[stage]; ok {
			return false, pipeline.StagesLine, pipeline.StagesColumn, errors.New("syntax error: duplicated stages")
		} else {
			stages[stage] = make(map[string]*JobConfiguration)
		}
	}

	// Check jobs
	if len(pipeline.Jobs) == 0 {
		return false, pipeline.JobsLine, pipeline.JobsColumn, errors.New("syntax error: empty jobs")
	}

	for _, job := range pipeline.Jobs {
		// Check format
		// Name
		if job.Name == nil || *job.Name == "" {
			return false, job.NameLine, job.NameColumn, errors.New("syntax error: empty job name")
		}
		// Stage
		if job.Stage == nil {
			return false, job.StageLine, job.StageColumn, errors.New("syntax error: empty job stage")
		}
		if stages[*job.Stage] == nil {
			return false, job.StageLine, job.StageColumn, errors.New("syntax error: stage `" + *job.Stage + "` must be defined in stages")
		}
		if stages[*job.Stage][*job.Name] != nil {
			return false, job.StageLine, job.StageColumn, errors.New("syntax error: duplicated job name within a stage")
		}
		// Image
		if job.Image == nil || *job.Image == "" {
			return false, job.ImageLine, job.ImageColumn, errors.New("syntax error: empty job image")
		}
		// Script
		if len(job.Script) == 0 {
			return false, job.ScriptLine, job.ScriptColumn, errors.New("syntax error: empty job script")
		}

		// TODO: Check logic
		stages[*job.Stage][*job.Name] = &job
	}

	pipeline.ExecOrder = make(map[string][][]string)
	for stage, jobs := range stages {
		// check stages with empty jobs
		if len(jobs) == 0 {
			return false, pipeline.StagesLine, pipeline.StagesColumn, errors.New("syntax error: stage `" + stage + "` must have at least one job")
		}

		// check cyclic dependencies among jobs within a stage
		pipeline.ExecOrder[stage] = make([][]string, 0)
		indegrees := make(map[string]int)
		graph := make(map[string][]string)
		for name, job := range jobs {
			if indegrees[name] == 0 {
				indegrees[name] = 0
			}
			for _, dependency := range job.Dependencies {
				// dependency job not exist
				if jobs[dependency] == nil {
					return false, 0, 0, errors.New("syntax error: dependency job not exist")
				}
				indegrees[dependency] += 1
				graph[*job.Name] = append(graph[*job.Name], dependency)
			}
		}

		// Execution order
		var parallel [][]string

		for len(indegrees) > 0 {
			parallel = append(parallel, make([]string, 0))

			for job, degree := range indegrees {
				if degree == 0 {
					// add job to execution order
					parallel[len(parallel)-1] = append(parallel[len(parallel)-1], job)
				}
			}

			if len(parallel[len(parallel)-1]) == 0 {
				return false, 0, 0, errors.New("syntax error: cyclic dependencies detected")
			}

			// update indegrees
			for _, job := range parallel[len(parallel)-1] {
				delete(indegrees, job)
				for _, childJob := range graph[job] {
					indegrees[childJob]--
				}
			}
		}

		pipeline.ExecOrder[stage] = parallel
	}

	return true, 0, 0, nil
}
