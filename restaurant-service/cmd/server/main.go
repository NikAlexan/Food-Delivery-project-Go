package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"food-delivery/restaurant-service/internal/handler"
	"food-delivery/restaurant-service/internal/repository"
	"food-delivery/restaurant-service/internal/usecase"
	pb "food-delivery/restaurant-service/proto/pb"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	dsn := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/restaurantdb?sslmode=disable")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	natsURL := getEnv("NATS_URL", nats.DefaultURL)
	grpcAddr := getEnv("GRPC_ADDR", ":50052")

	// ── PostgreSQL ────────────────────────────────────────────────────────────
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Error("db open", "error", err)
		os.Exit(1)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := waitForDB(db, logger); err != nil {
		logger.Error("db ping failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// ── Redis ─────────────────────────────────────────────────────────────────
	rdb := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Warn("redis ping failed, caching disabled", "error", err)
	}
	defer rdb.Close()

	// ── NATS JetStream ────────────────────────────────────────────────────────
	nc, err := nats.Connect(natsURL,
		nats.MaxReconnects(10),
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			logger.Warn("NATS disconnected", "error", err)
		}),
	)
	if err != nil {
		logger.Warn("NATS connect failed, event publishing disabled", "error", err)
	}

	var js jetstream.JetStream
	if nc != nil {
		js, err = jetstream.New(nc)
		if err != nil {
			logger.Warn("JetStream init failed", "error", err)
		} else {
			// Ensure the stream exists; ignore "already exists" error.
			_, streamErr := js.CreateStream(context.Background(), jetstream.StreamConfig{
				Name:     "RESTAURANT",
				Subjects: []string{"restaurant.>"},
			})
			if streamErr != nil && !errors.Is(streamErr, jetstream.ErrStreamNameAlreadyInUse) {
				logger.Warn("JetStream stream create", "error", streamErr)
			}
		}
		defer nc.Drain()
	}

	// ── Wire up layers ────────────────────────────────────────────────────────
	repo := repository.NewPostgresRepo(db)
	uc := usecase.NewRestaurantUsecase(repo, rdb, js, logger)
	h := handler.NewRestaurantHandler(uc)

	// ── gRPC server ───────────────────────────────────────────────────────────
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Error("listen", "error", err)
		os.Exit(1)
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(loggingInterceptor(logger)),
	)
	pb.RegisterRestaurantServiceServer(srv, h)

	// Health check endpoint (used by Docker and load-balancers).
	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	go func() {
		logger.Info("restaurant-service gRPC listening", "addr", grpcAddr)
		if err := srv.Serve(lis); err != nil {
			logger.Error("serve", "error", err)
		}
	}()

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down restaurant-service...")
	healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	srv.GracefulStop()
	logger.Info("shutdown complete")
}

// waitForDB retries db.Ping up to 10 times with 2-second intervals.
func waitForDB(db *sql.DB, logger *slog.Logger) error {
	for i := range 10 {
		if err := db.Ping(); err == nil {
			return nil
		} else if i < 9 {
			logger.Warn("db not ready, retrying...", "attempt", i+1, "error", err)
			time.Sleep(2 * time.Second)
		} else {
			return err
		}
	}
	return nil
}

func loggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		logger.Info("grpc",
			"method", info.FullMethod,
			"duration_ms", time.Since(start).Milliseconds(),
			"error", err,
		)
		return resp, err
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
