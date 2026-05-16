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

	"food-delivery/api-gateway/internal/proxy"
	"food-delivery/api-gateway/internal/router"
)

func main() {
	userServiceAddr := getEnv("USER_SERVICE_ADDR", "localhost:50051")
	restaurantServiceAddr := getEnv("RESTAURANT_SERVICE_ADDR", "localhost:50052")
	orderServiceAddr := getEnv("ORDER_SERVICE_ADDR", "localhost:50053")
	deliveryServiceAddr := getEnv("DELIVERY_SERVICE_ADDR", "localhost:50054")
	jwtSecret := getEnv("JWT_SECRET", "change-me-in-production")
	httpAddr := getEnv("HTTP_ADDR", ":8080")

	userProxy, err := proxy.NewUserProxy(userServiceAddr)
	if err != nil {
		log.Fatalf("connect user-service: %v", err)
	}
	defer userProxy.Close()

	restaurantProxy, err := proxy.NewRestaurantProxy(restaurantServiceAddr)
	if err != nil {
		log.Fatalf("connect restaurant-service: %v", err)
	}
	defer restaurantProxy.Close()

	deliveryProxy, err := proxy.NewDeliveryProxy(deliveryServiceAddr)
	if err != nil {
		log.Fatalf("connect delivery-service: %v", err)
	}
	defer deliveryProxy.Close()

	orderProxy, err := proxy.NewOrderProxy(orderServiceAddr)
	if err != nil {
		log.Fatalf("connect order-service: %v", err)
	}
	defer orderProxy.Close()

	srv := &http.Server{
		Addr:    httpAddr,
		Handler: router.New(userProxy, deliveryProxy, restaurantProxy, orderProxy, jwtSecret),
	}

	go func() {
		log.Printf("api-gateway listening on %s", httpAddr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down api-gateway...")
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
