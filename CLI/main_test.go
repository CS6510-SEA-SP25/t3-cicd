package main

import (
	"bytes"
	"log"
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	// Redirect log output to a buffer
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// Set up command-line arguments for testing
	os.Args = []string{"pipeci", "--help"}

	// Run the main function
	main()
}
