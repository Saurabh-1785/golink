package api //This file belongs to the api package.

import (
	"log"
	"net/http" //Go's built-in HTTP package. Used to: http.ListenAndServe(...) -> which starts the web server.

	"github.com/gorilla/mux" //A router library. Incoming request go to router -> router matches the path and method -> calls the appropriate handler function.
)

// NewRouter creates and returns the HTTP router with all routes.
func NewRouter() *mux.Router { // * -> returns a pointer to a Router object.
	r := mux.NewRouter() //Creates an empty router

	// POST /shorten — create a new short URL
	// Rate limited (10 req/min per IP): creates rows in the DB.
	r.Handle("/shorten", RateLimitMiddleware(http.HandlerFunc(ShortenHandler))).Methods("POST") //route1: when a POST request is made to /shorten, the ShortenHandler function will be called.

	// POST /suggest-slugs — get AI-generated slug suggestions
	// Rate limited: each call hits the paid Gemini API and fetches a remote page.
	r.Handle("/suggest-slugs", RateLimitMiddleware(http.HandlerFunc(SuggestSlugsHandler))).Methods("POST") //route2: when a POST request is made to /suggest-slugs, the SuggestSlugsHandler function will be called. Uses Gemini AI to generate meaningful short URL slugs.

	// GET /analytics/{shortCode} — get click stats
	// This must be before the redirect catch-all
	r.HandleFunc("/analytics/{shortCode}", AnalyticsHandler).Methods("GET") //route3: when a GET request is made to /analytics/{shortCode}, the AnalyticsHandler function will be called. {shortCode} is a path variable that can be accessed in the handler.

	// GET /{shortCode} — redirect to original URL
	r.HandleFunc("/{shortCode}", RedirectHandler).Methods("GET") //route4: when a GET request is made to /{shortCode}, the RedirectHandler function will be called. This is a catch-all route for any short code, so it must be defined after the more specific routes.

	return r
}

// StartServer starts the HTTP server on the given port.
func StartServer(port string) {
	router := NewRouter()
	log.Println("URL Service running on port", port)
	// Rate limiting is applied per-route in NewRouter (write/AI endpoints only),
	// so redirects — the hot path — are never throttled.
	log.Fatal(http.ListenAndServe(":"+port, router))
}


/*
When main.go executes:

api.StartServer("8080")


StartServer()
     |
     v
NewRouter()
     |
     +--> Register POST /shorten
     |
     +--> Register GET /analytics/{shortCode}
     |
     +--> Register GET /{shortCode}
     |
     v
Return Router
     |
     v
Print startup log
     |
     v
ListenAndServe(":8080")
     |
     v
Waiting for requests...
*/