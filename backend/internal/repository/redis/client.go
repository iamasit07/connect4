package redis

import (
	"context"
	"log"
	"time"

	"github.com/iamasit07/4-in-a-row/backend/internal/config"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var redisEnabled bool

// InitRedis initializes Redis connection
func InitRedis() error {
	addr := config.GetEnv("REDIS_URL", "localhost:6379")
	password := config.GetEnv("REDIS_PASSWORD", "")

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	// Test connection
	ctx := context.Background()
	err := RedisClient.Ping(ctx).Err()
	if err != nil {
		log.Printf("[REDIS] Warning: Could not connect to Redis: %v. Falling back to PostgreSQL only.", err)
		redisEnabled = false
		return nil // Don't fail startup if Redis is unavailable
	}

	redisEnabled = true
	log.Println("[REDIS] Connected successfully")
	return nil
}

// IsRedisEnabled returns whether Redis is available
func IsRedisEnabled() bool {
	return redisEnabled
}

// CloseRedis closes the Redis connection
func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

// RedisCache acts as a wrapper around redis.Client to implement CacheRepository interface
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new RedisCache instance
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

// Set stores a key-value pair with expiration
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value by key
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Del deletes keys
func (r *RedisCache) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}
