package cmd_test

import (
	"cicd/gocc/cmd"
	"os"
	"strings"
	"testing"
)

/*
Test running the CLI in Git-initialized directory
*/
func TestGitRepository(t *testing.T) {

	// Store the original directory to restore later
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get original directory: %v", err)
	}
	defer os.Chdir(originalDir) // Restore original directory after test

	// Change to a wrong directory (assume it's root for test purposes)
	err = os.Chdir("/")
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
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get original directory: %v", err)
	}
	defer os.Chdir(originalDir) // Restore original directory after test

	// Change to a wrong directory (assume it's root for test purposes)
	err = os.Chdir("../../")
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// * TEST CASES //
	// Default config file
	err = cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error message: %v", err)
	}

	cmd.RootCmd.Flags().Set("filename", "")
	err = cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error message: %v", err)
	}

	// Config file not exists
	cmd.RootCmd.Flags().Set("filename", "./not_exist.yaml")
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
	cmd.RootCmd.Flags().Set("filename", "./README.md")
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
	cmd.RootCmd.Flags().Set("filename", ".pipelines/pipeline.yaml")
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
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get original directory: %v", err)
	}
	defer os.Chdir(originalDir) // Restore original directory after test

	// Change to a wrong directory (assume it's root for test purposes)
	err = os.Chdir("../../")
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd.RootCmd.Flags().Set("check", "true")
	err = cmd.Execute()
	if err != nil {
		t.Errorf("unexpected error message: %v", err)
	}
}

/*
Validate config file
*/
func TestCyclicDeps(t *testing.T) {
	// Store the original directory to restore later
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get original directory: %v", err)
	}
	defer os.Chdir(originalDir) // Restore original directory after test

	// Change to a wrong directory (assume it's root for test purposes)
	err = os.Chdir("../../")
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cmd.RootCmd.Flags().Set("filename", "./.pipelines/test/cyclic_deps.yaml")
	cmd.RootCmd.Flags().Set("check", "true")
	err = cmd.Execute()
	if err == nil {
		t.Errorf("expected an error but got none")
	} else {
		expectedError := "cyclic dependencies detected"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("unexpected error message: %v", err)
		}
	}
}
