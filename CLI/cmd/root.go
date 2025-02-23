/*
Copyright © 2025 Minh Nguyen minh160302@gmail.com
*/
package cmd

import (
	"bytes"
	"cicd/pipeci/apis"
	schema "cicd/pipeci/schema"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

/*
Command line variables
*/
var (
	// Global flags
	filename   string
	check      bool
	showDryRun bool
	isLocal    bool
	repo       string
	commit     string

	// report subFlags
	reportPipelineName string
	reportRunCounter   int

	pipeline schema.PipelineConfiguration

	GlobalDirectory string = "." // Current directory
)

/* Base handler for all commands under root */
func mandatoryProcess(cmd *cobra.Command) error {
	// validate
	err := isGitRoot()
	if err != nil {
		return errors.New("current directory must be root of a Git repository")
	}

	// FLAGS PROCESSING
	// Validate configuration during `run`
	if cmd.Use == "run" {
		check = true
	}
	// Validate configuration file during dry-run
	if showDryRun {
		check = true
	}

	// flags
	err = HandleRepoFlag()
	if err != nil {
		return err
	}

	err = HandleFilenameFlag()
	if err != nil {
		return err
	}

	err = HandleCheckFlag()
	if err != nil {
		return err
	}

	err = HandleDryRunFlag()
	if err != nil {
		return err
	}
	return nil
}

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:           "pipeci",
	Short:         "A CLI application to run pipelines locally.",
	Long:          `pipeci helps you execute your CI/CD pipelines on both local and remote environments.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	// pipeci [flags]
	RunE: func(cmd *cobra.Command, args []string) error {
		return mandatoryProcess(cmd)
	},
}

/*
Validate current directory
*/
// Checks if the current directory is the root of a Git repository.
// Check if .git directory exists in current path
func isGitRoot() error {
	_, err := os.Stat(".git")
	return err
}

// Check YAML file
func isYAMLFile(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".yaml" || ext == ".yml"
}

// Check file exists
func doesfileExist(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// Get GitHub local repository info with authentication using Personal Access Token (PAT)
func getLocalGitRepo() (schema.Repository, error) {
	repository := schema.Repository{}

	// Get the Git remote URL
	url, err := runGitCommand("config", "--get", "remote.origin.url")
	if err != nil {
		return repository, fmt.Errorf("failed to get repository URL: %v", err)
	}
	repository.Url = strings.Trim(url, "\n")

	// ! ❯ export GITHUB_TOKEN=''
	token := os.Getenv("GITHUB_TOKEN")

	// Configure GITHUB_TOKEN if exists
	if token != "" {
		repository.Url = strings.Replace(repository.Url, "https://", "https://"+token+"@", 1)
	} else {
		log.Println("Warning: No GitHub token provided for authentication.")
	}

	// Get the latest commit hash if not specified
	if commit == "" {
		commitHash, err := runGitCommand("rev-parse", "HEAD")
		if err != nil {
			return repository, fmt.Errorf("failed to get commit hash: %v", err)
		}
		repository.CommitHash = strings.Trim(commitHash, "\n")
	} else {
		repository.CommitHash = commit
	}

	return repository, nil
}

/*
Executes a Git command at the current/specified directory
*/
func runGitCommand(args ...string) (string, error) {
	if GlobalDirectory != "." {
		args = append([]string{"-C", GlobalDirectory}, args...)
	}
	cmd := exec.Command("git", args...)
	cmd.Stderr = os.Stderr

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

/* Flag handlers */
// --filename | -f
func HandleFilenameFlag() error {
	// Default filename if not provided
	if filename == "" {
		filename = ".pipelines/pipeline.yaml"
		log.Printf("Using default configuration file at %v\n", filename)
	} else {
		log.Printf("Using input configuration file at %v\n", filename)
	}

	// Check file exists
	if !doesfileExist(filename) {
		return errors.New("no such file or directory")
	}

	// Check file extension
	if !isYAMLFile(filename) {
		return errors.New("configuration file must be a YAML file")
	}

	return nil
}

/* Validate configuration then exit. */
// --check | -c
func HandleCheckFlag() error {
	if check {
		// Parse configuration file
		pConfig, location, err := schema.ParseYAMLFile(filename)
		if err != nil {
			return fmt.Errorf("%s:%d:%d: %s", filename, location.Line, location.Column, err.Error())
		}
		pipeline = *pConfig

		// Validate configuration
		location, validateErr := pipeline.ValidateConfiguration()
		if validateErr != nil {
			// Format error message
			return fmt.Errorf("%s:%d:%d: %s", filename, location.Line, location.Column, validateErr.Error())
		} else {
			log.Print("Pipeline configuration is valid.")
		}
	}
	return nil
}

/* Show the jobs execution order. */
// --dry-run
func HandleDryRunFlag() error {
	if showDryRun {
		if pipeline.ExecOrder == nil {
			panic("empty excution order when pipeline configuration is valid")
		} else {
			var orders []string
			for _, stageName := range pipeline.StageOrder {
				jobs := pipeline.ExecOrder[stageName]
				var stageOrder []string
				stageOrder = append(stageOrder, stageName+":")

				for _, level := range jobs {
					for _, jobName := range level {
						job := pipeline.Stages.Value[stageName].Value[jobName]
						var jobOrder []string
						// name
						jobOrder = append(jobOrder, "\t"+jobName+":")

						// image
						jobOrder = append(jobOrder, "\t\timage: "+job.Image.Value)

						// script
						var jobScript []string
						jobScript = append(jobScript, "\t\tscript:")
						for _, script := range job.Script.Value {
							jobScript = append(jobScript, "\t\t\t- "+script)
						}
						jobOrder = append(jobOrder, strings.Join(jobScript, "\n"))

						// needs
						if job.Dependencies != nil && len(job.Dependencies.Value) > 0 {
							var jobDependencies []string
							jobDependencies = append(jobDependencies, "\t\tneeds:")
							for _, dep := range job.Dependencies.Value {
								jobDependencies = append(jobDependencies, "\t\t\t- "+dep)
							}
							jobOrder = append(jobOrder, strings.Join(jobDependencies, "\n"))
						}

						stageOrder = append(stageOrder, strings.Join(jobOrder, "\n"))
					}
				}
				orders = append(orders, strings.Join(stageOrder, "\n"))
			}
			log.Println(strings.Join(orders, "\n"))
		}
	}
	return nil
}

/*
Specifies the location of the repository to use (--repo must a local directory)
*/
// --repo
func HandleRepoFlag() error {
	// Default repo if not provided
	if repo == "" {
		GlobalDirectory = "."
		return nil
	}

	// Check if repo is a local directory
	fileInfo, err := os.Stat(repo)
	if err == nil && fileInfo.IsDir() {
		// If it's a local directory, change to it
		GlobalDirectory = repo
		return nil
	} else {
		return fmt.Errorf("--repo value must be a valid local directory")
	}
}

// Sub-command: pipeci run
var RunCmd = &cobra.Command{
	Use:           "run",
	Short:         "usage: pipeci run",
	Long:          "Execute the pipeline on local machine when pipeline configuration is valid",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := mandatoryProcess(cmd)

		if err != nil {
			return err
		}

		// TODO 1: --local: pull image and run
		// TODO 2: no local: run the deployed version
		if isLocal {
			var repository schema.Repository
			repository, err = getLocalGitRepo()
			if err != nil {
				return fmt.Errorf("error while getting local repository info: %v", err)
			}
			err = apis.ExecuteLocal(pipeline, repository)
		}

		return err
	},
}

// Sub-command: pipeci report
var ReportCmd = &cobra.Command{
	Use:           "report",
	Short:         "usage: pipeci report",
	Long:          "Report on past pipeline execution by input parameters",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := mandatoryProcess(cmd)
		if err != nil {
			return err
		}

		if isLocal {
			var repository schema.Repository
			repository, err = getLocalGitRepo()
			if err != nil {
				return fmt.Errorf("error while getting local repository info: %v", err)
			}

			// Show summary all past pipeline runs for the local repository if no pipeline name specified
			if reportPipelineName == "" {
				err = apis.ReportPastExecutionsLocal(repository)
			}
		}

		return err
	},
}

// Init function
func init() {
	// --filename | -f
	RootCmd.PersistentFlags().StringVarP(&filename, "filename", "f", ".pipelines/pipeline.yaml", "Path to the pipeline configuration file.")

	// --check | -c
	RootCmd.PersistentFlags().BoolVarP(&check, "check", "c", false, "Validate the pipeline configuration file")

	// --dry-run
	RootCmd.PersistentFlags().BoolVar(&showDryRun, "dry-run", false, "Show the jobs execution order.")

	// --local
	RootCmd.PersistentFlags().BoolVar(&isLocal, "local", false, "Execute the pipeline locally.")

	// TODO: check repo and commit
	// --repo
	RootCmd.PersistentFlags().StringVar(&repo, "repo", "", "Specify GitHub repository.")

	// --commit
	RootCmd.PersistentFlags().StringVar(&commit, "commit", "", "Specify Git commit hash.")

	// report --pipeline "code-review"
	ReportCmd.LocalFlags().StringVar(&reportPipelineName, "pipeline", "", "Returns the list of all pipeline runs for the specified pipeline")

	// report --run 2
	ReportCmd.LocalFlags().IntVar(&reportRunCounter, "run", 0, "Run number i-th for a specified pipeline name")

	// run
	RootCmd.AddCommand(RunCmd)

	// report
	RootCmd.AddCommand(ReportCmd)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	err := RootCmd.Execute()
	return err
}
