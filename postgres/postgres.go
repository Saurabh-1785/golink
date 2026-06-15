package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// DB is the global database connection.
var DB *sql.DB

// Connect opens a connection to PostgreSQL using env variables.
func Connect() {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "golink")
	password := getEnv("DB_PASSWORD", "golink123")
	dbname := getEnv("DB_NAME", "golink")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to open DB:", err)
	}

	// Retry connection (wait for DB to be ready in Docker)
	for i := 0; i < 10; i++ {
		err = DB.Ping()
		if err == nil {
			log.Println("Connected to PostgreSQL")
			return
		}
		log.Printf("DB not ready, retrying in 2s... (%d/10)\n", i+1)
		time.Sleep(2 * time.Second)
	}
	log.Fatal("Could not connect to PostgreSQL:", err)
}

// SaveURL inserts a new short URL mapping into the database.
// expiresAt is optional — pass nil for no expiration.
func SaveURL(shortCode, longURL string, expiresAt *time.Time) error {
	_, err := DB.Exec(
		"INSERT INTO urls (short_code, long_url, expires_at) VALUES ($1, $2, $3)",
		shortCode, longURL, expiresAt,
	)
	return err
}

// GetLongURL retrieves the original URL for a given short code.
// Also returns the expiry time (nil if never expires) so callers can
// set the Redis TTL precisely, avoiding future DB re-checks on cache HITs.
// Returns an error if the URL has expired.
func GetLongURL(shortCode string) (string, *time.Time, error) {
	var longURL string
	var expiresAt sql.NullTime

	err := DB.QueryRow(
		"SELECT long_url, expires_at FROM urls WHERE short_code = $1",
		shortCode,
	).Scan(&longURL, &expiresAt)
	if err != nil {
		return "", nil, err
	}

	// Check expiration
	if expiresAt.Valid && time.Now().After(expiresAt.Time) {
		return "", nil, fmt.Errorf("URL has expired")
	}

	if expiresAt.Valid {
		return longURL, &expiresAt.Time, nil
	}
	return longURL, nil, nil
}

// ShortCodeExists checks if a short code is already taken.
func ShortCodeExists(shortCode string) bool {
	var id int
	err := DB.QueryRow("SELECT id FROM urls WHERE short_code = $1", shortCode).Scan(&id)
	return err == nil
}

// InsertAnalytics creates an analytics row for a new short code.
func InsertAnalytics(shortCode string) error {
	_, err := DB.Exec(
		"INSERT INTO analytics (short_code, click_count) VALUES ($1, 0)",
		shortCode,
	)
	return err
}

// IncrementClick adds 1 to the click count and updates the timestamp.
func IncrementClick(shortCode string) error {
	_, err := DB.Exec(
		`UPDATE analytics 
		 SET click_count = click_count + 1, last_clicked_at = NOW() 
		 WHERE short_code = $1`,
		shortCode,
	)
	return err
}

// GetAnalytics returns click count and last clicked time for a short code.
func GetAnalytics(shortCode string) (int64, *time.Time, error) {
	var clickCount int64
	var lastClicked sql.NullTime

	err := DB.QueryRow(
		"SELECT click_count, last_clicked_at FROM analytics WHERE short_code = $1",
		shortCode,
	).Scan(&clickCount, &lastClicked)
	if err != nil {
		return 0, nil, err
	}

	if lastClicked.Valid {
		return clickCount, &lastClicked.Time, nil
	}
	return clickCount, nil, nil
}

// getEnv reads an env variable with a fallback default.
func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
