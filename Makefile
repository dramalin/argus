.DEFAULT_GOAL := help

BINARY_NAME=argus
MAIN_PATH=./cmd/argus/main.go
GO_FILES=$(shell find . -name "*.go" -type f -not -path "./vendor/*")
FRONTEND_DIR=./web/argus-react
RELEASE_DIR=./release

help:
	@echo "Argus System Monitor - Available targets:"
	@echo "  build          - Build both backend and frontend"
	@echo "  frontend-build - Build frontend and copy to release directory"
	@echo "  build-backend  - Build Go backend only"
	@echo "  clean         - Clean build artifacts"

build: frontend-build build-backend

frontend-build:
	@echo "Building frontend..."
	cd $(FRONTEND_DIR) && npm install && npm run build
	@echo "Copying frontend build to release directory..."
	mkdir -p $(RELEASE_DIR)/web
	rm -rf $(RELEASE_DIR)/web/*
	cp -r $(FRONTEND_DIR)/dist/* $(RELEASE_DIR)/web/
	@echo "Frontend build completed"

build-backend:
	@echo "Building Go backend..."
	mkdir -p $(RELEASE_DIR)/bin
	go build -o $(RELEASE_DIR)/bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Copying configuration and documentation..."
	cp config.example.yaml $(RELEASE_DIR)/config.yaml
	chmod +x $(RELEASE_DIR)/bin/$(BINARY_NAME)
	@echo "Creating startup script..."
	@echo '#!/bin/bash\ncd "$$(dirname "$$0")"\n./bin/argus "$$@"' > $(RELEASE_DIR)/start.sh
	@chmod +x $(RELEASE_DIR)/start.sh
	@echo "Backend build completed"

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(RELEASE_DIR)
	@echo "Clean complete"

.PHONY: help build frontend-build build-backend clean
