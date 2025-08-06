#!/usr/bin/make

## help: Print help.
.PHONY: help
help:
	@echo Possible commands:
	@cat Makefile | grep '##' | grep -v "Makefile" | sed -e 's/^##/  -/'

# Catch-all target to prevent Make from complaining about unknown targets.
%:
	@:


## install_dev: Install dependencies for development.
.PHONY: install_dev
install_dev:
	uv venv
	uv pip install pre-commit
	uv run pre-commit install

## dev: Run the application directly for development. Use -- before args.
.PHONY: dev
dev:
	go run ./cmd/todo $(filter-out $@,$(MAKECMDGOALS))


## all: Default target.
.PHONY: all
all: check build

## check: Run all checks.
.PHONY: check
check: fmt vet lint test


## build: Build the application.
.PHONY: build
build:
	go build -o bin/todo ./cmd/todo

## clean: Clean build artifacts.
.PHONY: clean
clean:
	rm -rf bin/
	rm -f cover.out coverage.html

## fmt: Format code.
.PHONY: fmt
fmt:
	gofumpt -w .

## install: Install the application to GOPATH/bin.
.PHONY: install
install:
	go install ./cmd/todo

## lint: Run linter.
.PHONY: lint
lint:
	golangci-lint run

## run: Run the application. Use -- before args.
.PHONY: run
run: build
	./bin/todo $(filter-out $@,$(MAKECMDGOALS))

## test: Run tests.
.PHONY: test
test:
	go test -v ./...

## test-coverage: Run tests with coverage.
.PHONY: test-coverage
test-coverage:
	go test -v -coverprofile=cover.out ./...
	go tool cover -html=cover.out -o coverage.html

## vet: Vet code.
.PHONY: vet
vet:
	go vet ./...


## build_docker: Build the Docker image.
.PHONY: build_docker
build_docker:
	docker build \
		--file Dockerfile \
		--tag togo \
		${PWD}

## run_pre_commit: Run pre-commit.
.PHONY: run_pre_commit
run_pre_commit: build_docker
	docker run --rm \
		--volume ${PWD}:/app \
		togo \
		-c "uv run pre-commit run --all-files"
