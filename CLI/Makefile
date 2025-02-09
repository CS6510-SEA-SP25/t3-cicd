test:
	go test cicd/gocc/cmd

coverage:
	./scripts/coverage.sh

linter:
	golangci-lint run

build:
	go build

install:
	go install .


all: test coverage linter build install