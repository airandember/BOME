package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheManager handles multi-level caching
type CacheManager struct {
	redis   *redis.Client
	l1Cache *sync.Map // In-memory L1 cache
	l2Cache *sync.Map // Persistent L2 cache
	stats   *CacheStats
}

type CacheStats struct {
	L1Hits      int64
	L1Misses    int64
	L2Hits      int64
	L2Misses    int64
	RedisHits   int64
	RedisMisses int64
}

// NewCacheManager creates a new cache manager
func NewCacheManager(redisClient *redis.Client) *CacheManager {
	return &CacheManager{
		redis:   redisClient,
		l1Cache: &sync.Map{},
		l2Cache: &sync.Map{},
		stats:   &CacheStats{},
	}
}

// Get retrieves data from cache with fallback strategy
func (cm *CacheManager) Get(ctx context.Context, key string, dest interface{}) error {
	// Try L1 cache first
	if value, ok := cm.l1Cache.Load(key); ok {
		cm.stats.L1Hits++
		return json.Unmarshal(value.([]byte), dest)
	}
	cm.stats.L1Misses++

	// Try L2 cache
	if value, ok := cm.l2Cache.Load(key); ok {
		cm.stats.L2Hits++
		data := value.([]byte)
		cm.l1Cache.Store(key, data) // Promote to L1
		return json.Unmarshal(data, dest)
	}
	cm.stats.L2Misses++

	// Try Redis
	result := cm.redis.Get(ctx, key)
	if result.Err() == nil {
		cm.stats.RedisHits++
		data, _ := result.Bytes()
		cm.l1Cache.Store(key, data) // Promote to L1
		cm.l2Cache.Store(key, data) // Store in L2
		return json.Unmarshal(data, dest)
	}
	cm.stats.RedisMisses++

	return fmt.Errorf("cache miss for key: %s", key)
}

// Set stores data in all cache levels
func (cm *CacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// Store in all levels
	cm.l1Cache.Store(key, data)
	cm.l2Cache.Store(key, data)

	// Store in Redis with TTL
	return cm.redis.Set(ctx, key, data, ttl).Err()
}

// GetStats returns cache statistics
func (cm *CacheManager) GetStats() *CacheStats {
	return cm.stats
}
