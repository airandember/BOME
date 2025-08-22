package routes

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"bome-backend/internal/database"

	"github.com/gin-gonic/gin"
)

// SetupTagRoutes configures tag and category routes for the category modal
func SetupTagRoutes(router *gin.Engine, db *database.DB) {
	// Tag Categories Management - Primary focus for category modal
	tagCategories := router.Group("/api/v1/tag-categories")
	{
		// Get all tag categories
		tagCategories.GET("", func(c *gin.Context) {
			categories, err := db.GetTagCategoriesBySubsite(1) // Hardcode streaming subsite for now
			if err != nil {
				log.Printf("❌ Failed to get tag categories: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to get tag categories: %v", err),
				})
				return
			}
			log.Printf("✅ Retrieved %d tag categories", len(categories))
			c.JSON(200, gin.H{"success": true, "result": categories})
		})

		// Get tag category by ID
		tagCategories.GET("/:id", func(c *gin.Context) {
			categoryID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(400, gin.H{"success": false, "error": "Invalid category ID"})
				return
			}

			category, err := db.GetTagCategoryByID(categoryID)
			if err != nil {
				log.Printf("❌ Failed to get category %d: %v", categoryID, err)
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Tag category not found: %v", err),
				})
				return
			}

			log.Printf("✅ Retrieved category %d: %s", categoryID, category.Name)
			c.JSON(200, gin.H{"success": true, "result": category})
		})

		// Create new tag category
		tagCategories.POST("", func(c *gin.Context) {
			var req struct {
				Name        string `json:"name" binding:"required"`
				Color       string `json:"color" binding:"required"`
				Description string `json:"description"`
				SubsiteID   *int   `json:"subsite_id"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"success": false, "error": err.Error()})
				return
			}

			category, err := db.CreateTagCategory(req.Name, req.Color, req.Description, req.SubsiteID)
			if err != nil {
				log.Printf("❌ Failed to create category '%s': %v", req.Name, err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to create tag category: %v", err),
				})
				return
			}

			log.Printf("✅ Created category %d: %s", category.ID, category.Name)
			c.JSON(201, gin.H{"success": true, "result": category})
		})

		// Update tag category details
		tagCategories.PUT("/:id", func(c *gin.Context) {
			categoryID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(400, gin.H{"success": false, "error": "Invalid category ID"})
				return
			}

			var req struct {
				Name        string `json:"name"`
				Color       string `json:"color"`
				Description string `json:"description"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"success": false, "error": err.Error()})
				return
			}

			err = db.UpdateTagCategoryFields(categoryID, req.Name, req.Color, req.Description)
			if err != nil {
				log.Printf("❌ Failed to update category %d: %v", categoryID, err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to update tag category: %v", err),
				})
				return
			}

			log.Printf("✅ Updated category %d", categoryID)
			c.JSON(200, gin.H{"success": true, "message": "Tag category updated successfully"})
		})

		// Delete tag category
		tagCategories.DELETE("/:id", func(c *gin.Context) {
			categoryID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(400, gin.H{"success": false, "error": "Invalid category ID"})
				return
			}

			err = db.DeleteTagCategory(categoryID)
			if err != nil {
				log.Printf("❌ Failed to delete category %d: %v", categoryID, err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to delete tag category: %v", err),
				})
				return
			}

			log.Printf("✅ Deleted category %d", categoryID)
			c.JSON(200, gin.H{"success": true, "message": "Tag category deleted successfully"})
		})

		// Add tag to category (single operation)
		tagCategories.POST("/:id/tags/:tagId", func(c *gin.Context) {
			categoryID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(400, gin.H{"success": false, "error": "Invalid category ID"})
				return
			}

			tagID, err := strconv.Atoi(c.Param("tagId"))
			if err != nil {
				c.JSON(400, gin.H{"success": false, "error": "Invalid tag ID"})
				return
			}

			err = db.AddTagToCategory(tagID, categoryID)
			if err != nil {
				log.Printf("❌ Failed to add tag %d to category %d: %v", tagID, categoryID, err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to add tag to category: %v", err),
				})
				return
			}

			log.Printf("✅ Added tag %d to category %d", tagID, categoryID)
			c.JSON(200, gin.H{"success": true, "message": "Tag added to category successfully"})
		})

		// Remove tag from category (single operation)
		tagCategories.DELETE("/:id/tags/:tagId", func(c *gin.Context) {
			categoryID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(400, gin.H{"success": false, "error": "Invalid category ID"})
				return
			}

			tagID, err := strconv.Atoi(c.Param("tagId"))
			if err != nil {
				c.JSON(400, gin.H{"success": false, "error": "Invalid tag ID"})
				return
			}

			err = db.TagRemoveFromCategory(tagID, categoryID)
			if err != nil {
				log.Printf("❌ Failed to remove tag %d from category %d: %v", tagID, categoryID, err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to remove tag from category: %v", err),
				})
				return
			}

			log.Printf("✅ Removed tag %d from category %d", tagID, categoryID)
			c.JSON(200, gin.H{"success": true, "message": "Tag removed from category successfully"})
		})

		// Get all tags for a category
		tagCategories.GET("/:id/tags", func(c *gin.Context) {
			categoryID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(400, gin.H{"success": false, "error": "Invalid category ID"})
				return
			}

			tags, err := db.GetTagsByCategory(categoryID)
			if err != nil {
				log.Printf("❌ Failed to get tags for category %d: %v", categoryID, err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to get category tags: %v", err),
				})
				return
			}

			log.Printf("✅ Retrieved %d tags for category %d", len(tags), categoryID)
			c.JSON(200, gin.H{"success": true, "result": tags})
		})

		// Batch update tag-category relationships (for modal save)
		tagCategories.POST("/batch-update", func(c *gin.Context) {
			var req struct {
				Changes []struct {
					TagID      int    `json:"tagId" binding:"required"`
					CategoryID *int   `json:"categoryId"`
					Action     string `json:"action" binding:"required"`
				} `json:"changes" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				log.Printf("❌ Invalid batch update request: %v", err)
				c.JSON(400, gin.H{"success": false, "error": err.Error()})
				return
			}

			log.Printf("🔄 Processing %d batch changes", len(req.Changes))

			// Process each change
			for i, change := range req.Changes {
				log.Printf("  Change %d: Tag %d -> %s (Category: %v)",
					i+1, change.TagID, change.Action, change.CategoryID)

				switch change.Action {
				case "add":
					if change.CategoryID == nil {
						log.Printf("❌ Add action requires category ID")
						c.JSON(400, gin.H{
							"success": false,
							"error":   fmt.Sprintf("Add action for tag %d requires category ID", change.TagID),
						})
						return
					}

					err := db.AddTagToCategory(change.TagID, *change.CategoryID)
					if err != nil {
						log.Printf("❌ Failed to add tag %d to category %d: %v",
							change.TagID, *change.CategoryID, err)
						c.JSON(http.StatusInternalServerError, gin.H{
							"success": false,
							"error":   fmt.Sprintf("Failed to add tag %d to category: %v", change.TagID, err),
						})
						return
					}

				case "remove":
					if change.CategoryID == nil {
						// Remove from all categories
						err := db.RemoveTagFromAllCategories(change.TagID)
						if err != nil {
							log.Printf("❌ Failed to remove tag %d from all categories: %v", change.TagID, err)
							c.JSON(http.StatusInternalServerError, gin.H{
								"success": false,
								"error":   fmt.Sprintf("Failed to remove tag %d from all categories: %v", change.TagID, err),
							})
							return
						}
					} else {
						// Remove from specific category
						err := db.TagRemoveFromCategory(change.TagID, *change.CategoryID)
						if err != nil {
							log.Printf("❌ Failed to remove tag %d from category %d: %v",
								change.TagID, *change.CategoryID, err)
							c.JSON(http.StatusInternalServerError, gin.H{
								"success": false,
								"error":   fmt.Sprintf("Failed to remove tag %d from category: %v", change.TagID, err),
							})
							return
						}
					}

				default:
					log.Printf("❌ Invalid action: %s", change.Action)
					c.JSON(400, gin.H{
						"success": false,
						"error":   fmt.Sprintf("Invalid action '%s' for tag %d", change.Action, change.TagID),
					})
					return
				}
			}

			log.Printf("✅ Successfully processed %d batch changes", len(req.Changes))
			c.JSON(200, gin.H{
				"success": true,
				"message": fmt.Sprintf("Successfully processed %d changes", len(req.Changes)),
			})
		})
	}

	// Tags Management - Basic CRUD for universal tags
	tags := router.Group("/api/v1/tags")
	{
		// Get all tags
		tags.GET("", func(c *gin.Context) {
			tags, err := db.GetTagsBySubsite(1) // Hardcode streaming subsite for now
			if err != nil {
				log.Printf("❌ Failed to get tags: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to get tags: %v", err),
				})
				return
			}
			log.Printf("✅ Retrieved %d tags", len(tags))
			c.JSON(200, gin.H{"success": true, "result": tags})
		})

		// Get tag by ID
		tags.GET("/:id", func(c *gin.Context) {
			tagID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(400, gin.H{"success": false, "error": "Invalid tag ID"})
				return
			}

			tag, err := db.GetTagByID(tagID)
			if err != nil {
				log.Printf("❌ Failed to get tag %d: %v", tagID, err)
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Tag not found: %v", err),
				})
				return
			}

			log.Printf("✅ Retrieved tag %d: %s", tagID, tag.Word)
			c.JSON(200, gin.H{"success": true, "result": tag})
		})

		// Get categories for a tag
		tags.GET("/:id/categories", func(c *gin.Context) {
			tagID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(400, gin.H{"success": false, "error": "Invalid tag ID"})
				return
			}

			categories, err := db.GetCategoriesForTag(tagID)
			if err != nil {
				log.Printf("❌ Failed to get categories for tag %d: %v", tagID, err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to get tag categories: %v", err),
				})
				return
			}

			log.Printf("✅ Retrieved %d categories for tag %d", len(categories), tagID)
			c.JSON(200, gin.H{"success": true, "result": categories})
		})
	}
}
