package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DesignToken represents a design token from Figma
type DesignToken struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Value       string    `json:"value" gorm:"type:text"`
	Type        string    `json:"type" gorm:"not null"` // color, spacing, typography, shadow, border, size
	Category    string    `json:"category" gorm:"not null"`
	Description string    `json:"description"`
	FigmaID     string    `json:"figmaId"`
	ThemeID     uint      `json:"themeId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// StyleTheme represents a collection of design tokens
type StyleTheme struct {
	ID          uint          `json:"id" gorm:"primaryKey"`
	Name        string        `json:"name" gorm:"not null"`
	Description string        `json:"description"`
	IsActive    bool          `json:"isActive" gorm:"default:false"`
	FigmaFileID string        `json:"figmaFileId"`
	FigmaNodeID string        `json:"figmaNodeId"`
	Tokens      []DesignToken `json:"tokens" gorm:"foreignKey:ThemeID;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
}

// FigmaAPIResponse represents the structure of Figma API response
type FigmaAPIResponse struct {
	Document struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Children []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Type     string `json:"type"`
			Children []struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Type  string `json:"type"`
				Fills []struct {
					Type  string `json:"type"`
					Color struct {
						R float64 `json:"r"`
						G float64 `json:"g"`
						B float64 `json:"b"`
						A float64 `json:"a"`
					} `json:"color"`
				} `json:"fills"`
				Strokes []struct {
					Type  string `json:"type"`
					Color struct {
						R float64 `json:"r"`
						G float64 `json:"g"`
						B float64 `json:"b"`
						A float64 `json:"a"`
					} `json:"color"`
				} `json:"strokes"`
				Effects []struct {
					Type   string  `json:"type"`
					Radius float64 `json:"radius"`
					Color  struct {
						R float64 `json:"r"`
						G float64 `json:"g"`
						B float64 `json:"b"`
						A float64 `json:"a"`
					} `json:"color"`
					Offset struct {
						X float64 `json:"x"`
						Y float64 `json:"y"`
					} `json:"offset"`
				} `json:"effects"`
				Style struct {
					FontFamily string  `json:"fontFamily"`
					FontSize   float64 `json:"fontSize"`
					FontWeight float64 `json:"fontWeight"`
					LineHeight struct {
						Value float64 `json:"value"`
						Unit  string  `json:"unit"`
					} `json:"lineHeight"`
					LetterSpacing struct {
						Value float64 `json:"value"`
						Unit  string  `json:"unit"`
					} `json:"letterSpacing"`
				} `json:"style"`
			} `json:"children"`
		} `json:"children"`
	} `json:"document"`
	Styles map[string]struct {
		Key         string `json:"key"`
		Name        string `json:"name"`
		Description string `json:"description"`
		StyleType   string `json:"styleType"`
	} `json:"styles"`
}

