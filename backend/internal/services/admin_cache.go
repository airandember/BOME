package services

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AdminCacheService provides specialized caching for admin data
type AdminCacheService struct {
	mu           sync.RWMutex
	cache        map[string]*AdminCacheEntry
	adminMetrics *AdminCacheMetrics
	ctx          context.Context
	cancel       context.CancelFunc
}

// AdminCacheEntry represents a cached admin data entry
type AdminCacheEntry struct {
	Data        interface{}
	ExpiresAt   time.Time
	CreatedAt   time.Time
	AccessCount int
	LastAccess  time.Time
	Size        int64
	Type        string
}

// AdminCacheMetrics tracks admin cache performance
type AdminCacheMetrics struct {
	mu              sync.RWMutex
	Hits            int64
	Misses          int64
	Evictions       int64
	TotalSize       int64
	MaxSize         int64
	AverageLoadTime time.Duration
}

// AdminCacheConfig holds configuration for admin caching
type AdminCacheConfig struct {
	MaxSize         int64
	DefaultTTL      time.Duration
	CleanupInterval time.Duration
	MaxLoadTime     time.Duration
}

// NewAdminCacheService creates a new admin cache service
func NewAdminCacheService(config *AdminCacheConfig) *AdminCacheService {
	if config == nil {
		config = &AdminCacheConfig{
			MaxSize:         100 * 1024 * 1024, // 100MB
			DefaultTTL:      5 * time.Minute,
			CleanupInterval: 1 * time.Minute,
			MaxLoadTime:     30 * time.Second,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	service := &AdminCacheService{
		cache: make(map[string]*AdminCacheEntry),
		adminMetrics: &AdminCacheMetrics{
			MaxSize: config.MaxSize,
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Start cleanup goroutine
	go service.cleanupRoutine(config.CleanupInterval)

	return service
}

// Get retrieves data from admin cache
func (acs *AdminCacheService) Get(key string) (interface{}, bool) {
	acs.mu.RLock()
	defer acs.mu.RUnlock()

	entry, exists := acs.cache[key]
	if !exists {
		acs.adminMetrics.recordMiss()
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		acs.adminMetrics.recordMiss()
		return nil, false
	}

	// Update access metrics
	entry.AccessCount++
	entry.LastAccess = time.Now()
	acs.adminMetrics.recordHit()

	return entry.Data, true
}

// Set stores data in admin cache with admin-specific TTL
func (acs *AdminCacheService) Set(key string, data interface{}, ttl time.Duration, dataType string) error {
	acs.mu.Lock()
	defer acs.mu.Unlock()

	// Calculate entry size (approximate)
	size := acs.calculateSize(data)

	// Check if we need to evict entries to make space
	if acs.adminMetrics.TotalSize+size > acs.adminMetrics.MaxSize {
		acs.evictEntries(size)
	}

	entry := &AdminCacheEntry{
		Data:        data,
		ExpiresAt:   time.Now().Add(ttl),
		CreatedAt:   time.Now(),
		AccessCount: 0,
		LastAccess:  time.Now(),
		Size:        size,
		Type:        dataType,
	}

	// Remove existing entry if it exists
	if existing, exists := acs.cache[key]; exists {
		acs.adminMetrics.TotalSize -= existing.Size
	}

	acs.cache[key] = entry
	acs.adminMetrics.TotalSize += size

	return nil
}

// Delete removes an entry from admin cache
func (acs *AdminCacheService) Delete(key string) {
	acs.mu.Lock()
	defer acs.mu.Unlock()

	if entry, exists := acs.cache[key]; exists {
		acs.adminMetrics.TotalSize -= entry.Size
		delete(acs.cache, key)
	}
}

// Clear removes all entries from admin cache
func (acs *AdminCacheService) Clear() {
	acs.mu.Lock()
	defer acs.mu.Unlock()

	acs.cache = make(map[string]*AdminCacheEntry)
	acs.adminMetrics.TotalSize = 0
}

// GetMetrics returns admin cache metrics
func (acs *AdminCacheService) GetMetrics() *AdminCacheMetrics {
	acs.adminMetrics.mu.RLock()
	defer acs.adminMetrics.mu.RUnlock()

	return &AdminCacheMetrics{
		Hits:            acs.adminMetrics.Hits,
		Misses:          acs.adminMetrics.Misses,
		Evictions:       acs.adminMetrics.Evictions,
		TotalSize:       acs.adminMetrics.TotalSize,
		MaxSize:         acs.adminMetrics.MaxSize,
		AverageLoadTime: acs.adminMetrics.AverageLoadTime,
	}
}

// InvalidateByType removes all entries of a specific type
func (acs *AdminCacheService) InvalidateByType(dataType string) {
	acs.mu.Lock()
	defer acs.mu.Unlock()

	for key, entry := range acs.cache {
		if entry.Type == dataType {
			acs.adminMetrics.TotalSize -= entry.Size
			delete(acs.cache, key)
		}
	}
}

// GetWithLoader loads data using a loader function if not in cache
func (acs *AdminCacheService) GetWithLoader(key string, loader func() (interface{}, error), ttl time.Duration, dataType string) (interface{}, error) {
	// Try to get from cache first
	if data, found := acs.Get(key); found {
		return data, nil
	}

	// Load data using the loader function
	start := time.Now()
	data, err := loader()
	loadTime := time.Since(start)

	if err != nil {
		return nil, err
	}

	// Cache the loaded data
	err = acs.Set(key, data, ttl, dataType)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to cache admin data for key %s: %v\n", key, err)
	}

	// Update average load time
	acs.adminMetrics.mu.Lock()
	if acs.adminMetrics.AverageLoadTime == 0 {
		acs.adminMetrics.AverageLoadTime = loadTime
	} else {
		acs.adminMetrics.AverageLoadTime = (acs.adminMetrics.AverageLoadTime + loadTime) / 2
	}
	acs.adminMetrics.mu.Unlock()

	return data, nil
}

// Helper methods

func (acs *AdminCacheService) calculateSize(data interface{}) int64 {
	// Simple size estimation
	switch v := data.(type) {
	case string:
		return int64(len(v))
	case []byte:
		return int64(len(v))
	case map[string]interface{}:
		size := int64(0)
		for k, val := range v {
			size += int64(len(k))
			switch valStr := val.(type) {
			case string:
				size += int64(len(valStr))
			case []byte:
				size += int64(len(valStr))
			}
		}
		return size
	default:
		// Default size estimation
		return 1024 // 1KB default
	}
}

func (acs *AdminCacheService) evictEntries(neededSize int64) {
	// Simple LRU eviction - remove oldest entries first
	type entryInfo struct {
		key   string
		entry *AdminCacheEntry
	}

	var entries []entryInfo
	for key, entry := range acs.cache {
		entries = append(entries, entryInfo{key, entry})
	}

	// Sort by last access time (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].entry.LastAccess.After(entries[j].entry.LastAccess) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Remove entries until we have enough space
	freedSize := int64(0)
	for _, entryInfo := range entries {
		if acs.adminMetrics.TotalSize-freedSize+neededSize <= acs.adminMetrics.MaxSize {
			break
		}
		acs.adminMetrics.TotalSize -= entryInfo.entry.Size
		freedSize += entryInfo.entry.Size
		delete(acs.cache, entryInfo.key)
		acs.adminMetrics.Evictions++
	}
}

func (acs *AdminCacheService) cleanupRoutine(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			acs.cleanup()
		case <-acs.ctx.Done():
			return
		}
	}
}

func (acs *AdminCacheService) cleanup() {
	acs.mu.Lock()
	defer acs.mu.Unlock()

	now := time.Now()
	for key, entry := range acs.cache {
		if now.After(entry.ExpiresAt) {
			acs.adminMetrics.TotalSize -= entry.Size
			delete(acs.cache, key)
		}
	}
}

func (am *AdminCacheMetrics) recordHit() {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.Hits++
}

func (am *AdminCacheMetrics) recordMiss() {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.Misses++
}

// Admin-specific cache keys
const (
	AdminCacheKeyVideos     = "admin_videos"
	AdminCacheKeyVideoStats = "admin_video_stats"
	AdminCacheKeyCDNUsage   = "admin_cdn_usage"
	AdminCacheKeyStorage    = "admin_storage_usage"
	AdminCacheKeySyncStatus = "admin_sync_status"
)

// Admin cache TTLs (different from user cache)
var AdminCacheTTLs = struct {
	Videos     time.Duration
	Stats      time.Duration
	CDNUsage   time.Duration
	Storage    time.Duration
	SyncStatus time.Duration
}{
	Videos:     2 * time.Minute,  // Shorter TTL for admin data
	Stats:      1 * time.Minute,  // Very short TTL for stats
	CDNUsage:   5 * time.Minute,  // Medium TTL for CDN data
	Storage:    5 * time.Minute,  // Medium TTL for storage data
	SyncStatus: 30 * time.Second, // Very short TTL for sync status
}
