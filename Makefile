# Project variables
GO_VERSION=1.26
LINTER_VERSION=v1.64.8
BINARY_NAME=WhatsAppJpegRepair

ifeq ($(OS),Windows_NT)
    BINARY_EXT=.exe
else
    BINARY_EXT=
endif
FULL_BINARY_NAME=$(BINARY_NAME)$(BINARY_EXT)

.PHONY: all help update lint test build clean

# Default target: shows help
all: help

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## update: Update project to actual go version and tidy modules
update:
	@echo "Updating project to Go $(GO_VERSION)..."
	go mod edit -go=$(GO_VERSION)
	go mod tidy
	@echo "Update complete."

## lint: Run golangci-lint with the local configuration
lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...

## test: Run all tests with race detector and coverage
test:
	@echo "Running tests..."
	go test -v -race -cover ./...

## build: Compile the binary for the current OS/Arch
build: clean
	@echo "Building binary..."
	go build -trimpath -ldflags="-s -w" -o bin/$(FULL_BINARY_NAME) ./cmd/whatsapp-jpeg-repair

## clean: Remove build artifacts
clean:
	@echo "Cleaning up..."
ifeq ($(OS),Windows_NT)
	@if exist bin rmdir /s /q bin
else
	@rm -rf bin
endif
	@echo "Cleaned."