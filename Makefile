# go-dctrack-client Makefile
# Professional development automation following go-ldap-redhat patterns

.PHONY: help build test coverage clean fmt vet check install cli run-cli dev quick release-check
.DEFAULT_GOAL := help

# Configuration
GO_VERSION := 1.21
MODULE_NAME := github.com/Jethzabell/go-dctrack-client
CLI_NAME := dctrackcheck
COVERAGE_DIR := coverage
COVERAGE_FILE := $(COVERAGE_DIR)/coverage.out
COVERAGE_HTML := $(COVERAGE_DIR)/coverage.html

# Colors for output
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
RESET := \033[0m

# Build targets
build: ## Build the library
	@echo "$(BLUE)Building go-dctrack-client library...$(RESET)"
	@go build ./...
	@echo "$(GREEN)Build completed successfully$(RESET)"

cli: ## Build the CLI tool
	@echo "$(BLUE)Building dctrackcheck CLI tool...$(RESET)"
	@go build -o bin/$(CLI_NAME) ./cmd/$(CLI_NAME)
	@echo "$(GREEN)CLI tool built: bin/$(CLI_NAME)$(RESET)"

install: ## Install dependencies
	@echo "$(BLUE)Installing dependencies...$(RESET)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)Dependencies installed$(RESET)"

# Testing targets
test: ## Run all tests
	@echo "$(BLUE)Running all tests...$(RESET)"
	@go test -v ./...
	@echo "$(GREEN)All tests passed$(RESET)"

test-short: ## Run short tests (skip integration)
	@echo "$(BLUE)Running short tests...$(RESET)"
	@go test -short -v ./...

test-race: ## Run tests with race detection
	@echo "$(BLUE)Running tests with race detection...$(RESET)"
	@go test -race -v ./...

benchmark: ## Run benchmarks
	@echo "$(BLUE)Running benchmarks...$(RESET)"
	@go test -bench=. -benchmem ./...

coverage: ## Generate test coverage report
	@echo "$(BLUE)Generating coverage report...$(RESET)"
	@mkdir -p $(COVERAGE_DIR)
	@go test -coverprofile=$(COVERAGE_FILE) ./...
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@go tool cover -func=$(COVERAGE_FILE)
	@echo "$(GREEN)Coverage report generated: $(COVERAGE_HTML)$(RESET)"

# Code quality targets
fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(RESET)"
	@go fmt ./...
	@echo "$(GREEN)Code formatted$(RESET)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(RESET)"
	@go vet ./...
	@echo "$(GREEN)Vet checks passed$(RESET)"

check: ## Run all code quality checks
	@echo "$(BLUE)Running code quality checks...$(RESET)"
	@$(MAKE) fmt
	@$(MAKE) vet
	@echo "$(GREEN)All quality checks passed$(RESET)"

# Development targets
dev: ## Development build and test cycle
	@echo "$(BLUE)Running development cycle...$(RESET)"
	@$(MAKE) check
	@$(MAKE) test-short
	@$(MAKE) build
	@$(MAKE) cli
	@echo "$(GREEN)Development cycle completed$(RESET)"

quick: ## Quick validation (fmt, vet, build)
	@echo "$(BLUE)Running quick validation...$(RESET)"
	@$(MAKE) fmt
	@$(MAKE) vet
	@$(MAKE) build
	@echo "$(GREEN)Quick validation passed$(RESET)"

# CLI targets
run-cli: ## Run CLI tool with example (set DCTRACK_URL, DCTRACK_USER, DCTRACK_PASS, SEARCH_QUERY)
	@echo "$(BLUE)Running dctrackcheck CLI...$(RESET)"
	@if [ -z "$(DCTRACK_URL)" ]; then \
		echo "$(RED)Error: DCTRACK_URL environment variable is required$(RESET)"; \
		echo "$(YELLOW)Example: make run-cli DCTRACK_URL=https://dctrack.company.com/api/v2 DCTRACK_USER=username DCTRACK_PASS=password SEARCH_QUERY=PowerEdge$(RESET)"; \
		exit 1; \
	fi
	@if [ -z "$(DCTRACK_USER)" ]; then \
		echo "$(RED)Error: DCTRACK_USER environment variable is required$(RESET)"; \
		exit 1; \
	fi
	@if [ -z "$(DCTRACK_PASS)" ]; then \
		echo "$(RED)Error: DCTRACK_PASS environment variable is required$(RESET)"; \
		exit 1; \
	fi
	@export DCTRACK_URL=$(DCTRACK_URL) && \
	 export DCTRACK_USERNAME=$(DCTRACK_USER) && \
	 export DCTRACK_PASSWORD=$(DCTRACK_PASS) && \
	 ./bin/$(CLI_NAME) $(SEARCH_QUERY)

# Release targets
release-check: ## Check if ready for release
	@echo "$(BLUE)Checking release readiness...$(RESET)"
	@$(MAKE) check
	@$(MAKE) test
	@$(MAKE) coverage
	@echo "$(GREEN)Release checks passed - ready for release$(RESET)"

tag: ## Create a new version tag (requires VERSION)
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)Error: VERSION is required. Example: make tag VERSION=v1.0.1$(RESET)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Creating tag $(VERSION)...$(RESET)"
	@git tag $(VERSION)
	@git push origin $(VERSION)
	@echo "$(GREEN)Tag $(VERSION) created and pushed$(RESET)"

