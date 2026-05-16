COMPOSE = docker compose -f compose.yml
GO      = docker run --rm -v $(PWD)/order-service:/app -w /app golang:1.26-alpine go

.PHONY: up down restart logs proto build migrate-up migrate-down test lint tidy

# ── Compose ───────────────────────────────────────────────────────────────────

up:
	$(COMPOSE) up --build -d

down:
	$(COMPOSE) down

restart:
	$(COMPOSE) restart

logs:
	$(COMPOSE) logs -f

logs-order:
	$(COMPOSE) logs -f order-service

logs-gateway:
	$(COMPOSE) logs -f order-gateway

# ── Proto ─────────────────────────────────────────────────────────────────────

proto:
	docker run --rm -v $(PWD)/order-service:/app -w /app golang:1.26-alpine \
		sh -c "apk add --no-cache protobuf > /dev/null && \
		       go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
		       go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
		       export PATH=\$$PATH:\$$(go env GOPATH)/bin && \
		       rm -rf proto/pb && mkdir -p proto/pb && \
		       protoc --proto_path=proto \
		              --go_out=proto/pb --go_opt=paths=source_relative \
		              --go-grpc_out=proto/pb --go-grpc_opt=paths=source_relative \
		              proto/order.proto && \
		       chown -R $$(id -u):$$(id -g) proto/pb"

# ── Build ─────────────────────────────────────────────────────────────────────

build:
	$(GO) build ./...

# ── Migrations ────────────────────────────────────────────────────────────────

GOOSE = docker run --rm -v $(PWD)/order-service:/app -w /app \
        --network food-delivery-project-go_default \
        -e DATABASE_URL=postgres://postgres:postgres@postgres-orders:5432/orderdb?sslmode=disable \
        golang:1.26-alpine \
        go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres $$DATABASE_URL

migrate-up:
	$(GOOSE) up

migrate-down:
	$(GOOSE) down

migrate-status:
	$(GOOSE) status

# ── Tests ─────────────────────────────────────────────────────────────────────

test:
	$(GO) test ./... -v -count=1

# ── Lint + Tidy ───────────────────────────────────────────────────────────────

lint:
	docker run --rm -v $(PWD)/order-service:/app -w /app \
		golangci/golangci-lint:latest golangci-lint run ./...

tidy:
	$(GO) mod tidy
