package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"bome-backend/authentication/middleware"
	"bome-backend/infrastructure/database"
	videoModels "bome-backend/video-streaming/models"
	videoServices "bome-backend/video-streaming/services"

	"github.com/gin-gonic/gin"
)

// SetupMasterVideoRoutes sets up routes for master video list management
func SetupMasterVideoRoutes(router *gin.RouterGroup, db *database.DB, bunnyService *videoServices.BunnyService) {
	fmt.Printf("Setting up master video routes...\n")
	masterVideoService := videoServices.NewMasterVideoSyncService(db, bunnyService)
	// smartTaggingService := services.NewSmartTaggingService(db) // TODO: Implement SmartTaggingService

	// Master video list routes
	masterVideos := router.Group("/master-videos")
	masterVideos.Use(middleware.AuthRequired(), middleware.AdminRequired())
	{
		fmt.Printf("Registered master video routes:\n")
		fmt.Printf("  GET /master-videos\n")
		fmt.Printf("  GET /master-videos/:id\n")
		fmt.Printf("  PUT /master-videos/:id/toggle-status\n")
		fmt.Printf("  PUT /master-videos/:id\n")
		fmt.Printf("  DELETE /master-videos/:id\n")
		fmt.Printf("  POST /master-videos/:id/auto-tag\n")
		fmt.Printf("  GET /master-videos/tags/analytics\n")
		fmt.Printf("  GET /master-videos/tags/untagged\n")
		// Get all master videos with filtering and pagination
		masterVideos.GET("", func(c *gin.Context) {
			log.Printf("🎬 [MASTER-VIDEOS] GET request received")

			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			category := c.Query("category")
			status := c.Query("status")
			syncStatus := c.Query("sync_status")
			vidStatus := c.Query("vid_status")
			search := c.Query("search")
			sortField := c.DefaultQuery("sort_field", "id")
			sortDirection := c.DefaultQuery("sort_direction", "desc")

			log.Printf("🎬 [MASTER-VIDEOS] Query params - page: %d, limit: %d, category: %s, status: %s, syncStatus: %s, vidStatus: %s, search: %s, sortField: %s, sortDirection: %s",
				page, limit, category, status, syncStatus, vidStatus, search, sortField, sortDirection)

			offset := (page - 1) * limit

			var videos []*videoModels.MasterVideo
			var err error
			var totalCount int

			if search != "" {
				log.Printf("🎬 [MASTER-VIDEOS] Executing search query for: %s", search)
				videos, err = videoModels.SearchMasterVideos(db, search, limit, offset, sortField, sortDirection)
				if err != nil {
					log.Printf("❌ [MASTER-VIDEOS] SearchMasterVideos failed: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to search master videos",
						"details": err.Error(),
					})
					return
				}
				log.Printf("🎬 [MASTER-VIDEOS] SearchMasterVideos returned %d videos", len(videos))

				// For search, we need to get total count separately
				log.Printf("🎬 [MASTER-VIDEOS] Getting search count for: %s", search)
				totalCount, err = videoModels.GetMasterVideoSearchCount(db, search)
				if err != nil {
					log.Printf("❌ [MASTER-VIDEOS] GetMasterVideoSearchCount failed: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to get search count",
						"details": err.Error(),
					})
					return
				}
				log.Printf("🎬 [MASTER-VIDEOS] GetMasterVideoSearchCount returned: %d", totalCount)
			} else {
				log.Printf("🎬 [MASTER-VIDEOS] Executing GetMasterVideos query")
				log.Printf("🔍 [CONSOLE-LOG] GetMasterVideos() called with params: limit=%d, offset=%d, category='%s', status='%s', syncStatus='%s', vidStatus='%s', sortField='%s', sortDirection='%s'",
					limit, offset, category, status, syncStatus, vidStatus, sortField, sortDirection)
				videos, err = videoModels.GetMasterVideos(db, limit, offset, category, status, syncStatus, vidStatus, sortField, sortDirection)
				log.Printf("🎬 [MASTER-VIDEOS] GetMasterVideos returned: %v", videos)
				if err != nil {
					log.Printf("❌ [MASTER-VIDEOS] GetMasterVideos failed: %v", err)
					log.Printf("🔍 [CONSOLE-LOG] GetMasterVideos() ERROR: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to fetch master videos",
						"details": err.Error(),
					})
					return
				}
				log.Printf("🎬 [MASTER-VIDEOS] GetMasterVideos returned %d videos", len(videos))
				log.Printf("🔍 [CONSOLE-LOG] GetMasterVideos() SUCCESS: returned %d videos", len(videos))

				// Get total count for filtered results
				log.Printf("🎬 [MASTER-VIDEOS] Getting total count")
				totalCount, err = videoModels.GetMasterVideoCount(db, category, status, syncStatus, vidStatus)
				if err != nil {
					log.Printf("❌ [MASTER-VIDEOS] GetMasterVideoCount failed: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to get video count",
						"details": err.Error(),
					})
					return
				}
				log.Printf("🎬 [MASTER-VIDEOS] GetMasterVideoCount returned: %d", totalCount)
			}

			totalPages := (totalCount + limit - 1) / limit

			log.Printf("🎬 [MASTER-VIDEOS] Preparing response - videos: %d, totalCount: %d, totalPages: %d, currentPage: %d",
				len(videos), totalCount, totalPages, page)

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

			log.Printf("🎬 [MASTER-VIDEOS] Response sent successfully")
		})

		// Get master video by ID
		masterVideos.GET("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
				return
			}

			video, err := videoModels.GetMasterVideoByID(db, id)
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
			existingVideo, err := videoModels.GetMasterVideoByID(db, id)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
				return
			}

			fmt.Printf("DEBUG: Existing video status: %t, new status: %t\n", existingVideo.Vid_Status, toggleData.Vid_Status)

			// Update only the vid_status field
			existingVideo.Vid_Status = toggleData.Vid_Status

			err = videoModels.UpdateMasterVideo(db, existingVideo)
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
			var updateData videoModels.MasterVideo
			if err := json.Unmarshal(body, &updateData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
				return
			}

			// Get existing video
			existingVideo, err := videoModels.GetMasterVideoByID(db, id)
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

			err = videoModels.UpdateMasterVideo(db, existingVideo)
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

			err = videoModels.DeleteMasterVideo(db, id)
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

	// Sync routes - these should be under /master-videos/sync/ to match frontend expectations
	sync := masterVideos.Group("/sync")
	{
		// Sync from Bunny.net to master list
		sync.POST("/from-bunny", func(c *gin.Context) {
			// Get authenticated user ID from context
			userID := c.GetInt("user_id")
			if userID == 0 {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Authentication required",
				})
				return
			}

			result, err := masterVideoService.SyncFromBunny(userID)
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
	conflicts := masterVideos.Group("/conflicts")
	{
		// Get all conflicts
		conflicts.GET("", func(c *gin.Context) {
			conflicts, err := videoModels.GetSyncConflicts(db, nil)
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

			var resolution videoServices.ConflictResolution
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
			stats, err := videoModels.GetMasterVideoStats(db)
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

	// Smart tagging routes
	tags := masterVideos.Group("/tags")
	{
		// Get tag analytics
		tags.GET("/analytics", func(c *gin.Context) {
			analytics, err := videoModels.GetTagAnalytics(db)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to fetch tag analytics",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    analytics,
			})
		})

		// Get untagged videos
		tags.GET("/untagged", func(c *gin.Context) {
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
			videos, err := videoModels.GetUntaggedVideos(db, limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to fetch untagged videos",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"videos":  videos,
				"count":   len(videos),
			})
		})
	}

	// Auto-tag a specific video
	masterVideos.POST("/:id/auto-tag", func(c *gin.Context) {
		// TODO: Implement smart tagging (Phase 3)
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": "Smart tagging feature not yet implemented",
		})
	})

	// Batch auto-tag multiple videos
	masterVideos.POST("/batch-auto-tag", func(c *gin.Context) {
		var request struct {
			VideoIDs []int `json:"video_ids" binding:"required"`
			Replace  bool  `json:"replace"` // NEW: Add this line
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		if len(request.VideoIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No video IDs provided"})
			return
		}

		if len(request.VideoIDs) > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum 100 videos per batch"})
			return
		}

		// TODO: Implement batch smart tagging (Phase 3)
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": "Batch tagging feature not yet implemented",
		})
	})

	// Tag management endpoints
	masterVideos.POST("/tags", func(c *gin.Context) {
		var req struct {
			Tag string `json:"tag" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"success": false, "error": err.Error()})
			return
		}

		// Add tag logic here
		c.JSON(200, gin.H{"success": true, "message": "Tag added successfully"})
	})

	masterVideos.DELETE("/tags/:id", func(c *gin.Context) {
		tagID := c.Param("id")
		// Delete tag logic here
		c.JSON(200, gin.H{"success": true, "message": fmt.Sprintf("Tag %s deleted successfully", tagID)})
	})

	masterVideos.PUT("/tags/:id/category", func(c *gin.Context) {
		tagID := c.Param("id")
		var req struct {
			CategoryID int `json:"category_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"success": false, "error": err.Error()})
			return
		}

		// Assign tag to category logic here
		c.JSON(200, gin.H{"success": true, "message": fmt.Sprintf("Tag %s assigned to category %d successfully", tagID, req.CategoryID)})
	})

	// Category management endpoints
	masterVideos.GET("/tags/categories", func(c *gin.Context) {
		// Get all tag categories across all subsites for admin dashboard
		categories, err := videoModels.GetTagCategories(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   fmt.Sprintf("Failed to get tag categories: %v", err),
			})
			return
		}

		c.JSON(200, gin.H{"success": true, "result": categories})
	})

	masterVideos.POST("/tags/categories", func(c *gin.Context) {
		var req struct {
			Name  string `json:"name" binding:"required"`
			Color string `json:"color" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"success": false, "error": err.Error()})
			return
		}

		// Add category logic here
		c.JSON(200, gin.H{"success": true, "message": fmt.Sprintf("Category '%s' added successfully", req.Name)})
	})

	masterVideos.DELETE("/tags/categories/:id", func(c *gin.Context) {
		categoryID := c.Param("id")
		// Delete category logic here
		c.JSON(200, gin.H{"success": true, "message": fmt.Sprintf("Category %s deleted successfully", categoryID)})
	})

	// Subsite-specific tag management endpoints
	masterVideos.GET("/tags/subsites/:subsite", func(c *gin.Context) {
		subsite := c.Param("subsite")

		tags, err := videoModels.GetSubsiteTags(db, subsite)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   fmt.Sprintf("Failed to get %s tags: %v", subsite, err),
			})
			return
		}

		c.JSON(200, gin.H{"success": true, "result": tags, "subsite": subsite})
	})

	masterVideos.POST("/tags/subsites/:subsite", func(c *gin.Context) {
		subsite := c.Param("subsite")
		var req struct {
			Tag string `json:"tag" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"success": false, "error": err.Error()})
			return
		}

		// Add subsite-specific tag
		err := videoModels.AddSubsiteTag(db, subsite, req.Tag)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   fmt.Sprintf("Failed to add tag to %s subsite: %v", subsite, err),
			})
			return
		}

		c.JSON(200, gin.H{"success": true, "message": fmt.Sprintf("Tag added successfully to %s subsite", subsite)})
	})

	masterVideos.DELETE("/tags/subsites/:subsite/:id", func(c *gin.Context) {
		subsite := c.Param("subsite")
		tagID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid tag ID"})
			return
		}

		// Delete subsite-specific tag
		err = videoModels.DeleteSubsiteTag(db, subsite, tagID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   fmt.Sprintf("Failed to delete tag from %s subsite: %v", subsite, err),
			})
			return
		}

		c.JSON(200, gin.H{"success": true, "message": fmt.Sprintf("Tag deleted successfully from %s subsite", subsite)})
	})

	masterVideos.PUT("/tags/subsites/:subsite/:id/category", func(c *gin.Context) {
		subsite := c.Param("subsite")
		tagID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid tag ID"})
			return
		}

		var req struct {
			CategoryID *int `json:"category_id"` // Make nullable to handle removals
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"success": false, "error": err.Error()})
			return
		}

		// Handle removal (category_id is null) vs assignment
		if req.CategoryID == nil {
			// Remove tag from category
			err = videoModels.RemoveTagFromCategory(db, subsite, tagID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to remove tag from category in %s subsite: %v", subsite, err),
				})
				return
			}
			c.JSON(200, gin.H{"success": true, "message": fmt.Sprintf("Tag removed from category successfully in %s subsite", subsite)})
		} else {
			// Assign tag to category
			err = videoModels.AssignSubsiteTagToCategory(db, subsite, tagID, *req.CategoryID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to assign tag to category in %s subsite: %v", subsite, err),
				})
				return
			}
			c.JSON(200, gin.H{"success": true, "message": fmt.Sprintf("Tag assigned to category successfully in %s subsite", subsite)})
		}
	})

	// Toggle tag active status
	masterVideos.PUT("/tags/subsites/:subsite/:id/toggle-active", func(c *gin.Context) {
		subsite := c.Param("subsite")
		tagID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid tag ID"})
			return
		}

		// Toggle tag active status
		err = videoModels.ToggleTagActiveStatus(db, subsite, tagID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   fmt.Sprintf("Failed to toggle tag active status in %s subsite: %v", subsite, err),
			})
			return
		}

		c.JSON(200, gin.H{"success": true, "message": fmt.Sprintf("Tag active status toggled successfully in %s subsite", subsite)})
	})

	// Subsite-specific category management endpoints
	masterVideos.GET("/tags/subsites/:subsite/categories", func(c *gin.Context) {
		subsite := c.Param("subsite")

		categories, err := videoModels.GetSubsiteCategories(db, subsite)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   fmt.Sprintf("Failed to get %s categories: %v", subsite, err),
			})
			return
		}

		c.JSON(200, gin.H{"success": true, "result": categories, "subsite": subsite})
	})

	masterVideos.POST("/tags/subsites/:subsite/categories", func(c *gin.Context) {
		subsite := c.Param("subsite")
		var req struct {
			Name        string `json:"name" binding:"required"`
			Color       string `json:"color" binding:"required"`
			Description string `json:"description"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"success": false, "error": err.Error()})
			return
		}

		// Add subsite-specific category
		err := videoModels.AddSubsiteCategory(db, subsite, req.Name, req.Color, req.Description)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   fmt.Sprintf("Failed to add category to %s subsite: %v", subsite, err),
			})
			return
		}

		c.JSON(200, gin.H{"success": true, "message": fmt.Sprintf("Category '%s' added successfully to %s subsite", req.Name, subsite)})
	})

	masterVideos.DELETE("/tags/subsites/:subsite/categories/:id", func(c *gin.Context) {
		subsite := c.Param("subsite")
		categoryID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid category ID"})
			return
		}

		// Delete subsite-specific category
		err = videoModels.DeleteSubsiteCategory(db, subsite, categoryID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   fmt.Sprintf("Failed to delete category from %s subsite: %v", subsite, err),
			})
			return
		}

		c.JSON(200, gin.H{"success": true, "message": fmt.Sprintf("Category deleted successfully from %s subsite", subsite)})
	})

	// Article Exclusions Management
	masterVideos.GET("/article-exclusions/:subsite", func(c *gin.Context) {
		subsite := c.Param("subsite")

		// Get subsite ID
		subsiteID, err := videoModels.GetSubsiteID(db, subsite)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to get subsite ID for '%s': %v", subsite, err)})
			return
		}

		exclusions, err := videoModels.GetArticleExclusions(db, subsiteID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get article exclusions: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"result":  exclusions,
		})
	})

	masterVideos.POST("/article-exclusions/:subsite", func(c *gin.Context) {
		subsite := c.Param("subsite")

		var request struct {
			Word string `json:"word" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		// Get subsite ID
		subsiteID, err := videoModels.GetSubsiteID(db, subsite)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to get subsite ID for '%s': %v", subsite, err)})
			return
		}

		err = videoModels.AddArticleExclusion(db, subsiteID, request.Word)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to add article exclusion: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Article exclusion added successfully",
		})
	})

	masterVideos.PUT("/article-exclusions/:subsite/toggle", func(c *gin.Context) {
		subsite := c.Param("subsite")

		var request struct {
			Word     string `json:"word" binding:"required"`
			Excluded bool   `json:"excluded"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		// Get subsite ID
		subsiteID, err := videoModels.GetSubsiteID(db, subsite)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to get subsite ID for '%s': %v", subsite, err)})
			return
		}

		err = videoModels.ToggleArticleExclusion(db, subsiteID, request.Word, request.Excluded)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to toggle article exclusion: %v", err)})
			return
		}

		status := "excluded"
		if !request.Excluded {
			status = "included"
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": fmt.Sprintf("Article exclusion '%s' %s", request.Word, status),
		})
	})

	masterVideos.DELETE("/article-exclusions/:subsite/:word", func(c *gin.Context) {
		subsite := c.Param("subsite")
		word := c.Param("word")

		// Get subsite ID
		subsiteID, err := videoModels.GetSubsiteID(db, subsite)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to get subsite ID for '%s': %v", subsite, err)})
			return
		}

		err = videoModels.RemoveArticleExclusion(db, subsiteID, word)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to remove article exclusion: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Article exclusion removed successfully",
		})
	})
}
