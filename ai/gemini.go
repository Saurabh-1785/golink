package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/genai"
)

// geminiClient is the global Gemini client (nil if API key not set).
var geminiClient *genai.Client

// geminiModel is the model to use for slug generation.
const geminiModel = "gemini-2.5-flash-lite"

// InitGemini initializes the Gemini client using the GEMINI_API_KEY env var.
// Fails gracefully — logs a warning if the key is missing.
func InitGemini() {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("Warning: GEMINI_API_KEY not set — AI slug suggestions will be unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Println("Warning: Could not initialize Gemini client:", err)
		return
	}

	geminiClient = client
	log.Println("Gemini AI client initialized (model:", geminiModel, ")")
}

// IsAvailable returns true if the Gemini client was successfully initialized.
func IsAvailable() bool {
	return geminiClient != nil
}

// GenerateSlugs sends the webpage metadata to Gemini and returns 3 slug suggestions.
// Returns an error if the client is not initialized or the API call fails.
func GenerateSlugs(title, description string) ([]string, error) {
	if geminiClient == nil {
		return nil, fmt.Errorf("Gemini AI is not available (API key not set)")
	}

	prompt := buildPrompt(title, description)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := geminiClient.Models.GenerateContent(ctx, geminiModel, genai.Text(prompt), nil)
	if err != nil {
		return nil, fmt.Errorf("Gemini API error: %w", err)
	}

	// Extract the text response
	text := result.Text()
	if text == "" {
		return nil, fmt.Errorf("Gemini returned empty response")
	}

	// Parse the JSON array from the response
	slugs, err := parseSlugs(text)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	return slugs, nil
}

// buildPrompt creates the structured prompt for Gemini.
func buildPrompt(title, description string) string {
	descPart := ""
	if description != "" {
		descPart = fmt.Sprintf("\nDescription: %s", description)
	}

	return fmt.Sprintf(`Generate exactly 3 short URL slugs for a webpage.

Rules:
- Lowercase only
- Use hyphens to separate words (no underscores or spaces)
- Maximum 20 characters each
- Make them memorable and meaningful
- SEO-friendly
- No special characters except hyphens

Title: %s%s

Respond ONLY with a JSON array of strings, nothing else.
Example: ["learn-go", "go-basics", "go-start"]`, title, descPart)
}

// parseSlugs extracts a JSON string array from the Gemini response text.
// Handles cases where the response might have markdown code fences.
func parseSlugs(text string) ([]string, error) {
	// Strip markdown code fences if present (```json ... ```)
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		// Remove opening fence (```json or ```)
		if idx := strings.Index(text, "\n"); idx != -1 {
			text = text[idx+1:]
		}
		// Remove closing fence
		if idx := strings.LastIndex(text, "```"); idx != -1 {
			text = text[:idx]
		}
		text = strings.TrimSpace(text)
	}

	var slugs []string
	if err := json.Unmarshal([]byte(text), &slugs); err != nil {
		return nil, fmt.Errorf("invalid JSON in response %q: %w", text, err)
	}

	if len(slugs) == 0 {
		return nil, fmt.Errorf("no slugs in response")
	}

	// Validate and clean each slug
	var valid []string
	for _, s := range slugs {
		s = strings.TrimSpace(s)
		s = strings.ToLower(s)
		if s != "" && len(s) <= 20 {
			valid = append(valid, s)
		}
	}

	if len(valid) == 0 {
		return nil, fmt.Errorf("no valid slugs after filtering")
	}

	return valid, nil
}
