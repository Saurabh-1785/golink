package ai

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// PageMetadata holds the extracted title and description from a webpage.
type PageMetadata struct {
	Title       string
	Description string
}

// FetchMetadata fetches the <title> and <meta name="description"> from a URL.
// Uses a 5-second timeout to avoid blocking the request for too long.
// If the fetch fails, it falls back to extracting info from the URL path itself.
func FetchMetadata(url string) PageMetadata {
	meta := PageMetadata{}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		// Fallback: use the URL path as a rough title
		meta.Title = fallbackTitleFromURL(url)
		return meta
	}
	defer resp.Body.Close()

	// Parse HTML to extract title and meta description
	tokenizer := html.NewTokenizer(resp.Body)
	var inTitle bool

	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			// End of document or error — return what we have
			if meta.Title == "" {
				meta.Title = fallbackTitleFromURL(url)
			}
			return meta

		case html.StartTagToken:
			t := tokenizer.Token()
			if t.Data == "title" {
				inTitle = true
			}
			// Check for <meta name="description" content="...">
			if t.Data == "meta" {
				var name, content string
				for _, attr := range t.Attr {
					if strings.EqualFold(attr.Key, "name") {
						name = attr.Val
					}
					if strings.EqualFold(attr.Key, "content") {
						content = attr.Val
					}
				}
				if strings.EqualFold(name, "description") && content != "" {
					meta.Description = content
				}
			}
			// Stop parsing after </head> — no need to parse the body
			if t.Data == "body" {
				if meta.Title == "" {
					meta.Title = fallbackTitleFromURL(url)
				}
				return meta
			}

		case html.TextToken:
			if inTitle {
				meta.Title = strings.TrimSpace(tokenizer.Token().Data)
				inTitle = false
			}
		}
	}
}

// fallbackTitleFromURL extracts a readable title from the URL path.
// Example: "https://go.dev/doc/tutorial/getting-started" → "doc tutorial getting-started"
func fallbackTitleFromURL(rawURL string) string {
	// Strip protocol
	u := rawURL
	if idx := strings.Index(u, "://"); idx != -1 {
		u = u[idx+3:]
	}
	// Strip domain — take everything after the first /
	if idx := strings.Index(u, "/"); idx != -1 {
		u = u[idx+1:]
	} else {
		return rawURL // No path, just return the URL
	}
	// Strip query params and fragments
	if idx := strings.IndexAny(u, "?#"); idx != -1 {
		u = u[:idx]
	}
	// Replace slashes with spaces
	u = strings.ReplaceAll(u, "/", " ")
	u = strings.TrimSpace(u)

	if u == "" {
		return fmt.Sprintf("page from %s", rawURL)
	}
	return u
}
