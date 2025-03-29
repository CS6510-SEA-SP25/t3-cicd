package db

import (
	"os"
	"testing"
)

// TestInit tests the normal initialization of the database connection
func TestInit(t *testing.T) {
	// Setup environment variables
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "root")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_NAME", "CicdApplication")
	os.Setenv("DB_PASSWORD", "root")
	os.Setenv("DB_SSL_MODE", "false")
	defer func() {
		// Cleanup
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_SSL_MODE")
	}()

	Init()

	if Instance == nil {
		t.Error("Expected database instance to be initialized, got nil")
	}

	err := Instance.Ping()
	if err != nil {
		t.Errorf("Expected successful ping, got error: %v", err)
	}
}
