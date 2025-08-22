package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"bome-backend/internal/database"

	"github.com/gin-gonic/gin"
)

// SetupTagRoutes configures all tag-related routes
func SetupTagRoutes(router *gin.Engine, db *database.DB) {
	tags := router.Group("/api/v1/tags")
	{
		// Get all tags
		tags.GET("", func(c *gin.Context) {
			tags, err := db.GetTags()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to get tags: %v", err),
				})
				return
			}
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
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Tag not found: %v", err),
				})
				return
			}
			c.JSON(200, gin.H{"success": true, "result": tag})
		})

		// Create new tag
		tags.POST("", func(c *gin.Context) {
			var req struct {
				Word string `json:"word" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"success": false, "error": err.Error()})
				return
			}
			tag, err := db.CreateTag(req.Word, nil)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to create tag: %v", err),
				})
				return
			}
			c.JSON(201, gin.H{"success": true, "result": tag})
		})

		// Update tag
		tags.PUT("/:id", func(c *gin.Context) {
			tagID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(400, gin.H{"success": false, "error": "Invalid tag ID"})
				return
			}
			var req struct {
				Word string `json:"word"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"success": false, "error": err.Error()})
				return
			}
			err = db.UpdateTag(tagID, req.Word)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to update tag: %v", err),
				})
				return
			}
			c.JSON(200, gin.H{"success": true, "message": "Tag updated successfully"})
		})

		// Delete tag
		tags.DELETE("/:id", func(c *gin.Context) {
			tagID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(400, gin.H{"success": false, "error": "Invalid tag ID"})
				return
			}
			err = db.DeleteTag(tagID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to delete tag: %v", err),
				})
				return
			}
			c.JSON(200, gin.H{"success": true, "message": "Tag deleted successfully"})
		})

		// Tag-Category Management
		tags.PUT("/:id/categories", func(c *gin.Context) {
			tagID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(400, gin.H{"success": false, "error": "Invalid tag ID"})
				return
			}
			var req struct {
				CategoryIDs []int `json:"category_ids"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"success": false, "error": err.Error()})
				return
			}
			err = db.UpdateTagCategories(tagID, req.CategoryIDs)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to update tag categories: %v", err),
				})
				return
			}
			c.JSON(200, gin.H{"success": true, "message": "Tag categories updated successfully"})
		})

		// Get tag categories
		tags.GET("/:id/categories", func(c *gin.Context) {
			tagID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(400, gin.H{"success": false, "error": "Invalid tag ID"})
				return
			}
			tag, err := db.GetTagByID(tagID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Tag not found: %v", err),
				})
				return
			}
			c.JSON(200, gin.H{"success": true, "result": tag.CategoryIDs})
		})
	}

	// Tag Categories Management
	tagCategories := router.Group("/api/v1/tag-categories")
	{
		// Get all tag categories
		tagCategories.GET("", func(c *gin.Context) {
			categories, err := db.GetTagCategories()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to get tag categories: %v", err),
				})
				return
			}
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
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Tag category not found: %v", err),
				})
				return
			}
			c.JSON(200, gin.H{"success": true, "result": category})
		})

		// Create new tag category
		tagCategories.POST("", func(c *gin.Context) {
			var req struct {
				Name        string `json:"name" binding:"required"`
				Color       string `json:"color" binding:"required"`
				Description string `json:"description"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"success": false, "error": err.Error()})
				return
			}
			category, err := db.CreateTagCategory(req.Name, req.Color, req.Description, nil)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to create tag category: %v", err),
				})
				return
			}
			c.JSON(201, gin.H{"success": true, "result": category})
		})

		// Update tag category
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
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to update tag category: %v", err),
				})
				return
			}
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
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to delete tag category: %v", err),
				})
				return
			}
			c.JSON(200, gin.H{"success": true, "message": "Tag category deleted successfully"})
		})

		// Tag-Category Relationship Management
		tagCategories.PUT("/:id/tags", func(c *gin.Context) {
			// For now, we'll use the batch update endpoint instead
			// This endpoint can be implemented later if needed
			c.JSON(501, gin.H{"success": false, "error": "Not implemented - use batch-update endpoint"})
		})

		// Get category tags
		tagCategories.GET("/:id/tags", func(c *gin.Context) {
			categoryID, err := strconv.Atoi(c.Param("id"))
			if err != nil {
				c.JSON(400, gin.H{"success": false, "error": "Invalid category ID"})
				return
			}
			tags, err := db.GetCategoryTags(categoryID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to get category tags: %v", err),
				})
				return
			}
			c.JSON(200, gin.H{"success": true, "result": tags})
		})

		// Batch update tag-category relationships
		tagCategories.POST("/batch-update", func(c *gin.Context) {
			var req struct {
				Changes []map[string]interface{} `json:"changes"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"success": false, "error": err.Error()})
				return
			}
			err := db.BatchUpdateTagCategories(req.Changes)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Failed to batch update: %v", err),
				})
				return
			}
			c.JSON(200, gin.H{"success": true, "message": "Batch update completed successfully"})
		})
	}
}
