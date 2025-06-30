package routes

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bome-backend/internal/database"
	"bome-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// UpdateUserRequest represents a user update payload
type UpdateUserRequest struct {
	Role string `json:"role" binding:"required"`
}

// GetUsersHandler handles retrieving users for admin
func GetUsersHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Mock user data for development mode
		if db == nil {
			users := []map[string]interface{}{
				{
					"id":                 1,
					"email":              "admin@bome.test",
					"firstName":          "Test",
					"lastName":           "Administrator",
					"role":               "admin",
					"emailVerified":      true,
					"createdAt":          "2024-01-15T10:30:00Z",
					"lastLogin":          "2024-01-20T14:22:00Z",
					"status":             "active",
					"subscriptionStatus": "premium",
				},
				{
					"id":                 2,
					"email":              "john.doe@example.com",
					"firstName":          "John",
					"lastName":           "Doe",
					"role":               "user",
					"emailVerified":      true,
					"createdAt":          "2024-01-18T09:15:00Z",
					"lastLogin":          "2024-01-20T11:45:00Z",
					"status":             "active",
					"subscriptionStatus": "basic",
				},
				{
					"id":                 3,
					"email":              "jane.smith@example.com",
					"firstName":          "Jane",
					"lastName":           "Smith",
					"role":               "user",
					"emailVerified":      true,
					"createdAt":          "2024-01-19T16:20:00Z",
					"lastLogin":          "2024-01-20T08:30:00Z",
					"status":             "active",
					"subscriptionStatus": "premium",
				},
				{
					"id":                 4,
					"email":              "mike.wilson@example.com",
					"firstName":          "Mike",
					"lastName":           "Wilson",
					"role":               "user",
					"emailVerified":      false,
					"createdAt":          "2024-01-20T12:00:00Z",
					"lastLogin":          nil,
					"status":             "pending",
					"subscriptionStatus": "free",
				},
			}

			// Handle pagination
			page := c.DefaultQuery("page", "1")
			limit := c.DefaultQuery("limit", "10")
			search := c.Query("search")
			// Note: role and status filtering not implemented in mock data

			// Mock filtering (in real implementation, this would be done in the database)
			filteredUsers := users
			if search != "" {
				// Simple search simulation
				filteredUsers = []map[string]interface{}{}
				for _, user := range users {
					if strings.Contains(strings.ToLower(user["email"].(string)), strings.ToLower(search)) ||
						strings.Contains(strings.ToLower(user["firstName"].(string)), strings.ToLower(search)) ||
						strings.Contains(strings.ToLower(user["lastName"].(string)), strings.ToLower(search)) {
						filteredUsers = append(filteredUsers, user)
					}
				}
			}

			c.JSON(http.StatusOK, gin.H{
				"users": filteredUsers,
				"pagination": gin.H{
					"page":       page,
					"limit":      limit,
					"total":      len(filteredUsers),
					"totalPages": 1,
				},
			})
			return
		}

		// Real database implementation
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		search := c.Query("search")
		// Note: role and status not used in current database implementation

		offset := (page - 1) * limit
		users, err := db.GetUsers(limit, offset, "", search)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
			return
		}

		// Calculate total pages (mock for now)
		total := len(users)
		totalPages := (total + limit - 1) / limit

		c.JSON(http.StatusOK, gin.H{
			"users": users,
			"pagination": gin.H{
				"page":       page,
				"limit":      limit,
				"total":      total,
				"totalPages": totalPages,
			},
		})
	}
}

// GetUserHandler handles retrieving a single user for admin
func GetUserHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		user, err := db.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"user": user})
	}
}

// UpdateUserHandler handles updating a user for admin
func UpdateUserHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var req UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		adminID := c.GetInt("user_id")
		if err := db.UpdateUserRole(userID, req.Role); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}

		// Log admin action
		go db.CreateAdminLog(&adminID, "user_updated", "user", &userID, map[string]interface{}{"new_role": req.Role}, c.ClientIP(), c.GetHeader("User-Agent"))

		c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
	}
}

// DeleteUserHandler handles deleting a user for admin
func DeleteUserHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		adminID := c.GetInt("user_id")
		if err := db.DeleteUser(userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
			return
		}

		// Log admin action
		go db.CreateAdminLog(&adminID, "user_deleted", "user", &userID, nil, c.ClientIP(), c.GetHeader("User-Agent"))

		c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
	}
}

// GetAnalyticsHandler handles retrieving analytics data
func GetAnalyticsHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get query parameters for filtering
		period := c.DefaultQuery("period", "7d")

		// Mock analytics data for development mode when database is not available
		if db == nil {
			analytics := map[string]interface{}{
				"total_advertisers": 23,
				"active_campaigns":  12,
				"total_revenue":     15420.80,
				"pending_approvals": 5,
				"top_performing_placements": []map[string]interface{}{
					{
						"placement_id": 1,
						"name":         "Header Banner",
						"revenue":      8520.30,
						"impressions":  45230,
					},
					{
						"placement_id": 2,
						"name":         "Sidebar Large",
						"revenue":      4890.50,
						"impressions":  28940,
					},
					{
						"placement_id": 3,
						"name":         "Between Videos",
						"revenue":      2010.00,
						"impressions":  15680,
					},
				},
				"revenue_by_month": []map[string]interface{}{
					{"month": "Jan", "revenue": 12450.80, "advertisers": 18},
					{"month": "Feb", "revenue": 13890.20, "advertisers": 20},
					{"month": "Mar", "revenue": 15420.80, "advertisers": 23},
					{"month": "Apr", "revenue": 18230.50, "advertisers": 25},
					{"month": "May", "revenue": 21340.90, "advertisers": 28},
					{"month": "Jun", "revenue": 19850.60, "advertisers": 26},
				},
				"users": map[string]interface{}{
					"total":          1247,
					"new_today":      23,
					"new_week":       156,
					"new_month":      892,
					"active_today":   445,
					"growth_rate":    0.125,
					"churn_rate":     0.032,
					"retention_rate": 0.87,
				},
				"videos": map[string]interface{}{
					"total":          342,
					"published":      298,
					"pending":        12,
					"draft":          32,
					"total_views":    15678,
					"total_likes":    3456,
					"total_comments": 1234,
					"total_shares":   567,
					"avg_rating":     4.2,
					"top_categories": []map[string]interface{}{
						{"name": "Archaeology", "count": 89, "views": 4567},
						{"name": "History", "count": 76, "views": 3890},
						{"name": "Science", "count": 65, "views": 3245},
						{"name": "Geography", "count": 43, "views": 2134},
						{"name": "Linguistics", "count": 25, "views": 1842},
					},
				},
			}
			c.JSON(http.StatusOK, gin.H{"data": analytics, "period": period})
			return
		}

		// Real database implementation would go here
		userCount, err := db.GetUserCount()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user count"})
			return
		}

		videoCount, err := db.GetVideoCount()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get video count"})
			return
		}

		totalViews, err := db.GetTotalViews()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get total views"})
			return
		}

		totalLikes, err := db.GetTotalLikes()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get total likes"})
			return
		}

		recentActivity, err := db.GetRecentActivity(10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recent activity"})
			return
		}

		analytics := map[string]interface{}{
			"users":    userCount,
			"videos":   videoCount,
			"views":    totalViews,
			"likes":    totalLikes,
			"activity": recentActivity,
			"period":   period,
		}

		c.JSON(http.StatusOK, gin.H{"data": analytics, "period": period})
	}
}

