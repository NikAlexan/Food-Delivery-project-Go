FROM golang:1.26-alpine AS builder
WORKDIR /workspace
COPY user-service/go.mod user-service/go.sum ./user-service/
COPY restaurant-service/go.mod restaurant-service/go.sum ./restaurant-service/
COPY order-service/go.mod order-service/go.sum ./order-service/
COPY delivery-service/go.mod delivery-service/go.sum ./delivery-service/
COPY api-gateway/go.mod api-gateway/go.sum ./api-gateway/
WORKDIR /workspace/api-gateway
RUN go mod download
WORKDIR /workspace
COPY user-service ./user-service
COPY restaurant-service ./restaurant-service
COPY order-service ./order-service
COPY delivery-service ./delivery-service
COPY api-gateway ./api-gateway
WORKDIR /workspace/api-gateway
RUN go build -o /gateway ./cmd

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /gateway ./gateway
ENTRYPOINT ["./gateway"]
