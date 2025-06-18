package database

import (
	"context"
	"fmt"
	"log"

	"bome-backend/internal/config"

	"github.com/redis/go-redis/v9"
)

// Redis wraps the Redis client
type Redis struct {
	*redis.Client
}

// NewRedis creates a new Redis connection
func NewRedis(cfg *config.Config) (*Redis, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Test the connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	log.Println("Redis connection established")

	return &Redis{rdb}, nil
}

// Close closes the Redis connection
func (r *Redis) Close() error {
	return r.Client.Close()
}
