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

	"cicd/gocc/schema"

	"github.com/spf13/cobra"
)

/*
Command line variables
*/
var (
	filename string
	check    bool
	pipeline schema.PipelineConfiguration
)

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:           "gocc",
	Short:         "A CLI application to run pipelines locally.",
	Long:          `GoCC helps you execute your CI/CD pipelines on both local and remote environments.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	// gocc [flags]
	RunE: func(cmd *cobra.Command, args []string) error {
		// validate
		err := isGitRoot()
		if err != nil {
			return errors.New("current directory must be root of a Git repository")
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

func isYAMLFile(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".yaml" || ext == ".yml"
}

/*
Flag handlers
*/
// --filename | -f
func HandleFilenameFlag() error {
	// Default filename if not provided
	if filename == "" {
		filename = ".pipelines/pipeline.yaml"
		log.Printf("Using default configuration file at %v\n", filename)
	} else {
		log.Printf("Using input configuration file at %v\n", filename)
	}

	// Check file extension
	if !isYAMLFile(filename) {
		return errors.New("configuration file must be a YAML file")
	}

	// Parse configuration file
	pConfig, err := schema.ParseYAMLFile(filename)
	if err != nil {
		return err
	}
	pipeline = *pConfig
	return nil
}

// --check | -c
// Validate configuration then exit.
func HandleCheckFlag() error {
	if check {
		// Validate configuration
		location, validateErr := pipeline.ValidateConfiguration()
		if validateErr != nil {
			// Format error message
			return fmt.Errorf("%s:%d:%d: %s", filename, location.Line, location.Column, validateErr.Error())
		} else {
			log.Print("Pipeline configuration is valid.")
		}

		log.Printf("Execution Order: %#v\n", pipeline.ExecOrder)
	}
	return nil
}

// Init function
func init() {
	// --filename | -f
	RootCmd.PersistentFlags().StringVarP(&filename, "filename", "f", ".pipelines/pipeline.yaml", "Path to the pipeline configuration file.")

	// --check | -c
	RootCmd.PersistentFlags().BoolVarP(&check, "check", "c", false, "Validate the pipeline configuration file.")
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	err := RootCmd.Execute()
	return err
}
