test:
	@echo TEST AND COVERAGE
	./script_coverage.sh
	@echo "--------------------------------------------"

linter:
	@echo STATIC CODE ANALYSIS
	golangci-lint -c .golangci.yaml run
	@echo "--------------------------------------------"

build:
	@echo BUILD
	go build
	@echo "--------------------------------------------"

install:
	@echo INSTALL EXECUTABLE
	go install .
	@echo "--------------------------------------------"


all: test linter build install