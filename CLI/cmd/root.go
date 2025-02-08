/*
Copyright © 2025 Minh Nguyen minh160302@gmail.com
*/
package cmd

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"

	"cicd/gocc/schema"

	"github.com/spf13/cobra"
)

/*
Command line variables
*/
var (
	filename string
	check    bool
	verbose  bool
	pipeline schema.PipelineConfiguration
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gocc",
	Short: "A CLI application to run pipelines locally.",
	Long:  `GoCC helps you execute your CI/CD pipelines on both local and remote environments.`,

	// gocc [flags]
	Run: func(cmd *cobra.Command, args []string) {
		// verbose mod
		HandleVerboseFlag()

		// validate
		if !isGitRoot() {
			log.Fatal("Current directory must be root of a Git repository")
		}

		// flags
		HandleFilenameFlag()
		HandleCheckFlag()
	},
}

// Format error message
func logPipelineErr(err error, line int, column int) {
	log.New(os.Stderr, "", 0).Fatalf("%s:%d:%d: %s", filename, line, column, err.Error())
}

/*
Validate current directory
*/
// Checks if the current directory is the root of a Git repository.
func isGitRoot() bool {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current directory: %v", err)
		return false
	}

	// Run 'git rev-parse --show-toplevel' to get the root of the repository
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err = cmd.Run()
	if err != nil {
		return false
	}

	gitRoot := strings.TrimSpace(out.String())
	return cwd == gitRoot
}

/*
Flag handlers
*/
// --filename | -f
func HandleFilenameFlag() {
	// Default filename if not provided
	if filename == "" {
		filename = ".pipelines/pipeline.yaml"
		log.Printf("Using default configuration file at %v\n", filename)
	} else {
		log.Printf("Using input configuration file at %v\n", filename)
	}

	// Parse configuration file
	pConfig, err := schema.ParseYAMLFile(filename)
	pipeline = *pConfig
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
}

// --check | -c
// Validate configuration then exit.
func HandleCheckFlag() {
	if check {
		// Validate configuration
		isPipelineValid, errLine, errColumn, validateErr := pipeline.ValidateConfiguration()
		if validateErr != nil {
			// log.Fatal(validateErr)
			logPipelineErr(validateErr, errLine, errColumn)
		} else {
			if isPipelineValid {
				log.Print("Pipeline configuration is valid.")
			} else {
				log.Print("Pipeline configuration is invalid.")
			}
		}

		log.Printf("Execution Order: %#v\n", pipeline.ExecOrder)
	}
}

// --verbose | -v
func HandleVerboseFlag() {
	// if verbose {
	// 	fmt.Println("Verbose mode enabled")
	// }
}

// Init function
func init() {
	// --filename | -f
	rootCmd.PersistentFlags().StringVarP(&filename, "filename", "f", ".pipelines/pipeline.yaml", "Path to the pipeline configuration file.")

	// --check | -c
	rootCmd.PersistentFlags().BoolVarP(&check, "check", "c", false, "Validate the pipeline configuration file.")

	// --verbose | -v
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output.")
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
