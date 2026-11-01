package api

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// visitor holds the request count and the window start time for one IP.
type visitor struct {
	count       int
	windowStart time.Time
}

const (
	rateLimit  = 10              // max requests per window
	rateWindow = 1 * time.Minute // window duration
)

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

func init() {
	// Periodically drop visitors whose window has expired so the map
	// doesn't grow forever with one entry per IP ever seen.
	go func() {
		ticker := time.NewTicker(rateWindow)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.windowStart) > rateWindow {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

// clientIP returns the client IP without the port. r.RemoteAddr is "ip:port",
// and the port changes per connection, so keying on it would let every new
// connection bypass the limit.
//
// X-Forwarded-For is intentionally NOT trusted: without a known reverse proxy
// in front, clients could spoof it to evade the limiter.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// RateLimitMiddleware allows at most rateLimit requests per IP per rateWindow.
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)

		mu.Lock()
		v, exists := visitors[ip]
		if !exists || time.Since(v.windowStart) > rateWindow {
			// New visitor or window expired — reset
			visitors[ip] = &visitor{count: 1, windowStart: time.Now()}
			mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}

		if v.count >= rateLimit {
			mu.Unlock()
			http.Error(w, `{"error": "rate limit exceeded, try again later"}`, http.StatusTooManyRequests)
			return
		}

		v.count++
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
