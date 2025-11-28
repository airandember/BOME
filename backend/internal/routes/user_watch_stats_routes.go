package routes

import (
	"bome-backend/internal/database"
	"bome-backend/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RegisterUserWatchStatsRoutes registers all user watch statistics routes
func RegisterUserWatchStatsRoutes(router *gin.RouterGroup, db *database.DB) {
	service := services.NewUserWatchStatsService(db)

	// User stats routes (requires authentication)
	stats := router.Group("/user/stats")
	{
		// Get comprehensive watch statistics
		stats.GET("", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
				return
			}

			watchStats, err := service.GetUserWatchStats(userID.(int))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch watch statistics"})
				return
			}

			c.JSON(http.StatusOK, watchStats)
		})

		// Get top watched videos
		stats.GET("/top-videos", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
				return
			}

			limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
			if err != nil || limit <= 0 {
				limit = 10
			}

			topVideos, err := service.GetTopVideos(userID.(int), limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch top videos"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"videos": topVideos, "count": len(topVideos)})
		})

		// Get watching sessions
		stats.GET("/sessions", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
				return
			}

			limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
			if err != nil || limit <= 0 {
				limit = 10
			}

			sessions, err := service.GetWatchingSessions(userID.(int), limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sessions"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"sessions": sessions, "count": len(sessions)})
		})
	}
}

