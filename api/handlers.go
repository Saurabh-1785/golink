package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/Saurabh/golink/ai"
	"github.com/Saurabh/golink/postgres"
	goredis "github.com/Saurabh/golink/redis"

	pb "github.com/Saurabh/golink/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ShortenRequest is the JSON body for POST /shorten.
type ShortenRequest struct {
	URL           string `json:"url"`
	CustomAlias   string `json:"custom_alias,omitempty"`    // optional
	ExpiresInHours int   `json:"expires_in_hours,omitempty"` // optional; 0 = never
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

// SuggestSlugsRequest is the JSON body for POST /suggest-slugs.
type SuggestSlugsRequest struct {
	URL string `json:"url"`
}

// SuggestSlugsResponse is the JSON response for POST /suggest-slugs.
type SuggestSlugsResponse struct {
	Suggestions []string `json:"suggestions"`
}

// aliasPattern restricts custom aliases to URL-safe characters, 3-30 long
// (matches the VARCHAR(30) short_code column).
var aliasPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{3,30}$`)

// maxCodeAttempts bounds retries when a generated short code collides.
const maxCodeAttempts = 5

// baseURL is the public prefix for short links, e.g. "https://go.example.com".
// Configured via BASE_URL; defaults to local dev.
var baseURL = strings.TrimRight(envOr("BASE_URL", "http://localhost:8080"), "/")

// envOr reads an env variable with a fallback default.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
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

	// Validate custom alias: letters, digits, hyphen, underscore; 3-30 chars.
	// Hyphens are allowed so AI-suggested slugs (e.g. "learn-go") can be used.
	if req.CustomAlias != "" && !aliasPattern.MatchString(req.CustomAlias) {
		http.Error(w, `{"error": "custom_alias must be 3-30 characters of letters, digits, '-' or '_'"}`, http.StatusBadRequest)
		return
	}

	// Compute optional expiry time
	var expiresAt *time.Time
	if req.ExpiresInHours > 0 {
		t := time.Now().Add(time.Duration(req.ExpiresInHours) * time.Hour)
		expiresAt = &t
	}

	// Save to PostgreSQL. The UNIQUE constraint is the source of truth for
	// availability (no check-then-insert race):
	//   - custom alias taken      -> 409
	//   - generated code collides -> retry with a fresh code
	var shortCode string
	var err error
	for attempt := 0; attempt < maxCodeAttempts; attempt++ {
		shortCode = req.CustomAlias
		if shortCode == "" {
			shortCode = GenerateShortCode()
		}
		err = postgres.SaveURL(shortCode, req.URL, expiresAt)
		if !errors.Is(err, postgres.ErrDuplicateShortCode) || req.CustomAlias != "" {
			break
		}
		log.Println("Short code collision, retrying:", shortCode)
	}
	if errors.Is(err, postgres.ErrDuplicateShortCode) && req.CustomAlias != "" {
		http.Error(w, `{"error": "custom alias already taken"}`, http.StatusConflict)
		return
	}
	if err != nil {
		log.Println("DB save error:", err)
		http.Error(w, `{"error": "could not save URL"}`, http.StatusInternalServerError)
		return
	}

	// Create analytics row
	if err := postgres.InsertAnalytics(shortCode); err != nil {
		log.Println("Analytics insert error:", err)
	}

	// Cache in Redis — pass expiresAt so Redis TTL mirrors URL lifetime
	goredis.SetCachedURL(shortCode, req.URL, expiresAt)

	// Return response
	resp := ShortenResponse{
		ShortCode: shortCode,
		ShortURL:  baseURL + "/" + shortCode,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// RedirectHandler handles GET /{shortCode}
// It uses cache-aside: check Redis first, then fall back to DB.
func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	shortCode := mux.Vars(r)["shortCode"]

	// Step 1: Cache HIT — Redis TTL already enforces expiry, no DB check needed.
	// If the URL was expired, Redis would have evicted the key already.
	if longURL, found := goredis.GetCachedURL(shortCode); found {
		log.Println("Cache HIT for", shortCode)
		recordClick(shortCode)
		http.Redirect(w, r, longURL, http.StatusFound)
		return
	}

	// Step 2: Cache MISS — query PostgreSQL (checks expiry, returns expiresAt for cache TTL)
	log.Println("Cache MISS for", shortCode)
	longURL, expiresAt, err := postgres.GetLongURL(shortCode)
	if err != nil {
		if err.Error() == "URL has expired" {
			http.Error(w, `{"error": "URL has expired"}`, http.StatusGone)
			return
		}
		http.Error(w, `{"error": "short URL not found"}`, http.StatusNotFound)
		return
	}

	// Step 3: Refill cache — set Redis TTL to match URL's remaining lifetime
	goredis.SetCachedURL(shortCode, longURL, expiresAt)

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

// SuggestSlugsHandler handles POST /suggest-slugs
// It fetches webpage metadata, sends it to Gemini, and returns AI-generated slug suggestions.
func SuggestSlugsHandler(w http.ResponseWriter, r *http.Request) {
	// Check if AI is available
	if !ai.IsAvailable() {
		http.Error(w, `{"error": "AI slug suggestions are unavailable (GEMINI_API_KEY not configured)"}`, http.StatusServiceUnavailable)
		return
	}

	var req SuggestSlugsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, `{"error": "invalid request, provide 'url' field"}`, http.StatusBadRequest)
		return
	}

	// Step 1: Fetch webpage metadata (title + description)
	log.Println("Fetching metadata for:", req.URL)
	meta := ai.FetchMetadata(req.URL)
	log.Printf("Metadata — Title: %q, Description: %q\n", meta.Title, meta.Description)

	// Step 2: Generate slug suggestions via Gemini
	slugs, err := ai.GenerateSlugs(meta.Title, meta.Description)
	if err != nil {
		log.Println("AI slug generation error:", err)
		http.Error(w, `{"error": "failed to generate slug suggestions"}`, http.StatusInternalServerError)
		return
	}

	// Step 3: Filter out slugs that already exist in the database
	var available []string
	for _, slug := range slugs {
		if !postgres.ShortCodeExists(slug) {
			available = append(available, slug)
		}
	}

	// If all suggestions are taken, return what we have with a note
	if len(available) == 0 {
		available = slugs // Return them anyway — user can still edit
		log.Println("Warning: All suggested slugs already exist in DB")
	}

	// Return suggestions
	resp := SuggestSlugsResponse{
		Suggestions: available,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
