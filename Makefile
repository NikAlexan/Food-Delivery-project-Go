COMPOSE = docker compose
GO      = docker run --rm -v "$(PWD)/user-service:/app" -w /app golang:1.26-alpine go
GO_RS   = docker run --rm -v "$(PWD)/restaurant-service:/app" -w /app golang:1.26-alpine go
GO_DELIVERY = docker run --rm -v "$(PWD)/delivery-service:/app" -w /app golang:1.26-alpine go
GO_ORDER    = docker run --rm -v "$(PWD)/order-service:/app" -w /app golang:1.26-alpine go

.PHONY: up down restart logs \
        proto proto-user proto-delivery proto-restaurant build-user build-delivery build-gateway build-restaurant \
        migrate-up migrate-down migrate-status \
        migrate-restaurant-up migrate-restaurant-down migrate-restaurant-status \
        test-user test-user-integration test-delivery test-restaurant test-gateway \
        test-order test-order-integration \
        lint-user lint-delivery lint-restaurant tidy

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

proto-restaurant:
	docker run --rm -v $(PWD)/restaurant-service:/app -w /app golang:1.26-alpine \
		sh -c "apk add --no-cache protobuf > /dev/null && \
		       go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
		       go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
		       export PATH=\$$PATH:\$$(go env GOPATH)/bin && \
		       rm -rf proto/pb && mkdir -p proto/pb && \
		       protoc --proto_path=proto \
		              --go_out=proto/pb --go_opt=paths=source_relative \
		              --go-grpc_out=proto/pb --go-grpc_opt=paths=source_relative \
		              proto/restaurant.proto && \
		       chown -R $$(id -u):$$(id -g) proto/pb"

proto-order:
	docker run --rm -v "$(PWD)/order-service:/app" -w /app golang:1.26-alpine \
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

build-restaurant:
	$(GO_RS) build ./...

build-delivery:
	$(GO_DELIVERY) build ./...

build-gateway:
	docker run --rm \
		-v "$(PWD)/api-gateway:/app/api-gateway" \
		-v "$(PWD)/delivery-service:/app/delivery-service" \
		-v "$(PWD)/user-service:/app/user-service" \
		-v "$(PWD)/restaurant-service:/app/restaurant-service" \
		-v "$(PWD)/order-service:/app/order-service" \
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

GOOSE_RS = docker run --rm -v $(PWD)/restaurant-service:/app -w /app \
           --network food-delivery-project-go_default \
           -e DATABASE_URL=postgres://postgres:postgres@postgres:5432/restaurantdb?sslmode=disable \
           golang:1.26-alpine \
           go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres $$DATABASE_URL

migrate-restaurant-up:
	$(GOOSE_RS) up

migrate-restaurant-down:
	$(GOOSE_RS) down

migrate-restaurant-status:
	$(GOOSE_RS) status

# ── Dev helpers ───────────────────────────────────────────────────────────────

tidy:
	$(GO) mod tidy
	$(GO_RS) mod tidy
	$(GO_DELIVERY) mod tidy
	docker run --rm \
		-v $(PWD)/api-gateway:/app/api-gateway \
		-v $(PWD)/delivery-service:/app/delivery-service \
		-v $(PWD)/user-service:/app/user-service \
		-v $(PWD)/restaurant-service:/app/restaurant-service \
		-w /app/api-gateway golang:1.26-alpine \
		go mod tidy

test-user:
	$(GO) test ./internal/usecase/... -v -count=1

test-user-integration:
	docker run --rm \
		-v "$(PWD)/user-service:/app" \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-w /app golang:1.26-alpine \
		go test ./internal/repository/... -v -count=1 -timeout 120s

test-restaurant:
	$(GO_RS) test ./... -v -count=1

test-delivery:
	$(GO_DELIVERY) test ./... -v -count=1

test-gateway:
	docker run --rm \
		-v "$(PWD)/api-gateway:/app/api-gateway" \
		-v "$(PWD)/delivery-service:/app/delivery-service" \
		-v "$(PWD)/user-service:/app/user-service" \
		-v "$(PWD)/restaurant-service:/app/restaurant-service" \
		-v "$(PWD)/order-service:/app/order-service" \
		-w /app/api-gateway golang:1.26-alpine \
		go test ./... -v -count=1

test-order:
	$(GO_ORDER) test ./internal/usecase/... -v -count=1

test-order-integration:
	docker run --rm \
		-v "$(PWD)/order-service:/app" \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-w /app golang:1.26-alpine \
		go test ./internal/repository/... -v -count=1 -timeout 120s

lint-user:
	docker run --rm -v $(PWD)/user-service:/app -w /app \
		golangci/golangci-lint:latest golangci-lint run ./...

lint-restaurant:
	docker run --rm -v $(PWD)/restaurant-service:/app -w /app \
		golangci/golangci-lint:latest golangci-lint run ./...

lint-delivery:
	docker run --rm -v $(PWD)/delivery-service:/app -w /app \
		golangci/golangci-lint:latest golangci-lint run ./...
