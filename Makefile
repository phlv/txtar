.PHONY: build clean test install run help fmt vet lint release

BINARY_NAME=txtar
VERSION?=dev
BUILD_DIR=build
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  install    - Install the binary to GOPATH/bin"
	@echo "  test       - Run tests"
	@echo "  clean      - Remove build artifacts"
	@echo "  run        - Run the application"
	@echo "  fmt        - Format code"
	@echo "  vet        - Run go vet"
	@echo "  lint       - Run golangci-lint (if installed)"
	@echo "  release    - Create and push a new tag (auto-increment)"
	@echo "  all        - Format, vet, test, and build"

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(LDFLAGS) .
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

test:
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	@echo "Coverage report saved to coverage.out"

coverage: test
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

run: build
	@$(BUILD_DIR)/$(BINARY_NAME)

fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

vet:
	@echo "Running go vet..."
	$(GO) vet ./...

lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Skipping..."; \
	fi

all: fmt vet test build
	@echo "All tasks complete"

release:
	@echo "Creating new release..."
	@LATEST_TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	echo "Latest tag: $$LATEST_TAG"; \
	NEW_TAG=$$(echo $$LATEST_TAG | awk -F. '{$$NF = $$NF + 1;} 1' | sed 's/ /./g'); \
	echo "New tag: $$NEW_TAG"; \
	git tag $$NEW_TAG && \
	git push origin $$NEW_TAG && \
	echo "Successfully created and pushed tag $$NEW_TAG"

.DEFAULT_GOAL := help
