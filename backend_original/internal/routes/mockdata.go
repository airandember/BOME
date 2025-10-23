package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Video represents a video with comprehensive metadata
type Video struct {
	ID           int      `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	ThumbnailURL string   `json:"thumbnailUrl"`
	VideoURL     string   `json:"videoUrl"`
	Duration     int      `json:"duration"`
	ViewCount    int      `json:"viewCount"`
	LikeCount    int      `json:"likeCount"`
	Category     string   `json:"category"`
	Tags         []string `json:"tags"`
	CreatedAt    string   `json:"createdAt"`
	UpdatedAt    string   `json:"updatedAt"`
}

// VideoCategory represents a video category
type VideoCategory struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	VideoCount  int    `json:"videoCount"`
}

// VideoComment represents a video comment
type VideoComment struct {
	ID        int    `json:"id"`
	VideoID   int    `json:"videoId"`
	UserName  string `json:"userName"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

// Article represents a blog article
type Article struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Content     string   `json:"content"`
	Excerpt     string   `json:"excerpt"`
	FeaturedImg string   `json:"featuredImg"`
	CategoryID  int      `json:"categoryId"`
	Category    string   `json:"category"`
	AuthorID    int      `json:"authorId"`
	Author      Author   `json:"author"`
	Tags        []string `json:"tags"`
	Featured    bool     `json:"featured"`
	Published   bool     `json:"published"`
	ViewCount   int      `json:"viewCount"`
	ReadTime    int      `json:"readTime"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

// Author represents an article author
type Author struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Bio      string `json:"bio"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
	Verified bool   `json:"verified"`
}

// DashboardData represents user dashboard data
type DashboardData struct {
	Stats struct {
		TotalWatchTime  int `json:"totalWatchTime"`
		VideosWatched   int `json:"videosWatched"`
		FavoriteVideos  int `json:"favoriteVideos"`
		CompletedSeries int `json:"completedSeries"`
	} `json:"stats"`
	RecentActivity    []RecentActivity `json:"recentActivity"`
	RecommendedVideos []Video          `json:"recommendedVideos"`
	FavoriteVideos    []Video          `json:"favoriteVideos"`
	ContinueWatching  []VideoProgress  `json:"continueWatching"`
}

// RecentActivity represents recent user activity
type RecentActivity struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	Timestamp string `json:"timestamp"`
}

// VideoProgress represents video progress for continue watching
type VideoProgress struct {
	Video       Video   `json:"video"`
	Progress    float64 `json:"progress"`
	LastWatched string  `json:"lastWatched"`
}

