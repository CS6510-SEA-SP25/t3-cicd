package cmd_test

import (
	"bytes"
	"cicd/pipeci/cmd"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"strings"
	"testing"
)

/*
Test running the CLI in Git-initialized directory
*/
func TestGitRepository(t *testing.T) {
	// Store the original directory to restore later
	originalDir, _ := os.Getwd()
	// Restore original directory after test
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Failed to return to original directory: %v\n", err)
		}
	}()

	// Change to a wrong directory (assume it's root for test purposes)
	err := os.Chdir("/")
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Execute the command
	err = cmd.Execute()

	if err == nil {
		t.Errorf("expected an error but got none")
	}

	expectedError := "current directory must be root of a Git repository"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("unexpected error message: %v", err)
	}
}

/*
Test --filename | -f
*/
func TestFileNameFlag(t *testing.T) {
	// Store the original directory to restore later
	originalDir, _ := os.Getwd()
	// Restore original directory after test
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Failed to return to original directory: %v\n", err)
		}
	}()

	// Change to a wrong directory (assume it's root for test purposes)
	err := os.Chdir("../../")
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// * TEST CASES //
	// Default config file
	err = cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error message: %v", err)
	}

	err = cmd.RootCmd.Flags().Set("filename", "")
	if err != nil {
		t.Errorf("unexpected error message: %v", err)
	}

	err = cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error message: %v", err)
	}

	// Config file not exists
	err = cmd.RootCmd.Flags().Set("filename", "./not_exist.yaml")
	if err != nil {
		t.Errorf("unexpected error message: %v", err)
	}

	err = cmd.Execute()
	if err == nil {
		t.Errorf("expected an error but got none")
	} else {
		expectedError := "no such file or directory"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("unexpected error message: %v", err)
		}
	}

	// Config file exists but not YAML
	err = cmd.RootCmd.Flags().Set("filename", "./README.md")
	if err != nil {
		t.Errorf("unexpected error message: %v", err)
	}

	err = cmd.Execute()
	if err == nil {
		t.Errorf("expected an error but got none")
	} else {
		expectedError := "configuration file must be a YAML file"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("unexpected error message: %v", err)
		}
	}

	// Default config file
	err = cmd.RootCmd.Flags().Set("filename", ".pipelines/pipeline.yaml")
	if err != nil {
		t.Errorf("unexpected error message: %v", err)
	}

	err = cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error message: %v", err)
	}
}

/*
Test --check | -c
*/
func TestCheckFlag(t *testing.T) {
	// Store the original directory to restore later
	originalDir, _ := os.Getwd()
	// Restore original directory after test
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Failed to return to original directory: %v\n", err)
		}
	}()

	// Change to a wrong directory (assume it's root for test purposes)
	err := os.Chdir("../../")
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	err = cmd.RootCmd.Flags().Set("check", "true")
	if err != nil {
		t.Errorf("unexpected error message: %v", err)
	}

	err = cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error message: %v", err)
	}
}

/*
Test --dry-run no dependencies
*/
func TestDryRunOne(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)      // Redirect log output to buffer
	defer log.SetOutput(nil) // Reset after test

	// Store the original directory to restore later
	originalDir, _ := os.Getwd()
	// Restore original directory after test
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Failed to return to original directory: %v\n", err)
		}
	}()

	// Change to a wrong directory (assume it's root for test purposes)
	err := os.Chdir("../../")
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	err = cmd.RootCmd.PersistentFlags().Set("filename", "./.pipelines/test/dry_run_success.yaml")
	if err != nil {
		t.Errorf("unexpected error message: %v", err)
	}

	err = cmd.RootCmd.PersistentFlags().Set("dry-run", "true")
	if err != nil {
		t.Errorf("unexpected error message: %v", err)
	}

	err = cmd.Execute()
	assert.NoError(t, err)
	// 	if err != nil {
	// 		t.Errorf("unexpected error message: %v", err)
	// 	} else {
	// 		// Get log output as a string
	// 		loggedOutput := buf.String()

	// 		// Assert output
	// 		expected := `build:
	// 	compile:
	// 		image: gradle:8.12-jdk21
	// 		script:
	// 			- ./gradlew classes
	// test:
	// 	unittests:
	// 		image: gradle:8.12-jdk21
	// 		script:
	// 			- ./gradlew test
	// 	reports:
	// 		image: gradle:8.12-jdk21
	// 		script:
	// 			- ./gradlew check
	// docs:
	// 	javadoc:
	// 		image: gradle:8.12-jdk21
	// 		script:
	// 			- ./gradlew javadoc`

	// 		// cleanedText := strings.ReplaceAll(loggedOutput, "	", " ")
	// 		// cleanedText = strings.ReplaceAll(cleanedText, "\n", " ")
	// 		// cleanedText = strings.ReplaceAll(cleanedText, " ", "")
	// 		// cleanedExpected := strings.ReplaceAll(expected, "\t", " ")
	// 		// cleanedExpected = strings.ReplaceAll(cleanedExpected, "\n", " ")
	// 		// cleanedExpected = strings.ReplaceAll(cleanedExpected, " ", "")

	// 		// if !strings.Contains(cleanedText, cleanedExpected) {
	// 		// 	t.Errorf("expected %q but got %q", cleanedExpected, cleanedText)
	// 		// }

	// }
}