// GetSystemHealthHandler handles retrieving system health information
func GetSystemHealthHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Mock system health data for development mode
		if db == nil {
			health := map[string]interface{}{
				"status": "healthy",
				"uptime": "15 days, 4 hours",
				"database": map[string]interface{}{
					"status":          "connected",
					"connections":     12,
					"max_connections": 100,
					"response_time":   "2ms",
				},
				"redis": map[string]interface{}{
					"status":        "disconnected",
					"memory_used":   "0MB",
					"memory_total":  "0MB",
					"response_time": "N/A",
				},
				"storage": map[string]interface{}{
					"disk_used":     "45.2GB",
					"disk_total":    "100GB",
					"disk_free":     "54.8GB",
					"usage_percent": 45.2,
				},
				"memory": map[string]interface{}{
					"used":    "2.1GB",
					"total":   "8GB",
					"free":    "5.9GB",
					"percent": 26.25,
				},
				"cpu": map[string]interface{}{
					"usage":    "15.3%",
					"load_avg": 0.45,
					"cores":    4,
				},
				"network": map[string]interface{}{
					"bandwidth_in":  "1.2MB/s",
					"bandwidth_out": "3.4MB/s",
					"total_in":      "45.6GB",
					"total_out":     "123.4GB",
				},
				"services": map[string]interface{}{
					"api_server":      "running",
					"video_processor": "running",
					"email_service":   "running",
					"backup_service":  "running",
				},
				"last_backup": "2 hours ago",
				"next_backup": "in 22 hours",
			}
			c.JSON(http.StatusOK, gin.H{"health": health})
			return
		}

		// Real database implementation would go here
		health, err := db.GetSystemHealth()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get system health"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"health": health})
	}
}

// GetDetailedAnalyticsHandler handles retrieving detailed analytics for specific metrics
func GetDetailedAnalyticsHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		metric := c.Param("metric")
		period := c.DefaultQuery("period", "7d")

		// Mock detailed analytics for development mode
		if db == nil {
			var data map[string]interface{}

			switch metric {
			case "users":
				data = map[string]interface{}{
					"daily_signups": []map[string]interface{}{
						{"date": "2024-06-12", "signups": 23, "conversions": 18},
						{"date": "2024-06-13", "signups": 34, "conversions": 27},
						{"date": "2024-06-14", "signups": 28, "conversions": 22},
						{"date": "2024-06-15", "signups": 45, "conversions": 36},
						{"date": "2024-06-16", "signups": 32, "conversions": 25},
						{"date": "2024-06-17", "signups": 38, "conversions": 30},
						{"date": "2024-06-18", "signups": 29, "conversions": 23},
					},
					"demographics": map[string]interface{}{
						"age_groups": []map[string]interface{}{
							{"range": "18-24", "count": 234, "percentage": 18.8},
							{"range": "25-34", "count": 456, "percentage": 36.6},
							{"range": "35-44", "count": 298, "percentage": 23.9},
							{"range": "45-54", "count": 167, "percentage": 13.4},
							{"range": "55+", "count": 92, "percentage": 7.4},
						},
						"gender": []map[string]interface{}{
							{"type": "Male", "count": 678, "percentage": 54.4},
							{"type": "Female", "count": 489, "percentage": 39.2},
							{"type": "Other", "count": 80, "percentage": 6.4},
						},
					},
				}
			case "videos":
				data = map[string]interface{}{
					"performance": []map[string]interface{}{
						{"title": "Archaeological Evidence for the Book of Mormon", "views": 4567, "likes": 234, "completion_rate": 0.78},
						{"title": "DNA and the Book of Mormon", "views": 3890, "likes": 198, "completion_rate": 0.72},
						{"title": "Ancient American Civilizations", "views": 3245, "likes": 156, "completion_rate": 0.65},
						{"title": "Nephite Metallurgy Evidence", "views": 2134, "likes": 123, "completion_rate": 0.81},
						{"title": "Book of Mormon Geography", "views": 1842, "likes": 98, "completion_rate": 0.69},
					},
					"upload_trends": []map[string]interface{}{
						{"date": "2024-06-12", "uploads": 5, "approved": 4, "rejected": 1},
						{"date": "2024-06-13", "uploads": 7, "approved": 6, "rejected": 1},
						{"date": "2024-06-14", "uploads": 3, "approved": 3, "rejected": 0},
						{"date": "2024-06-15", "uploads": 8, "approved": 7, "rejected": 1},
						{"date": "2024-06-16", "uploads": 4, "approved": 4, "rejected": 0},
						{"date": "2024-06-17", "uploads": 6, "approved": 5, "rejected": 1},
						{"date": "2024-06-18", "uploads": 2, "approved": 2, "rejected": 0},
					},
				}
			case "revenue":
				data = map[string]interface{}{
					"subscription_trends": []map[string]interface{}{
						{"date": "2024-06-12", "new_subs": 12, "cancellations": 3, "upgrades": 2, "downgrades": 1},
						{"date": "2024-06-13", "new_subs": 18, "cancellations": 4, "upgrades": 3, "downgrades": 1},
						{"date": "2024-06-14", "new_subs": 15, "cancellations": 2, "upgrades": 1, "downgrades": 0},
						{"date": "2024-06-15", "new_subs": 23, "cancellations": 5, "upgrades": 4, "downgrades": 2},
						{"date": "2024-06-16", "new_subs": 16, "cancellations": 3, "upgrades": 2, "downgrades": 1},
						{"date": "2024-06-17", "new_subs": 20, "cancellations": 4, "upgrades": 3, "downgrades": 1},
						{"date": "2024-06-18", "new_subs": 14, "cancellations": 2, "upgrades": 1, "downgrades": 0},
					},
					"churn_analysis": map[string]interface{}{
						"reasons": []map[string]interface{}{
							{"reason": "Price too high", "count": 23, "percentage": 34.3},
							{"reason": "Not enough content", "count": 18, "percentage": 26.9},
							{"reason": "Technical issues", "count": 12, "percentage": 17.9},
							{"reason": "Found alternative", "count": 8, "percentage": 11.9},
							{"reason": "Other", "count": 6, "percentage": 9.0},
						},
						"monthly_churn_rate": []map[string]interface{}{
							{"month": "2024-01", "rate": 0.032},
							{"month": "2024-02", "rate": 0.028},
							{"month": "2024-03", "rate": 0.035},
							{"month": "2024-04", "rate": 0.029},
							{"month": "2024-05", "rate": 0.031},
							{"month": "2024-06", "rate": 0.027},
						},
					},
				}
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metric"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"data": data, "metric": metric, "period": period})
			return
		}

		// Real database implementation would go here
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Database analytics not implemented"})
	}
}

