.PHONY: help build run test test-integration vet fmt compose-up compose-down compose-logs smoke clean

GO  ?= go
BIN := build/_output/workbench-mcp

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | sort | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

build: ## Compile the workbench-mcp binary
	@mkdir -p $(dir $(BIN))
	$(GO) build -o $(BIN) ./cmd/workbench-mcp

run: ## Run workbench-mcp directly
	$(GO) run ./cmd/workbench-mcp

test: ## Run unit tests (skip integration with -short)
	$(GO) test -short -race ./...

test-integration: ## Run all tests (requires `make compose-up`)
	$(GO) test -race ./...

vet: ## Run go vet
	$(GO) vet ./...

fmt: ## Run gofmt -s -w
	gofmt -s -w .

compose-up: ## Start Postgres + pgvector via docker compose (waits for healthy)
	docker compose up -d --wait

compose-down: ## Stop and remove docker compose stack (including volumes)
	docker compose down -v

compose-logs: ## Tail docker compose logs
	docker compose logs -f

smoke: ## Boot the binary and exercise /healthz, /readyz, and MCP initialize
	bash scripts/smoke.sh

clean: ## Remove build output
	rm -rf build/_output
