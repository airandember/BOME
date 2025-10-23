package health

import (
	"bome-backend/infrastructure/database"
	emailService "bome-backend/services/communication/email"
	stripeService "bome-backend/services/payment/stripe"
	bunnyService "bome-backend/video-streaming/services"
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Services  map[string]ServiceInfo `json:"services"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
}

type ServiceInfo struct {
	Status       string                 `json:"status"`
	ResponseTime float64                `json:"response_time_ms"`
	Message      string                 `json:"message,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

var serverStartTime = time.Now()

// HealthCheckHandler returns comprehensive health status of all services
func HealthCheckHandler(
	db *database.DB,
	email *emailService.EmailService,
	stripe *stripeService.StripeService,
	bunny *bunnyService.BunnyService,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		health := HealthStatus{
			Status:    "healthy",
			Timestamp: time.Now().Format(time.RFC3339),
			Services:  make(map[string]ServiceInfo),
			Version:   "1.0.0",
			Uptime:    time.Since(serverStartTime).String(),
		}

		overallHealthy := true

		// Check PostgreSQL Database
		start := time.Now()
		if db != nil && db.DB != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			if err := db.PingContext(ctx); err != nil {
				health.Services["database"] = ServiceInfo{
					Status:       "unhealthy",
					ResponseTime: time.Since(start).Seconds() * 1000,
					Message:      "Database ping failed: " + err.Error(),
				}
				overallHealthy = false
			} else {
				// Get connection stats
				stats := db.Stats()
				health.Services["database"] = ServiceInfo{
					Status:       "healthy",
					ResponseTime: time.Since(start).Seconds() * 1000,
					Details: map[string]interface{}{
						"open_connections": stats.OpenConnections,
						"in_use":           stats.InUse,
						"idle":             stats.Idle,
					},
				}
			}
		} else {
			health.Services["database"] = ServiceInfo{
				Status:  "unavailable",
				Message: "Database not configured",
			}
			overallHealthy = false
		}

		// Check Redis
		start = time.Now()
		if db != nil && db.Redis != nil && db.Redis.Client != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			if err := db.Redis.Client.Ping(ctx).Err(); err != nil {
				health.Services["redis"] = ServiceInfo{
					Status:       "unhealthy",
					ResponseTime: time.Since(start).Seconds() * 1000,
					Message:      "Redis ping failed: " + err.Error(),
				}
				// Redis is optional, don't mark overall as unhealthy
			} else {
				health.Services["redis"] = ServiceInfo{
					Status:       "healthy",
					ResponseTime: time.Since(start).Seconds() * 1000,
				}
			}
		} else {
			health.Services["redis"] = ServiceInfo{
				Status:  "disabled",
				Message: "Redis not configured (development mode)",
			}
		}

		// Check Stripe Service
		start = time.Now()
		if stripe != nil {
			if stripe.IsEnabled() {
				health.Services["stripe"] = ServiceInfo{
					Status:       "healthy",
					ResponseTime: time.Since(start).Seconds() * 1000,
					Message:      "Stripe service enabled",
				}
			} else {
				health.Services["stripe"] = ServiceInfo{
					Status:  "disabled",
					Message: "Stripe service not configured",
				}
			}
		} else {
			health.Services["stripe"] = ServiceInfo{
				Status:  "unavailable",
				Message: "Stripe service not initialized",
			}
		}

		// Check Bunny Service
		start = time.Now()
		if bunny != nil {
			health.Services["bunny"] = ServiceInfo{
				Status:       "healthy",
				ResponseTime: time.Since(start).Seconds() * 1000,
				Message:      "Bunny.net service initialized",
			}
		} else {
			health.Services["bunny"] = ServiceInfo{
				Status:  "unavailable",
				Message: "Bunny.net service not configured",
			}
		}

		// Check Email Service
		start = time.Now()
		if email != nil {
			if email.IsConfigured() {
				health.Services["email"] = ServiceInfo{
					Status:       "healthy",
					ResponseTime: time.Since(start).Seconds() * 1000,
					Message:      "Email service configured",
				}
			} else {
				health.Services["email"] = ServiceInfo{
					Status:  "degraded",
					Message: "Email service partially configured",
				}
			}
		} else {
			health.Services["email"] = ServiceInfo{
				Status:  "unavailable",
				Message: "Email service not initialized",
			}
		}

		// Determine overall status
		if !overallHealthy {
			health.Status = "unhealthy"
		} else {
			// Check if any critical service is degraded
			for _, service := range health.Services {
				if service.Status == "unhealthy" {
					health.Status = "unhealthy"
					break
				} else if service.Status == "degraded" {
					if health.Status != "unhealthy" {
						health.Status = "degraded"
					}
				}
			}
		}

		statusCode := 200
		if health.Status == "unhealthy" {
			statusCode = 503
		}

		c.JSON(statusCode, health)
	}
}

// LivenessProbe for Kubernetes/Docker health checks
func LivenessProbe() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "alive",
			"uptime": time.Since(serverStartTime).String(),
		})
	}
}

// ReadinessProbe checks if all critical services are ready
func ReadinessProbe(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if database is ready
		if db != nil && db.DB != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			if err := db.PingContext(ctx); err != nil {
				c.JSON(503, gin.H{
					"status": "not ready",
					"reason": "database not responding",
				})
				return
			}
		} else {
			c.JSON(503, gin.H{
				"status": "not ready",
				"reason": "database not configured",
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ready",
		})
	}
}
