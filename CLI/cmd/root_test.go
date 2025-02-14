package cmd_test

import (
	"cicd/pipeci/cmd"
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
