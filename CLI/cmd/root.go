/*
Copyright © 2025 Minh Nguyen minh160302@gmail.com
*/
package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"cicd/pipeci/schema"

	"github.com/spf13/cobra"
)

/*
Command line variables
*/
var (
	filename   string
	check      bool
	showDryRun bool
	pipeline   schema.PipelineConfiguration
)

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:           "pipeci",
	Short:         "A CLI application to run pipelines locally.",
	Long:          `pipeci helps you execute your CI/CD pipelines on both local and remote environments.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	// pipeci [flags]
	RunE: func(cmd *cobra.Command, args []string) error {
		// validate
		err := isGitRoot()
		if err != nil {
			return errors.New("current directory must be root of a Git repository")
		}

		// FLAGS PROCESSING
		if showDryRun {
			check = true
		}

		// flags
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
						jobOrder = append(jobOrder, "\t"+jobName+":")
						jobOrder = append(jobOrder, "\t\timage: "+job.Image.Value)

						var jobScript []string
						jobScript = append(jobScript, "\t\tscript:")
						for _, script := range job.Script.Value {
							jobScript = append(jobScript, "\t\t\t- "+script)
						}

						jobOrder = append(jobOrder, strings.Join(jobScript, "\n"))
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

// Init function
func init() {
	// --filename | -f
	RootCmd.PersistentFlags().StringVarP(&filename, "filename", "f", ".pipelines/pipeline.yaml", "Path to the pipeline configuration file.")

	// --check | -c
	RootCmd.PersistentFlags().BoolVarP(&check, "check", "c", false, "Validate the pipeline configuration file.")

	// --dry-run
	RootCmd.PersistentFlags().BoolVar(&showDryRun, "dry-run", false, "Show the jobs execution order.")
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	err := RootCmd.Execute()
	return err
}
