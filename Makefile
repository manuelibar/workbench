.PHONY: build run test vet fmt dev-up dev-down dev-logs dev-ps db-url

build:
	go build -buildvcs=false -o build/workbench-mcp ./cmd/workbench-mcp

run:
	go run -buildvcs=false ./cmd/workbench-mcp

test:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

dev-up:
	docker compose up -d postgres nats

dev-down:
	docker compose down

dev-logs:
	docker compose logs -f postgres nats

dev-ps:
	docker compose ps

db-url:
	@printf '%s\n' "$${WORKBENCH_DATABASE_URL:-postgres://workbench:workbench@localhost:5432/workbench?sslmode=disable}"
