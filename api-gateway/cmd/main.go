package main

import (
	"log"
	"os"

	"food-delivery/api-gateway/internal/proxy"
	"food-delivery/api-gateway/internal/router"
)

func main() {
	userServiceAddr := getEnv("USER_SERVICE_ADDR", "localhost:50051")
	jwtSecret := getEnv("JWT_SECRET", "change-me-in-production")
	httpAddr := getEnv("HTTP_ADDR", ":8080")

	userProxy, err := proxy.NewUserProxy(userServiceAddr)
	if err != nil {
		log.Fatalf("connect user-service: %v", err)
	}

	r := router.New(userProxy, jwtSecret)

	log.Printf("api-gateway listening on %s", httpAddr)
	if err := r.Run(httpAddr); err != nil {
		log.Fatalf("run: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}