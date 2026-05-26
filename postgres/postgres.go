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
func SaveURL(shortCode, longURL string) error {
	_, err := DB.Exec(
		"INSERT INTO urls (short_code, long_url) VALUES ($1, $2)",
		shortCode, longURL,
	)
	return err
}

// GetLongURL retrieves the original URL for a given short code.
func GetLongURL(shortCode string) (string, error) {
	var longURL string
	err := DB.QueryRow(
		"SELECT long_url FROM urls WHERE short_code = $1",
		shortCode,
	).Scan(&longURL)
	return longURL, err
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
