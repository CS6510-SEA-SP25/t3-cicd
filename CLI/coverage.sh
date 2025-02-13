#!/bin/bash

# Clean test cache
go clean -testcache
# run tests and create a coverprofile
go test ./... -coverprofile=./cover.out
# open the interactive UI to check the Coverage Repor
go tool cover -html=./cover.out
