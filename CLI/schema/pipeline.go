package schema

import (
	"errors"
	"os"
	"strings"

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
		return nil, err
	}

	parsePipelineConfig(&root, &pipeline)

	return &pipeline, nil
}

// Validate Pipeline configuration
func (pipeline *PipelineConfiguration) ValidateConfiguration() (int, int, error) {
	// Validate version
	if pipeline.Version == nil || *pipeline.Version != "v0" {
		return pipeline.VersionLine, pipeline.PipelineColumn, errors.New("syntax error: version")
	}

	// Validate pipeline info
	if pipeline.Pipeline == nil {
		return pipeline.PipelineLine, pipeline.PipelineColumn, errors.New("syntax error: missing key `pipeline`")
	}
	if pipeline.Pipeline.Name == nil || *pipeline.Pipeline.Name == "" {
		return pipeline.Pipeline.NameLine, pipeline.Pipeline.NameColumn, errors.New("syntax error: pipeline name is required")
	}

	// Validate stages and jobs
	stages := make(map[string]map[string]*JobConfiguration)

	// Check stages
	if pipeline.Stages == nil {
		return pipeline.StagesLine, pipeline.StagesColumn, errors.New("syntax error: missing key `stages`")
	}
	if len(pipeline.Stages) == 0 {
		return pipeline.StagesLine, pipeline.StagesColumn, errors.New("syntax error: stages must have at least one item")
	}

	for _, stage := range pipeline.Stages {
		if stage == "" {
			return pipeline.StagesLine, pipeline.StagesColumn, errors.New("syntax error: stage name must be a non-empty string")
		} else if _, ok := stages[stage]; ok {
			return pipeline.StagesLine, pipeline.StagesColumn, errors.New("syntax error: duplicated stages")
		} else {
			stages[stage] = make(map[string]*JobConfiguration)
		}
	}

	// Check jobs
	if len(pipeline.Jobs) == 0 {
		return pipeline.JobsLine, pipeline.JobsColumn, errors.New("syntax error: empty jobs")
	}

	for _, job := range pipeline.Jobs {
		// Check format
		// Name
		if job.Name == nil {
			return job.NameLine, job.NameColumn, errors.New("syntax error: missing job name")
		}
		if *job.Name == "" {
			return pipeline.JobsLine, pipeline.JobsColumn, errors.New("syntax error: job name must be a non-empty string")
		}

		// Stage
		if job.Stage == nil {
			return job.NameLine, job.NameColumn, errors.New("syntax error: job `" + *job.Name + "` is missing stage")
		}
		if *job.Stage == "" {
			return job.StageLine, job.StageColumn, errors.New("syntax error: job stage must be a non-empty string")
		}
		if stages[*job.Stage] == nil {
			return job.StageLine, job.StageColumn, errors.New("syntax error: stage `" + *job.Stage + "` must be defined in stages")
		}
		if stages[*job.Stage][*job.Name] != nil {
			return job.StageLine, job.StageColumn, errors.New("syntax error: duplicated job name within a stage")
		}

		// Image
		if job.Image == nil {
			return job.NameLine, job.NameColumn, errors.New("syntax error: job `" + *job.Name + "` is missing image")
		}
		if *job.Image == "" {
			return job.ImageLine, job.ImageColumn, errors.New("syntax error: job image must be a non-empty string")
		}

		// Script
		if job.Script == nil {
			return job.NameLine, job.NameColumn, errors.New("syntax error: job `" + *job.Name + "` is missing script")
		}
		if len(job.Script) == 0 {
			return job.ScriptLine, job.ScriptColumn, errors.New("syntax error: empty job script")
		}

		stages[*job.Stage][*job.Name] = &job
	}

	// Validate logic
	pipeline.ExecOrder = make(map[string][][]string)
	for stage, jobs := range stages {
		// check stages with empty jobs
		if len(jobs) == 0 {
			return pipeline.StagesLine, pipeline.StagesColumn, errors.New("syntax error: stage `" + stage + "` must have at least one job")
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
					return 0, 0, errors.New("syntax error: dependency job not exist")
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
				// find cycle
				cycle := make([]string, 0)
				for remainingJobName := range indegrees {
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
				return jobs[cycleHead].NameLine, jobs[cycleHead].NameColumn, errors.New("syntax error: cyclic dependencies detected: " + cycleStr)
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

	return 0, 0, nil
}

// TODO: Refactor this code
// func () getExecutionOrder() (int, int, error) {

// }

/*
Trace the dependencies cycle.
*/
func (job *JobConfiguration) checkCycle(visited map[string]bool, jobs map[string]*JobConfiguration, cycle *[]string) bool {
	if visited[*job.Name] {
		return true
	}

	visited[*job.Name] = true
	for _, name := range job.Dependencies {
		child := jobs[name]
		*cycle = append(*cycle, name)
		if child.checkCycle(visited, jobs, cycle) {
			return true
		}
		*cycle = (*cycle)[:len(*cycle)-1]
	}

	return false
}
