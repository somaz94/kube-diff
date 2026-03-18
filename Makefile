.PHONY: build clean test test-unit test-integration cover cover-html bench lint fmt vet demo demo-clean help

BINARY_NAME=kube-diff
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/somaz94/kube-diff/cmd/cli.version=$(VERSION) -X github.com/somaz94/kube-diff/cmd/cli.commit=$(COMMIT) -X github.com/somaz94/kube-diff/cmd/cli.date=$(DATE)"

## Build

build: ## Build the binary
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/main.go

clean: ## Remove build artifacts and coverage files
	rm -f $(BINARY_NAME) coverage.out

## Test

test: test-unit ## Run unit tests (alias)

test-unit: ## Run unit tests with coverage
	go test ./... -v -race -cover

test-integration: ## Run integration tests (requires helm, kustomize)
	go test ./... -tags=integration -v -race

## Coverage

cover: ## Generate coverage report
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

cover-html: cover ## Open coverage report in browser
	go tool cover -html=coverage.out

## Benchmark

bench: ## Run benchmarks
	go test -bench=. -benchmem ./...

## Quality

lint: ## Run golangci-lint
	golangci-lint run ./...

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

## Demo

demo: ## Run demo (deploy resources, compare with kube-diff)
	@./scripts/demo.sh

demo-clean: ## Clean up demo resources from cluster
	@./scripts/demo-clean.sh

## Help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
