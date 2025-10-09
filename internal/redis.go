package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// Global Redis client instance
var RedisClient *redis.Client

/**
 * InitRedis initializes the Redis connection
 * @returns {redis.Client, error} Redis client and error if any
 * @description
 * - Creates Redis client connection
 * - Tests connection with ping
 * - Configures connection pool settings
 * - Sets default options for connection
 * @throws
 * - Redis connection errors
 * - Ping errors
 */
func InitRedis() (*redis.Client, error) {
	// Get Redis configuration from environment or use defaults
	addr := "localhost:6379" // Default Redis address
	password := ""           // Default Redis password
	db := 0                  // Default Redis database

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
		IdleTimeout:  5 * time.Minute,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Store global instance
	RedisClient = client

	return client, nil
}

/**
 * GetRedis returns the global Redis client instance
 * @returns {redis.Client} Redis client
 * @description
 * - Provides access to the global Redis client
 * - Returns nil if Redis is not initialized
 */
func GetRedis() *redis.Client {
	return RedisClient
}

/**
 * CloseRedis closes the Redis connection
 * @description
 * - Closes the Redis connection
 * - Should be called on application shutdown
 * @throws
 * - Redis close errors
 */
func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

/**
 * CacheSet sets a value in Redis cache with expiration
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} key - Cache key
 * @param {string} value - Cache value
 * @param {time.Duration} expiration - Cache expiration time
 * @returns {error} Error if operation fails
 * @description
 * - Sets a key-value pair in Redis
 * - Sets expiration time for the key
 * - Uses JSON serialization for complex values
 * @throws
 * - Redis operation errors
 */
func CacheSet(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if RedisClient == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	err := RedisClient.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

/**
 * CacheGet gets a value from Redis cache
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} key - Cache key
 * @returns {string, error} Cache value and error if any
 * @description
 * - Retrieves a value by key from Redis
 * - Returns nil if key doesn't exist
 * - Handles Redis errors gracefully
 * @throws
 * - Redis operation errors
 */
func CacheGet(ctx context.Context, key string) (string, error) {
	if RedisClient == nil {
		return "", fmt.Errorf("Redis client is not initialized")
	}

	val, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // Key doesn't exist
		}
		return "", fmt.Errorf("failed to get cache: %w", err)
	}

	return val, nil
}

/**
 * CacheDelete deletes a value from Redis cache
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} key - Cache key
 * @returns {error} Error if operation fails
 * @description
 * - Deletes a key from Redis
 * - Handles Redis errors gracefully
 * @throws
 * - Redis operation errors
 */
func CacheDelete(ctx context.Context, key string) error {
	if RedisClient == nil {
		return fmt.Errorf("Redis client is not initialized")
	}

	err := RedisClient.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	return nil
}

/**
 * CacheExists checks if a key exists in Redis cache
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} key - Cache key
 * @returns {bool, error} True if key exists, false otherwise
 * @description
 * - Checks if a key exists in Redis
 * - Handles Redis errors gracefully
 * @throws
 * - Redis operation errors
 */
func CacheExists(ctx context.Context, key string) (bool, error) {
	if RedisClient == nil {
		return false, fmt.Errorf("Redis client is not initialized")
	}

	val, err := RedisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache existence: %w", err)
	}

	return val > 0, nil
}

/**
 * CacheInvalidatePattern deletes all keys matching a pattern
 * @param {context.Context} ctx - Context for request cancellation
 * @param {string} pattern - Pattern to match keys
 * @returns {int64, error} Number of keys deleted
 * @description
 * - Deletes all keys matching the given pattern
 * - Uses SCAN for performance with large datasets
 * - Handles Redis errors gracefully
 * @throws
 * - Redis operation errors
 */
func CacheInvalidatePattern(ctx context.Context, pattern string) (int64, error) {
	if RedisClient == nil {
		return 0, fmt.Errorf("Redis client is not initialized")
	}

	var deleted int64
	iter := RedisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		err := RedisClient.Del(ctx, key).Err()
		if err != nil {
			logrus.WithError(err).WithField("key", key).Error("Failed to delete cache key")
			continue
		}
		deleted++
	}

	if err := iter.Err(); err != nil {
		return deleted, fmt.Errorf("failed to scan cache keys: %w", err)
	}

	return deleted, nil
}
