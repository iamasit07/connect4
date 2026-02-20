package redis

import (
	"context"
	"log"
	"time"

	"github.com/iamasit07/connect4/backend/internal/config"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var redisEnabled bool

func InitRedis() error {
	url := config.GetEnv("REDIS_URL", "localhost:6379")

	opts, err := redis.ParseURL(url)
	if err != nil {
		opts = &redis.Options{
			Addr:     url,
			Password: config.GetEnv("REDIS_PASSWORD", ""),
		}
	}

	RedisClient = redis.NewClient(opts)

	// Verify the connection with a Ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		log.Printf("[REDIS] Warning: Could not connect to Redis: %v. Falling back to PostgreSQL only.", err)
		redisEnabled = false
		return nil
	}

	redisEnabled = true
	log.Println("[REDIS] Connected successfully")
	return nil
}

func IsRedisEnabled() bool {
	return redisEnabled
}

func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisCache) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}