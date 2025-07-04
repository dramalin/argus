# Argus System Monitor Makefile

BINARY_NAME=argus
BINARY_PATH=./bin/$(BINARY_NAME)
MAIN_PATH=./cmd/argus/main.go
GO_FILES=$(shell find . -name "*.go" -type f -not -path "./vendor/*")
FRONTEND_DIR=./web/argus-react
RELEASE_DIR=./release
TEST_TIMEOUT=30s
COVERAGE_OUT=coverage.out

.DEFAULT_GOAL := help

help:
	@echo "Argus System Monitor - Available targets:"
	@echo "  build          - Build both backend and frontend"
	@echo "  frontend-build - Build frontend and copy to release directory"
	@echo "  build-backend  - Build Go backend only"

build: frontend-build build-backend

frontend-build:
	@echo "Building frontend..."
	cd $(FRONTEND_DIR) && npm install && npm run build
	@echo "Copying frontend build to release directory..."
	mkdir -p $(RELEASE_DIR)
	rm -rf $(RELEASE_DIR)/*
	cp -r $(FRONTEND_DIR)/dist/* $(RELEASE_DIR)/

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BINARY_PATH) $(RELEASE_DIR)/*
	@echo "Clean complete"

.PHONY: help build frontend-build clean