// CreateThemeRequest represents the request to create a theme from Figma
type CreateThemeRequest struct {
	FigmaFileID string `json:"figmaFileId" binding:"required"`
	FigmaNodeID string `json:"figmaNodeId"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ImportThemeRequest represents the request to import a theme
type ImportThemeRequest struct {
	ThemeData string `json:"themeData" binding:"required"`
}

// ActivateThemeRequest represents the request to activate a theme
type ActivateThemeRequest struct {
	ThemeID uint `json:"themeId" binding:"required"`
}

// DesignSystemRoutes sets up design system related routes
func DesignSystemRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// Auto-migrate tables
	db.AutoMigrate(&StyleTheme{}, &DesignToken{})

	designSystem := router.Group("/design-system")
	{
		// Theme management
		designSystem.GET("/themes", getThemes(db))
		designSystem.POST("/themes", createTheme(db))
		designSystem.PUT("/themes/:id", updateTheme(db))
		designSystem.DELETE("/themes/:id", deleteTheme(db))
		designSystem.POST("/themes/activate", activateTheme(db))
		designSystem.GET("/themes/:id/tokens", getThemeTokens(db))

		// Figma integration
		designSystem.POST("/figma/import", importFromFigma(db))
		designSystem.POST("/figma/sync/:id", syncWithFigma(db))
		designSystem.GET("/figma/preview", previewFigmaTokens())

		// Theme operations
		designSystem.POST("/themes/import", importTheme(db))
		designSystem.GET("/themes/:id/export", exportTheme(db))
		designSystem.GET("/active", getActiveTheme(db))

		// Token management
		designSystem.GET("/tokens", getAllTokens(db))
		designSystem.POST("/tokens", createToken(db))
		designSystem.PUT("/tokens/:id", updateToken(db))
		designSystem.DELETE("/tokens/:id", deleteToken(db))
	}
}

// getThemes returns all themes
func getThemes(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var themes []StyleTheme

		if err := db.Preload("Tokens").Find(&themes).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch themes",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"themes": themes,
			"count":  len(themes),
		})
	}
}

// createTheme creates a new theme
func createTheme(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var theme StyleTheme

		if err := c.ShouldBindJSON(&theme); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid theme data",
				"details": err.Error(),
			})
			return
		}

		// Ensure only one active theme
		if theme.IsActive {
			db.Model(&StyleTheme{}).Where("is_active = ?", true).Update("is_active", false)
		}

		if err := db.Create(&theme).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create theme",
				"details": err.Error(),
			})
			return
		}

		// Load tokens for response
		db.Preload("Tokens").First(&theme, theme.ID)

		c.JSON(http.StatusCreated, gin.H{
			"message": "Theme created successfully",
			"theme":   theme,
		})
	}
}

// updateTheme updates an existing theme
func updateTheme(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		themeID := c.Param("id")

		var theme StyleTheme
		if err := db.First(&theme, themeID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Theme not found",
			})
			return
		}

		var updateData StyleTheme
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid update data",
				"details": err.Error(),
			})
			return
		}

		// Ensure only one active theme
		if updateData.IsActive && !theme.IsActive {
			db.Model(&StyleTheme{}).Where("is_active = ?", true).Update("is_active", false)
		}

		if err := db.Model(&theme).Updates(updateData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update theme",
				"details": err.Error(),
			})
			return
		}

		// Load updated theme with tokens
		db.Preload("Tokens").First(&theme, theme.ID)

		c.JSON(http.StatusOK, gin.H{
			"message": "Theme updated successfully",
			"theme":   theme,
		})
	}
}

// deleteTheme deletes a theme
func deleteTheme(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		themeID := c.Param("id")

		var theme StyleTheme
		if err := db.First(&theme, themeID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Theme not found",
			})
			return
		}

		if err := db.Delete(&theme).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to delete theme",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Theme deleted successfully",
		})
	}
}

// activateTheme activates a specific theme
func activateTheme(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request ActivateThemeRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request data",
				"details": err.Error(),
			})
			return
		}

		// Deactivate all themes
		if err := db.Model(&StyleTheme{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to deactivate existing themes",
				"details": err.Error(),
			})
			return
		}

		// Activate the selected theme
		var theme StyleTheme
		if err := db.First(&theme, request.ThemeID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Theme not found",
			})
			return
		}

		theme.IsActive = true
		if err := db.Save(&theme).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to activate theme",
				"details": err.Error(),
			})
			return
		}

		// Load theme with tokens
		db.Preload("Tokens").First(&theme, theme.ID)

		c.JSON(http.StatusOK, gin.H{
			"message": "Theme activated successfully",
			"theme":   theme,
		})
	}
}

// getActiveTheme returns the currently active theme
func getActiveTheme(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var theme StyleTheme

		if err := db.Preload("Tokens").Where("is_active = ?", true).First(&theme).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusOK, gin.H{
					"theme":   nil,
					"message": "No active theme found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch active theme",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"theme": theme,
		})
	}
}

// importFromFigma creates a theme from Figma data
func importFromFigma(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request CreateThemeRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request data",
				"details": err.Error(),
			})
			return
		}

		// Fetch data from Figma API (mock implementation)
		figmaData, err := fetchFigmaData(request.FigmaFileID, request.FigmaNodeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch Figma data",
				"details": err.Error(),
			})
			return
		}

		// Parse Figma data into design tokens
		tokens, err := parseFigmaData(figmaData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to parse Figma data",
				"details": err.Error(),
			})
			return
		}

		// Create theme
		theme := StyleTheme{
			Name:        getThemeName(request.Name, figmaData),
			Description: getThemeDescription(request.Description, figmaData),
			FigmaFileID: request.FigmaFileID,
			FigmaNodeID: request.FigmaNodeID,
			IsActive:    false,
			Tokens:      tokens,
		}

		if err := db.Create(&theme).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create theme",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Theme created from Figma successfully",
			"theme":   theme,
		})
	}
}

// syncWithFigma updates a theme with latest Figma data
func syncWithFigma(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		themeID := c.Param("id")

		var theme StyleTheme
		if err := db.Preload("Tokens").First(&theme, themeID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Theme not found",
			})
			return
		}

		if theme.FigmaFileID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Theme is not linked to Figma",
			})
			return
		}

		// Fetch latest data from Figma
		figmaData, err := fetchFigmaData(theme.FigmaFileID, theme.FigmaNodeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch Figma data",
				"details": err.Error(),
			})
			return
		}

		// Parse new tokens
		newTokens, err := parseFigmaData(figmaData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to parse Figma data",
				"details": err.Error(),
			})
			return
		}

		// Delete existing tokens
		if err := db.Where("theme_id = ?", theme.ID).Delete(&DesignToken{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to clear existing tokens",
				"details": err.Error(),
			})
			return
		}

		// Add new tokens
		for i := range newTokens {
			newTokens[i].ThemeID = theme.ID
		}

		if err := db.Create(&newTokens).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create new tokens",
				"details": err.Error(),
			})
			return
		}

		// Update theme timestamp
		theme.UpdatedAt = time.Now()
		db.Save(&theme)

		// Reload theme with new tokens
		db.Preload("Tokens").First(&theme, theme.ID)

		c.JSON(http.StatusOK, gin.H{
			"message": "Theme synced with Figma successfully",
			"theme":   theme,
		})
	}
}

// previewFigmaTokens previews tokens from Figma without saving
func previewFigmaTokens() gin.HandlerFunc {
	return func(c *gin.Context) {
		figmaFileID := c.Query("fileId")
		figmaNodeID := c.Query("nodeId")

		if figmaFileID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Figma file ID is required",
			})
			return
		}

		// Fetch data from Figma API
		figmaData, err := fetchFigmaData(figmaFileID, figmaNodeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch Figma data",
				"details": err.Error(),
			})
			return
		}

		// Parse tokens without saving
		tokens, err := parseFigmaData(figmaData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to parse Figma data",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tokens":  tokens,
			"count":   len(tokens),
			"preview": true,
		})
	}
}

// importTheme imports a theme from JSON data
func importTheme(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request ImportThemeRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request data",
				"details": err.Error(),
			})
			return
		}

		var theme StyleTheme
		if err := json.Unmarshal([]byte(request.ThemeData), &theme); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid theme data format",
				"details": err.Error(),
			})
			return
		}

		// Reset IDs to avoid conflicts
		theme.ID = 0
		theme.IsActive = false
		theme.CreatedAt = time.Now()
		theme.UpdatedAt = time.Now()

		for i := range theme.Tokens {
			theme.Tokens[i].ID = 0
			theme.Tokens[i].ThemeID = 0
			theme.Tokens[i].CreatedAt = time.Now()
			theme.Tokens[i].UpdatedAt = time.Now()
		}

		if err := db.Create(&theme).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to import theme",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Theme imported successfully",
			"theme":   theme,
		})
	}
}

// exportTheme exports a theme as JSON
func exportTheme(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		themeID := c.Param("id")

		var theme StyleTheme
		if err := db.Preload("Tokens").First(&theme, themeID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Theme not found",
			})
			return
		}

		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"theme-%s.json\"", theme.Name))
		c.JSON(http.StatusOK, theme)
	}
}

// getThemeTokens returns tokens for a specific theme
func getThemeTokens(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		themeID := c.Param("id")

		var tokens []DesignToken
		if err := db.Where("theme_id = ?", themeID).Find(&tokens).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch tokens",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tokens": tokens,
			"count":  len(tokens),
		})
	}
}

// getAllTokens returns all tokens across all themes
func getAllTokens(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokens []DesignToken

		query := db.Model(&DesignToken{})

		// Filter by theme if specified
		if themeID := c.Query("themeId"); themeID != "" {
			query = query.Where("theme_id = ?", themeID)
		}

		// Filter by type if specified
		if tokenType := c.Query("type"); tokenType != "" {
			query = query.Where("type = ?", tokenType)
		}

		// Filter by category if specified
		if category := c.Query("category"); category != "" {
			query = query.Where("category = ?", category)
		}

		if err := query.Find(&tokens).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to fetch tokens",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tokens": tokens,
			"count":  len(tokens),
		})
	}
}

// createToken creates a new design token
func createToken(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token DesignToken

		if err := c.ShouldBindJSON(&token); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid token data",
				"details": err.Error(),
			})
			return
		}

		if err := db.Create(&token).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create token",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Token created successfully",
			"token":   token,
		})
	}
}

// updateToken updates an existing design token
func updateToken(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenID := c.Param("id")

		var token DesignToken
		if err := db.First(&token, tokenID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Token not found",
			})
			return
		}

		var updateData DesignToken
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid update data",
				"details": err.Error(),
			})
			return
		}

		if err := db.Model(&token).Updates(updateData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update token",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Token updated successfully",
			"token":   token,
		})
	}
}

// deleteToken deletes a design token
func deleteToken(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenID := c.Param("id")

		var token DesignToken
		if err := db.First(&token, tokenID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Token not found",
			})
			return
		}

		if err := db.Delete(&token).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to delete token",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Token deleted successfully",
		})
	}
}

// Helper functions

// fetchFigmaData fetches design data from Figma API
func fetchFigmaData(fileID, nodeID string) (*FigmaAPIResponse, error) {
	// In a real implementation, this would make an actual API call to Figma
	// For now, return mock data
	mockResponse := &FigmaAPIResponse{
		Document: struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Children []struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				Type     string `json:"type"`
				Children []struct {
					ID    string `json:"id"`
					Name  string `json:"name"`
					Type  string `json:"type"`
					Fills []struct {
						Type  string `json:"type"`
						Color struct {
							R float64 `json:"r"`
							G float64 `json:"g"`
							B float64 `json:"b"`
							A float64 `json:"a"`
						} `json:"color"`
					} `json:"fills"`
					Strokes []struct {
						Type  string `json:"type"`
						Color struct {
							R float64 `json:"r"`
							G float64 `json:"g"`
							B float64 `json:"b"`
							A float64 `json:"a"`
						} `json:"color"`
					} `json:"strokes"`
					Effects []struct {
						Type   string  `json:"type"`
						Radius float64 `json:"radius"`
						Color  struct {
							R float64 `json:"r"`
							G float64 `json:"g"`
							B float64 `json:"b"`
							A float64 `json:"a"`
						} `json:"color"`
						Offset struct {
							X float64 `json:"x"`
							Y float64 `json:"y"`
						} `json:"offset"`
					} `json:"effects"`
					Style struct {
						FontFamily string  `json:"fontFamily"`
						FontSize   float64 `json:"fontSize"`
						FontWeight float64 `json:"fontWeight"`
						LineHeight struct {
							Value float64 `json:"value"`
							Unit  string  `json:"unit"`
						} `json:"lineHeight"`
						LetterSpacing struct {
							Value float64 `json:"value"`
							Unit  string  `json:"unit"`
						} `json:"letterSpacing"`
					} `json:"style"`
				} `json:"children"`
			} `json:"children"`
		}{
			ID:   fileID,
			Name: "Design System",
			Children: []struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				Type     string `json:"type"`
				Children []struct {
					ID    string `json:"id"`
					Name  string `json:"name"`
					Type  string `json:"type"`
					Fills []struct {
						Type  string `json:"type"`
						Color struct {
							R float64 `json:"r"`
							G float64 `json:"g"`
							B float64 `json:"b"`
							A float64 `json:"a"`
						} `json:"color"`
					} `json:"fills"`
					Strokes []struct {
						Type  string `json:"type"`
						Color struct {
							R float64 `json:"r"`
							G float64 `json:"g"`
							B float64 `json:"b"`
							A float64 `json:"a"`
						} `json:"color"`
					} `json:"strokes"`
					Effects []struct {
						Type   string  `json:"type"`
						Radius float64 `json:"radius"`
						Color  struct {
							R float64 `json:"r"`
							G float64 `json:"g"`
							B float64 `json:"b"`
							A float64 `json:"a"`
						} `json:"color"`
						Offset struct {
							X float64 `json:"x"`
							Y float64 `json:"y"`
						} `json:"offset"`
					} `json:"effects"`
					Style struct {
						FontFamily string  `json:"fontFamily"`
						FontSize   float64 `json:"fontSize"`
						FontWeight float64 `json:"fontWeight"`
						LineHeight struct {
							Value float64 `json:"value"`
							Unit  string  `json:"unit"`
						} `json:"lineHeight"`
						LetterSpacing struct {
							Value float64 `json:"value"`
							Unit  string  `json:"unit"`
						} `json:"letterSpacing"`
					} `json:"style"`
				} `json:"children"`
			}{
				{
					ID:   "colors-frame",
					Name: "Colors",
					Type: "FRAME",
					Children: []struct {
						ID    string `json:"id"`
						Name  string `json:"name"`
						Type  string `json:"type"`
						Fills []struct {
							Type  string `json:"type"`
							Color struct {
								R float64 `json:"r"`
								G float64 `json:"g"`
								B float64 `json:"b"`
								A float64 `json:"a"`
							} `json:"color"`
						} `json:"fills"`
						Strokes []struct {
							Type  string `json:"type"`
							Color struct {
								R float64 `json:"r"`
								G float64 `json:"g"`
								B float64 `json:"b"`
								A float64 `json:"a"`
							} `json:"color"`
						} `json:"strokes"`
						Effects []struct {
							Type   string  `json:"type"`
							Radius float64 `json:"radius"`
							Color  struct {
								R float64 `json:"r"`
								G float64 `json:"g"`
								B float64 `json:"b"`
								A float64 `json:"a"`
							} `json:"color"`
							Offset struct {
								X float64 `json:"x"`
								Y float64 `json:"y"`
							} `json:"offset"`
						} `json:"effects"`
						Style struct {
							FontFamily string  `json:"fontFamily"`
							FontSize   float64 `json:"fontSize"`
							FontWeight float64 `json:"fontWeight"`
							LineHeight struct {
								Value float64 `json:"value"`
								Unit  string  `json:"unit"`
							} `json:"lineHeight"`
							LetterSpacing struct {
								Value float64 `json:"value"`
								Unit  string  `json:"unit"`
							} `json:"letterSpacing"`
						} `json:"style"`
					}{
						{
							ID:   "primary-500",
							Name: "primary-500",
							Type: "RECTANGLE",
							Fills: []struct {
								Type  string `json:"type"`
								Color struct {
									R float64 `json:"r"`
									G float64 `json:"g"`
									B float64 `json:"b"`
									A float64 `json:"a"`
								} `json:"color"`
							}{
								{
									Type: "SOLID",
									Color: struct {
										R float64 `json:"r"`
										G float64 `json:"g"`
										B float64 `json:"b"`
										A float64 `json:"a"`
									}{
										R: 0.4,
										G: 0.49,
										B: 0.92,
										A: 1.0,
									},
								},
							},
						},
						{
							ID:   "secondary-500",
							Name: "secondary-500",
							Type: "RECTANGLE",
							Fills: []struct {
								Type  string `json:"type"`
								Color struct {
									R float64 `json:"r"`
									G float64 `json:"g"`
									B float64 `json:"b"`
									A float64 `json:"a"`
								} `json:"color"`
							}{
								{
									Type: "SOLID",
									Color: struct {
										R float64 `json:"r"`
										G float64 `json:"g"`
										B float64 `json:"b"`
										A float64 `json:"a"`
									}{
										R: 0.94,
										G: 0.58,
										B: 0.98,
										A: 1.0,
									},
								},
							},
						},
					},
				},
			},
		},
		Styles: map[string]struct {
			Key         string `json:"key"`
			Name        string `json:"name"`
			Description string `json:"description"`
			StyleType   string `json:"styleType"`
		}{
			"primary-color": {
				Key:         "primary-500",
				Name:        "Primary 500",
				Description: "Primary brand color",
				StyleType:   "FILL",
			},
		},
	}

	return mockResponse, nil
}

// parseFigmaData converts Figma API response to design tokens
func parseFigmaData(figmaData *FigmaAPIResponse) ([]DesignToken, error) {
	var tokens []DesignToken

	// Parse colors from document structure
	for _, child := range figmaData.Document.Children {
		if child.Name == "Colors" {
			for _, colorChild := range child.Children {
				if len(colorChild.Fills) > 0 && colorChild.Fills[0].Type == "SOLID" {
					color := colorChild.Fills[0].Color
					colorValue := fmt.Sprintf("rgb(%d, %d, %d)",
						int(color.R*255),
						int(color.G*255),
						int(color.B*255),
					)

					tokens = append(tokens, DesignToken{
						Name:     sanitizeTokenName(colorChild.Name),
						Value:    colorValue,
						Type:     "color",
						Category: "colors",
						FigmaID:  colorChild.ID,
					})
				}
			}
		}

		// Parse typography
		if child.Name == "Typography" {
			for _, textChild := range child.Children {
				if textChild.Style.FontFamily != "" {
					typographyValue, _ := json.Marshal(map[string]interface{}{
						"fontFamily":    textChild.Style.FontFamily,
						"fontSize":      textChild.Style.FontSize,
						"fontWeight":    textChild.Style.FontWeight,
						"lineHeight":    textChild.Style.LineHeight.Value,
						"letterSpacing": textChild.Style.LetterSpacing.Value,
					})

					tokens = append(tokens, DesignToken{
						Name:     sanitizeTokenName(textChild.Name),
						Value:    string(typographyValue),
						Type:     "typography",
						Category: "typography",
						FigmaID:  textChild.ID,
					})
				}
			}
		}

		// Parse shadows/effects
		if child.Name == "Effects" {
			for _, effectChild := range child.Children {
				if len(effectChild.Effects) > 0 {
					shadowValue, _ := json.Marshal(effectChild.Effects)

					tokens = append(tokens, DesignToken{
						Name:     sanitizeTokenName(effectChild.Name),
						Value:    string(shadowValue),
						Type:     "shadow",
						Category: "shadows",
						FigmaID:  effectChild.ID,
					})
				}
			}
		}
	}

	return tokens, nil
}

// sanitizeTokenName converts Figma names to CSS-friendly token names
func sanitizeTokenName(name string) string {
	// Convert to lowercase and replace spaces/special chars with hyphens
	result := ""
	for _, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			result += string(char)
		} else if char >= 'A' && char <= 'Z' {
			result += string(char + 32) // Convert to lowercase
		} else if char == ' ' || char == '_' || char == '.' {
			result += "-"
		}
	}
	return result
}

// getThemeName returns theme name with fallback
func getThemeName(customName string, figmaData *FigmaAPIResponse) string {
	if customName != "" {
		return customName
	}
	if figmaData.Document.Name != "" {
		return figmaData.Document.Name
	}
	return "Figma Theme " + time.Now().Format("2006-01-02")
}

// getThemeDescription returns theme description with fallback
func getThemeDescription(customDescription string, figmaData *FigmaAPIResponse) string {
	if customDescription != "" {
		return customDescription
	}
	return fmt.Sprintf("Design system imported from Figma on %s", time.Now().Format("January 2, 2006"))
}

// Real Figma API integration (commented out for mock implementation)
/*
func fetchFigmaDataReal(fileID, nodeID, accessToken string) (*FigmaAPIResponse, error) {
	url := fmt.Sprintf("https://api.figma.com/v1/files/%s", fileID)
	if nodeID != "" {
		url += fmt.Sprintf("/nodes?ids=%s", nodeID)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Figma-Token", accessToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Figma API error: %s", string(body))
	}

	var figmaResponse FigmaAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&figmaResponse); err != nil {
		return nil, err
	}

	return &figmaResponse, nil
}
*/
