package main

import (
	"log"
	"os"

	"github.com/Saurabh/golink/api"
	"github.com/Saurabh/golink/postgres"
	goredis "github.com/Saurabh/golink/redis"

	// Import proto package to register the JSON codec for gRPC
	_ "github.com/Saurabh/golink/proto"
)

func main() {
	log.Println("Starting GoLink URL Service...")

	// Connect to PostgreSQL
	postgres.Connect()

	// Connect to Redis
	goredis.Connect()

	// Connect to Analytics gRPC Service
	grpcAddr := getEnv("ANALYTICS_GRPC_ADDR", "localhost:50051")
	api.InitGRPCClient(grpcAddr)

	// Start HTTP server
	port := getEnv("PORT", "8080")
	api.StartServer(port)
}

// getEnv reads an env variable with a fallback default.
func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