// GetRealTimeAnalyticsHandler handles retrieving real-time analytics data
func GetRealTimeAnalyticsHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Mock real-time data for development mode
		if db == nil {
			realTimeData := map[string]interface{}{
				"active_users":         234,
				"current_streams":      89,
				"server_load":          0.45,
				"bandwidth_usage":      "1.2 GB/s",
				"recent_signups":       5,
				"recent_subscriptions": 2,
				"error_rate":           0.002,
				"response_time":        245,
				"live_events": []map[string]interface{}{
					{"time": "2024-06-18T15:30:00Z", "event": "User signup", "details": "New user from Utah"},
					{"time": "2024-06-18T15:28:00Z", "event": "Video upload", "details": "New video: Ancient Civilizations"},
					{"time": "2024-06-18T15:25:00Z", "event": "Subscription", "details": "Premium subscription upgrade"},
					{"time": "2024-06-18T15:22:00Z", "event": "High traffic", "details": "Popular video trending"},
					{"time": "2024-06-18T15:20:00Z", "event": "User signup", "details": "New user from California"},
				},
				"top_content_now": []map[string]interface{}{
					{"title": "Archaeological Evidence for the Book of Mormon", "viewers": 45},
					{"title": "DNA and the Book of Mormon", "viewers": 32},
					{"title": "Ancient American Civilizations", "viewers": 28},
				},
			}

			c.JSON(http.StatusOK, gin.H{"real_time": realTimeData})
			return
		}

		// Real database implementation would go here
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Real-time analytics not implemented"})
	}
}

// ExportAnalyticsHandler handles exporting analytics data
func ExportAnalyticsHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		format := c.DefaultQuery("format", "csv")
		metric := c.Query("metric")
		period := c.DefaultQuery("period", "30d")

		// Mock export for development mode
		if db == nil {
			switch format {
			case "csv":
				c.Header("Content-Type", "text/csv")
				c.Header("Content-Disposition", "attachment; filename=analytics_export.csv")
				csvData := "Date,Users,Revenue,Videos,Engagement\n"
				csvData += "2024-06-12,234,234.50,5,0.78\n"
				csvData += "2024-06-13,267,345.75,7,0.82\n"
				csvData += "2024-06-14,245,289.25,3,0.75\n"
				c.String(200, csvData)
			case "json":
				c.Header("Content-Type", "application/json")
				c.Header("Content-Disposition", "attachment; filename=analytics_export.json")
				exportData := map[string]interface{}{
					"export_date": "2024-06-18T15:30:00Z",
					"period":      period,
					"metric":      metric,
					"data": []map[string]interface{}{
						{"date": "2024-06-12", "users": 234, "revenue": 234.50, "videos": 5, "engagement": 0.78},
						{"date": "2024-06-13", "users": 267, "revenue": 345.75, "videos": 7, "engagement": 0.82},
						{"date": "2024-06-14", "users": 245, "revenue": 289.25, "videos": 3, "engagement": 0.75},
					},
				}
				c.JSON(200, exportData)
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported format"})
			}
			return
		}

		// Real database implementation would go here
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Export not implemented"})
	}
}

// Video management handlers

