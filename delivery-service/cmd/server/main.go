package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	"food-delivery/delivery-service/internal/cache"
	"food-delivery/delivery-service/internal/email"
	"food-delivery/delivery-service/internal/handler"
	natsclient "food-delivery/delivery-service/internal/nats"
	"food-delivery/delivery-service/internal/repository"
	"food-delivery/delivery-service/internal/usecase"
	pb "food-delivery/delivery-service/proto/pb"
)

func main() {
	dsn := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/userdb?sslmode=disable")
	grpcAddr := getEnv("GRPC_ADDR", ":50054")
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	smtpHost := getEnv("SMTP_HOST", "localhost")
	smtpPort := getEnvInt("SMTP_PORT", 587)
	smtpUser := getEnv("SMTP_USER", "")
	smtpPass := getEnv("SMTP_PASS", "")
	emailFrom := getEnv("EMAIL_FROM", "no-reply@example.com")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}
	defer db.Close()

	repo := repository.NewPostgresDeliveryRepo(db)
	locationCache, err := cache.NewRedisLocationCache(redisAddr, 5*time.Minute)
	if err != nil {
		log.Fatalf("redis connect: %v", err)
	}
	defer locationCache.Close()

	mailer, err := email.NewSMTPClient(smtpHost, smtpUser, smtpPass, emailFrom, smtpPort)
	if err != nil {
		log.Fatalf("smtp config: %v", err)
	}

	natsConn, err := natsclient.NewClient(natsURL)
	if err != nil {
		log.Fatalf("nats connect: %v", err)
	}
	defer natsConn.Close()

	uc := usecase.NewDeliveryUsecase(repo, mailer, natsConn, locationCache)
	if err := natsConn.StartOrderConsumers(uc); err != nil {
		log.Fatalf("nats subscribe: %v", err)
	}

	h := handler.NewDeliveryHandler(uc)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterDeliveryServiceServer(srv, h)

	log.Printf("delivery-service gRPC listening on %s", grpcAddr)
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

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
