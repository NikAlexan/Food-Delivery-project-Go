COMPOSE = docker compose
GO      = docker run --rm -v $(PWD)/user-service:/app -w /app golang:1.26-alpine go
GO_DELIVERY = docker run --rm -v $(PWD)/delivery-service:/app -w /app golang:1.26-alpine go

.PHONY: up down restart logs \
        proto proto-user proto-delivery build-user build-delivery build-gateway \
        migrate-up migrate-down migrate-status \
        test-user test-delivery lint-user lint-delivery tidy

# ── Compose ──────────────────────────────────────────────────────────────────

up:
	$(COMPOSE) up --build -d

down:
	$(COMPOSE) down

restart:
	$(COMPOSE) restart

logs:
	$(COMPOSE) logs -f

logs-user:
	$(COMPOSE) logs -f user-service

logs-delivery:
	$(COMPOSE) logs -f delivery-service

logs-gateway:
	$(COMPOSE) logs -f api-gateway

# ── Proto ─────────────────────────────────────────────────────────────────────

proto: proto-user proto-delivery

proto-user:
	docker run --rm -v $(PWD)/user-service:/app -w /app golang:1.26-alpine \
		sh -c "apk add --no-cache protobuf > /dev/null && \
		       go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
		       go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
		       export PATH=\$$PATH:\$$(go env GOPATH)/bin && \
		       rm -rf proto/pb && mkdir -p proto/pb && \
		       protoc --proto_path=proto \
		              --go_out=proto/pb --go_opt=paths=source_relative \
		              --go-grpc_out=proto/pb --go-grpc_opt=paths=source_relative \
		              proto/user.proto && \
		       chown -R $$(id -u):$$(id -g) proto/pb"

proto-delivery:
	docker run --rm -v $(PWD)/delivery-service:/app -w /app golang:1.26-alpine \
		sh -c "apk add --no-cache protobuf > /dev/null && \
		       go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
		       go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
		       export PATH=\$$PATH:\$$(go env GOPATH)/bin && \
		       rm -rf proto/pb && mkdir -p proto/pb && \
		       protoc --proto_path=proto \
		              --go_out=proto/pb --go_opt=paths=source_relative \
		              --go-grpc_out=proto/pb --go-grpc_opt=paths=source_relative \
		              proto/delivery.proto && \
		       chown -R $$(id -u):$$(id -g) proto/pb"

# ── Build ─────────────────────────────────────────────────────────────────────

build-user:
	$(GO) build ./...

build-delivery:
	$(GO_DELIVERY) build ./...

build-gateway:
	docker run --rm \
		-v $(PWD)/api-gateway:/app/api-gateway \
		-v $(PWD)/delivery-service:/app/delivery-service \
		-v $(PWD)/user-service:/app/user-service \
		-w /app/api-gateway golang:1.26-alpine \
		go build ./...

# ── Migrations ────────────────────────────────────────────────────────────────

GOOSE = docker run --rm -v $(PWD)/user-service:/app -w /app \
        --network food-delivery-project-go_default \
        -e DATABASE_URL=postgres://postgres:postgres@postgres:5432/userdb?sslmode=disable \
        golang:1.26-alpine \
        go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres $$DATABASE_URL

migrate-up:
	$(GOOSE) up

migrate-down:
	$(GOOSE) down

migrate-status:
	$(GOOSE) status

# ── Dev helpers ───────────────────────────────────────────────────────────────

tidy:
	$(GO) mod tidy
	$(GO_DELIVERY) mod tidy
	docker run --rm \
		-v $(PWD)/api-gateway:/app/api-gateway \
		-v $(PWD)/delivery-service:/app/delivery-service \
		-v $(PWD)/user-service:/app/user-service \
		-w /app/api-gateway golang:1.26-alpine \
		go mod tidy

test-user:
	$(GO) test ./... -v -count=1

test-delivery:
	$(GO_DELIVERY) test ./... -v -count=1

lint-user:
	docker run --rm -v $(PWD)/user-service:/app -w /app \
		golangci/golangci-lint:latest golangci-lint run ./...

lint-delivery:
	docker run --rm -v $(PWD)/delivery-service:/app -w /app \
		golangci/golangci-lint:latest golangci-lint run ./...
