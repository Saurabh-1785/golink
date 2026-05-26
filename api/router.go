package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// NewRouter creates and returns the HTTP router with all routes.
func NewRouter() *mux.Router {
	r := mux.NewRouter()

	// POST /shorten — create a new short URL
	r.HandleFunc("/shorten", ShortenHandler).Methods("POST")

	// GET /analytics/{shortCode} — get click stats
	// This must be before the redirect catch-all
	r.HandleFunc("/analytics/{shortCode}", AnalyticsHandler).Methods("GET")

	// GET /{shortCode} — redirect to original URL
	r.HandleFunc("/{shortCode}", RedirectHandler).Methods("GET")

	return r
}

// StartServer starts the HTTP server on the given port.
func StartServer(port string) {
	router := NewRouter()
	log.Println("URL Service running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
