# Variables
APP_NAME := minfo
BUILD_DIR := .
GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_TAG    = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)

ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif


# Default target
all: build

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags "-s -w -X main.GitCommit=$(GIT_COMMIT) -X main.GitVersion=$(GIT_TAG)" -o $(BUILD_DIR)/$(APP_NAME)
	@echo "Build complete. Binary is located at $(BUILD_DIR)/$(APP_NAME)"

# Run the application
run: build
	@echo "Running $(APP_NAME)..."
	@./$(BUILD_DIR)/$(APP_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f $(BUILD_DIR)/$(APP_NAME)
	@echo "Clean complete."

# Tidy up Go dependencies (optional)
.PHONY: tidy
tidy:
	@echo "Tidying up Go dependencies..."
	@go mod tidy
