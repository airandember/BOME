package middleware

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// AnalyticsResilience provides circuit breaker and graceful degradation for analytics
type AnalyticsResilience struct {
	// Circuit breaker state
	isOpen          atomic.Bool
	failureCount    atomic.Int64
	successCount    atomic.Int64
	lastFailureTime atomic.Value // time.Time
	
	// Configuration
	failureThreshold int64         // Open circuit after N failures
	successThreshold int64         // Close circuit after N successes
	timeout          time.Duration // How long circuit stays open
	
	// Sampling (when degraded)
	sampleRate    atomic.Int64 // Track 1 in N events when degraded
	requestCount  atomic.Int64
	
	mu sync.RWMutex
}

// NewAnalyticsResilience creates a new resilience middleware
func NewAnalyticsResilience() *AnalyticsResilience {
	ar := &AnalyticsResilience{
		failureThreshold: 10,             // Open after 10 failures
		successThreshold: 5,              // Close after 5 successes
		timeout:          30 * time.Second, // Stay open for 30s
	}
	ar.sampleRate.Store(1) // Default: track all events
	return ar
}

// ShouldAllowRequest determines if analytics tracking should proceed
func (ar *AnalyticsResilience) ShouldAllowRequest() bool {
	// Check if circuit is open
	if ar.isOpen.Load() {
		// Check if timeout has elapsed
		lastFailure := ar.lastFailureTime.Load()
		if lastFailure != nil {
			if time.Since(lastFailure.(time.Time)) > ar.timeout {
				// Try half-open state
				log.Printf("🔄 [Analytics Resilience] Circuit half-open, trying request")
				return true
			}
		}
		
		// Circuit still open, apply sampling
		ar.requestCount.Add(1)
		sampleRate := ar.sampleRate.Load()
		if ar.requestCount.Load()%sampleRate == 0 {
			log.Printf("📊 [Analytics Resilience] Sampling 1/%d requests", sampleRate)
			return true
		}
		
		return false // Drop this request
	}
	
	return true // Circuit closed, allow request
}

// RecordSuccess records a successful analytics operation
func (ar *AnalyticsResilience) RecordSuccess() {
	ar.failureCount.Store(0) // Reset failures
	
	if ar.isOpen.Load() {
		// Circuit was open, increment success count
		successes := ar.successCount.Add(1)
		if successes >= ar.successThreshold {
			// Close circuit
			ar.isOpen.Store(false)
			ar.successCount.Store(0)
			ar.sampleRate.Store(1) // Reset to full tracking
			log.Printf("✅ [Analytics Resilience] Circuit closed after %d successes", successes)
		}
	}
}

// RecordFailure records a failed analytics operation
func (ar *AnalyticsResilience) RecordFailure() {
	failures := ar.failureCount.Add(1)
	ar.successCount.Store(0) // Reset successes
	
	if failures >= ar.failureThreshold && !ar.isOpen.Load() {
		// Open circuit
		ar.isOpen.Store(true)
		ar.lastFailureTime.Store(time.Now())
		ar.sampleRate.Store(10) // Sample 1 in 10 requests
		log.Printf("🚨 [Analytics Resilience] Circuit opened after %d failures, sampling 1/10", failures)
	}
}

// GetStatus returns the current circuit breaker status
func (ar *AnalyticsResilience) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"circuit_open":   ar.isOpen.Load(),
		"failure_count":  ar.failureCount.Load(),
		"success_count":  ar.successCount.Load(),
		"sample_rate":    ar.sampleRate.Load(),
		"request_count":  ar.requestCount.Load(),
	}
}

// Middleware wraps Gin handlers with resilience logic
func (ar *AnalyticsResilience) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Store resilience instance in context for use by handlers
		c.Set("analytics_resilience", ar)
		c.Next()
	}
}

// GetFromContext retrieves the resilience instance from Gin context
func GetFromContext(c *gin.Context) *AnalyticsResilience {
	if ar, exists := c.Get("analytics_resilience"); exists {
		return ar.(*AnalyticsResilience)
	}
	return nil
}

