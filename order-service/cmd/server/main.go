package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	"food-delivery/order-service/internal/handler"
	appnats "food-delivery/order-service/internal/nats"
	"food-delivery/order-service/internal/repository"
	"food-delivery/order-service/internal/usecase"
	pb "food-delivery/order-service/proto/pb"
)

func main() {
	dsn := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/orderdb?sslmode=disable")
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")
	grpcAddr := getEnv("GRPC_ADDR", ":50053")

	// ── Database ──────────────────────────────────────────────────────────────
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}
	defer db.Close()

	// ── NATS ──────────────────────────────────────────────────────────────────
	publisher, err := appnats.NewPublisher(natsURL)
	if err != nil {
		log.Fatalf("nats: %v", err)
	}
	defer publisher.Close()

	// ── Dependency injection (Clean Architecture) ─────────────────────────────
	repo := repository.NewPostgresOrderRepo(db)
	uc := usecase.NewOrderUsecase(repo, publisher)
	h := handler.NewOrderHandler(uc)

	// ── gRPC server ───────────────────────────────────────────────────────────
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterOrderServiceServer(srv, h)

	log.Printf("order-service gRPC listening on %s", grpcAddr)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