// GetAdminVideosHandler handles retrieving all videos for admin with pagination and filtering
func GetAdminVideosHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Mock video data for development mode
		if db == nil {
			videos := []map[string]interface{}{
				{
					"id":          1,
					"title":       "Archaeological Evidence of the Book of Mormon",
					"description": "Exploring recent archaeological discoveries that support Book of Mormon narratives",
					"duration":    "15:42",
					"thumbnail":   "https://example.com/thumb1.jpg",
					"status":      "published",
					"category":    "Archaeology",
					"uploaded_by": map[string]interface{}{
						"id":    2,
						"name":  "Dr. John Smith",
						"email": "john.smith@byu.edu",
					},
					"upload_date": "2024-01-15T10:30:00Z",
					"views":       1247,
					"likes":       89,
					"comments":    23,
					"file_size":   "145.6 MB",
					"resolution":  "1080p",
					"tags":        []string{"archaeology", "evidence", "ancient-america"},
				},
				{
					"id":          2,
					"title":       "DNA and the Book of Mormon",
					"description": "Scientific perspectives on DNA evidence and Book of Mormon populations",
					"duration":    "22:15",
					"thumbnail":   "https://example.com/thumb2.jpg",
					"status":      "pending",
					"category":    "Science",
					"uploaded_by": map[string]interface{}{
						"id":    3,
						"name":  "Dr. Sarah Johnson",
						"email": "sarah.johnson@byu.edu",
					},
					"upload_date": "2024-01-18T14:20:00Z",
					"views":       0,
					"likes":       0,
					"comments":    0,
					"file_size":   "298.4 MB",
					"resolution":  "1080p",
					"tags":        []string{"dna", "science", "genetics"},
				},
				{
					"id":          3,
					"title":       "Mesoamerican Connections",
					"description": "Examining cultural and geographical connections between Mesoamerica and the Book of Mormon",
					"duration":    "18:33",
					"thumbnail":   "https://example.com/thumb3.jpg",
					"status":      "published",
					"category":    "Geography",
					"uploaded_by": map[string]interface{}{
						"id":    4,
						"name":  "Dr. Michael Brown",
						"email": "michael.brown@byu.edu",
					},
					"upload_date": "2024-01-20T09:45:00Z",
					"views":       856,
					"likes":       67,
					"comments":    15,
					"file_size":   "187.2 MB",
					"resolution":  "1080p",
					"tags":        []string{"mesoamerica", "geography", "culture"},
				},
				{
					"id":          4,
					"title":       "Linguistic Analysis of Book of Mormon Names",
					"description": "Scholarly analysis of Hebrew and Egyptian linguistic patterns in Book of Mormon names",
					"duration":    "25:18",
					"thumbnail":   "https://example.com/thumb4.jpg",
					"status":      "draft",
					"category":    "Linguistics",
					"uploaded_by": map[string]interface{}{
						"id":    5,
						"name":  "Dr. Rachel Davis",
						"email": "rachel.davis@byu.edu",
					},
					"upload_date": "2024-01-22T16:12:00Z",
					"views":       0,
					"likes":       0,
					"comments":    0,
					"file_size":   "324.1 MB",
					"resolution":  "1080p",
					"tags":        []string{"linguistics", "hebrew", "names"},
				},
			}

			// Handle pagination and filtering
			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
			search := c.Query("search")
			category := c.Query("category")
			status := c.Query("status")
			sortBy := c.DefaultQuery("sort", "upload_date")
			sortOrder := c.DefaultQuery("order", "desc")

			// Mock filtering
			filteredVideos := videos
			if search != "" {
				filteredVideos = []map[string]interface{}{}
				for _, video := range videos {
					if strings.Contains(strings.ToLower(video["title"].(string)), strings.ToLower(search)) ||
						strings.Contains(strings.ToLower(video["description"].(string)), strings.ToLower(search)) {
						filteredVideos = append(filteredVideos, video)
					}
				}
			}

			if category != "" {
				temp := []map[string]interface{}{}
				for _, video := range filteredVideos {
					if video["category"].(string) == category {
						temp = append(temp, video)
					}
				}
				filteredVideos = temp
			}

			if status != "" {
				temp := []map[string]interface{}{}
				for _, video := range filteredVideos {
					if video["status"].(string) == status {
						temp = append(temp, video)
					}
				}
				filteredVideos = temp
			}

			// Mock pagination
			start := (page - 1) * limit
			end := start + limit
			if start > len(filteredVideos) {
				start = len(filteredVideos)
			}
			if end > len(filteredVideos) {
				end = len(filteredVideos)
			}

			paginatedVideos := filteredVideos[start:end]

			c.JSON(http.StatusOK, gin.H{
				"videos": paginatedVideos,
				"pagination": gin.H{
					"page":       page,
					"limit":      limit,
					"total":      len(filteredVideos),
					"totalPages": (len(filteredVideos) + limit - 1) / limit,
				},
				"filters": gin.H{
					"search":   search,
					"category": category,
					"status":   status,
					"sort":     sortBy,
					"order":    sortOrder,
				},
			})
			return
		}

		// Real database implementation
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		category := c.Query("category")
		status := c.Query("status")

		offset := (page - 1) * limit
		videos, err := db.GetVideos(limit, offset, category, status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get videos"})
			return
		}

		// Get total count for pagination
		totalVideos, err := db.GetVideoCount()
		if err != nil {
			totalVideos = len(videos)
		}

		totalPages := (totalVideos + limit - 1) / limit

		c.JSON(http.StatusOK, gin.H{
			"videos": videos,
			"pagination": gin.H{
				"page":       page,
				"limit":      limit,
				"total":      totalVideos,
				"totalPages": totalPages,
			},
		})
	}
}

