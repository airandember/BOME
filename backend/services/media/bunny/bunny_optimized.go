package stripe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Global optimized service instance for metrics access
var globalOptimizedBunnyService *OptimizedBunnyService
var globalServiceMutex sync.RWMutex

// SetGlobalOptimizedBunnyService sets the global optimized service instance
func SetGlobalOptimizedBunnyService(service *OptimizedBunnyService) {
	globalServiceMutex.Lock()
	defer globalServiceMutex.Unlock()
	globalOptimizedBunnyService = service
}

// GetGlobalOptimizedBunnyService returns the global optimized service instance
func GetGlobalOptimizedBunnyService() *OptimizedBunnyService {
	globalServiceMutex.RLock()
	defer globalServiceMutex.RUnlock()
	return globalOptimizedBunnyService
}

// OptimizedBunnyService with advanced caching and connection pooling
type OptimizedBunnyService struct {
	*BunnyService // Embed existing service

	// Advanced caching
	l1Cache    *sync.Map // In-memory L1 cache
	l2Cache    *sync.Map // Persistent L2 cache
	cacheStats *CacheStats

	// Connection pooling
	httpClient *http.Client

	// Rate limiting
	rateLimiter *BunnyRateLimiter

	// Metrics
	metrics *ServiceMetrics
}

// Ensure OptimizedBunnyService implements the same interface as BunnyService
var _ BunnyServiceInterface = (*OptimizedBunnyService)(nil)

// BunnyServiceInterface defines the interface that both services must implement
type BunnyServiceInterface interface {
	GetVideo(videoID string) (*BunnyVideo, error)
	GetVideoPlayData(videoID string) (*VideoPlayData, error)
	GetStreamURL(videoID string) string
	GetThumbnailURL(videoID string) string
	GetThumbnailURLWithFilename(videoID, filename string) string
	GetIframeURL(videoID string) string
	GetDirectPlayURL(videoID string) string
	GetStreamLibrary() string
	GetRegion() string
	GetStreamAPIKey() string
	GetCollections(page int, perPage int) (*BunnyCollectionsResponse, error)
	GetCollection(collectionID string) (*BunnyCollection, error)
	GetVideosByCollection(collectionID string, page, itemsPerPage int) ([]BunnyVideo, int, error)
	GetCDNHostname(videoID string) string
}

type CacheStats struct {
	Hits      int64
	Misses    int64
	Evictions int64
	mu        sync.RWMutex
}

type BunnyRateLimiter struct {
	requests chan struct{}
	ticker   *time.Ticker
	mu       sync.Mutex
}

type ServiceMetrics struct {
	RequestCount    int64
	ErrorCount      int64
	AvgResponseTime time.Duration
	mu              sync.RWMutex
}

// NewOptimizedBunnyService creates an optimized Bunny service
func NewOptimizedBunnyService(original *BunnyService) *OptimizedBunnyService {
	// Create optimized HTTP client
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableCompression:  false,
		ForceAttemptHTTP2:   true,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	// Create rate limiter (100 requests per minute)
	rateLimiter := &BunnyRateLimiter{
		requests: make(chan struct{}, 100),
		ticker:   time.NewTicker(time.Minute),
	}

	// Fill rate limiter initially
	for i := 0; i < 100; i++ {
		rateLimiter.requests <- struct{}{}
	}

	// Start rate limiter refill
	go func() {
		for range rateLimiter.ticker.C {
			// Refill rate limiter
			for i := 0; i < 100; i++ {
				select {
				case rateLimiter.requests <- struct{}{}:
				default:
					// Channel full, skip
				}
			}
		}
	}()

	return &OptimizedBunnyService{
		BunnyService: original,
		l1Cache:      &sync.Map{},
		l2Cache:      &sync.Map{},
		cacheStats:   &CacheStats{},
		httpClient:   httpClient,
		rateLimiter:  rateLimiter,
		metrics:      &ServiceMetrics{},
	}
}

// Enhanced cache entry with metadata
type EnhancedCacheEntry struct {
	Data        interface{}
	ExpiresAt   time.Time
	CreatedAt   time.Time
	AccessCount int64
	LastAccess  time.Time
	Size        int64
}

