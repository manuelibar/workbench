.PHONY: help build run test test-race vet fmt smoke clean

GO  ?= go
BIN := build/_output/workbench-mcp

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | sort | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

build: ## Compile the workbench-mcp binary
	@mkdir -p $(dir $(BIN))
	$(GO) build -o $(BIN) ./cmd/workbench-mcp

run: ## Run workbench-mcp over stdio
	$(GO) run ./cmd/workbench-mcp

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
