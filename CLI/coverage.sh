#!/bin/bash

set -e

# export DB_PASSWORD=16032002

# Clean test cache
# go clean -testcache

# run tests and create a coverprofile
# go test ./... -coverprofile=./cover.out

# open the interactive UI to check the Coverage Report
# go tool cover -html=./cover.out
go tool cover -html=./cover.out -o reports/test-coverage.html

# Set the minimum acceptable coverage percentage
MIN_COVERAGE=80

# Extract coverage percentage
coverage=$(go tool cover -func=./cover.out | tail -n 1 | awk '{print $3}' | tr -d '%')

# Compare with expected coverage
if (($(echo "$coverage < $MIN_COVERAGE" | bc -l))); then
    echo "Error: Test coverage is $coverage%, which is less than the required $MIN_COVERAGE%"
    exit 0
else
    echo "Test coverage is $coverage%, which meets the minimum requirement of $MIN_COVERAGE%"
    exit 0
fi