// GetAdminVideoHandler handles retrieving a single video for admin
func GetAdminVideoHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
			return
		}

		// Mock video data for development mode
		if db == nil {
			video := map[string]interface{}{
				"id":          videoID,
				"title":       "Archaeological Evidence in Mesoamerica",
				"description": "Detailed analysis of archaeological findings in Mesoamerica",
				"duration":    "45:12",
				"thumbnail":   "https://example.com/thumb1.jpg",
				"status":      "published",
				"category":    "Archaeology",
				"uploaded_by": map[string]interface{}{
					"id":    1,
					"name":  "Dr. John Smith",
					"email": "john.smith@byu.edu",
				},
				"upload_date": time.Now().AddDate(0, 0, -7).Format(time.RFC3339),
				"views":       1234,
				"likes":       89,
				"comments":    23,
				"file_size":   "256.4 MB",
				"resolution":  "1080p",
				"tags":        []string{"archaeology", "mesoamerica", "evidence"},
				"analytics": map[string]interface{}{
					"avg_watch_time":  "12:34",
					"completion_rate": 0.78,
					"engagement_rate": 0.45,
					"shares":          12,
					"unique_viewers":  987,
					"peak_viewers":    156,
					"demographics": map[string]interface{}{
						"age_groups": []map[string]interface{}{
							{"range": "18-24", "percentage": 15},
							{"range": "25-34", "percentage": 35},
							{"range": "35-44", "percentage": 25},
							{"range": "45-54", "percentage": 15},
							{"range": "55+", "percentage": 10},
						},
						"countries": []map[string]interface{}{
							{"name": "United States", "percentage": 45},
							{"name": "Mexico", "percentage": 20},
							{"name": "Canada", "percentage": 15},
							{"name": "United Kingdom", "percentage": 10},
							{"name": "Other", "percentage": 10},
						},
					},
				},
			}
			c.JSON(http.StatusOK, gin.H{"video": video})
			return
		}

		// Real database implementation
		video, err := db.GetVideoByID(videoID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"video": video})
	}
}

// UpdateVideoHandler handles updating video details for admin
func UpdateVideoHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
			return
		}

		var updateData map[string]interface{}
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		adminID := c.GetInt("user_id")

		// Update video in database
		if err := db.UpdateVideo(videoID, updateData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update video"})
			return
		}

		// Log admin action
		go db.CreateAdminLog(&adminID, "video_updated", "video", &videoID, updateData, c.ClientIP(), c.GetHeader("User-Agent"))

		c.JSON(http.StatusOK, gin.H{"message": "Video updated successfully"})
	}
}

// DeleteVideoHandler handles deleting a video for admin
func DeleteVideoHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
			return
		}

		adminID := c.GetInt("user_id")

		// Get video info before deletion for logging
		video, err := db.GetVideoByID(videoID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
			return
		}

		// Delete video from database
		if err := db.DeleteVideo(videoID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete video"})
			return
		}

		// Log admin action
		go db.CreateAdminLog(&adminID, "video_deleted", "video", &videoID, map[string]interface{}{"title": video.Title}, c.ClientIP(), c.GetHeader("User-Agent"))

		c.JSON(http.StatusOK, gin.H{"message": "Video deleted successfully"})
	}
}

