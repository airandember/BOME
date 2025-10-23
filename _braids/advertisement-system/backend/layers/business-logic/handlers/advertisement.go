package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"bome-backend/internal/database"
	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// GetRevenueAnalyticsHandler handles retrieving comprehensive revenue analytics
func GetRevenueAnalyticsHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		period := c.DefaultQuery("period", "30d")
		campaignID := c.Query("campaign_id")

		// Track analytics access
		trackAdAnalyticsAccess(c, "revenue_analytics", map[string]interface{}{
			"period":      period,
			"campaign_id": campaignID,
			"timestamp":   time.Now().Unix(),
		})

		analytics := map[string]interface{}{
			"period":      period,
			"campaign_id": campaignID,
			"revenue_metrics": map[string]interface{}{
				"total_revenue":            getTotalAdRevenue(adService, period, campaignID),
				"revenue_by_campaign":      getRevenueByCampaign(adService, period),
				"revenue_by_ad":            getRevenueByAd(adService, period, campaignID),
				"revenue_trends":           getRevenueTrends(adService, period),
				"average_revenue_per_user": getAverageRevenuePerUser(adService, period),
			},
			"performance_metrics": map[string]interface{}{
				"top_performing_campaigns": getTopPerformingCampaigns(adService, period),
				"top_performing_ads":       getTopPerformingAds(adService, period, campaignID),
				"conversion_funnel":        getAdConversionFunnel(adService, period),
				"attribution_analysis":     getAttributionAnalysis(adService, period),
			},
			"optimization_insights": map[string]interface{}{
				"optimal_timing":        getOptimalAdTiming(adService, period),
				"audience_segments":     getAudienceSegmentPerformance(adService, period),
				"placement_performance": getPlacementPerformance(adService, period),
				"recommendations":       getAdOptimizationRecommendations(adService, period),
			},
		}

		c.JSON(http.StatusOK, gin.H{"data": analytics})
	}
}

// SetupAdvertisementRoutes configures advertisement-related routes
func SetupAdvertisementRoutes(
	router *gin.RouterGroup,
	adService *services.AdvertisementService,
) {
	fmt.Printf("Setting up advertisement routes with router group: %s\n", router.BasePath())
	// Advertiser routes (require advertiser role)
	advertiser := router.Group("/advertiser")
	advertiser.Use(middleware.AuthRequired(), RoleRequired(
		"advertiser",            // Level 3: Advertiser
		"marketing_specialist",  // Level 4: Marketing Specialist
		"advertisement_manager", // Level 7: Advertisement Manager
		"super_admin",           // Level 10: Super Administrator
		"system_admin",          // Level 9: System Administrator
		"content_manager",       // Level 8: Content Manager
		"articles_manager",      // Level 7: Articles Manager
		"youtube_manager",       // Level 7: YouTube Manager
		"streaming_manager",     // Level 7: Video Streaming Manager
		"events_manager",        // Level 7: Events Manager
		"user_manager",          // Level 7: User Account Manager
		"analytics_manager",     // Level 7: Analytics Manager
		"financial_admin",       // Level 7: Financial Administrator
		"admin",                 // Legacy admin role
	))
	{
		// Advertiser account management
		advertiser.POST("/account", createAdvertiserAccountHandler(adService))
		advertiser.GET("/account", getAdvertiserAccountHandler(adService))

		// Campaign management
		advertiser.POST("/campaigns", createCampaignHandler(adService))
		advertiser.GET("/campaigns", getCampaignsHandler(adService))
		advertiser.GET("/campaigns/:id", getCampaignHandler(adService))

		// Advertisement management
		advertiser.POST("/campaigns/:campaignId/ads", createAdvertisementHandler(adService))
		advertiser.GET("/ads/:id", getAdvertisementHandler(adService))

		// Analytics
		advertiser.GET("/ads/:id/analytics", getAdAnalyticsHandler(adService))
		advertiser.GET("/campaigns/:id/analytics", getCampaignAnalyticsHandler(adService))
	}

	// Admin routes (require admin authentication)
	admin := router.Group("/admin/ads")
	admin.Use(middleware.AuthRequired(), middleware.AdminRequired())
	{
		// Advertiser account management
		admin.GET("/advertisers", getAdvertisersHandler(adService))
		admin.GET("/advertisers/:id", getAdvertiserHandler(adService))
		admin.POST("/advertisers/:id/approve", approveAdvertiserHandler(adService))
		admin.POST("/advertisers/:id/reject", rejectAdvertiserHandler(adService))

		// Campaign management
		admin.GET("/campaigns", getAllCampaignsHandler(adService))
		admin.GET("/campaigns/:id", getAdminCampaignHandler(adService))
		admin.POST("/campaigns/:id/approve", approveCampaignHandler(adService))
		admin.POST("/campaigns/:id/reject", rejectCampaignHandler(adService))

		// Ad placement management
		admin.GET("/placements", getPlacementsHandler(adService))
		admin.POST("/placements", createPlacementHandler(adService))
		admin.PUT("/placements/:id", updatePlacementHandler(adService))
	}

	// Public routes (no authentication required)
	public := router.Group("/ads")
	fmt.Printf("Setting up public ads routes with base path: %s\n", public.BasePath())
	{
		// Test route
		public.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Ads routes are working!"})
		})

		// Ad serving
		public.GET("/serve/:placementId", serveAdHandler(adService))
		public.POST("/impression/:adId", recordImpressionHandler(adService))
		public.POST("/click/:adId", recordClickHandler(adService))
		fmt.Printf("Registered public ad routes: /test, /serve/:placementId, /impression/:adId, /click/:adId\n")
	}
}

