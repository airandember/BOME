package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"bome-backend/infrastructure/config"

	"github.com/redis/go-redis/v9"
)

// Redis wraps the Redis client
type Redis struct {
	*redis.Client
}

// NewRedis creates a new Redis connection with optimized settings
func NewRedis(cfg *config.Config) (*Redis, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		PoolSize:     50,               // Connection pool size
		MinIdleConns: 10,               // Minimum idle connections
		MaxIdleConns: 20,               // Maximum idle connections
		PoolTimeout:  30 * time.Second, // Pool timeout
		ReadTimeout:  5 * time.Second,  // Read timeout
		WriteTimeout: 5 * time.Second,  // Write timeout
	})

	// Test the connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	log.Printf("Redis connection established with pool size: %d", 50)
	return &Redis{rdb}, nil
}

// Close closes the Redis connection
func (r *Redis) Close() error {
	return r.Client.Close()
}
