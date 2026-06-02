GO ?= go
BIN_DIR ?= bin

.PHONY: build build-all test test-race fmt fmt-check vet lint ci clean help

build: ## Build the dashboard binary.
	$(GO) build -o $(BIN_DIR)/curse ./cmd/dashboard

build-all: ## Build all command entrypoints.
	$(GO) build ./cmd/...

test: ## Run all tests.
	$(GO) test ./...

test-race: ## Run tests with race detector (supported platforms only).
	$(GO) test -race ./...

fmt: ## Format all Go files in-place.
	$(GO) fmt ./...

fmt-check: ## Check if Go files are formatted (fails if changes are needed).
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "These files need gofmt:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

vet: ## Run go vet checks.
	$(GO) vet ./...

lint: ## Run golangci-lint (requires golangci-lint in PATH).
	golangci-lint run ./...

ci: fmt-check vet build-all test ## Run the same core checks as CI.

clean: ## Remove build artifacts.
	rm -rf $(BIN_DIR)

help: ## Show available targets.
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z0-9_.-]+:.*## / {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)