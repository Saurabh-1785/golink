package redis

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client is the global Redis client.
var Client *redis.Client

// Ctx is a shared context for Redis operations.
var Ctx = context.Background()

// CacheTTL is how long URLs stay cached in Redis.
const CacheTTL = 10 * time.Minute

// Connect initializes the Redis client.
func Connect() {
	addr := getEnv("REDIS_ADDR", "localhost:6379")

	Client = redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Retry connection (wait for Redis to be ready in Docker)
	for i := 0; i < 10; i++ {
		_, err := Client.Ping(Ctx).Result()
		if err == nil {
			log.Println("Connected to Redis")
			return
		}
		log.Printf("Redis not ready, retrying in 2s... (%d/10)\n", i+1)
		time.Sleep(2 * time.Second)
	}
	log.Fatal("Could not connect to Redis")
}

// GetCachedURL checks Redis for a cached long URL.
// Returns the URL and true if found, empty string and false if not.
func GetCachedURL(shortCode string) (string, bool) {
	val, err := Client.Get(Ctx, "url:"+shortCode).Result()
	if err != nil {
		return "", false
	}
	return val, true
}

// SetCachedURL stores a URL mapping in Redis with TTL.
func SetCachedURL(shortCode, longURL string) {
	err := Client.Set(Ctx, "url:"+shortCode, longURL, CacheTTL).Err()
	if err != nil {
		log.Println("Redis SET error:", err)
	}
}

// DeleteCachedURL removes a short code from the Redis cache.
func DeleteCachedURL(shortCode string) {
	Client.Del(Ctx, "url:"+shortCode)
}

// getEnv reads an env variable with a fallback default.
func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
