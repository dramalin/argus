# Argus System Monitor Makefile
# Go project build automation

# Variables
BINARY_NAME=argus
BINARY_PATH=./bin/$(BINARY_NAME)
MAIN_PATH=./cmd/argus/main.go
GO_FILES=$(shell find . -name "*.go" -type f -not -path "./vendor/*")
WEBAPP_DIR=./webapp
TEST_TIMEOUT=30s
COVERAGE_OUT=coverage.out

# Go build flags
LDFLAGS=-ldflags "-X main.version=$(shell git describe --tags --always --dirty) -X main.buildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"
BUILD_FLAGS=-trimpath $(LDFLAGS)

# Default target
.DEFAULT_GOAL := help

## help: Show this help message
.PHONY: help
help:
	@echo "Argus System Monitor - Available targets:"
	@echo ""
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
	@echo ""

## build: Build both backend and frontend
.PHONY: build
build: web-build build-backend

## build-backend: Build Go backend only
.PHONY: build-backend
build-backend:
	@echo "Building Go backend..."
	@mkdir -p bin
	go build $(BUILD_FLAGS) -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Binary built: $(BINARY_PATH)"

## dev: Run both backend and frontend dev servers
.PHONY: dev
dev:
	@echo "Starting backend (auto-reload if air is available)..."
	$(MAKE) -j2 dev-backend web-dev

## dev-backend: Run backend dev server only
.PHONY: dev-backend
dev-backend:
	@if command -v air > /dev/null; then \
		echo "Starting backend with air..."; \
		hair; \
	else \
		echo "Air not found. Falling back to go run..."; \
		go run $(MAIN_PATH); \
	fi

## run: Run the application in development mode
.PHONY: run
run:
	@echo "Starting Argus System Monitor..."
	go run $(MAIN_PATH)

## test: Run all tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v -timeout $(TEST_TIMEOUT) ./...

## test-race: Run tests with race detector
.PHONY: test-race
test-race:
	@echo "Running tests with race detector..."
	go test -race -timeout $(TEST_TIMEOUT) ./...

## test-cover: Run tests with coverage
.PHONY: test-cover
test-cover:
	@echo "Running tests with coverage..."
	go test -coverprofile=$(COVERAGE_OUT) -timeout $(TEST_TIMEOUT) ./...
	go tool cover -html=$(COVERAGE_OUT) -o coverage.html
	@echo "Coverage report generated: coverage.html"

## bench: Run benchmarks
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

## lint: Run linters
.PHONY: lint
lint:
	@if command -v golangci-lint > /dev/null; then \
		echo "Running golangci-lint..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with:"; \
		echo "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## fmt: Format Go code
.PHONY: fmt
fmt:
	@echo "Formatting Go code..."
	gofmt -s -w $(GO_FILES)
	goimports -w $(GO_FILES)

## vet: Run go vet
.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

## mod-tidy: Tidy up module dependencies
.PHONY: mod-tidy
mod-tidy:
	@echo "Tidying module dependencies..."
	go mod tidy
	go mod verify

## mod-download: Download module dependencies
.PHONY: mod-download
mod-download:
	@echo "Downloading module dependencies..."
	go mod download

## web-build: Build the React frontend for production
.PHONY: web-build
web-build:
	@echo "Building React frontend..."
	cd web/argus-react && npm ci && npm run build

## web-dev: Start React frontend dev server (Vite)
.PHONY: web-dev
web-dev:
	@echo "Starting React frontend dev server..."
	cd web/argus-react && npm ci && npm run dev

## web-lint: Lint React frontend code
.PHONY: web-lint
web-lint:
	@echo "Linting React frontend..."
	cd web/argus-react && npm run lint

## web-clean: Clean React frontend build artifacts
.PHONY: web-clean
web-clean:
	@echo "Cleaning React frontend build artifacts..."
	cd web/argus-react && rm -rf dist/

## web-deps: Install React frontend dependencies
.PHONY: web-deps
web-deps:
	@echo "Installing React frontend dependencies..."
	cd web/argus-react && npm ci

## clean: Clean all build artifacts (backend and frontend)
.PHONY: clean
clean: web-clean
	@echo "Cleaning backend build artifacts..."
	@rm -rf bin/
	@rm -f $(COVERAGE_OUT) coverage.html
	@rm -f *.log
	@echo "Clean completed"

## install: Install the binary to GOPATH/bin
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(BUILD_FLAGS) $(MAIN_PATH)
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

## docker-build: Build Docker image
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t argus:latest .

## docker-run: Run Docker container
.PHONY: docker-run
docker-run: docker-build
	@echo "Running Docker container..."
	docker run -p 8080:8080 --rm argus:latest

## api-test: Test API endpoints
.PHONY: api-test
api-test:
	@if [ -f test_api.sh ]; then \
		echo "Testing API endpoints..."; \
		chmod +x test_api.sh; \
		./test_api.sh; \
	else \
		echo "test_api.sh not found"; \
	fi

## deps-check: Check for outdated dependencies
.PHONY: deps-check
deps-check:
	@echo "Checking for outdated dependencies..."
	go list -u -m all

## deps-update: Update all dependencies
.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

## security: Run security checks
.PHONY: security
security:
	@if command -v gosec > /dev/null; then \
		echo "Running security checks..."; \
		gosec ./...; \
	else \
		echo "gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

## release-build: Build release binaries for multiple platforms
.PHONY: release-build
release-build: clean
	@echo "Building release binaries..."
	@mkdir -p bin/releases
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/releases/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o bin/releases/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/releases/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o bin/releases/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/releases/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Release binaries built in bin/releases/"

## all: Run quality checks and build
.PHONY: all
all: fmt vet lint test build
	@echo "All checks passed and binary built successfully"

## ci: Continuous integration checks
.PHONY: ci
ci: mod-tidy fmt vet lint test-race test-cover
	@echo "CI checks completed"

## setup-dev: Setup development environment
.PHONY: setup-dev
setup-dev:
	@echo "Setting up development environment..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "Development tools installed"

## start: Quick start (build and run)
.PHONY: start
start: build
	@echo "Starting $(BINARY_NAME)..."
	$(BINARY_PATH)

## watch: Watch for changes and restart (requires air)
.PHONY: watch
watch: dev

# Include git hooks setup
## git-hooks: Setup git hooks
.PHONY: git-hooks
git-hooks:
	@echo "Setting up git hooks..."
	@mkdir -p .git/hooks
	@echo '#!/bin/sh\nmake fmt vet' > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Git hooks installed"

## deps: Install all dependencies (Go and frontend)
.PHONY: deps
deps: mod-download web-deps

## docker-up: Start all services with docker-compose
.PHONY: docker-up
docker-up:
	docker-compose up -d

## docker-down: Stop all services with docker-compose
.PHONY: docker-down
docker-down:
	docker-compose down

## docker-logs: Show logs from all docker-compose services
.PHONY: docker-logs
docker-logs:
	docker-compose logs -f 