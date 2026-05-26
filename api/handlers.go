package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/Saurabh/golink/postgres"
	goredis "github.com/Saurabh/golink/redis"

	pb "github.com/Saurabh/golink/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ShortenRequest is the JSON body for POST /shorten.
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse is the JSON response for POST /shorten.
type ShortenResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
}

// AnalyticsResponse is the JSON response for GET /analytics/{shortCode}.
type AnalyticsResponse struct {
	ShortCode    string `json:"short_code"`
	ClickCount   int64  `json:"click_count"`
	LastClicked  string `json:"last_clicked_at,omitempty"`
}

// grpcClient connects to the analytics gRPC service.
var grpcClient pb.AnalyticsServiceClient

// InitGRPCClient sets up the gRPC connection to the analytics service.
func InitGRPCClient(addr string) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Println("Warning: Could not connect to Analytics Service:", err)
		return
	}
	grpcClient = pb.NewAnalyticsServiceClient(conn)
	log.Println("Connected to Analytics gRPC Service at", addr)
}

// ShortenHandler handles POST /shorten
// It generates a short code, saves to DB, and returns the short URL.
func ShortenHandler(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, `{"error": "invalid request, provide 'url' field"}`, http.StatusBadRequest)
		return
	}

	// Generate a unique short code
	shortCode := GenerateShortCode()

	// Save to PostgreSQL
	if err := postgres.SaveURL(shortCode, req.URL); err != nil {
		log.Println("DB save error:", err)
		http.Error(w, `{"error": "could not save URL"}`, http.StatusInternalServerError)
		return
	}

	// Create analytics row
	if err := postgres.InsertAnalytics(shortCode); err != nil {
		log.Println("Analytics insert error:", err)
	}

	// Cache in Redis
	goredis.SetCachedURL(shortCode, req.URL)

	// Return response
	resp := ShortenResponse{
		ShortCode: shortCode,
		ShortURL:  "http://localhost:8080/" + shortCode,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// RedirectHandler handles GET /{shortCode}
// It uses cache-aside: check Redis first, then fall back to DB.
func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	shortCode := mux.Vars(r)["shortCode"]

	// Step 1: Check Redis cache
	if longURL, found := goredis.GetCachedURL(shortCode); found {
		log.Println("Cache HIT for", shortCode)
		recordClick(shortCode) // async analytics via gRPC
		http.Redirect(w, r, longURL, http.StatusFound)
		return
	}

	// Step 2: Cache MISS — query PostgreSQL
	log.Println("Cache MISS for", shortCode)
	longURL, err := postgres.GetLongURL(shortCode)
	if err != nil {
		http.Error(w, `{"error": "short URL not found"}`, http.StatusNotFound)
		return
	}

	// Step 3: Refill cache
	goredis.SetCachedURL(shortCode, longURL)

	// Record the click via gRPC
	recordClick(shortCode)

	// Redirect to the original URL
	http.Redirect(w, r, longURL, http.StatusFound)
}

// AnalyticsHandler handles GET /analytics/{shortCode}
// It fetches click stats via gRPC from the Analytics Service.
func AnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	shortCode := mux.Vars(r)["shortCode"]

	// Try gRPC first
	if grpcClient != nil {
		stats, err := grpcClient.GetStats(r.Context(), &pb.GetStatsRequest{
			ShortCode: shortCode,
		})
		if err == nil {
			resp := AnalyticsResponse{
				ShortCode:   stats.ShortCode,
				ClickCount:  stats.ClickCount,
				LastClicked: stats.LastClickedAt,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		log.Println("gRPC GetStats error, falling back to DB:", err)
	}

	// Fallback: query DB directly
	clickCount, lastClicked, err := postgres.GetAnalytics(shortCode)
	if err != nil {
		http.Error(w, `{"error": "analytics not found"}`, http.StatusNotFound)
		return
	}

	resp := AnalyticsResponse{
		ShortCode:  shortCode,
		ClickCount: clickCount,
	}
	if lastClicked != nil {
		resp.LastClicked = lastClicked.Format("2006-01-02 15:04:05")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// recordClick sends a click event to the Analytics Service via gRPC.
// It runs in a goroutine so it doesn't block the redirect.
func recordClick(shortCode string) {
	if grpcClient == nil {
		// Fallback: update DB directly if gRPC is not available
		postgres.IncrementClick(shortCode)
		return
	}

	go func() {
		_, err := grpcClient.RecordClick(
			// using background context since this is async
			context.Background(),
			&pb.RecordClickRequest{ShortCode: shortCode},
		)
		if err != nil {
			log.Println("gRPC RecordClick error:", err)
			// Fallback to direct DB update
			postgres.IncrementClick(shortCode)
		}
	}()
}
