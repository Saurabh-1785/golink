package main

import (
	"context"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/Saurabh/golink/postgres"
	pb "github.com/Saurabh/golink/proto"
)

// server implements the AnalyticsServiceServer interface.
type server struct {
	pb.UnimplementedAnalyticsServiceServer
}

// RecordClick increments the click count for a short code.
func (s *server) RecordClick(ctx context.Context, req *pb.RecordClickRequest) (*pb.RecordClickResponse, error) {
	log.Println("RecordClick called for:", req.ShortCode)

	err := postgres.IncrementClick(req.ShortCode)
	if err != nil {
		log.Println("Error recording click:", err)
		return &pb.RecordClickResponse{Success: false}, err
	}

	return &pb.RecordClickResponse{Success: true}, nil
}

// GetStats returns click analytics for a short code.
func (s *server) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	log.Println("GetStats called for:", req.ShortCode)

	clickCount, lastClicked, err := postgres.GetAnalytics(req.ShortCode)
	if err != nil {
		return nil, err
	}

	resp := &pb.GetStatsResponse{
		ShortCode:  req.ShortCode,
		ClickCount: clickCount,
	}
	if lastClicked != nil {
		resp.LastClickedAt = lastClicked.Format("2006-01-02 15:04:05")
	}

	return resp, nil
}

func main() {
	// Connect to PostgreSQL
	postgres.Connect()

	// Start gRPC server
	port := getEnv("GRPC_PORT", "50051")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAnalyticsServiceServer(grpcServer, &server{})

	log.Println("Analytics gRPC Service running on port", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("Failed to serve:", err)
	}
}

// getEnv reads an env variable with a fallback default.
func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