// Advertisement Mock Data Structures
type AdvertiserAccount struct {
	ID              int    `json:"id"`
	UserID          int    `json:"user_id"`
	CompanyName     string `json:"company_name"`
	BusinessEmail   string `json:"business_email"`
	ContactName     string `json:"contact_name"`
	ContactPhone    string `json:"contact_phone,omitempty"`
	BusinessAddress string `json:"business_address,omitempty"`
	TaxID           string `json:"tax_id,omitempty"`
	Website         string `json:"website,omitempty"`
	Industry        string `json:"industry,omitempty"`
	Status          string `json:"status"`
	ApprovedBy      *int   `json:"approved_by,omitempty"`
	ApprovedAt      string `json:"approved_at,omitempty"`
	RejectedBy      *int   `json:"rejected_by,omitempty"`
	RejectedAt      string `json:"rejected_at,omitempty"`
	CancelledBy     *int   `json:"cancelled_by,omitempty"`
	CancelledAt     string `json:"cancelled_at,omitempty"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type AdCampaign struct {
	ID             int     `json:"id"`
	AdvertiserID   int     `json:"advertiser_id"`
	Name           string  `json:"name"`
	Description    string  `json:"description,omitempty"`
	Status         string  `json:"status"`
	StartDate      string  `json:"start_date"`
	EndDate        string  `json:"end_date,omitempty"`
	Budget         float64 `json:"budget"`
	SpentAmount    float64 `json:"spent_amount"`
	TargetAudience string  `json:"target_audience,omitempty"`
	BillingType    string  `json:"billing_type"`
	BillingRate    float64 `json:"billing_rate"`
	ApprovalNotes  string  `json:"approval_notes,omitempty"`
	ApprovedBy     *int    `json:"approved_by,omitempty"`
	ApprovedAt     string  `json:"approved_at,omitempty"`
	RejectedBy     *int    `json:"rejected_by,omitempty"`
	RejectedAt     string  `json:"rejected_at,omitempty"`
	CancelledBy    *int    `json:"cancelled_by,omitempty"`
	CancelledAt    string  `json:"cancelled_at,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

type AdminAnalytics struct {
	Overview struct {
		TotalUsers     int     `json:"totalUsers"`
		ActiveUsers    int     `json:"activeUsers"`
		TotalVideos    int     `json:"totalVideos"`
		TotalRevenue   float64 `json:"totalRevenue"`
		MonthlyRevenue float64 `json:"monthlyRevenue"`
	} `json:"overview"`
	Advertisement struct {
		TotalAdvertisers int     `json:"totalAdvertisers"`
		ActiveCampaigns  int     `json:"activeCampaigns"`
		PendingApprovals int     `json:"pendingApprovals"`
		MonthlyAdRevenue float64 `json:"monthlyAdRevenue"`
		TopPerformingAds int     `json:"topPerformingAds"`
	} `json:"advertisement"`
	UserEngagement struct {
		AverageWatchTime int     `json:"averageWatchTime"`
		EngagementRate   float64 `json:"engagementRate"`
		RetentionRate    float64 `json:"retentionRate"`
	} `json:"userEngagement"`
}

type UserProfile struct {
	ID               int    `json:"id"`
	Username         string `json:"username"`
	Email            string `json:"email"`
	FullName         string `json:"full_name"`
	Avatar           string `json:"avatar"`
	Role             string `json:"role"`
	SubscriptionTier string `json:"subscription_tier"`
	IsActive         bool   `json:"is_active"`
	LastLogin        string `json:"last_login"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// COMPREHENSIVE MOCK DATA DEFINITIONS

var MOCK_VIDEOS = []Video{
	{
		ID:           1,
		Title:        "Archaeological Evidence for the Book of Mormon",
		Description:  "Exploring recent archaeological discoveries that support Book of Mormon narratives, including ancient civilizations, metallurgy, and cultural practices found in the Americas.",
		ThumbnailURL: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		VideoURL:     "https://vz-12345678-123.b-cdn.net/videos/archaeological-evidence-bom/playlist.m3u8",
		Duration:     942,
		ViewCount:    24567,
		LikeCount:    1234,
		Category:     "Archaeology",
		Tags:         []string{"archaeology", "evidence", "ancient-america", "civilizations"},
		CreatedAt:    "2024-01-15T10:30:00Z",
		UpdatedAt:    "2024-01-15T10:30:00Z",
	},
	{
		ID:           2,
		Title:        "DNA and the Book of Mormon: Scientific Perspectives",
		Description:  "A comprehensive look at DNA evidence and its relationship to Book of Mormon populations, examining recent genetic studies and their implications.",
		ThumbnailURL: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		VideoURL:     "https://vz-12345678-123.b-cdn.net/videos/dna-book-mormon-science/playlist.m3u8",
		Duration:     1335,
		ViewCount:    18934,
		LikeCount:    892,
		Category:     "Science",
		Tags:         []string{"dna", "science", "genetics", "populations"},
		CreatedAt:    "2024-01-18T14:20:00Z",
		UpdatedAt:    "2024-01-18T14:20:00Z",
	},
	{
		ID:           3,
		Title:        "Mesoamerican Connections to Book of Mormon Geography",
		Description:  "Examining cultural and geographical connections between Mesoamerica and the Book of Mormon, including recent discoveries and scholarly research.",
		ThumbnailURL: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		VideoURL:     "https://vz-12345678-123.b-cdn.net/videos/mesoamerican-connections/playlist.m3u8",
		Duration:     1113,
		ViewCount:    31245,
		LikeCount:    1567,
		Category:     "Geography",
		Tags:         []string{"mesoamerica", "geography", "culture", "maya"},
		CreatedAt:    "2024-01-20T09:45:00Z",
		UpdatedAt:    "2024-01-20T09:45:00Z",
	},
	// Adding more videos...
	{
		ID:           4,
		Title:        "Linguistic Analysis of Book of Mormon Names",
		Description:  "Scholarly analysis of Hebrew and Egyptian linguistic patterns in Book of Mormon names and their ancient Near Eastern connections.",
		ThumbnailURL: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		VideoURL:     "https://vz-12345678-123.b-cdn.net/videos/linguistic-analysis-names/playlist.m3u8",
		Duration:     1518,
		ViewCount:    15678,
		LikeCount:    743,
		Category:     "Linguistics",
		Tags:         []string{"linguistics", "hebrew", "names", "ancient-languages"},
		CreatedAt:    "2024-01-22T16:12:00Z",
		UpdatedAt:    "2024-01-22T16:12:00Z",
	},
	{
		ID:           5,
		Title:        "Metallurgy in Ancient America: Book of Mormon Evidence",
		Description:  "Evidence of advanced metallurgy in pre-Columbian America and its relationship to Book of Mormon descriptions of metalworking.",
		ThumbnailURL: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		VideoURL:     "https://vz-12345678-123.b-cdn.net/videos/ancient-metallurgy/playlist.m3u8",
		Duration:     1267,
		ViewCount:    12456,
		LikeCount:    623,
		Category:     "Archaeology",
		Tags:         []string{"metallurgy", "ancient-technology", "archaeology", "metals"},
		CreatedAt:    "2024-01-25T11:30:00Z",
		UpdatedAt:    "2024-01-25T11:30:00Z",
	},
}

var MOCK_CATEGORIES = []VideoCategory{
	{ID: 1, Name: "Archaeology", Description: "Archaeological evidence and discoveries supporting Book of Mormon narratives", VideoCount: 4},
	{ID: 2, Name: "Science", Description: "Scientific perspectives and research related to Book of Mormon claims", VideoCount: 5},
	{ID: 3, Name: "Geography", Description: "Geographical studies and location theories for Book of Mormon lands", VideoCount: 3},
	{ID: 4, Name: "Linguistics", Description: "Language analysis and linguistic evidence from the Book of Mormon", VideoCount: 3},
	{ID: 5, Name: "History", Description: "Historical context and parallels to Book of Mormon events", VideoCount: 3},
	{ID: 6, Name: "Culture", Description: "Cultural practices and social structures in ancient America", VideoCount: 1},
	{ID: 7, Name: "Agriculture", Description: "Agricultural techniques and practices in ancient American civilizations", VideoCount: 1},
	{ID: 8, Name: "Economics", Description: "Economic systems and trade networks in ancient America", VideoCount: 2},
	{ID: 9, Name: "Religion", Description: "Religious practices and beliefs in ancient American cultures", VideoCount: 2},
	{ID: 10, Name: "Technology", Description: "Technological achievements and innovations in ancient America", VideoCount: 1},
	{ID: 11, Name: "Zoology", Description: "Animal life and fauna mentioned in the Book of Mormon", VideoCount: 1},
	{ID: 12, Name: "Geology", Description: "Geological evidence and natural phenomena in Book of Mormon accounts", VideoCount: 1},
}

var MOCK_COMMENTS = []VideoComment{
	{ID: 1, VideoID: 1, UserName: "Dr. Sarah Mitchell", Content: "Excellent analysis of the archaeological evidence. The connections to ancient civilizations are particularly compelling.", CreatedAt: "2024-03-28T14:30:00Z"},
	{ID: 2, VideoID: 1, UserName: "Michael Johnson", Content: "This really helps me understand the historical context better. Thank you for this thorough presentation.", CreatedAt: "2024-03-29T09:15:00Z"},
	{ID: 3, VideoID: 2, UserName: "Dr. Robert Chen", Content: "The DNA research findings are fascinating. It's important to consider all perspectives in this discussion.", CreatedAt: "2024-03-30T16:45:00Z"},
}

// MOCK DATA ARRAYS
var MOCK_ADVERTISER_ACCOUNTS = []AdvertiserAccount{
	{
		ID: 1, UserID: 101, CompanyName: "Sacred Sites Tourism", BusinessEmail: "marketing@sacredsites.com",
		ContactName: "Jennifer Martinez", ContactPhone: "+1-555-0123", Industry: "Tourism",
		Status: "approved", ApprovedBy: intPtr(1), ApprovedAt: "2024-03-15T10:30:00Z",
		CreatedAt: "2024-03-10T14:20:00Z", UpdatedAt: "2024-03-15T10:30:00Z",
	},
	{
		ID: 2, UserID: 102, CompanyName: "Covenant Communications", BusinessEmail: "ads@covenant-comm.com",
		ContactName: "Robert Thompson", ContactPhone: "+1-555-0234", Industry: "Publishing",
		Status: "approved", ApprovedBy: intPtr(1), ApprovedAt: "2024-03-20T16:45:00Z",
		CreatedAt: "2024-03-18T09:15:00Z", UpdatedAt: "2024-03-20T16:45:00Z",
	},
	{
		ID: 3, UserID: 103, CompanyName: "LDS Family Services", BusinessEmail: "outreach@ldsfamilyservices.org",
		ContactName: "Sarah Williams", ContactPhone: "+1-555-0345", Industry: "Non-profit",
		Status: "pending", CreatedAt: "2024-03-25T11:00:00Z", UpdatedAt: "2024-03-25T11:00:00Z",
	},
	{
		ID: 4, UserID: 104, CompanyName: "Deseret Book Company", BusinessEmail: "marketing@deseretbook.com",
		ContactName: "Michael Davis", ContactPhone: "+1-555-0456", Industry: "Retail",
		Status: "approved", ApprovedBy: intPtr(1), ApprovedAt: "2024-03-22T13:20:00Z",
		CreatedAt: "2024-03-20T08:30:00Z", UpdatedAt: "2024-03-22T13:20:00Z",
	},
	{
		ID: 5, UserID: 105, CompanyName: "FAIR Mormon", BusinessEmail: "contact@fairmormon.org",
		ContactName: "David Brown", ContactPhone: "+1-555-0567", Industry: "Research",
		Status: "rejected", RejectedBy: intPtr(1), RejectedAt: "2024-03-18T14:15:00Z",
		CreatedAt: "2024-03-15T16:45:00Z", UpdatedAt: "2024-03-18T14:15:00Z",
	},
}

var MOCK_AD_CAMPAIGNS = []AdCampaign{
	{
		ID: 1, AdvertiserID: 1, Name: "Spring Holy Land Tours 2024",
		Description: "Promote spiritual tours to Jerusalem and surrounding biblical sites",
		Status:      "active", StartDate: "2024-04-01T00:00:00Z", EndDate: "2024-06-30T23:59:59Z",
		Budget: 5000.00, SpentAmount: 1250.75, BillingType: "weekly", BillingRate: 200.00,
		ApprovedBy: intPtr(1), ApprovedAt: "2024-03-15T10:35:00Z",
		CreatedAt: "2024-03-15T10:32:00Z", UpdatedAt: "2024-03-30T09:20:00Z",
	},
	{
		ID: 2, AdvertiserID: 2, Name: "New Book Release Campaign",
		Description: "Promoting the release of 'Evidences of the Book of Mormon'",
		Status:      "active", StartDate: "2024-03-20T00:00:00Z", EndDate: "2024-05-20T23:59:59Z",
		Budget: 3000.00, SpentAmount: 890.25, BillingType: "weekly", BillingRate: 150.00,
		ApprovedBy: intPtr(1), ApprovedAt: "2024-03-20T16:50:00Z",
		CreatedAt: "2024-03-20T16:47:00Z", UpdatedAt: "2024-03-29T14:10:00Z",
	},
	{
		ID: 3, AdvertiserID: 4, Name: "Easter Book Collection",
		Description: "Seasonal promotion of Easter and spring-themed religious books",
		Status:      "pending", StartDate: "2024-04-10T00:00:00Z", EndDate: "2024-04-25T23:59:59Z",
		Budget: 2500.00, SpentAmount: 0.00, BillingType: "daily", BillingRate: 75.00,
		CreatedAt: "2024-03-28T11:15:00Z", UpdatedAt: "2024-03-28T11:15:00Z",
	},
}

var MOCK_ADMIN_USERS = []UserProfile{
	{
		ID: 1, Username: "admin", Email: "admin@bome.com", FullName: "System Administrator",
		Avatar: "/avatars/admin.png", Role: "super_admin", SubscriptionTier: "admin",
		IsActive: true, LastLogin: "2024-03-30T16:30:00Z",
		CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-03-30T16:30:00Z",
	},
	{
		ID: 2, Username: "content_manager", Email: "content@bome.com", FullName: "Content Manager",
		Avatar: "/avatars/content.png", Role: "content_manager", SubscriptionTier: "admin",
		IsActive: true, LastLogin: "2024-03-30T14:15:00Z",
		CreatedAt: "2024-01-15T00:00:00Z", UpdatedAt: "2024-03-30T14:15:00Z",
	},
	{
		ID: 3, Username: "ad_manager", Email: "ads@bome.com", FullName: "Advertisement Manager",
		Avatar: "/avatars/ads.png", Role: "advertisement_manager", SubscriptionTier: "admin",
		IsActive: true, LastLogin: "2024-03-30T12:45:00Z",
		CreatedAt: "2024-02-01T00:00:00Z", UpdatedAt: "2024-03-30T12:45:00Z",
	},
}

var MOCK_ADMIN_ANALYTICS = AdminAnalytics{
	Overview: struct {
		TotalUsers     int     `json:"totalUsers"`
		ActiveUsers    int     `json:"activeUsers"`
		TotalVideos    int     `json:"totalVideos"`
		TotalRevenue   float64 `json:"totalRevenue"`
		MonthlyRevenue float64 `json:"monthlyRevenue"`
	}{
		TotalUsers:     1247,
		ActiveUsers:    856,
		TotalVideos:    125,
		TotalRevenue:   45678.90,
		MonthlyRevenue: 12450.75,
	},
	Advertisement: struct {
		TotalAdvertisers int     `json:"totalAdvertisers"`
		ActiveCampaigns  int     `json:"activeCampaigns"`
		PendingApprovals int     `json:"pendingApprovals"`
		MonthlyAdRevenue float64 `json:"monthlyAdRevenue"`
		TopPerformingAds int     `json:"topPerformingAds"`
	}{
		TotalAdvertisers: len(MOCK_ADVERTISER_ACCOUNTS),
		ActiveCampaigns:  2,
		PendingApprovals: 1,
		MonthlyAdRevenue: 8950.25,
		TopPerformingAds: 12,
	},
	UserEngagement: struct {
		AverageWatchTime int     `json:"averageWatchTime"`
		EngagementRate   float64 `json:"engagementRate"`
		RetentionRate    float64 `json:"retentionRate"`
	}{
		AverageWatchTime: 1247,
		EngagementRate:   0.78,
		RetentionRate:    0.85,
	},
}

// Utility function for pointer to int
func intPtr(i int) *int {
	return &i
}

// MOCK DATA ENDPOINTS

// GetVideosHandler returns paginated videos with optional filtering
func GetMockVideosHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	category := c.Query("category")
	search := c.Query("search")

	videos := MOCK_VIDEOS

	// Filter by category if specified
	if category != "" {
		filtered := []Video{}
		for _, video := range videos {
			if video.Category == category {
				filtered = append(filtered, video)
			}
		}
		videos = filtered
	}

	// Filter by search if specified
	if search != "" {
		filtered := []Video{}
		lowerSearch := strings.ToLower(search)
		for _, video := range videos {
			if strings.Contains(strings.ToLower(video.Title), lowerSearch) ||
				strings.Contains(strings.ToLower(video.Description), lowerSearch) {
				filtered = append(filtered, video)
			}
		}
		videos = filtered
	}

	// Calculate pagination
	total := len(videos)
	start := (page - 1) * limit
	end := start + limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedVideos := videos[start:end]

	c.JSON(http.StatusOK, gin.H{
		"videos": paginatedVideos,
		"pagination": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total":        total,
			"total_pages":  (total + limit - 1) / limit,
		},
	})
}

// GetVideoHandler returns a single video by ID
func GetMockVideoHandler(c *gin.Context) {
	videoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	for _, video := range MOCK_VIDEOS {
		if video.ID == videoID {
			c.JSON(http.StatusOK, gin.H{"video": video})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
}

// GetCategoriesHandler returns all video categories
func GetMockCategoriesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"categories": MOCK_CATEGORIES})
}

// GetCommentsHandler returns comments for a video
func GetMockCommentsHandler(c *gin.Context) {
	videoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	var videoComments []VideoComment
	for _, comment := range MOCK_COMMENTS {
		if comment.VideoID == videoID {
			videoComments = append(videoComments, comment)
		}
	}

	c.JSON(http.StatusOK, gin.H{"comments": videoComments})
}

// GetDashboardDataHandler returns user dashboard data
func GetDashboardDataHandler(c *gin.Context) {
	dashboardData := DashboardData{
		Stats: struct {
			TotalWatchTime  int `json:"totalWatchTime"`
			VideosWatched   int `json:"videosWatched"`
			FavoriteVideos  int `json:"favoriteVideos"`
			CompletedSeries int `json:"completedSeries"`
		}{
			TotalWatchTime:  1247,
			VideosWatched:   23,
			FavoriteVideos:  8,
			CompletedSeries: 3,
		},
		RecentActivity: []RecentActivity{
			{Type: "video_watched", Title: "Archaeological Evidence for the Book of Mormon", Timestamp: "2024-03-30T14:30:00Z"},
			{Type: "video_liked", Title: "DNA and the Book of Mormon", Timestamp: "2024-03-29T16:45:00Z"},
			{Type: "video_favorited", Title: "Mesoamerican Connections", Timestamp: "2024-03-28T11:20:00Z"},
		},
		RecommendedVideos: MOCK_VIDEOS[:6],
		FavoriteVideos:    MOCK_VIDEOS[:4],
		ContinueWatching: []VideoProgress{
			{
				Video:       MOCK_VIDEOS[0],
				Progress:    0.65,
				LastWatched: "2024-03-30T14:30:00Z",
			},
			{
				Video:       MOCK_VIDEOS[2],
				Progress:    0.23,
				LastWatched: "2024-03-29T20:15:00Z",
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{"data": dashboardData})
}

// ADVERTISEMENT ENDPOINTS

// GetAdvertiserAccountsHandler returns all advertiser accounts
func GetAdvertiserAccountsHandler(c *gin.Context) {
	status := c.Query("status")
	var filteredAccounts []AdvertiserAccount

	if status != "" {
		for _, account := range MOCK_ADVERTISER_ACCOUNTS {
			if account.Status == status {
				filteredAccounts = append(filteredAccounts, account)
			}
		}
	} else {
		filteredAccounts = MOCK_ADVERTISER_ACCOUNTS
	}

	c.JSON(http.StatusOK, gin.H{
		"advertisers": filteredAccounts,
		"total":       len(filteredAccounts),
	})
}

// GetAdvertiserAccountHandler returns a single advertiser account
func GetAdvertiserAccountHandler(c *gin.Context) {
	advertiserID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid advertiser ID"})
		return
	}

	for _, account := range MOCK_ADVERTISER_ACCOUNTS {
		if account.ID == advertiserID {
			c.JSON(http.StatusOK, gin.H{"advertiser": account})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Advertiser not found"})
}

// GetAdCampaignsHandler returns ad campaigns, optionally filtered by advertiser
func GetAdCampaignsHandler(c *gin.Context) {
	advertiserID := c.Query("advertiser_id")
	status := c.Query("status")
	var filteredCampaigns []AdCampaign

	for _, campaign := range MOCK_AD_CAMPAIGNS {
		includeThis := true

		if advertiserID != "" {
			id, _ := strconv.Atoi(advertiserID)
			if campaign.AdvertiserID != id {
				includeThis = false
			}
		}

		if status != "" && campaign.Status != status {
			includeThis = false
		}

		if includeThis {
			filteredCampaigns = append(filteredCampaigns, campaign)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"campaigns": filteredCampaigns,
		"total":     len(filteredCampaigns),
	})
}

// GetAdCampaignHandler returns a single ad campaign
func GetAdCampaignHandler(c *gin.Context) {
	campaignID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid campaign ID"})
		return
	}

	for _, campaign := range MOCK_AD_CAMPAIGNS {
		if campaign.ID == campaignID {
			c.JSON(http.StatusOK, gin.H{"campaign": campaign})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
}

// GetAdminAnalyticsHandler returns admin analytics data
func GetAdminAnalyticsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"analytics": MOCK_ADMIN_ANALYTICS})
}

// GetAdminUsersHandler returns admin users
func GetAdminUsersHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"users": MOCK_ADMIN_USERS,
		"total": len(MOCK_ADMIN_USERS),
	})
}

// SetupMockDataRoutes sets up all mock data routes
func SetupMockDataRoutes(router *gin.RouterGroup) {
	api := router.Group("/mock")
	{
		// Video endpoints
		api.GET("/videos", GetMockVideosHandler)
		api.GET("/videos/:id", GetMockVideoHandler)
		api.GET("/videos/:id/comments", GetMockCommentsHandler)
		api.GET("/videos/categories", GetMockCategoriesHandler)

		// Dashboard endpoint
		api.GET("/dashboard", GetDashboardDataHandler)

		// Advertisement endpoints
		api.GET("/advertisers", GetAdvertiserAccountsHandler)
		api.GET("/advertisers/:id", GetAdvertiserAccountHandler)
		api.GET("/campaigns", GetAdCampaignsHandler)
		api.GET("/campaigns/:id", GetAdCampaignHandler)
		api.GET("/analytics", GetAdminAnalyticsHandler)
		api.GET("/users", GetAdminUsersHandler)
	}
}
