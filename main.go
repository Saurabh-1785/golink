package main //Package Declaration:Every executable Go application must have a main package.

import (
	"log" //simple console logging during startup
	"os" // Used for interacting with the operating system like read environment variables

	"github.com/Saurabh/golink/api" //starts the HTTP server and holds the gRPC client init
	"github.com/Saurabh/golink/postgres" //handles PostgreSQL connection
	goredis "github.com/Saurabh/golink/redis" //handles Redis connection
	// goredis isa alias for the redis package to avoid naming conflicts and improve code readability.

	// Import proto package to register the JSON codec for gRPC
	_ "github.com/Saurabh/golink/proto"
	// blank(_) is used to Import this package only for its side effects.
)

func main() {
	log.Println("Starting GoLink URL Service...")

	// Connect to PostgreSQL: Open Database Connection -> Check Connection ->Store DB Instance
	postgres.Connect()

	// Connect to Redis
	goredis.Connect()

	// Connect to Analytics gRPC Service: reads ANALYTICS_GRPC_ADDR, If not found:localhost:50051 is used.
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


// Execution flow:
/*
main()
 │
 ├── Print startup message
 │
 ├── Connect PostgreSQL
 │
 ├── Connect Redis
 │
 ├── Read Analytics gRPC address
 │
 ├── Connect Analytics Service
 │
 ├── Read HTTP Port
 │
 └── Start HTTP Server

*/
