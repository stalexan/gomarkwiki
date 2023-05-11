.DEFAULT_GOAL = help

BINARY = gomarkwiki
BUILD_DIR = build
TMP_DIR = tmp
GO = go
VERSION = $(shell git describe --tags)

.PHONY: build
build:
	mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags "-X 'main.version=$(VERSION)'" -o $(BUILD_DIR)/$(BINARY) ./cmd/main.go

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(TMP_DIR)

.PHONY: test
test:
	$(GO) test -v ./internal/generator

.PHONY:fmt
fmt:
	$(GO) fmt ./...

.PHONY:vet
vet:
	$(GO) vet ./...

.PHONY:staticcheck
staticcheck:
	staticcheck ./...

.PHONY: help
help: ## Print this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@echo "  build          Build the binary"
	@echo "  clean          Clean the build directory"
	@echo "  fmt            Format the code"
	@echo "  staticcheck    Check the code using staticcheck"
	@echo "  test           Run the tests"
	@echo "  vet            Check the code using vet"
	@echo "  help           Print this help message"

