#!/bin/bash

git add .
git commit -m "test release attempt 5"

git tag v0.0.0
git push origin v0.0.0

goreleaser release --clean