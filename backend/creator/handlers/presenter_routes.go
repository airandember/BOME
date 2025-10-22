package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"bome-backend/creator/models"
	"bome-backend/creator/services"
	"bome-backend/infrastructure/database"
)

// SetupPresenterRoutes registers all presenter-related routes
func SetupPresenterRoutes(router *gin.RouterGroup, db *database.DB) {
	service := services.NewPresenterService(db)
	
	// Presenter CRUD
	router.GET("/presenters", GetPresentersHandler(service))
	router.POST("/presenters", CreatePresenterHandler(service))
	router.GET("/presenters/:id", GetPresenterByIDHandler(service))
	router.PUT("/presenters/:id", UpdatePresenterHandler(service))
	router.DELETE("/presenters/:id", DeletePresenterHandler(service))
	router.POST("/presenters/:id/verify", VerifyPresenterHandler(service))
	router.GET("/presenters/stats", GetPresenterStatsHandler(service))
	
	// Video-Presenter Links
	router.GET("/presenters/:id/videos", GetPresenterVideosHandler(service))
	router.GET("/videos/:id/presenters", GetVideoPresentersHandler(service))
	router.POST("/video-presenters", LinkPresenterToVideoHandler(service))
	router.PUT("/video-presenters/:id", UpdateVideoPresenterHandler(service))
	router.DELETE("/video-presenters/:id", UnlinkPresenterFromVideoHandler(service))
	
	// Statistics
	router.POST("/presenters/:id/update-stats", UpdatePresenterStatsHandler(service))
	router.POST("/presenters/update-all-stats", UpdateAllPresenterStatsHandler(service))
	
	log.Println("[PRESENTER-ROUTES] ✅ Registered 14 presenter management endpoints")
}

// GetPresentersHandler retrieves all presenters
func GetPresentersHandler(service *services.PresenterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		activeOnly := c.Query("active") == "true"
		verifiedOnly := c.Query("verified") == "true"
		
		presenters, err := service.GetPresenters(activeOnly, verifiedOnly)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"presenters": presenters,
			"count":      len(presenters),
		})
	}
}

// CreatePresenterHandler creates a new presenter
func CreatePresenterHandler(service *services.PresenterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.CreatePresenterInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		presenter, err := service.CreatePresenter(&input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusCreated, presenter)
	}
}

// GetPresenterByIDHandler retrieves a presenter by ID
func GetPresenterByIDHandler(service *services.PresenterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid presenter ID"})
			return
		}
		
		presenter, err := service.GetPresenterByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "presenter not found"})
			return
		}
		
		c.JSON(http.StatusOK, presenter)
	}
}

// UpdatePresenterHandler updates a presenter
func UpdatePresenterHandler(service *services.PresenterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid presenter ID"})
			return
		}
		
		var input models.UpdatePresenterInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		presenter, err := service.UpdatePresenter(id, &input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, presenter)
	}
}

// DeletePresenterHandler soft-deletes a presenter
func DeletePresenterHandler(service *services.PresenterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid presenter ID"})
			return
		}
		
		err = service.DeletePresenter(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"message": "Presenter deleted successfully"})
	}
}

// VerifyPresenterHandler marks a presenter as verified
func VerifyPresenterHandler(service *services.PresenterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid presenter ID"})
			return
		}
		
		// Get admin user ID from context (set by auth middleware)
		adminUserID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}
		
		err = service.VerifyPresenter(id, adminUserID.(int))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"message": "Presenter verified successfully"})
	}
}

// GetPresenterStatsHandler retrieves presenter statistics
func GetPresenterStatsHandler(service *services.PresenterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := service.GetPresenterStats()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, stats)
	}
}

// GetPresenterVideosHandler retrieves all videos for a presenter
func GetPresenterVideosHandler(service *services.PresenterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid presenter ID"})
			return
		}
		
		videos, err := service.GetPresenterVideos(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"videos": videos,
			"count":  len(videos),
		})
	}
}

// GetVideoPresentersHandler retrieves all presenters for a video
func GetVideoPresentersHandler(service *services.PresenterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid video ID"})
			return
		}
		
		presenters, err := service.GetVideoPresenters(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"presenters": presenters,
			"count":      len(presenters),
		})
	}
}

// LinkPresenterToVideoHandler links a presenter to a video
func LinkPresenterToVideoHandler(service *services.PresenterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.CreateVideoPresenterInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Get admin user ID from context
		adminUserID, exists := c.Get("user_id")
		if exists {
			input.AddedBy = adminUserID.(int)
		}
		
		videoPresenter, err := service.LinkPresenterToVideo(&input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusCreated, videoPresenter)
	}
}

// UpdateVideoPresenterHandler updates a video-presenter link
func UpdateVideoPresenterHandler(service *services.PresenterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid link ID"})
			return
		}
		
		var input models.UpdateVideoPresenterInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		videoPresenter, err := service.UpdateVideoPresenter(id, &input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, videoPresenter)
	}
}

// UnlinkPresenterFromVideoHandler removes a presenter from a video
func UnlinkPresenterFromVideoHandler(service *services.PresenterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid link ID"})
			return
		}
		
		err = service.UnlinkPresenterFromVideo(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"message": "Presenter unlinked from video successfully"})
	}
}

// UpdatePresenterStatsHandler updates statistics for a presenter
func UpdatePresenterStatsHandler(service *services.PresenterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid presenter ID"})
			return
		}
		
		err = service.UpdatePresenterStatistics(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"message": "Presenter statistics updated successfully"})
	}
}

// UpdateAllPresenterStatsHandler updates statistics for all presenters
func UpdateAllPresenterStatsHandler(service *services.PresenterService) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := service.UpdateAllPresenterStatistics()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"message": "All presenter statistics updated successfully"})
	}
}

