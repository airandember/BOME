package routes

import (
	"net/http"
	"time"

	"bome-backend/internal/database"

	"github.com/gin-gonic/gin"
)

// DatabaseMonitoringResponse represents the complete database monitoring data
type DatabaseMonitoringResponse struct {
	Timestamp        time.Time                     `json:"timestamp"`
	ConnectionPool   *database.ConnectionPoolStats `json:"connection_pool"`
	PostgreSQLHealth *PostgreSQLHealthCheck        `json:"postgresql_health"`
	Recommendations  []string                      `json:"recommendations"`
}

// PostgreSQLHealthCheck represents PostgreSQL server health metrics
type PostgreSQLHealthCheck struct {
	Status            string        `json:"status"`
	ResponseTime      time.Duration `json:"response_time_ms"`
	ActiveConnections int           `json:"active_connections,omitempty"`
	MaxConnections    int           `json:"max_connections,omitempty"`
	DatabaseSize      string        `json:"database_size,omitempty"`
	Error             string        `json:"error,omitempty"`
}

// RegisterDatabaseMonitoringRoutes registers database monitoring endpoints
func RegisterDatabaseMonitoringRoutes(router *gin.RouterGroup, db *database.DB) {
	monitoring := router.Group("/monitoring")
	{
		// Connection pool statistics endpoint
		monitoring.GET("/db/pool", func(c *gin.Context) {
			getConnectionPoolStats(c, db)
		})

		// Complete database health check endpoint
		monitoring.GET("/db/health", func(c *gin.Context) {
			getDatabaseHealth(c, db)
		})

		// Lightweight ping endpoint for load balancers
		monitoring.GET("/db/ping", func(c *gin.Context) {
			pingDatabase(c, db)
		})
	}
}

// getConnectionPoolStats returns detailed connection pool statistics
func getConnectionPoolStats(c *gin.Context, db *database.DB) {
	stats := db.GetConnectionPoolStats()

	// Log the stats for server-side monitoring
	db.LogConnectionPoolStats()

	c.JSON(http.StatusOK, gin.H{
		"timestamp": time.Now(),
		"status":    "success",
		"data":      stats,
	})
}

// getDatabaseHealth performs comprehensive database health check
func getDatabaseHealth(c *gin.Context, db *database.DB) {
	start := time.Now()

	// Get connection pool stats
	poolStats := db.GetConnectionPoolStats()

	// Perform PostgreSQL health check
	pgHealth := checkPostgreSQLHealth(db)

	// Generate recommendations based on current state
	recommendations := generateRecommendations(poolStats, pgHealth)

	response := DatabaseMonitoringResponse{
		Timestamp:        time.Now(),
		ConnectionPool:   poolStats,
		PostgreSQLHealth: pgHealth,
		Recommendations:  recommendations,
	}

	// Log comprehensive health status
	db.LogConnectionPoolStats()

	// Determine overall HTTP status
	httpStatus := http.StatusOK
	if pgHealth.Status == "unhealthy" || poolStats.UtilizationPercentage > 90 {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"timestamp":                time.Now(),
		"status":                   "success",
		"data":                     response,
		"health_check_duration_ms": time.Since(start).Milliseconds(),
	})
}

// pingDatabase provides a lightweight health check for load balancers
func pingDatabase(c *gin.Context, db *database.DB) {
	start := time.Now()

	err := db.Ping()
	duration := time.Since(start)

	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":           "error",
			"message":          "Database ping failed",
			"error":            err.Error(),
			"response_time_ms": duration.Milliseconds(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           "healthy",
		"message":          "Database is responsive",
		"response_time_ms": duration.Milliseconds(),
	})
}

// checkPostgreSQLHealth performs detailed PostgreSQL server health checks
func checkPostgreSQLHealth(db *database.DB) *PostgreSQLHealthCheck {
	start := time.Now()

	health := &PostgreSQLHealthCheck{
		Status: "healthy",
	}

	// Test basic connectivity
	if err := db.Ping(); err != nil {
		health.Status = "unhealthy"
		health.Error = err.Error()
		health.ResponseTime = time.Since(start)
		return health
	}

	health.ResponseTime = time.Since(start)

	// Try to get additional PostgreSQL metrics (optional, non-critical)
	// These queries might fail due to permissions, but that's okay

	// Get active connections count
	var activeConns int
	if err := db.QueryRow("SELECT count(*) FROM pg_stat_activity WHERE state = 'active'").Scan(&activeConns); err == nil {
		health.ActiveConnections = activeConns
	}

	// Get max connections setting
	var maxConns int
	if err := db.QueryRow("SHOW max_connections").Scan(&maxConns); err == nil {
		health.MaxConnections = maxConns
	}

	// Get database size
	var dbSize string
	if err := db.QueryRow("SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&dbSize); err == nil {
		health.DatabaseSize = dbSize
	}

	return health
}

// generateRecommendations provides actionable recommendations based on current metrics
func generateRecommendations(poolStats *database.ConnectionPoolStats, pgHealth *PostgreSQLHealthCheck) []string {
	var recommendations []string

	// Connection pool recommendations
	if poolStats.UtilizationPercentage > 90 {
		recommendations = append(recommendations, "🔴 CRITICAL: Connection pool utilization is very high (>90%). Consider increasing max_connections or implementing connection pooling middleware like PgBouncer.")
	} else if poolStats.UtilizationPercentage > 75 {
		recommendations = append(recommendations, "🟡 WARNING: Connection pool utilization is high (>75%). Monitor closely and consider scaling database connections.")
	}

	if poolStats.WaitCount > 100 {
		recommendations = append(recommendations, "🟡 WARNING: High connection wait count detected. Applications are waiting for available connections. Consider optimizing query performance or increasing connection limits.")
	}

	if poolStats.WaitDuration > time.Second {
		recommendations = append(recommendations, "🟡 WARNING: Applications are experiencing significant wait times for database connections. Review connection pool configuration.")
	}

	// PostgreSQL server recommendations
	if pgHealth.Status == "unhealthy" {
		recommendations = append(recommendations, "🔴 CRITICAL: PostgreSQL server is not responding. Check database server status immediately.")
	}

	if pgHealth.ResponseTime > 100*time.Millisecond {
		recommendations = append(recommendations, "🟠 CAUTION: Database response time is elevated (>100ms). Check server load and query performance.")
	}

	if pgHealth.MaxConnections > 0 && pgHealth.ActiveConnections > 0 {
		serverUtilization := float64(pgHealth.ActiveConnections) / float64(pgHealth.MaxConnections) * 100
		if serverUtilization > 80 {
			recommendations = append(recommendations, "🟡 WARNING: PostgreSQL server connection utilization is high (>80%). Consider increasing max_connections or implementing connection pooling.")
		}
	}

	// Positive recommendations
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "🟢 HEALTHY: Database connection pool and server are operating normally. No immediate action required.")
	}

	return recommendations
}
