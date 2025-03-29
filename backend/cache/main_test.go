package cache

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Setup test environment
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Unsetenv("REDIS_USERNAME")
	os.Unsetenv("REDIS_PASSWORD")

	// Initialize Redis client for tests
	Init()

	// Run tests
	code := m.Run()

	// Cleanup
	Close()

	os.Exit(code)
}

func TestInit(t *testing.T) {
	t.Run("default connection", func(t *testing.T) {
		assert.NotNil(t, instance, "Redis client should be initialized")
	})

	t.Run("connection with auth", func(t *testing.T) {
		// Backup current instance
		oldInstance := instance

		// Set auth credentials
		os.Setenv("REDIS_USERNAME", "default")
		os.Setenv("REDIS_PASSWORD", "password")
		defer func() {
			os.Unsetenv("REDIS_USERNAME")
			os.Unsetenv("REDIS_PASSWORD")
			instance = oldInstance
		}()

		Init()
		assert.NotNil(t, instance, "Redis client should be initialized with auth")
	})
}

func TestSetAndGet(t *testing.T) {
	ctx := context.Background()
	key := "test_key"
	value := "test_value"

	err := Set(ctx, key, value, 0)
	require.NoError(t, err, "Set should not return error")

	t.Run("get existing key", func(t *testing.T) {
		val, err := Get(ctx, key)
		require.NoError(t, err, "Get should not return error")
		assert.Equal(t, value, val, "Retrieved value should match set value")
	})

	t.Run("get non-existent key", func(t *testing.T) {
		_, err := Get(ctx, "non_existent_key")
		require.Error(t, err, "Get should return error for non-existent key")
		assert.Contains(t, err.Error(), "does not exist", "Error should indicate key doesn't exist")
	})

	// Cleanup
	instance.Del(ctx, key)
}

func TestGetWithDefault(t *testing.T) {
	ctx := context.Background()
	key := "test_default_key"
	value := "test_value"
	defaultValue := "default_value"

	t.Run("existing key", func(t *testing.T) {
		err := Set(ctx, key, value, 0)
		require.NoError(t, err)

		val, err := GetWithDefault(ctx, key, defaultValue)
		require.NoError(t, err)
		assert.Equal(t, value, val, "Should return actual value when key exists")

		// Cleanup
		instance.Del(ctx, key)
	})

	t.Run("non-existent key", func(t *testing.T) {
		val, err := GetWithDefault(ctx, "non_existent_key", defaultValue)
		require.NoError(t, err)
		assert.Equal(t, defaultValue, val, "Should return default value when key doesn't exist")
	})
}

func TestSetWithExpiration(t *testing.T) {
	ctx := context.Background()
	key := "expiring_key"
	value := "temp_value"
	expiration := 1 * time.Second

	err := Set(ctx, key, value, expiration)
	require.NoError(t, err)

	t.Run("before expiration", func(t *testing.T) {
		val, err := Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, value, val, "Value should exist before expiration")
	})

	t.Run("after expiration", func(t *testing.T) {
		time.Sleep(expiration + 100*time.Millisecond)
		_, err := Get(ctx, key)
		require.Error(t, err, "Should return error after expiration")
		assert.Contains(t, err.Error(), "does not exist", "Error should indicate key doesn't exist")
	})
}

func TestClose(t *testing.T) {
	// Create a new client to test Close without affecting the main instance
	options := &redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	}
	client := redis.NewClient(options)

	// Verify connection works before close
	_, err := client.Ping(context.Background()).Result()
	require.NoError(t, err, "Should connect successfully")

	// Replace instance for test
	oldInstance := instance
	instance = client
	defer func() {
		instance = oldInstance
	}()

	err = Close()
	require.NoError(t, err, "Close should not return error")

	// Verify connection is closed
	_, err = client.Ping(context.Background()).Result()
	require.Error(t, err, "Should return error after connection closed")
}
