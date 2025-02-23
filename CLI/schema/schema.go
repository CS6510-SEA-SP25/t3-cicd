package schema

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
	Stages   *ConfigurationNode[map[string]*ConfigurationNode[map[string]*JobConfiguration]]

	/*
		Jobs execution order for each stage (topological)
		e.g. "build" -> [[job1, job2], [job3], [job4]]
	*/
	StageOrder []string
	ExecOrder  map[string][][]string
}

// GitHub repository configuration
type Repository struct {
	Url        string // Repository URL
	CommitHash string // Git commit hash value
}
