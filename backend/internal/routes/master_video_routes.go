package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"bome-backend/internal/database"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupMasterVideoRoutes sets up routes for master video list management
func SetupMasterVideoRoutes(router *gin.RouterGroup, db *database.DB, bunnyService *services.BunnyService) {
	fmt.Printf("Setting up master video routes...\n")
	masterVideoService := services.NewMasterVideoSyncService(db, bunnyService)

	// Master video list routes
	masterVideos := router.Group("/master-videos")
	{
		fmt.Printf("Registered master video routes:\n")
		fmt.Printf("  GET /master-videos\n")
		fmt.Printf("  GET /master-videos/:id\n")
		fmt.Printf("  PUT /master-videos/:id/toggle-status\n")
		fmt.Printf("  PUT /master-videos/:id\n")
		fmt.Printf("  DELETE /master-videos/:id\n")
		// Get all master videos with filtering and pagination
		masterVideos.GET("", func(c *gin.Context) {
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			category := c.Query("category")
			status := c.Query("status")
			syncStatus := c.Query("sync_status")
			vidStatus := c.Query("vid_status")
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
				videos, err = db.GetMasterVideos(limit, offset, category, status, syncStatus, vidStatus, sortField, sortDirection)
				// Get total count for filtered results
				totalCount, err = db.GetMasterVideoCount(category, status, syncStatus, vidStatus)
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

		// Toggle video status (vid_status) - No page reload
		masterVideos.PUT("/:id/toggle-status", func(c *gin.Context) {
			fmt.Printf("DEBUG: Toggle status route hit for video ID: %s\n", c.Param("id"))

			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
				return
			}

			var toggleData struct {
				Vid_Status bool `json:"vid_status"`
			}
			if err := c.ShouldBindJSON(&toggleData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
				return
			}

			fmt.Printf("DEBUG: Toggle data received: %+v\n", toggleData)

			// Get existing video
			existingVideo, err := db.GetMasterVideoByID(id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
				return
			}

			fmt.Printf("DEBUG: Existing video status: %t, new status: %t\n", existingVideo.Vid_Status, toggleData.Vid_Status)

			// Update only the vid_status field
			existingVideo.Vid_Status = toggleData.Vid_Status

			err = db.UpdateMasterVideo(existingVideo)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to update video status",
					"details": err.Error(),
				})
				return
			}

			statusText := "activated"
			if !toggleData.Vid_Status {
				statusText = "deactivated"
			}

			fmt.Printf("DEBUG: Video %s successfully\n", statusText)

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": fmt.Sprintf("Video %s successfully", statusText),
				"video":   existingVideo,
			})
		})

		// Update master video
		masterVideos.PUT("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
				return
			}

			// Read the raw request body to check which fields are provided
			body, err := c.GetRawData()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
				return
			}

			// Parse as map to see which fields are provided
			var rawData map[string]interface{}
			if err := json.Unmarshal(body, &rawData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
				return
			}

			// Parse into the struct
			var updateData database.MasterVideo
			if err := json.Unmarshal(body, &updateData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
				return
			}

			// Get existing video
			existingVideo, err := db.GetMasterVideoByID(id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
				return
			}

			// Debug: Log what we received
			fmt.Printf("DEBUG: Update request for video %d\n", id)
			fmt.Printf("DEBUG: Raw JSON data: %+v\n", rawData)
			fmt.Printf("DEBUG: Received updateData: %+v\n", updateData)

			// Track if we're updating any Bunny.net fields
			hasBunnyFieldChanges := false

			// Update fields and check if any Bunny.net fields are being updated
			if updateData.Title != "" {
				fmt.Printf("DEBUG: Updating Title (Bunny field) - hasBunnyFieldChanges = true\n")
				existingVideo.Title = updateData.Title
				hasBunnyFieldChanges = true
			}
			if updateData.Description != "" {
				fmt.Printf("DEBUG: Updating Description (Bunny field) - hasBunnyFieldChanges = true\n")
				existingVideo.Description = updateData.Description
				hasBunnyFieldChanges = true
			}
			if updateData.Category != "" {
				fmt.Printf("DEBUG: Updating Category (Bunny field) - hasBunnyFieldChanges = true\n")
				existingVideo.Category = updateData.Category
				hasBunnyFieldChanges = true
			}
			if len(updateData.Tags) > 0 {
				fmt.Printf("DEBUG: Updating Tags (Bunny field) - hasBunnyFieldChanges = true\n")
				existingVideo.Tags = updateData.Tags
				hasBunnyFieldChanges = true
			}
			if updateData.Duration > 0 {
				fmt.Printf("DEBUG: Updating Duration (Bunny field) - hasBunnyFieldChanges = true\n")
				existingVideo.Duration = updateData.Duration
				hasBunnyFieldChanges = true
			}
			if updateData.ThumbnailURL != "" {
				fmt.Printf("DEBUG: Updating ThumbnailURL (local field) - no sync change\n")
				existingVideo.ThumbnailURL = updateData.ThumbnailURL
			}
			if updateData.VideoURL != "" {
				fmt.Printf("DEBUG: Updating VideoURL (local field) - no sync change\n")
				existingVideo.VideoURL = updateData.VideoURL
			}
			if updateData.IframeSrc != "" {
				fmt.Printf("DEBUG: Updating IframeSrc (local field) - no sync change\n")
				existingVideo.IframeSrc = updateData.IframeSrc
			}
			if updateData.PlaybackURL != "" {
				fmt.Printf("DEBUG: Updating PlaybackURL (local field) - no sync change\n")
				existingVideo.PlaybackURL = updateData.PlaybackURL
			}
			if updateData.Status != "" {
				fmt.Printf("DEBUG: Updating Status (Bunny field) - hasBunnyFieldChanges = true\n")
				existingVideo.Status = updateData.Status
				hasBunnyFieldChanges = true
			}
			// Check if Views was actually provided in the JSON
			if _, viewsProvided := rawData["views"]; viewsProvided {
				fmt.Printf("DEBUG: Updating Views (Bunny field) - hasBunnyFieldChanges = true\n")
				existingVideo.Views = updateData.Views
				hasBunnyFieldChanges = true
			}
			// Check if Likes was actually provided in the JSON
			if _, likesProvided := rawData["likes"]; likesProvided {
				fmt.Printf("DEBUG: Updating Likes (local field) - no sync change\n")
				existingVideo.Likes = updateData.Likes
			}
			// Check if IsPublic was actually provided in the JSON
			if _, isPublicProvided := rawData["is_public"]; isPublicProvided {
				fmt.Printf("DEBUG: Updating IsPublic (Bunny field) - hasBunnyFieldChanges = true\n")
				existingVideo.IsPublic = updateData.IsPublic
				hasBunnyFieldChanges = true
			}
			if updateData.CollectionID != "" {
				fmt.Printf("DEBUG: Updating CollectionID (Bunny field) - hasBunnyFieldChanges = true\n")
				existingVideo.CollectionID = updateData.CollectionID
				hasBunnyFieldChanges = true
			}
			if updateData.SyncStatus != "" {
				fmt.Printf("DEBUG: Updating SyncStatus (local field) - no sync change\n")
				existingVideo.SyncStatus = updateData.SyncStatus
			}
			if updateData.SyncNotes != "" {
				fmt.Printf("DEBUG: Updating SyncNotes (local field) - no sync change\n")
				existingVideo.SyncNotes = updateData.SyncNotes
			}
			// Vid_Status is local-only, never affects sync status
			if _, vidStatusProvided := rawData["vid_status"]; vidStatusProvided {
				fmt.Printf("DEBUG: Updating Vid_Status (local field) - no sync change\n")
				existingVideo.Vid_Status = updateData.Vid_Status
			}

			fmt.Printf("DEBUG: Final hasBunnyFieldChanges = %t\n", hasBunnyFieldChanges)

			// Only update sync status if we're updating fields that exist in Bunny.net
			if hasBunnyFieldChanges {
				fmt.Printf("DEBUG: Setting sync status to 'needs_attention' due to non-local changes\n")
				existingVideo.SyncStatus = "needs_attention"
				existingVideo.SyncNotes = "Updated by admin"
			}

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
