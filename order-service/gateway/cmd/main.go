package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"food-delivery/order-service/gateway/internal/proxy"
	"food-delivery/order-service/gateway/internal/router"
)

func main() {
	orderServiceAddr := getEnv("ORDER_SERVICE_ADDR", "localhost:50053")
	jwtSecret := getEnv("JWT_SECRET", "change-me-in-production")
	httpAddr := getEnv("HTTP_ADDR", ":8083")

	orderProxy, err := proxy.NewOrderProxy(orderServiceAddr)
	if err != nil {
		log.Fatalf("connect order-service: %v", err)
	}
	defer orderProxy.Close()

	srv := &http.Server{
		Addr:    httpAddr,
		Handler: router.New(orderProxy, jwtSecret),
	}

	go func() {
		log.Printf("order-gateway listening on %s", httpAddr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down order-gateway...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
