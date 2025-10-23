package websocket

import (
	"log"
	"net/http"

	"bome-backend/authentication/middleware"

	"github.com/gin-gonic/gin"
)

// SetupWebSocketRoutes sets up WebSocket routes for admin real-time updates
func SetupWebSocketRoutes(router *gin.RouterGroup, hub *AdminHub) {
	log.Println("🔧 Setting up WebSocket routes...")

	// WebSocket endpoint for admin real-time updates
	// Requires authentication
	ws := router.Group("/ws")
	ws.Use(middleware.AuthRequired())
	{
		// Admin WebSocket endpoint
		ws.GET("/admin", func(c *gin.Context) {
			HandleAdminWebSocket(c, hub)
		})

		// WebSocket stats endpoint (for monitoring)
		ws.GET("/admin/stats", func(c *gin.Context) {
			stats := hub.GetStats()
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"stats":   stats,
			})
		})
	}

	log.Println("✅ WebSocket routes setup complete")
}

// HandleAdminWebSocket handles WebSocket connection for admin users
func HandleAdminWebSocket(c *gin.Context, hub *AdminHub) {
	// Get user info from context (set by AuthRequired middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		log.Println("❌ WebSocket: user_id not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	username, _ := c.Get("email")
	if username == nil {
		username = "unknown"
	}

	// Check if user is admin
	role, _ := c.Get("role")
	if role != "super_admin" && role != "admin" {
		log.Printf("❌ WebSocket: user %v is not admin (role: %v)", username, role)
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("❌ Failed to upgrade to WebSocket: %v", err)
		return
	}

	// Create client
	client := &Client{
		conn:     conn,
		send:     make(chan AdminEvent, 256),
		hub:      hub,
		userID:   userID.(int),
		username: username.(string),
	}

	// Register client
	hub.register <- client

	// Start read/write pumps
	go client.writePump()
	go client.readPump()

	log.Printf("✅ WebSocket connection established for admin: %s (ID: %d)", client.username, client.userID)
}
