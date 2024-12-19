# Variables
APP_NAME := minfo
SRC_DIR := ./src
DOC_DIR := ./doc
BUILD_DIR := ..
GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_TAG    = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
RONN := ronn

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
	@cd $(SRC_DIR) && go build -ldflags "-s -w -X main.GitCommit=$(GIT_COMMIT) -X main.GitVersion=$(GIT_TAG)" -o $(BUILD_DIR)/$(APP_NAME)
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
tidy:
	@echo "Tidying up Go dependencies..."
	@go mod tidy

doc:
	@command -v $(TOOL) >/dev/null 2>&1 || { \
		echo >&2 "Error: $(TOOL) is not installed."; \
		exit 1; \
	}
	@echo "Generating documentation..."
	@cd $(DOC_DIR) && ronn --roff $(APP_NAME).1.ronn

.PHONY: all build run clean tidy doc
