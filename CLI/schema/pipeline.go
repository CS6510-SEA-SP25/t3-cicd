package schema

import (
	"errors"
	"os"
	"strings"

	"fmt"

	"gopkg.in/yaml.v3"
)

// Location of ConfigurationNode in YAML file.
type YAMLFileLocation struct {
	Line   int
	Column int
}

// Configuration Node contains value and location of YAML element.
type ConfigurationNode[T any] struct {
	Value    T
	Location *YAMLFileLocation
}

// Job configuration.
type JobConfiguration struct {
	// (required) Name of job within a stage.
	Name *ConfigurationNode[string]
	// (required) Stage name.
	Stage *ConfigurationNode[string]
	// (required) Docker image to be used.
	Image *ConfigurationNode[string]
	// (required) List of scripts to be executed sequentially.
	Script *ConfigurationNode[[]string]
	// (optional) Jobs within the same statge that must complete successfully before this job can start executing.
	Dependencies *ConfigurationNode[[]string]
}

// Pipeline identifier info.
type PipelineInfo struct {
	Name *ConfigurationNode[string] // (required) Name of pipeline.
}

// Pipeline configuration
type PipelineConfiguration struct {
	Version  *ConfigurationNode[string] // (required) API version. Currently set at v0.
	Pipeline *ConfigurationNode[PipelineInfo]
	Stages   *ConfigurationNode[[]string]           // (required) List of stages. Also represent the order of execution.
	Jobs     *ConfigurationNode[[]JobConfiguration] // (required) List of jobs' configuration. Each stage has at least one job.

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
			config.Version = &ConfigurationNode[string]{Value: valueNode.Value, Location: &YAMLFileLocation{Line: keyNode.Line, Column: keyNode.Column}}
		case "pipeline":
			config.Pipeline = &ConfigurationNode[PipelineInfo]{Location: &YAMLFileLocation{Line: keyNode.Line, Column: keyNode.Column}}
			parsePipelineInfo(valueNode, &config.Pipeline.Value)
		case "stages":
			config.Stages = &ConfigurationNode[[]string]{Value: make([]string, 0), Location: &YAMLFileLocation{Line: keyNode.Line, Column: keyNode.Column}}
			if valueNode.Kind == yaml.SequenceNode {
				for _, item := range valueNode.Content {
					config.Stages.Value = append(config.Stages.Value, item.Value)
				}
			}
		case "jobs":
			config.Jobs = &ConfigurationNode[[]JobConfiguration]{Value: make([]JobConfiguration, 0), Location: &YAMLFileLocation{Line: keyNode.Line, Column: keyNode.Column}}
			if valueNode.Kind == yaml.SequenceNode {
				for _, jobNode := range valueNode.Content {
					var job JobConfiguration
					parseJobConfig(jobNode, &job)
					config.Jobs.Value = append(config.Jobs.Value, job)
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
			pipeline.Name = &ConfigurationNode[string]{Value: valueNode.Value, Location: &YAMLFileLocation{Line: keyNode.Line, Column: keyNode.Column}}
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
			job.Name = &ConfigurationNode[string]{Value: valueNode.Value, Location: &YAMLFileLocation{Line: keyNode.Line, Column: keyNode.Column}}
		case "stage":
			job.Stage = &ConfigurationNode[string]{Value: valueNode.Value, Location: &YAMLFileLocation{Line: keyNode.Line, Column: keyNode.Column}}
		case "image":
			job.Image = &ConfigurationNode[string]{Value: valueNode.Value, Location: &YAMLFileLocation{Line: keyNode.Line, Column: keyNode.Column}}
		case "script":
			job.Script = &ConfigurationNode[[]string]{Value: make([]string, 0), Location: &YAMLFileLocation{Line: keyNode.Line, Column: keyNode.Column}}
			if valueNode.Kind == yaml.SequenceNode {
				for _, item := range valueNode.Content {
					job.Script.Value = append(job.Script.Value, item.Value)
				}
			}
		case "needs":
			job.Dependencies = &ConfigurationNode[[]string]{Value: make([]string, 0), Location: &YAMLFileLocation{Line: keyNode.Line, Column: keyNode.Column}}
			if valueNode.Kind == yaml.SequenceNode {
				for _, item := range valueNode.Content {
					job.Dependencies.Value = append(job.Dependencies.Value, item.Value)
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
		return nil, err
	}

	parsePipelineConfig(&root, &pipeline)

	return &pipeline, nil
}

// Validate Pipeline configuration
func (pipeline *PipelineConfiguration) ValidateConfiguration() (YAMLFileLocation, error) {
	// Validate version
	if pipeline.Version == nil {
		return *pipeline.Pipeline.Location, errors.New("syntax error: missing key `pipeline`")
	}
	if pipeline.Version.Value != "v0" {
		return *pipeline.Version.Location, errors.New("syntax error: invalid version")
	}

	// Validate pipeline info
	if pipeline.Pipeline == nil {
		return *pipeline.Pipeline.Location, errors.New("syntax error: missing key `pipeline`")
	}
	if pipeline.Pipeline.Value.Name == nil || pipeline.Pipeline.Value.Name.Value == "" {
		return *pipeline.Pipeline.Value.Name.Location, errors.New("syntax error: pipeline name is required")
	}

	// Validate stages and jobs
	stages := make(map[string]map[string]*JobConfiguration)

	// Check stages
	if pipeline.Stages == nil {
		return *pipeline.Stages.Location, errors.New("syntax error: missing key `stages`")
	}
	if len(pipeline.Stages.Value) == 0 {
		return *pipeline.Stages.Location, errors.New("syntax error: stages must have at least one item")
	}

	for _, stage := range pipeline.Stages.Value {
		if stage == "" {
			return *pipeline.Stages.Location, errors.New("syntax error: stage name must be a non-empty string")
		} else if _, ok := stages[stage]; ok {
			return *pipeline.Stages.Location, errors.New("syntax error: duplicated stages")
		} else {
			stages[stage] = make(map[string]*JobConfiguration)
		}
	}

	// Check jobs
	if len(pipeline.Jobs.Value) == 0 {
		return *pipeline.Jobs.Location, errors.New("syntax error: empty jobs")
	}

	// Check format
	for _, job := range pipeline.Jobs.Value {
		// Name
		if job.Name == nil {
			return *job.Name.Location, errors.New("syntax error: missing job name")
		}
		if job.Name.Value == "" {
			return *pipeline.Jobs.Location, errors.New("syntax error: job name must be a non-empty string")
		}

		// Stage
		if job.Stage == nil {
			return *job.Name.Location, errors.New("syntax error: job `" + job.Name.Value + "` is missing stage")
		}
		if job.Stage.Value == "" {
			return *job.Stage.Location, errors.New("syntax error: job stage must be a non-empty string")
		}
		if stages[job.Stage.Value] == nil {
			return *job.Stage.Location, errors.New("syntax error: stage `" + job.Stage.Value + "` must be defined in stages")
		}
		if stages[job.Stage.Value][job.Name.Value] != nil {
			return *job.Stage.Location, errors.New("syntax error: duplicated job name within a stage")
		}

		// Image
		if job.Image == nil {
			return *job.Name.Location, errors.New("syntax error: job `" + job.Name.Value + "` is missing image")
		}
		if job.Image.Value == "" {
			return *job.Image.Location, errors.New("syntax error: job image must be a non-empty string")
		}

		// Script
		if job.Script == nil {
			return *job.Name.Location, errors.New("syntax error: job `" + job.Name.Value + "` is missing script")
		}
		if len(job.Script.Value) == 0 {
			return *job.Script.Location, errors.New("syntax error: empty job script")
		}

		stages[job.Stage.Value][job.Name.Value] = &job
	}

	// Validate logic
	pipeline.ExecOrder = make(map[string][][]string)
	for stage, jobs := range stages {
		// check stages with empty jobs
		if len(jobs) == 0 {
			return *pipeline.Stages.Location, errors.New("syntax error: stage `" + stage + "` must have at least one job")
		}

		pipeline.ExecOrder[stage] = make([][]string, 0)
		indegrees := make(map[string]int)
		graph := make(map[string][]string)
		for name, job := range jobs {
			if indegrees[name] == 0 {
				indegrees[name] = 0
			}

			if job.Dependencies != nil {
				for _, dependency := range job.Dependencies.Value {
					// dependency job not exist
					if jobs[dependency] == nil {
						return *job.Dependencies.Location, errors.New("syntax error: dependency job not exist")
					}
					indegrees[job.Name.Value] += 1
					graph[dependency] = append(graph[dependency], job.Name.Value)
				}
			}
		}

		// Execution order
		var parallel [][]string

		// check cyclic dependencies among jobs within a stage
		hasCycle, location, validateErr := detectCycle(&parallel, &indegrees, jobs, graph)
		if hasCycle && validateErr != nil {
			return location, validateErr
		}

		pipeline.ExecOrder[stage] = parallel
	}

	return *pipeline.Version.Location, nil
}

/*
Detect cyclic dependencies in a stage.
*/
func detectCycle(parallel *[][]string, indegrees *map[string]int, jobs map[string]*JobConfiguration, graph map[string][]string) (bool, YAMLFileLocation, error) {
	for len(*indegrees) > 0 {
		*parallel = append(*parallel, make([]string, 0))

		for job, degree := range *indegrees {
			if degree == 0 {
				// add job to execution order
				(*parallel)[len(*parallel)-1] = append((*parallel)[len(*parallel)-1], job)
			}
		}

		if len((*parallel)[len(*parallel)-1]) == 0 {
			// find cycle
			cycle := make([]string, 0)
			for remainingJobName := range *indegrees {
				visited := make(map[string]bool)
				if jobs[remainingJobName].checkCycle(visited, jobs, &cycle) {
					break
				}
			}
			if len(cycle) == 0 {
				panic("cycle is not supposed to be empty")
			}
			cycleHead := cycle[0]
			cycleStr := strings.Join(cycle, " -> ")
			return true, *jobs[cycleHead].Name.Location, errors.New("syntax error: cyclic dependencies detected: " + cycleStr)
		}

		// update indegrees
		for _, job := range (*parallel)[len(*parallel)-1] {
			delete(*indegrees, job)
			for _, childJob := range graph[job] {
				(*indegrees)[childJob]--
			}
		}
	}
	return false, YAMLFileLocation{}, nil
}

/*
Trace the dependency cycle.
*/
func (job *JobConfiguration) checkCycle(visited map[string]bool, jobs map[string]*JobConfiguration, cycle *[]string) bool {
	if visited[job.Name.Value] {
		return true
	}

	visited[job.Name.Value] = true
	for _, name := range job.Dependencies.Value {
		child := jobs[name]
		*cycle = append(*cycle, name)
		if child.checkCycle(visited, jobs, cycle) {
			return true
		}
		*cycle = (*cycle)[:len(*cycle)-1]
	}

	return false
}
