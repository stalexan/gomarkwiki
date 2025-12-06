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
	$(GO) test -v ./...

.PHONY:fmt
fmt:
	$(GO) fmt ./...

.PHONY:vet
vet:
	$(GO) vet ./...

.PHONY:staticcheck
staticcheck:
	staticcheck ./...

.PHONY: check-updates
check-updates: ## Check for dependency updates
	@echo "Checking for dependency updates..."
	@$(GO) list -u -m all 2>/dev/null | grep -E '\[.*\]' || echo "All dependencies are up to date."

.PHONY: update-deps
update-deps: ## Update all dependencies to their latest versions
	@echo "Updating dependencies..."
	$(GO) get -u ./...
	$(GO) mod tidy
	@echo "Dependencies updated. Run 'make test' to verify everything still works."

.PHONY: update-deps-patch
update-deps-patch: ## Update dependencies to latest patch versions only
	@echo "Updating dependencies (patch versions only)..."
	$(GO) get -u=patch ./...
	$(GO) mod tidy
	@echo "Dependencies updated. Run 'make test' to verify everything still works."

.PHONY: help
help: ## Print this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@echo "  build             Build the binary"
	@echo "  clean             Clean the build directory"
	@echo "  fmt               Format the code"
	@echo "  staticcheck       Check the code using staticcheck"
	@echo "  test              Run the tests"
	@echo "  vet               Check the code using vet"
	@echo "  check-updates     Check for dependency updates"
	@echo "  update-deps       Update all dependencies to latest versions"
	@echo "  update-deps-patch Update dependencies to latest patch versions only"
	@echo "  help              Print this help message"

