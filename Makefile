.PHONY: help build run test test-race vet fmt smoke clean

GO  ?= go
BIN := build/_output/workbench-mcp
STORAGE_BIN := build/_output/storage-service

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | sort | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

build: build-workbench build-storage ## Compile all binaries

build-workbench: ## Compile the workbench-mcp binary
	@mkdir -p $(dir $(BIN))
	$(GO) build -o $(BIN) ./cmd/workbench-mcp

build-storage: ## Compile the storage-service binary
	@mkdir -p $(dir $(STORAGE_BIN))
	$(GO) build -o $(STORAGE_BIN) ./cmd/storage-service

run: ## Run workbench-mcp over stdio
	$(GO) run ./cmd/workbench-mcp

run-storage: ## Run storage-service over HTTP
	$(GO) run ./cmd/storage-service

test: ## Run unit and integration tests
	$(GO) test ./...

test-race: ## Run tests with the race detector
	$(GO) test -race ./...

vet: ## Run go vet
	$(GO) vet ./...

fmt: ## Run gofmt
	gofmt -w .

smoke: ## Run the stdio MCP smoke test
	bash scripts/smoke.sh

clean: ## Remove build output
	rm -rf build/_output
