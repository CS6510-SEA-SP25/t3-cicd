package schema_test

import (
	"cicd/pipeci/cmd"
	"os"
	"strings"
	"testing"
)

/*
Validate config file with wrong format
*/
func testWrongConfigFile(t *testing.T, filename string, expectedError string) {
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

	err = cmd.RootCmd.PersistentFlags().Set("filename", filename)
	if err != nil {
		t.Errorf("unexpected error message: %v", err)
	}

	err = cmd.RootCmd.PersistentFlags().Set("check", "true")
	if err != nil {
		t.Errorf("unexpected error message: %v", err)
	}

	err = cmd.Execute()
	if err == nil {
		t.Errorf("expected an error but got none")
	} else {
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("unexpected error message: %v", err)
		}
	}
}

/*
Different incorrect format scenarios.
*/
func TestInvalidConfigFile(t *testing.T) {
	// Cyclic deps
	testWrongConfigFile(t, "./.pipelines/test/cyclic_deps.yaml", "cyclic dependencies detected")
	// Empty version
	testWrongConfigFile(t, "./.pipelines/test/empty_version.yaml", "missing key `version`")
	// Empty pipeline
	testWrongConfigFile(t, "./.pipelines/test/empty_pipeline.yaml", "missing key `pipeline`")
	// Empty stages
	testWrongConfigFile(t, "./.pipelines/test/empty_stages.yaml", "missing key `stages`")
	// Empty stages
	testWrongConfigFile(t, "./.pipelines/test/stage_not_exist.yaml", "stage `not_exist` must be defined in stages")
}
