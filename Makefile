GO_VERSION=1.26
LINTER_VERSION=v1.64.8

DIST_FOLDER=dist
BINARY_NAME=whatsapp-jpeg-repair

ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    # Принудительно используем cmd для Windows-блоков
    SHELL := cmd.exe
else
    DETECTED_OS := $(shell uname -s)
endif

ifeq ($(DETECTED_OS),Windows)
    BINARY_EXT=.exe
    # Убираем слэш в конце переменной, чтобы не было конфликтов с переносом строк
    COPY_SOURCE_FILES_DIR = xcopy /E /I /Y "whatsapp-files" "$(DIST_FOLDER)\whatsapp-files"
    MKDIR_REPAIRED = if not exist "$(DIST_FOLDER)\repaired-files" mkdir "$(DIST_FOLDER)\repaired-files"
    # Важно: кавычки защищают от проблем со слэшами
    COPY_SHELL_FILES = copy /Y "platform\windows\runme.bat" "$(DIST_FOLDER)"
    COPY_LICENSE_FILE = copy /Y "LICENSE.txt" "$(DIST_FOLDER)"
endif

ifeq ($(DETECTED_OS),Darwin)
    COPY_SOURCE_FILES_DIR = cp -r "whatsapp-files" "$(DIST_FOLDER)/whatsapp-files"
    MKDIR_REPAIRED = mkdir -p $(DIST_FOLDER)/repaired-files 
    COPY_SHELL_FILES = cp platform/mac/*.* $(DIST_FOLDER)/
    COPY_LICENSE_FILE = cp LICENSE.txt $(DIST_FOLDER)/
    BINARY_EXT=     
endif

ifeq ($(DETECTED_OS),Linux)
    COPY_SOURCE_FILES_DIR = cp -r "whatsapp-files" "$(DIST_FOLDER)/whatsapp-files"
    MKDIR_REPAIRED = mkdir -p $(DIST_FOLDER)/repaired-files 
    COPY_SHELL_FILES = cp platform/linux/*.* $(DIST_FOLDER)/
    COPY_LICENSE_FILE = cp LICENSE.txt $(DIST_FOLDER)/
    BINARY_EXT=     
endif

FULL_BINARY_NAME=$(BINARY_NAME)$(BINARY_EXT)

.PHONY: all help update lint test build clean

all: help

help:
	@echo Usage: make [target]
	@echo.
	@echo Targets:
# На Windows стандартные sed/column могут отсутствовать, поэтому делаем проверку
ifeq ($(DETECTED_OS),Windows)
	@echo  update, lint, test, build, clean
else
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
endif

## git-hooks: Install git hooks
git-hooks:
ifeq ($(DETECTED_OS),Windows)
	@for %%f in (.githooks\*) do copy /y "%%f" ".git\hooks\"
else
	@cp .githooks/* .git/hooks/
	@chmod +x .git/hooks/*
endif

## update: Update project to actual go version and tidy modules
update:
	go mod edit -go=$(GO_VERSION)
	go mod tidy

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## test: Run all tests
test:
	go test -v -race -cover ./...

## build: Compile binary
build: clean
	@echo Building for $(DETECTED_OS)...
	go build -trimpath -ldflags="-s -w" -o $(DIST_FOLDER)/$(FULL_BINARY_NAME) ./cmd/whatsapp-jpeg-repair
	$(COPY_SOURCE_FILES_DIR)
	$(MKDIR_REPAIRED)
	$(COPY_SHELL_FILES)
	$(COPY_LICENSE_FILE)

## clean: Remove build artifacts
clean:
ifeq ($(DETECTED_OS),Windows)
	@if exist $(DIST_FOLDER) rmdir /s /q $(DIST_FOLDER)
else
	@rm -rf $(DIST_FOLDER)
endif