// Advertiser account handlers
func createAdvertiserAccountHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var req services.CreateAdvertiserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		account, err := adService.CreateAdvertiserAccount(userID.(int), &req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    account,
		})
	}
}

func getAdvertiserAccountHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		account, err := adService.GetAdvertiserAccountByUserID(userID.(int))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Advertiser account not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    account,
		})
	}
}

// Campaign handlers
func createCampaignHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Get advertiser account
		account, err := adService.GetAdvertiserAccountByUserID(userID.(int))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Advertiser account not found"})
			return
		}

		var req services.CreateCampaignRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		campaign, err := adService.CreateAdCampaign(account.ID, &req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    campaign,
		})
	}
}

func getCampaignsHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Get advertiser account
		account, err := adService.GetAdvertiserAccountByUserID(userID.(int))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Advertiser account not found"})
			return
		}

		campaigns, err := adService.GetCampaignsByAdvertiser(account.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    campaigns,
		})
	}
}

func getCampaignHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		campaignID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID"})
			return
		}

		campaign, err := adService.GetAdCampaignByID(campaignID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
			return
		}

		// Verify ownership
		account, err := adService.GetAdvertiserAccountByUserID(userID.(int))
		if err != nil || account.ID != campaign.AdvertiserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    campaign,
		})
	}
}

// Advertisement handlers
func createAdvertisementHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		campaignID, err := strconv.Atoi(c.Param("campaignId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID"})
			return
		}

		// Verify campaign ownership
		campaign, err := adService.GetAdCampaignByID(campaignID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
			return
		}

		account, err := adService.GetAdvertiserAccountByUserID(userID.(int))
		if err != nil || account.ID != campaign.AdvertiserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		var req services.CreateAdRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		ad, err := adService.CreateAdvertisement(campaignID, &req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    ad,
		})
	}
}

func getAdvertisementHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		adID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ad ID"})
			return
		}

		ad, err := adService.GetAdvertisementByID(adID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Advertisement not found"})
			return
		}

		// Verify ownership through campaign
		campaign, err := adService.GetAdCampaignByID(ad.CampaignID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
			return
		}

		account, err := adService.GetAdvertiserAccountByUserID(userID.(int))
		if err != nil || account.ID != campaign.AdvertiserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    ad,
		})
	}
}

// Analytics handlers
func getAdAnalyticsHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		adID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ad ID"})
			return
		}

		// Parse date range
		startDate, endDate := parseDateRange(c)

		// Verify ownership
		ad, err := adService.GetAdvertisementByID(adID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Advertisement not found"})
			return
		}

		campaign, err := adService.GetAdCampaignByID(ad.CampaignID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
			return
		}

		account, err := adService.GetAdvertiserAccountByUserID(userID.(int))
		if err != nil || account.ID != campaign.AdvertiserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		analytics, err := adService.GetAdAnalytics(adID, startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    analytics,
		})
	}
}

func getCampaignAnalyticsHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		campaignID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID"})
			return
		}

		// Parse date range
		startDate, endDate := parseDateRange(c)

		// Verify ownership
		campaign, err := adService.GetAdCampaignByID(campaignID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
			return
		}

		account, err := adService.GetAdvertiserAccountByUserID(userID.(int))
		if err != nil || account.ID != campaign.AdvertiserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		analytics, err := adService.GetCampaignAnalytics(campaignID, startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    analytics,
		})
	}
}

// Admin handlers
func getAdvertisersHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement pagination and filtering
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []interface{}{},
			"message": "Admin advertiser list endpoint - to be implemented",
		})
	}
}

func getAdvertiserHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		advertiserID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid advertiser ID"})
			return
		}

		advertiser, err := adService.GetAdvertiserAccountByID(advertiserID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Advertiser not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    advertiser,
		})
	}
}

func approveAdvertiserHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		advertiserID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid advertiser ID"})
			return
		}

		var req struct {
			Notes string `json:"notes"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		err = adService.ApproveAdvertiserAccount(advertiserID, req.Notes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Advertiser approved successfully",
		})
	}
}

func rejectAdvertiserHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		advertiserID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid advertiser ID"})
			return
		}

		var req struct {
			Notes string `json:"notes"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		err = adService.RejectAdvertiserAccount(advertiserID, req.Notes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Advertiser rejected successfully",
		})
	}
}

func getAllCampaignsHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement admin campaign listing with pagination and filtering
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []interface{}{},
			"message": "Admin campaign list endpoint - to be implemented",
		})
	}
}

func getAdminCampaignHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		campaignID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID"})
			return
		}

		campaign, err := adService.GetAdCampaignByID(campaignID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    campaign,
		})
	}
}

func approveCampaignHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		campaignID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID"})
			return
		}

		var req struct {
			Notes string `json:"notes"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		err = adService.ApproveCampaign(campaignID, adminID.(int), req.Notes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Campaign approved successfully",
		})
	}
}

func rejectCampaignHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		campaignID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID"})
			return
		}

		var req struct {
			Notes string `json:"notes"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		err = adService.RejectCampaign(campaignID, adminID.(int), req.Notes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Campaign rejected successfully",
		})
	}
}

// Public ad serving handlers
func serveAdHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		placementID, err := strconv.Atoi(c.Param("placementId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid placement ID"})
			return
		}

		ads, err := adService.GetActiveAdsForPlacement(placementID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get placement information
		placement, err := adService.GetPlacementByID(placementID)
		if err != nil {
			// If placement doesn't exist, create a fallback placement
			placement = &database.AdPlacement{
				ID:          placementID,
				Name:        fmt.Sprintf("Placement %d", placementID),
				Description: "Default advertisement placement",
				Location:    "content",
				AdType:      "banner",
				MaxWidth:    728,
				MaxHeight:   90,
				BaseRate:    100.00,
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
		}

		// Return the first ad (highest priority + random) with proper response structure
		var ad interface{}
		if len(ads) > 0 {
			ad = ads[0]
		}

		// Create the proper response structure that matches AdServeResponse
		responseData := gin.H{
			"ad":        ad,
			"placement": placement,
			"tracking_data": gin.H{
				"impression_url": fmt.Sprintf("/api/v1/ads/impression/%d", getAdID(ad)),
				"click_url":      fmt.Sprintf("/api/v1/ads/click/%d", getAdID(ad)),
				"view_tracking":  true,
			},
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    responseData,
		})
	}
}

// Helper function to extract ad ID from ad interface
func getAdID(ad interface{}) int {
	if ad == nil {
		return 0
	}

	// Type assertion to get the ID field
	if adMap, ok := ad.(map[string]interface{}); ok {
		if id, exists := adMap["id"]; exists {
			if idInt, ok := id.(int); ok {
				return idInt
			}
			if idFloat, ok := id.(float64); ok {
				return int(idFloat)
			}
		}
	}

	return 0
}

func recordImpressionHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		adID, err := strconv.Atoi(c.Param("adId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ad ID"})
			return
		}

		var req struct {
			ViewDuration int `json:"view_duration"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			req.ViewDuration = 0 // Default to 0 if not provided
		}

		// Get user ID if authenticated (optional)
		var userID *int
		if uid, exists := c.Get("user_id"); exists {
			uidInt := uid.(int)
			userID = &uidInt
		}

		// Get client info
		ipAddress := getClientIP(c)
		userAgent := c.GetHeader("User-Agent")

		err = adService.RecordAdImpression(adID, userID, ipAddress, userAgent, req.ViewDuration)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Impression recorded",
		})
	}
}

func recordClickHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		adID, err := strconv.Atoi(c.Param("adId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ad ID"})
			return
		}

		// Get user ID if authenticated (optional)
		var userID *int
		if uid, exists := c.Get("user_id"); exists {
			uidInt := uid.(int)
			userID = &uidInt
		}

		// Get client info
		ipAddress := getClientIP(c)
		userAgent := c.GetHeader("User-Agent")
		referrer := c.GetHeader("Referer")

		err = adService.RecordAdClick(adID, userID, ipAddress, userAgent, referrer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Click recorded",
		})
	}
}

// Placement management handlers
func getPlacementsHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement placement listing
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []interface{}{},
			"message": "Placement list endpoint - to be implemented",
		})
	}
}

func createPlacementHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement placement creation
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Placement creation endpoint - to be implemented",
		})
	}
}

func updatePlacementHandler(adService *services.AdvertisementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement placement update
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Placement update endpoint - to be implemented",
		})
	}
}

// Helper functions
func parseDateRange(c *gin.Context) (startDate, endDate time.Time) {
	// Default to last 30 days
	endDate = time.Now()
	startDate = endDate.AddDate(0, 0, -30)

	if start := c.Query("start_date"); start != "" {
		if parsed, err := time.Parse("2006-01-02", start); err == nil {
			startDate = parsed
		}
	}

	if end := c.Query("end_date"); end != "" {
		if parsed, err := time.Parse("2006-01-02", end); err == nil {
			endDate = parsed
		}
	}

	return startDate, endDate
}

func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to ClientIP
	return c.ClientIP()
}

// Revenue Analytics Helper Functions

func getTotalAdRevenue(adService *services.AdvertisementService, period string, campaignID string) float64 {
	// This would query the database for total revenue in the period
	// For now, return a mock value
	return 1250.75
}

func getRevenueByCampaign(adService *services.AdvertisementService, period string) map[string]float64 {
	// This would query campaign revenue breakdown
	return map[string]float64{
		"campaign_1": 450.25,
		"campaign_2": 380.50,
		"campaign_3": 420.00,
	}
}

func getRevenueByAd(adService *services.AdvertisementService, period string, campaignID string) map[string]float64 {
	// This would query ad-level revenue breakdown
	return map[string]float64{
		"ad_1": 150.25,
		"ad_2": 200.50,
		"ad_3": 100.00,
	}
}

func getRevenueTrends(adService *services.AdvertisementService, period string) []map[string]interface{} {
	// This would query daily/weekly revenue trends
	return []map[string]interface{}{
		{"date": "2025-01-01", "revenue": 45.25},
		{"date": "2025-01-02", "revenue": 52.75},
		{"date": "2025-01-03", "revenue": 48.50},
		{"date": "2025-01-04", "revenue": 61.25},
		{"date": "2025-01-05", "revenue": 55.00},
	}
}

func getAverageRevenuePerUser(adService *services.AdvertisementService, period string) float64 {
	// This would calculate ARPU from ad interactions
	return 2.45
}

func getTopPerformingCampaigns(adService *services.AdvertisementService, period string) []map[string]interface{} {
	// This would query top campaigns by revenue/performance
	return []map[string]interface{}{
		{
			"campaign_id":     1,
			"name":            "Holiday Sale Campaign",
			"revenue":         450.25,
			"ctr":             3.2,
			"conversion_rate": 8.5,
		},
		{
			"campaign_id":     2,
			"name":            "Product Launch",
			"revenue":         380.50,
			"ctr":             2.8,
			"conversion_rate": 7.2,
		},
	}
}

func getTopPerformingAds(adService *services.AdvertisementService, period string, campaignID string) []map[string]interface{} {
	// This would query top ads by performance
	return []map[string]interface{}{
		{
			"ad_id":       1,
			"title":       "Premium Subscription Ad",
			"revenue":     200.50,
			"ctr":         4.1,
			"impressions": 5000,
		},
		{
			"ad_id":       2,
			"title":       "Free Trial Offer",
			"revenue":     150.25,
			"ctr":         3.8,
			"impressions": 4000,
		},
	}
}

func getAdConversionFunnel(adService *services.AdvertisementService, period string) map[string]int {
	// This would analyze the conversion funnel from impression to purchase
	return map[string]int{
		"impressions":        10000,
		"clicks":             320,
		"landing_page_views": 280,
		"signups":            45,
		"conversions":        12,
	}
}

func getAttributionAnalysis(adService *services.AdvertisementService, period string) map[string]interface{} {
	// This would analyze attribution across different touchpoints
	return map[string]interface{}{
		"first_touch": map[string]float64{
			"social_media": 35.2,
			"search":       28.5,
			"direct":       22.1,
			"email":        14.2,
		},
		"last_touch": map[string]float64{
			"search":       42.3,
			"direct":       31.7,
			"social_media": 18.5,
			"email":        7.5,
		},
		"multi_touch": map[string]float64{
			"search_social": 15.8,
			"email_search":  12.3,
			"direct_social": 8.9,
		},
	}
}

func getOptimalAdTiming(adService *services.AdvertisementService, period string) map[string]interface{} {
	// This would analyze optimal timing for ad delivery
	return map[string]interface{}{
		"best_hours": []int{9, 10, 11, 14, 15, 16, 19, 20},
		"best_days":  []string{"Tuesday", "Wednesday", "Thursday"},
		"peak_performance": map[string]interface{}{
			"hour":           15,
			"day":            "Wednesday",
			"ctr_multiplier": 1.8,
		},
	}
}

func getAudienceSegmentPerformance(adService *services.AdvertisementService, period string) map[string]interface{} {
	// This would analyze performance by audience segments
	return map[string]interface{}{
		"age_groups": map[string]float64{
			"18-24": 2.1,
			"25-34": 3.8,
			"35-44": 4.2,
			"45-54": 3.5,
			"55+":   2.8,
		},
		"interests": map[string]float64{
			"technology":    4.5,
			"entertainment": 3.8,
			"lifestyle":     3.2,
			"business":      4.1,
		},
		"geographic": map[string]float64{
			"North America": 4.2,
			"Europe":        3.8,
			"Asia":          3.5,
			"Other":         2.9,
		},
	}
}

func getPlacementPerformance(adService *services.AdvertisementService, period string) map[string]interface{} {
	// This would analyze performance by ad placement
	return map[string]interface{}{
		"homepage_banner": map[string]interface{}{
			"revenue":     450.25,
			"ctr":         2.8,
			"impressions": 15000,
		},
		"video_pre_roll": map[string]interface{}{
			"revenue":     380.50,
			"ctr":         4.2,
			"impressions": 9000,
		},
		"sidebar": map[string]interface{}{
			"revenue":     200.00,
			"ctr":         1.5,
			"impressions": 13000,
		},
		"in_content": map[string]interface{}{
			"revenue":     220.00,
			"ctr":         3.1,
			"impressions": 7000,
		},
	}
}

func getAdOptimizationRecommendations(adService *services.AdvertisementService, period string) []map[string]interface{} {
	// This would generate optimization recommendations
	return []map[string]interface{}{
		{
			"type":             "timing",
			"priority":         "high",
			"title":            "Increase ad frequency during peak hours",
			"description":      "CTR increases by 80% during 2-4 PM",
			"potential_impact": "+25% revenue",
		},
		{
			"type":             "audience",
			"priority":         "medium",
			"title":            "Expand targeting to 25-34 age group",
			"description":      "This segment shows highest conversion rates",
			"potential_impact": "+15% conversions",
		},
		{
			"type":             "creative",
			"priority":         "low",
			"title":            "A/B test new ad creative",
			"description":      "Current creative has been running for 30 days",
			"potential_impact": "+10% CTR",
		},
	}
}

func trackAdAnalyticsAccess(c *gin.Context, activity string, metadata map[string]interface{}) {
	// This would track analytics access for audit purposes
	userID, exists := c.Get("user_id")
	if !exists {
		return
	}

	// Log analytics access
	fmt.Printf("Analytics access: User %v accessed %s with metadata %+v\n", userID, activity, metadata)
}
