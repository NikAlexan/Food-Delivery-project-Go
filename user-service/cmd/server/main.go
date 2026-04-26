package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	"food-delivery/user-service/internal/handler"
	"food-delivery/user-service/internal/repository"
	"food-delivery/user-service/internal/usecase"
	pb "food-delivery/user-service/proto/pb"
)

func main() {
	dsn := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/userdb?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "change-me-in-production")
	grpcAddr := getEnv("GRPC_ADDR", ":50051")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}
	defer db.Close()

	repo := repository.NewPostgresUserRepo(db)
	uc := usecase.NewUserUsecase(repo, jwtSecret)
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