// BulkVideoOperationHandler handles bulk operations on videos
func BulkVideoOperationHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Operation string `json:"operation" binding:"required"`
			VideoIDs  []int  `json:"video_ids" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		adminID := c.GetInt("user_id")

		switch req.Operation {
		case "publish":
			for _, videoID := range req.VideoIDs {
				if err := db.UpdateVideoStatus(videoID, "published"); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish videos"})
					return
				}
			}
		case "unpublish":
			for _, videoID := range req.VideoIDs {
				if err := db.UpdateVideoStatus(videoID, "draft"); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unpublish videos"})
					return
				}
			}
		case "delete":
			for _, videoID := range req.VideoIDs {
				if err := db.DeleteVideo(videoID); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete videos"})
					return
				}
			}
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid operation"})
			return
		}

		// Log admin action
		go db.CreateAdminLog(&adminID, "bulk_video_operation", "video", nil, map[string]interface{}{
			"operation": req.Operation,
			"video_ids": req.VideoIDs,
		}, c.ClientIP(), c.GetHeader("User-Agent"))

		c.JSON(http.StatusOK, gin.H{"message": "Bulk operation completed successfully"})
	}
}

// GetVideoStatsHandler handles getting video statistics for admin
func GetVideoStatsHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Mock stats for development mode
		if db == nil {
			stats := map[string]interface{}{
				"total_videos":   342,
				"published":      298,
				"pending":        12,
				"draft":          32,
				"total_views":    15678,
				"total_likes":    3456,
				"total_comments": 1234,
				"total_duration": "2847:32", // in minutes:seconds format
				"storage_used":   "45.6 GB",
				"top_categories": []map[string]interface{}{
					{"name": "Archaeology", "count": 89, "views": 4567},
					{"name": "History", "count": 76, "views": 3890},
					{"name": "Science", "count": 65, "views": 3245},
					{"name": "Geography", "count": 43, "views": 2134},
					{"name": "Linguistics", "count": 25, "views": 1842},
				},
				"recent_uploads": []map[string]interface{}{
					{"date": "2024-01-22", "count": 3},
					{"date": "2024-01-21", "count": 5},
					{"date": "2024-01-20", "count": 2},
					{"date": "2024-01-19", "count": 4},
					{"date": "2024-01-18", "count": 1},
				},
			}

			c.JSON(http.StatusOK, gin.H{"stats": stats})
			return
		}

		// Real database implementation
		totalVideos, _ := db.GetVideoCount()
		totalViews, _ := db.GetTotalViews()
		totalLikes, _ := db.GetTotalLikes()

		stats := map[string]interface{}{
			"total_videos": totalVideos,
			"total_views":  totalViews,
			"total_likes":  totalLikes,
		}

		c.JSON(http.StatusOK, gin.H{"stats": stats})
	}
}

// GetVideoCategoriesHandler handles getting video categories for admin
func GetVideoCategoriesHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Mock categories for development mode
		if db == nil {
			categories := []map[string]interface{}{
				{"id": 1, "name": "Archaeology", "description": "Archaeological evidence and discoveries", "video_count": 89},
				{"id": 2, "name": "History", "description": "Historical context and analysis", "video_count": 76},
				{"id": 3, "name": "Science", "description": "Scientific perspectives and research", "video_count": 65},
				{"id": 4, "name": "Geography", "description": "Geographical and cultural studies", "video_count": 43},
				{"id": 5, "name": "Linguistics", "description": "Language and linguistic analysis", "video_count": 25},
			}

			c.JSON(http.StatusOK, gin.H{"categories": categories})
			return
		}

		// Real database implementation
		categories, err := db.GetVideoCategories()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get categories"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"categories": categories})
	}
}

// ScheduleVideoHandler handles scheduling a video for admin
func ScheduleVideoHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
			return
		}

		var scheduleReq struct {
			PublishDate string `json:"publish_date" binding:"required"`
		}
		if err := c.ShouldBindJSON(&scheduleReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		adminID := c.GetInt("user_id")

		// Parse the publish date
		publishDate, err := time.Parse("2006-01-02T15:04:05Z", scheduleReq.PublishDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use ISO 8601 format"})
			return
		}

		// Schedule video in database
		if err := db.ScheduleVideo(videoID, publishDate); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule video"})
			return
		}

		// Log admin action
		go db.CreateAdminLog(&adminID, "video_scheduled", "video", &videoID, map[string]interface{}{"publish_date": scheduleReq.PublishDate}, c.ClientIP(), c.GetHeader("User-Agent"))

		c.JSON(http.StatusOK, gin.H{"message": "Video scheduled successfully"})
	}
}

// UnscheduleVideoHandler handles unscheduling a video for admin
func UnscheduleVideoHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
			return
		}

		adminID := c.GetInt("user_id")

		// Unschedule video from database
		if err := db.UnscheduleVideo(videoID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unschedule video"})
			return
		}

		// Log admin action
		go db.CreateAdminLog(&adminID, "video_unscheduled", "video", &videoID, nil, c.ClientIP(), c.GetHeader("User-Agent"))

		c.JSON(http.StatusOK, gin.H{"message": "Video unscheduled successfully"})
	}
}

// GetScheduledVideosHandler handles retrieving scheduled videos for admin
func GetScheduledVideosHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Mock scheduled videos for development mode
		if db == nil {
			videos := []map[string]interface{}{
				{
					"id":          1,
					"title":       "Archaeological Evidence of the Book of Mormon",
					"description": "Exploring recent archaeological discoveries that support Book of Mormon narratives",
					"duration":    "15:42",
					"thumbnail":   "https://example.com/thumb1.jpg",
					"status":      "published",
					"category":    "Archaeology",
					"uploaded_by": map[string]interface{}{
						"id":    2,
						"name":  "Dr. John Smith",
						"email": "john.smith@byu.edu",
					},
					"upload_date": "2024-01-15T10:30:00Z",
					"views":       1247,
					"likes":       89,
					"comments":    23,
					"file_size":   "145.6 MB",
					"resolution":  "1080p",
					"tags":        []string{"archaeology", "evidence", "ancient-america"},
				},
				{
					"id":          2,
					"title":       "DNA and the Book of Mormon",
					"description": "Scientific perspectives on DNA evidence and Book of Mormon populations",
					"duration":    "22:15",
					"thumbnail":   "https://example.com/thumb2.jpg",
					"status":      "pending",
					"category":    "Science",
					"uploaded_by": map[string]interface{}{
						"id":    3,
						"name":  "Dr. Sarah Johnson",
						"email": "sarah.johnson@byu.edu",
					},
					"upload_date": "2024-01-18T14:20:00Z",
					"views":       0,
					"likes":       0,
					"comments":    0,
					"file_size":   "298.4 MB",
					"resolution":  "1080p",
					"tags":        []string{"dna", "science", "genetics"},
				},
				{
					"id":          3,
					"title":       "Mesoamerican Connections",
					"description": "Examining cultural and geographical connections between Mesoamerica and the Book of Mormon",
					"duration":    "18:33",
					"thumbnail":   "https://example.com/thumb3.jpg",
					"status":      "published",
					"category":    "Geography",
					"uploaded_by": map[string]interface{}{
						"id":    4,
						"name":  "Dr. Michael Brown",
						"email": "michael.brown@byu.edu",
					},
					"upload_date": "2024-01-20T09:45:00Z",
					"views":       856,
					"likes":       67,
					"comments":    15,
					"file_size":   "187.2 MB",
					"resolution":  "1080p",
					"tags":        []string{"mesoamerica", "geography", "culture"},
				},
				{
					"id":          4,
					"title":       "Linguistic Analysis of Book of Mormon Names",
					"description": "Scholarly analysis of Hebrew and Egyptian linguistic patterns in Book of Mormon names",
					"duration":    "25:18",
					"thumbnail":   "https://example.com/thumb4.jpg",
					"status":      "draft",
					"category":    "Linguistics",
					"uploaded_by": map[string]interface{}{
						"id":    5,
						"name":  "Dr. Rachel Davis",
						"email": "rachel.davis@byu.edu",
					},
					"upload_date": "2024-01-22T16:12:00Z",
					"views":       0,
					"likes":       0,
					"comments":    0,
					"file_size":   "324.1 MB",
					"resolution":  "1080p",
					"tags":        []string{"linguistics", "hebrew", "names"},
				},
			}

			c.JSON(http.StatusOK, gin.H{"videos": videos})
			return
		}

		// Real database implementation
		videos, err := db.GetScheduledVideos(time.Now())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get scheduled videos"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"videos": videos})
	}
}

// Advertisement Placement Handlers
func GetAdPlacementsHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Mock ad placement data for development
		placements := []map[string]interface{}{
			{
				"id":          1,
				"name":        "Homepage Banner",
				"description": "Main banner on homepage",
				"type":        "banner",
				"size":        "728x90",
				"active":      true,
				"position":    "header",
				"page":        "homepage",
				"created_at":  "2024-01-15T10:00:00Z",
				"updated_at":  "2024-06-15T10:00:00Z",
			},
			{
				"id":          2,
				"name":        "Video Player Pre-roll",
				"description": "Advertisement before video content",
				"type":        "video",
				"size":        "1920x1080",
				"active":      true,
				"position":    "pre-roll",
				"page":        "video-player",
				"created_at":  "2024-01-15T10:00:00Z",
				"updated_at":  "2024-06-15T10:00:00Z",
			},
			{
				"id":          3,
				"name":        "Sidebar Rectangle",
				"description": "Medium rectangle ad in sidebar",
				"type":        "banner",
				"size":        "300x250",
				"active":      true,
				"position":    "sidebar",
				"page":        "article",
				"created_at":  "2024-01-15T10:00:00Z",
				"updated_at":  "2024-06-15T10:00:00Z",
			},
			{
				"id":          4,
				"name":        "Mobile Banner",
				"description": "Mobile-optimized banner",
				"type":        "banner",
				"size":        "320x50",
				"active":      false,
				"position":    "bottom",
				"page":        "all",
				"created_at":  "2024-01-15T10:00:00Z",
				"updated_at":  "2024-06-15T10:00:00Z",
			},
		}

		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"placements": placements,
			"total":      len(placements),
		})
	}
}

func GetAdPlacementsPerformanceHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Mock placement performance data
		performance := map[string]interface{}{
			"total_impressions": 125430,
			"total_clicks":      3627,
			"total_revenue":     2847.65,
			"average_ctr":       2.89,
			"average_cpm":       22.70,
			"placements": []map[string]interface{}{
				{
					"id":          1,
					"name":        "Homepage Banner",
					"impressions": 45230,
					"clicks":      1205,
					"revenue":     1024.50,
					"ctr":         2.66,
					"cpm":         22.65,
					"fill_rate":   89.5,
				},
				{
					"id":          2,
					"name":        "Video Player Pre-roll",
					"impressions": 38920,
					"clicks":      1384,
					"revenue":     1186.20,
					"ctr":         3.56,
					"cpm":         30.47,
					"fill_rate":   94.2,
				},
				{
					"id":          3,
					"name":        "Sidebar Rectangle",
					"impressions": 32840,
					"clicks":      892,
					"revenue":     564.75,
					"ctr":         2.72,
					"cpm":         17.19,
					"fill_rate":   78.3,
				},
				{
					"id":          4,
					"name":        "Mobile Banner",
					"impressions": 8440,
					"clicks":      146,
					"revenue":     72.20,
					"ctr":         1.73,
					"cpm":         8.55,
					"fill_rate":   65.8,
				},
			},
			"daily_performance": []map[string]interface{}{
				{"date": "2024-06-17", "impressions": 8923, "clicks": 267, "revenue": 203.45, "ctr": 2.99},
				{"date": "2024-06-18", "impressions": 9104, "clicks": 284, "revenue": 218.30, "ctr": 3.12},
				{"date": "2024-06-19", "impressions": 8756, "clicks": 251, "revenue": 195.80, "ctr": 2.87},
				{"date": "2024-06-20", "impressions": 9287, "clicks": 298, "revenue": 224.15, "ctr": 3.21},
				{"date": "2024-06-21", "impressions": 8834, "clicks": 265, "revenue": 201.25, "ctr": 3.00},
				{"date": "2024-06-22", "impressions": 9145, "clicks": 289, "revenue": 216.90, "ctr": 3.16},
				{"date": "2024-06-23", "impressions": 8967, "clicks": 272, "revenue": 204.85, "ctr": 3.03},
			},
			"top_performers": []map[string]interface{}{
				{
					"placement_id":   2,
					"placement_name": "Video Player Pre-roll",
					"metric":         "highest_ctr",
					"value":          3.56,
				},
				{
					"placement_id":   2,
					"placement_name": "Video Player Pre-roll",
					"metric":         "highest_cpm",
					"value":          30.47,
				},
				{
					"placement_id":   1,
					"placement_name": "Homepage Banner",
					"metric":         "most_impressions",
					"value":          45230,
				},
			},
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    performance,
		})
	}
}

func CreateAdPlacementHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string `json:"name" binding:"required"`
			Description string `json:"description"`
			Type        string `json:"type" binding:"required"`
			Size        string `json:"size" binding:"required"`
			Position    string `json:"position" binding:"required"`
			Page        string `json:"page" binding:"required"`
			Active      bool   `json:"active"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Mock creation response
		placement := map[string]interface{}{
			"id":          5, // Mock new ID
			"name":        req.Name,
			"description": req.Description,
			"type":        req.Type,
			"size":        req.Size,
			"position":    req.Position,
			"page":        req.Page,
			"active":      req.Active,
			"created_at":  "2024-06-23T14:00:00Z",
			"updated_at":  "2024-06-23T14:00:00Z",
		}

		c.JSON(http.StatusCreated, gin.H{
			"success":   true,
			"placement": placement,
			"message":   "Ad placement created successfully",
		})
	}
}

func UpdateAdPlacementHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		placementID := c.Param("id")

		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Type        string `json:"type"`
			Size        string `json:"size"`
			Position    string `json:"position"`
			Page        string `json:"page"`
			Active      *bool  `json:"active"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Mock update response
		placement := map[string]interface{}{
			"id":          placementID,
			"name":        req.Name,
			"description": req.Description,
			"type":        req.Type,
			"size":        req.Size,
			"position":    req.Position,
			"page":        req.Page,
			"active":      req.Active,
			"updated_at":  "2024-06-23T14:00:00Z",
		}

		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"placement": placement,
			"message":   "Ad placement updated successfully",
		})
	}
}

// SetupAdminRoutes configures admin-related routes
func SetupAdminRoutes(router *gin.RouterGroup, db *database.DB) {
	// Users
	router.GET("/users", middleware.AuthRequired(), GetUsersHandler(db))
	router.GET("/users/:id", middleware.AuthRequired(), GetUserHandler(db))
	router.PUT("/users/:id", middleware.AuthRequired(), UpdateUserHandler(db))
	router.DELETE("/users/:id", middleware.AuthRequired(), DeleteUserHandler(db))

	// Videos
	router.GET("/videos", middleware.AuthRequired(), GetAdminVideosHandler(db))
	router.GET("/videos/:id", middleware.AuthRequired(), GetAdminVideoHandler(db))
	router.PUT("/videos/:id", middleware.AuthRequired(), UpdateVideoHandler(db))
	router.DELETE("/videos/:id", middleware.AuthRequired(), DeleteVideoHandler(db))
	router.POST("/videos/bulk", middleware.AuthRequired(), BulkVideoOperationHandler(db))
	router.GET("/videos/:id/stats", middleware.AuthRequired(), GetVideoStatsHandler(db))
	router.GET("/videos/categories", middleware.AuthRequired(), GetVideoCategoriesHandler(db))
	router.POST("/videos/:id/schedule", middleware.AuthRequired(), ScheduleVideoHandler(db))
	router.POST("/videos/:id/unschedule", middleware.AuthRequired(), UnscheduleVideoHandler(db))
	router.GET("/videos/scheduled", middleware.AuthRequired(), GetScheduledVideosHandler(db))

	// Ad Placements
	router.GET("/placements", middleware.AuthRequired(), GetAdPlacementsHandler(db))
	router.GET("/placements/performance", middleware.AuthRequired(), GetAdPlacementsPerformanceHandler(db))
	router.POST("/placements", middleware.AuthRequired(), CreateAdPlacementHandler(db))
	router.PUT("/placements/:id", middleware.AuthRequired(), UpdateAdPlacementHandler(db))

	// Design System Routes
	// Temporarily disabled for debugging
	log.Println("Skipping design system routes for debugging...")
	// SetupMockDesignSystemRoutes(router)
	log.Println("Design system routes skipped")
}

// SetupMockDesignSystemRoutes sets up mock design system routes for development
func SetupMockDesignSystemRoutes(router *gin.RouterGroup) {
	designSystem := router.Group("/design-system")
	{
		// Theme management
		designSystem.GET("/themes", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"themes": []gin.H{},
				"count":  0,
			})
		})

		designSystem.POST("/themes", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{
				"message": "Theme created successfully (mock)",
				"theme": gin.H{
					"id":          1,
					"name":        "Mock Theme",
					"description": "Mock theme for development",
					"isActive":    false,
					"tokens":      []gin.H{},
					"createdAt":   time.Now(),
					"updatedAt":   time.Now(),
				},
			})
		})

		designSystem.PUT("/themes/:id", func(c *gin.Context) {
			themeID := c.Param("id")
			c.JSON(http.StatusOK, gin.H{
				"message": "Theme updated successfully (mock)",
				"theme": gin.H{
					"id":          themeID,
					"name":        "Updated Mock Theme",
					"description": "Updated mock theme",
					"isActive":    true,
					"tokens":      []gin.H{},
					"updatedAt":   time.Now(),
				},
			})
		})

		designSystem.DELETE("/themes/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Theme deleted successfully (mock)",
			})
		})

		designSystem.POST("/themes/activate", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Theme activated successfully (mock)",
				"theme": gin.H{
					"id":          1,
					"name":        "Active Mock Theme",
					"description": "Activated mock theme",
					"isActive":    true,
					"tokens":      []gin.H{},
					"updatedAt":   time.Now(),
				},
			})
		})

		designSystem.GET("/active", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"theme":   nil,
				"message": "No active theme found (mock)",
			})
		})

		// Figma integration
		designSystem.POST("/figma/import", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{
				"message": "Theme created from Figma successfully (mock)",
				"theme": gin.H{
					"id":          1,
					"name":        "Figma Import Mock",
					"description": "Mock theme imported from Figma",
					"isActive":    false,
					"figmaFileId": "mock-file-id",
					"figmaNodeId": "mock-node-id",
					"tokens":      []gin.H{},
					"createdAt":   time.Now(),
					"updatedAt":   time.Now(),
				},
			})
		})

		designSystem.POST("/figma/sync/:id", func(c *gin.Context) {
			themeID := c.Param("id")
			c.JSON(http.StatusOK, gin.H{
				"message": "Theme updated from Figma successfully (mock)",
				"theme": gin.H{
					"id":          themeID,
					"name":        "Synced Mock Theme",
					"description": "Mock theme synced with Figma",
					"tokens":      []gin.H{},
					"updatedAt":   time.Now(),
				},
			})
		})

		designSystem.GET("/figma/preview", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"tokens":  []gin.H{},
				"count":   0,
				"preview": true,
			})
		})

		// Theme operations
		designSystem.POST("/themes/import", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{
				"message": "Theme imported successfully (mock)",
				"theme": gin.H{
					"id":          1,
					"name":        "Imported Mock Theme",
					"description": "Mock imported theme",
					"tokens":      []gin.H{},
					"createdAt":   time.Now(),
				},
			})
		})

		designSystem.GET("/themes/:id/export", func(c *gin.Context) {
			mockTheme := gin.H{
				"id":          c.Param("id"),
				"name":        "Mock Export Theme",
				"description": "Mock theme for export",
				"tokens":      []gin.H{},
				"createdAt":   time.Now(),
			}

			c.Header("Content-Type", "application/json")
			c.Header("Content-Disposition", "attachment; filename=theme-mock.json")
			c.JSON(http.StatusOK, mockTheme)
		})

		designSystem.GET("/themes/:id/tokens", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"tokens": []gin.H{},
				"count":  0,
			})
		})

		// Token management
		designSystem.GET("/tokens", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"tokens": []gin.H{},
				"count":  0,
			})
		})

		designSystem.POST("/tokens", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{
				"message": "Token created successfully (mock)",
				"token": gin.H{
					"id":          1,
					"name":        "mock-token",
					"value":       "#ffffff",
					"type":        "color",
					"category":    "primary",
					"description": "Mock color token",
					"createdAt":   time.Now(),
				},
			})
		})

		designSystem.PUT("/tokens/:id", func(c *gin.Context) {
			tokenID := c.Param("id")
			c.JSON(http.StatusOK, gin.H{
				"message": "Token updated successfully (mock)",
				"token": gin.H{
					"id":          tokenID,
					"name":        "updated-mock-token",
					"value":       "#000000",
					"type":        "color",
					"category":    "primary",
					"description": "Updated mock color token",
					"updatedAt":   time.Now(),
				},
			})
		})

		designSystem.DELETE("/tokens/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Token deleted successfully (mock)",
			})
		})
	}
}
