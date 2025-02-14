#!/bin/bash
set -e

for mod in $(find . -name "go.mod" -exec dirname {} \;); do
    echo "Running golangci-lint in $mod"
    (cd "$mod" && golangci-lint run --timeout=5m) || exit 1
done
