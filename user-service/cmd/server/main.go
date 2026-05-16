package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	"food-delivery/user-service/internal/cache"
	"food-delivery/user-service/internal/handler"
	natspkg "food-delivery/user-service/internal/nats"
	"food-delivery/user-service/internal/repository"
	"food-delivery/user-service/internal/usecase"
	pb "food-delivery/user-service/proto/pb"
)

func main() {
	dsn := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/userdb?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "change-me-in-production")
	grpcAddr := getEnv("GRPC_ADDR", ":50051")
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")
	redisAddr := getEnv("REDIS_URL", "redis:6379")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}
	defer db.Close()

	var natsPublisher natspkg.Publisher
	if p, err := natspkg.NewPublisher(natsURL); err != nil {
		log.Printf("warn: nats unavailable, registration events disabled: %v", err)
	} else {
		natsPublisher = p
		defer natsPublisher.Close()
	}

	userCache := cache.NewUserCache(redisAddr)

	repo := repository.NewPostgresUserRepo(db)
	uc := usecase.NewUserUsecase(repo, jwtSecret, natsPublisher, userCache)
	h := handler.NewUserHandler(uc)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterUserServiceServer(srv, h)

	log.Printf("user-service gRPC listening on %s", grpcAddr)
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