// GetVideoWithCache retrieves video with multi-level caching
func (obs *OptimizedBunnyService) GetVideoWithCache(ctx context.Context, videoID string) (*BunnyVideo, error) {
	start := time.Now()
	defer func() {
		obs.updateMetrics(time.Since(start))
	}()

	// Check L1 cache first
	if cached, ok := obs.l1Cache.Load(videoID); ok {
		if entry, ok := cached.(*EnhancedCacheEntry); ok {
			if time.Now().Before(entry.ExpiresAt) {
				obs.cacheStats.recordHit()
				entry.AccessCount++
				entry.LastAccess = time.Now()
				return entry.Data.(*BunnyVideo), nil
			}
			// Remove expired entry
			obs.l1Cache.Delete(videoID)
		}
	}

	// Check L2 cache
	if cached, ok := obs.l2Cache.Load(videoID); ok {
		if entry, ok := cached.(*EnhancedCacheEntry); ok {
			if time.Now().Before(entry.ExpiresAt) {
				obs.cacheStats.recordHit()
				// Promote to L1 cache
				obs.l1Cache.Store(videoID, entry)
				return entry.Data.(*BunnyVideo), nil
			}
			obs.l2Cache.Delete(videoID)
		}
	}

	obs.cacheStats.recordMiss()

	// Rate limit request
	select {
	case <-obs.rateLimiter.requests:
		// Proceed with request
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Fetch from Bunny.net
	video, err := obs.fetchVideoFromBunny(ctx, videoID)
	if err != nil {
		obs.metrics.recordError()
		return nil, err
	}

	// Cache the result
	entry := &EnhancedCacheEntry{
		Data:        video,
		ExpiresAt:   time.Now().Add(5 * time.Minute),
		CreatedAt:   time.Now(),
		AccessCount: 1,
		LastAccess:  time.Now(),
		Size:        int64(len(video.Title) + len(*video.Description)), // Approximate size
	}

	obs.l1Cache.Store(videoID, entry)
	obs.l2Cache.Store(videoID, entry)

	return video, nil
}

// Batch video fetching with concurrency control
func (obs *OptimizedBunnyService) GetVideosWithBatch(ctx context.Context, videoIDs []string) ([]*BunnyVideo, error) {
	const maxConcurrency = 10
	semaphore := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	results := make([]*BunnyVideo, len(videoIDs))
	errors := make([]error, len(videoIDs))

	for i, videoID := range videoIDs {
		wg.Add(1)
		go func(index int, id string) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				errors[index] = ctx.Err()
				return
			}

			video, err := obs.GetVideoWithCache(ctx, id)
			if err != nil {
				errors[index] = err
				return
			}
			results[index] = video
		}(i, videoID)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errors {
		if err != nil {
			return nil, fmt.Errorf("batch fetch failed: %w", err)
		}
	}

	return results, nil
}

// fetchVideoFromBunny performs the actual API call
func (obs *OptimizedBunnyService) fetchVideoFromBunny(ctx context.Context, videoID string) (*BunnyVideo, error) {
	url := fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos/%s",
		obs.BunnyService.GetStreamLibrary(), videoID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("AccessKey", obs.BunnyService.GetStreamAPIKey())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "BOME-Optimized/1.0")

	resp, err := obs.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var video BunnyVideo
	if err := json.NewDecoder(resp.Body).Decode(&video); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return &video, nil
}

// Cache management methods
func (cs *CacheStats) recordHit() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.Hits++
}

func (cs *CacheStats) recordMiss() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.Misses++
}

func (cs *CacheStats) GetStats() (hits, misses, evictions int64) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.Hits, cs.Misses, cs.Evictions
}

// Metrics methods
func (sm *ServiceMetrics) recordError() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.ErrorCount++
}

func (obs *OptimizedBunnyService) updateMetrics(duration time.Duration) {
	obs.metrics.mu.Lock()
	defer obs.metrics.mu.Unlock()

	obs.metrics.RequestCount++
	// Simple moving average
	obs.metrics.AvgResponseTime = (obs.metrics.AvgResponseTime + duration) / 2
}

// GetMetrics returns service performance metrics
func (obs *OptimizedBunnyService) GetMetrics() map[string]interface{} {
	obs.metrics.mu.RLock()
	defer obs.metrics.mu.RUnlock()

	hits, misses, evictions := obs.cacheStats.GetStats()

	return map[string]interface{}{
		"requests_total":    obs.metrics.RequestCount,
		"errors_total":      obs.metrics.ErrorCount,
		"avg_response_time": obs.metrics.AvgResponseTime,
		"cache_hits":        hits,
		"cache_misses":      misses,
		"cache_evictions":   evictions,
		"cache_hit_ratio":   float64(hits) / float64(hits+misses),
	}
}

// CleanupExpiredCache removes expired entries from cache
func (obs *OptimizedBunnyService) CleanupExpiredCache() {
	now := time.Now()

	// Cleanup L1 cache
	obs.l1Cache.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*EnhancedCacheEntry); ok {
			if now.After(entry.ExpiresAt) {
				obs.l1Cache.Delete(key)
				obs.cacheStats.mu.Lock()
				obs.cacheStats.Evictions++
				obs.cacheStats.mu.Unlock()
			}
		}
		return true
	})

	// Cleanup L2 cache
	obs.l2Cache.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*EnhancedCacheEntry); ok {
			if now.After(entry.ExpiresAt) {
				obs.l2Cache.Delete(key)
			}
		}
		return true
	})
}

// GetBunnyService returns the embedded BunnyService for compatibility
func (obs *OptimizedBunnyService) GetBunnyService() *BunnyService {
	return obs.BunnyService
}

// Override GetVideo to use the optimized version
func (obs *OptimizedBunnyService) GetVideo(videoID string) (*BunnyVideo, error) {
	ctx := context.Background()
	return obs.GetVideoWithCache(ctx, videoID)
}

// StartBackgroundTasks starts background maintenance tasks
func (obs *OptimizedBunnyService) StartBackgroundTasks() {
	// Cache cleanup every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			obs.CleanupExpiredCache()
		}
	}()
}

