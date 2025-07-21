package routes

import (
	"net/http"
	"strconv"

	"bome-backend/internal/database"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupMasterVideoRoutes sets up routes for master video list management
func SetupMasterVideoRoutes(router *gin.RouterGroup, db *database.DB, bunnyService *services.BunnyService) {
	masterVideoService := services.NewMasterVideoSyncService(db, bunnyService)

	// Master video list routes
	masterVideos := router.Group("/master-videos")
	{
		// Get all master videos with filtering and pagination
		masterVideos.GET("", func(c *gin.Context) {
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			category := c.Query("category")
			status := c.Query("status")
			syncStatus := c.Query("sync_status")
			search := c.Query("search")
			sortField := c.DefaultQuery("sort_field", "id")
			sortDirection := c.DefaultQuery("sort_direction", "desc")

			offset := (page - 1) * limit

			var videos []*database.MasterVideo
			var err error
			var totalCount int

			if search != "" {
				videos, err = db.SearchMasterVideos(search, limit, offset, sortField, sortDirection)
				// For search, we need to get total count separately
				totalCount, err = db.GetMasterVideoSearchCount(search)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to get search count",
						"details": err.Error(),
					})
					return
				}
			} else {
				videos, err = db.GetMasterVideos(limit, offset, category, status, syncStatus, sortField, sortDirection)
				// Get total count for filtered results
				totalCount, err = db.GetMasterVideoCount(category, status, syncStatus)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to get video count",
						"details": err.Error(),
					})
					return
				}
			}

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to fetch master videos",
					"details": err.Error(),
				})
				return
			}

			totalPages := (totalCount + limit - 1) / limit

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"videos":  videos,
				"pagination": gin.H{
					"current_page": page,
					"per_page":     limit,
					"total":        totalCount,
					"total_pages":  totalPages,
					"has_more":     page < totalPages,
				},
			})
		})

		// Get master video by ID
		masterVideos.GET("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
				return
			}

			video, err := db.GetMasterVideoByID(id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"video":   video,
			})
		})

		// Update master video
		masterVideos.PUT("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
				return
			}

			var updateData database.MasterVideo
			if err := c.ShouldBindJSON(&updateData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
				return
			}

			// Get existing video
			existingVideo, err := db.GetMasterVideoByID(id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
				return
			}

			// Update editable fields
			if updateData.Title != "" {
				existingVideo.Title = updateData.Title
			}
			if updateData.Description != "" {
				existingVideo.Description = updateData.Description
			}
			if updateData.Category != "" {
				existingVideo.Category = updateData.Category
			}
			if len(updateData.Tags) > 0 {
				existingVideo.Tags = updateData.Tags
			}
			if updateData.Duration > 0 {
				existingVideo.Duration = updateData.Duration
			}
			if updateData.ThumbnailURL != "" {
				existingVideo.ThumbnailURL = updateData.ThumbnailURL
			}
			if updateData.VideoURL != "" {
				existingVideo.VideoURL = updateData.VideoURL
			}
			if updateData.IframeSrc != "" {
				existingVideo.IframeSrc = updateData.IframeSrc
			}
			if updateData.PlaybackURL != "" {
				existingVideo.PlaybackURL = updateData.PlaybackURL
			}
			if updateData.Status != "" {
				existingVideo.Status = updateData.Status
			}
			if updateData.Views >= 0 {
				existingVideo.Views = updateData.Views
			}
			if updateData.Likes >= 0 {
				existingVideo.Likes = updateData.Likes
			}
			// IsPublic can be false, so we need to check if it's explicitly set
			if updateData.IsPublic != existingVideo.IsPublic {
				existingVideo.IsPublic = updateData.IsPublic
			}
			if updateData.CollectionID != "" {
				existingVideo.CollectionID = updateData.CollectionID
			}
			if updateData.SyncStatus != "" {
				existingVideo.SyncStatus = updateData.SyncStatus
			}
			if updateData.SyncNotes != "" {
				existingVideo.SyncNotes = updateData.SyncNotes
			}

			// Update sync status to indicate manual edit
			existingVideo.SyncStatus = "needs_attention"
			existingVideo.SyncNotes = "Updated by admin"

			err = db.UpdateMasterVideo(existingVideo)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to update video",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Video updated successfully",
				"video":   existingVideo,
			})
		})

		// Delete master video
		masterVideos.DELETE("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
				return
			}

			err = db.DeleteMasterVideo(id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to delete video",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Video deleted successfully",
			})
		})
	}

	// Sync routes
	sync := router.Group("/sync")
	{
		// Sync from Bunny.net to master list
		sync.POST("/from-bunny", func(c *gin.Context) {
			result, err := masterVideoService.SyncFromBunny()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Sync failed",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Sync completed",
				"result":  result,
			})
		})

		// Sync from master list to Bunny.net
		sync.POST("/to-bunny", func(c *gin.Context) {
			result, err := masterVideoService.SyncToBunny()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Sync failed",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Sync completed",
				"result":  result,
			})
		})

		// Check for conflicts
		sync.GET("/conflicts", func(c *gin.Context) {
			result, err := masterVideoService.CheckConflicts()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Conflict check failed",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"result":  result,
			})
		})
	}

	// Conflict resolution routes
	conflicts := router.Group("/conflicts")
	{
		// Get all conflicts
		conflicts.GET("", func(c *gin.Context) {
			conflicts, err := db.GetSyncConflicts(nil)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to fetch conflicts",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success":   true,
				"conflicts": conflicts,
			})
		})

		// Resolve a specific conflict
		conflicts.POST("/:id/resolve", func(c *gin.Context) {
			conflictID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conflict ID"})
				return
			}

			var resolution services.ConflictResolution
			if err := c.ShouldBindJSON(&resolution); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resolution data"})
				return
			}

			// Get user ID from context (assuming it's set by auth middleware)
			userID := 1 // Default to system user, should get from context

			err = masterVideoService.ResolveConflict(conflictID, &resolution, userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to resolve conflict",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Conflict resolved successfully",
			})
		})
	}

	// Stats routes
	stats := router.Group("/stats")
	{
		// Get master video statistics
		stats.GET("/master-videos", func(c *gin.Context) {
			stats, err := db.GetMasterVideoStats()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to fetch statistics",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"stats":   stats,
			})
		})
	}
}
