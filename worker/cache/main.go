package cache

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var instance *redis.Client

func Init() {
	var host string = os.Getenv("REDIS_HOST")
	var port string = os.Getenv("REDIS_PORT")
	var username string = os.Getenv("REDIS_USERNAME")
	var password string = os.Getenv("REDIS_PASSWORD")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "6379"
	}

	options := &redis.Options{
		Addr:     host + ":" + port,
		Username: username,
		Password: password, // Set password if required
		DB:       0,        // Default DB index
	}

	// Create Redis client
	instance = redis.NewClient(options)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := instance.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	fmt.Println("Redis connected!")
}

// Set stores a key-value pair in Redis with optional expiration
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := instance.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}
	return nil
}

// Get retrieves a value by key from Redis
func Get(ctx context.Context, key string) (string, error) {
	val, err := instance.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key %s does not exist", key)
	} else if err != nil {
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return val, nil
}

// GetWithDefault retrieves a value or returns a default if key doesn't exist
func GetWithDefault(ctx context.Context, key, defaultValue string) (string, error) {
	val, err := instance.Get(ctx, key).Result()
	if err == redis.Nil {
		return defaultValue, nil
	} else if err != nil {
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return val, nil
}

// Close cleans up the Redis connection
func Close() error {
	return instance.Close()
}
