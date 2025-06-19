package routes

import (
	"net/http"
	"strconv"
	"time"

	"bome-backend/internal/middleware"
	"bome-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupAdvertisementRoutes configures advertisement-related routes
func SetupAdvertisementRoutes(
	router *gin.RouterGroup,
	adService *services.AdvertisementService,
) {
	// Advertiser routes (require authentication)
	advertiser := router.Group("/advertiser")
	advertiser.Use(middleware.AuthMiddleware())
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
	admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
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
	{
		// Ad serving
		public.GET("/serve/:placementId", serveAdHandler(adService))
		public.POST("/impression/:adId", recordImpressionHandler(adService))
		public.POST("/click/:adId", recordClickHandler(adService))
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

		// Return the first ad (highest priority + random)
		var ad interface{}
		if len(ads) > 0 {
			ad = ads[0]
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    ad,
		})
	}
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