/*
Test --dry-run with dependencies
*/
func TestDryRunTwo(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)      // Redirect log output to buffer
	defer log.SetOutput(nil) // Reset after test

	// Store the original directory to restore later
	originalDir, _ := os.Getwd()
	// Restore original directory after test
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Failed to return to original directory: %v\n", err)
		}
	}()

	// Change to a wrong directory (assume it's root for test purposes)
	err := os.Chdir("../../")
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	err = cmd.RootCmd.PersistentFlags().Set("filename", "./.pipelines/pipeline.yaml")
	if err != nil {
		t.Errorf("unexpected error message: %v", err)
	}

	err = cmd.RootCmd.PersistentFlags().Set("dry-run", "true")
	if err != nil {
		t.Errorf("unexpected error message: %v", err)
	}

	err = cmd.Execute()
	assert.NoError(t, err)
	// 	if err != nil {
	// 		t.Errorf("unexpected error message: %v", err)
	// 	} else {
	// 		// Get log output as a string
	// 		loggedOutput := buf.String()

	// 		// Assert output
	// 		expected := `build:
	// 	compile:
	// 		image: maven:3.8.6
	// 		script:
	//     		- mvn clean install
	// test:
	// 	run-test:
	// 		image: maven:3.8.6
	// 		script:
	// 			- mvn test
	// 	checkstyle:
	// 		image: maven:3.8.6
	// 		script:
	// 			- mvn checkstyle:check
	// 			- mvn checkstyle:checkstyle
	// 		needs:
	// 			- run-test
	// 	check-coverage:
	// 		image: maven:3.8.6
	// 		script:
	// 			- mvn verify
	// 			- mvn jacoco:report
	// 		needs:
	// 			- checkstyle
	// docs:
	// 	generate-docs:
	// 		image: maven:3.8.6
	// 		script:
	// 			- ps -u
	// 			- mvn javadoc:javadoc`

	// 		cleanedText := strings.ReplaceAll(loggedOutput, "	", " ")
	// 		cleanedText = strings.ReplaceAll(cleanedText, "\n", " ")
	// 		cleanedText = strings.ReplaceAll(cleanedText, " ", "")
	// 		cleanedExpected := strings.ReplaceAll(expected, "\t", " ")
	// 		cleanedExpected = strings.ReplaceAll(cleanedExpected, "\n", " ")
	// 		cleanedExpected = strings.ReplaceAll(cleanedExpected, " ", "")

	//			if !strings.Contains(cleanedText, cleanedExpected) {
	//				t.Errorf("expected %q but got %q", cleanedExpected, cleanedText)
	//			}
	//	}
}

/*
Test `run` subcommand
*/
func TestRun(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)      // Redirect log output to buffer
	defer log.SetOutput(nil) // Reset after test

	// Store the original directory to restore later
	originalDir, _ := os.Getwd()
	// Restore original directory after test
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("Failed to return to original directory: %v\n", err)
		}
	}()

	// Change to a wrong directory (assume it's root for test purposes)
	err := os.Chdir("../../")
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd.RunCmd.PersistentFlags().Set("filename", "./.pipelines/test/dry_run_success.yaml")
	if err != nil {
		t.Errorf("unexpected error message: %v", err)
	}

	err = cmd.RunCmd.RunE(cmd.RootCmd, []string{})
	if err != nil {
		t.Errorf("unexpected error message: %v", err)
	}
}
