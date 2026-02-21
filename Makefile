GO_VERSION=1.26
LINTER_VERSION=v1.64.8

APP_VERSION=3.0.0
BIN_FOLDER=WhatsAppJpegRepair_$(APP_VERSION)
BINARY_NAME=WhatsAppJpegRepair

ifeq ($(OS),Windows_NT)
    BINARY_EXT=.exe
	COPY_SOURCE_FILES_DIR = xcopy /E /I /Y "whatsapp-files" "$(BIN_FOLDER)\whatsapp-files"
	MKDIR_REPAIRED = mkdir $(BIN_FOLDER)\repaired-files
	COPY_SHELL_FILES = copy /Y platform\win\runme.bat $(BIN_FOLDER)\	
	COPY_LICENSE_FILE = copy /Y LICENSE.txt $(BIN_FOLDER)\

else
	COPY_SOURCE_FILES_DIR = cp -r "whatsapp-files" "$(BIN_FOLDER)/whatsapp-files"
	MKDIR_REPAIRED = mkdir -p $(BIN_FOLDER)/repaired-files	
	COPY_SHELL_FILES = cp platform/nix/*.* $(BIN_FOLDER)/
	COPY_LICENSE_FILE = cp LICENSE.txt $(BIN_FOLDER)/
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


## git-hooks: Install git hooks
git-hooks:
	@echo "Installing git hooks..."
ifeq ($(OS),Windows_NT)
	@for %%f in (.githooks\*) do copy /y "%%f" ".git\hooks\"
else
	@cp .githooks/* .git/hooks/
	@chmod +x .git/hooks/*
endif
	@echo "Git hooks installed."

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
	go build -trimpath -ldflags="-s -w" -o $(BIN_FOLDER)/$(FULL_BINARY_NAME) ./cmd/whatsapp-jpeg-repair
	
	@echo "Copying assets..."
	$(COPY_SOURCE_FILES_DIR)
	$(MKDIR_REPAIRED)
	$(COPY_SHELL_FILES)
	$(COPY_LICENSE_FILE)
	@echo "Success!"

## clean: Remove build artifacts
clean:
	@echo "Cleaning up..."
ifeq ($(OS),Windows_NT)
	@if exist bin rmdir /s /q $(BIN_FOLDER)
else
	@rm -rf $(BIN_FOLDER)
endif
	@echo "Cleaned."