# Utility targets
clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(RESET)"
	@rm -rf bin/
	@rm -rf $(COVERAGE_DIR)/
	@go clean
	@echo "$(GREEN)Clean completed$(RESET)"

docs: ## Generate documentation
	@echo "$(BLUE)Generating documentation...$(RESET)"
	@go doc -all . > docs/API.md
	@echo "$(GREEN)Documentation generated$(RESET)"

example-basic: ## Run basic usage example
	@echo "$(BLUE)Running basic usage example...$(RESET)"
	@cd examples/basic_usage && go run main.go

example-advanced: ## Run advanced filtering example
	@echo "$(BLUE)Running advanced filtering example...$(RESET)"
	@cd examples/advanced_filtering && go run main.go

# Docker targets (if needed)
docker-build: ## Build Docker image for testing
	@echo "$(BLUE)Building Docker image...$(RESET)"
	@docker build -t go-dctrack-client:latest .

docker-test: ## Run tests in Docker
	@echo "$(BLUE)Running tests in Docker...$(RESET)"
	@docker run --rm go-dctrack-client:latest make test

# Security and validation
security-check: ## Run security checks
	@echo "$(BLUE)Running security checks...$(RESET)"
	@go mod verify
	@echo "$(GREEN)Security checks passed$(RESET)"

mod-update: ## Update Go module dependencies
	@echo "$(BLUE)Updating Go module dependencies...$(RESET)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)Dependencies updated$(RESET)"

# Information targets
version: ## Show version information
	@echo "Module: $(MODULE_NAME)"
	@echo "Version: $(shell cat VERSION 2>/dev/null || echo 'development')"
	@echo "Go Version: $(shell go version)"

env: ## Show environment information
	@echo "GOPATH: $(GOPATH)"
	@echo "GOROOT: $(GOROOT)"
	@echo "GO VERSION: $(shell go version)"
	@echo "PWD: $(PWD)"

help: ## Show this help message
	@echo "$(BLUE)go-dctrack-client Development Commands$(RESET)"
	@echo ""
	@echo "$(YELLOW)Build Commands:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E '(build|cli|install)' | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)Test Commands:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E '(test|coverage|benchmark)' | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)Code Quality:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E '(fmt|vet|check|security)' | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)Development:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E '(dev|quick|example|run-cli)' | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)Release:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E '(release|tag|clean)' | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)Utility:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E '(docs|version|env|help)' | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BLUE)Example Usage:$(RESET)"
	@echo "  make dev                    # Full development cycle"
	@echo "  make run-cli DCTRACK_URL=https://dctrack.company.com/api/v2 DCTRACK_USER=username DCTRACK_PASS=password SEARCH_QUERY=PowerEdge"
	@echo "  make coverage               # Generate coverage report"
	@echo "  make release-check          # Verify ready for release"
