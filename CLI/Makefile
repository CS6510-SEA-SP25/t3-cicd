test:
	@echo TEST
	go test cicd/gocc/cmd
	@echo "--------------------------------------------"

coverage:
	@echo CODE COVERAGE
	./scripts/coverage.sh
	@echo "--------------------------------------------"

linter:
	@echo STATIC CODE ANALYSIS
	golangci-lint run
	@echo "--------------------------------------------"

build:
	@echo BUILD
	go build
	@echo "--------------------------------------------"

install:
	@echo INSTALL EXECUTABLE
	go install .
	@echo "--------------------------------------------"


all: test coverage linter build install