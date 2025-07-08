package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// SystemHealth represents the system health metrics
type SystemHealth struct {
	Uptime             string    `json:"uptime"`
	ResponseTime       string    `json:"response_time"`
	ErrorRate          string    `json:"error_rate"`
	StorageUsed        string    `json:"storage_used"`
	BandwidthUsed      string    `json:"bandwidth_used"`
	CDNHits            string    `json:"cdn_hits"`
	DatabaseSize       string    `json:"database_size"`
	ActiveSessions     int       `json:"active_sessions"`
	LastWrite          time.Time `json:"last_write"`
	TotalEventsTracked int       `json:"total_events_tracked"`
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// TODO: Add proper origin check for production
			return true
		},
	}

	// Global connection management
	connections = make(map[*websocket.Conn]bool)
	connMutex   sync.RWMutex

	// Subscription management
	subscriptions = make(map[*websocket.Conn]map[string]bool)
	subMutex      sync.RWMutex
)

// WebSocketHandler handles WebSocket connections for real-time analytics
func WebSocketHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for authentication token in query parameters
		token := c.Query("token")
		if token != "" {
			// For now, just verify it's not empty - in production, validate against user database
			// This is mock authentication matching the frontend's mock token system
			if token != "mock-admin-token-1750747949368" && !strings.HasPrefix(token, "mock-admin-token") {
				log.Printf("WebSocket authentication failed for token: %s", token[:10]+"...")
				c.JSON(401, gin.H{"error": "Unauthorized"})
				return
			}
			log.Printf("WebSocket authenticated with token: %s", token[:10]+"...")
		} else {
			log.Printf("WebSocket connection without authentication token")
		}

		// Upgrade HTTP connection to WebSocket with better error handling
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			// Send a proper error response if possible
			if c.Writer.Written() {
				return
			}
			c.JSON(500, gin.H{"error": "Failed to establish WebSocket connection"})
			return
		}

		log.Printf("WebSocket connection established from %s", c.ClientIP())

		// Initialize connection with error handling
		connMutex.Lock()
		connections[ws] = true
		connMutex.Unlock()

		// Initialize subscriptions for this connection
		subMutex.Lock()
		subscriptions[ws] = make(map[string]bool)
		subMutex.Unlock()

		// Set up ping/pong handlers with error handling
		ws.SetPingHandler(func(data string) error {
			err := ws.WriteControl(websocket.PongMessage, []byte(data), time.Now().Add(10*time.Second))
			if err != nil {
				log.Printf("Failed to send pong: %v", err)
			}
			return err
		})

		ws.SetPongHandler(func(data string) error {
			// Reset read deadline on pong
			err := ws.SetReadDeadline(time.Now().Add(60 * time.Second))
			if err != nil {
				log.Printf("Failed to set read deadline: %v", err)
			}
			return err
		})

		// Set up close handler
		ws.SetCloseHandler(func(code int, text string) error {
			log.Printf("WebSocket close requested: %d %s", code, text)
			return nil
		})

		// Start ping ticker with error handling
		ticker := time.NewTicker(30 * time.Second)
		go func() {
			defer ticker.Stop()
			for range ticker.C {
				if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
					log.Printf("Failed to send ping: %v", err)
					return
				}
			}
		}()

		// Clean up on disconnect
		defer func() {
			log.Printf("WebSocket connection closed from %s", c.ClientIP())
			connMutex.Lock()
			delete(connections, ws)
			connMutex.Unlock()

			subMutex.Lock()
			delete(subscriptions, ws)
			subMutex.Unlock()

			// Close connection gracefully
			ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(5*time.Second))
			ws.Close()
			ticker.Stop()
		}()

		// Set initial read deadline
		if err := ws.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			log.Printf("Failed to set initial read deadline: %v", err)
			return
		}

		// Handle incoming messages with better error handling
		for {
			messageType, message, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				} else if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Printf("WebSocket closed normally: %v", err)
				} else {
					log.Printf("WebSocket read error: %v", err)
				}
				break
			}

			// Reset read deadline
			if err := ws.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
				log.Printf("Failed to reset read deadline: %v", err)
				break
			}

			if messageType == websocket.TextMessage {
				var msg struct {
					Type    string   `json:"type"`
					Metrics []string `json:"metrics,omitempty"`
				}

				if err := json.Unmarshal(message, &msg); err != nil {
					log.Printf("Failed to unmarshal message: %v", err)
					sendError(ws, "Invalid message format")
					continue
				}

				switch msg.Type {
				case "subscribe":
					handleSubscribe(ws, msg.Metrics)
				case "unsubscribe":
					handleUnsubscribe(ws, msg.Metrics)
				case "ping":
					// Respond to ping with pong
					if err := ws.WriteJSON(gin.H{"type": "pong", "timestamp": time.Now().Unix()}); err != nil {
						log.Printf("Failed to send pong response: %v", err)
					}
				default:
					sendError(ws, "Unknown message type: "+msg.Type)
				}
			} else {
				log.Printf("Received non-text message type: %d", messageType)
			}
		}
	}
}

// handleSubscribe handles metric subscription requests
func handleSubscribe(ws *websocket.Conn, metrics []string) {
	subMutex.Lock()
	defer subMutex.Unlock()

	subs := subscriptions[ws]
	for _, metric := range metrics {
		subs[metric] = true
	}
}

// handleUnsubscribe handles metric unsubscription requests
func handleUnsubscribe(ws *websocket.Conn, metrics []string) {
	subMutex.Lock()
	defer subMutex.Unlock()

	subs := subscriptions[ws]
	for _, metric := range metrics {
		delete(subs, metric)
	}
}

// sendError sends an error message to the client
func sendError(ws *websocket.Conn, message string) {
	ws.WriteJSON(gin.H{
		"type":    "error",
		"message": message,
	})
}

// BroadcastAnalyticsUpdate broadcasts analytics updates to subscribed clients
func BroadcastAnalyticsUpdate(metric string, data interface{}) {
	subMutex.RLock()
	defer subMutex.RUnlock()

	connMutex.RLock()
	defer connMutex.RUnlock()

	message := gin.H{
		"type":   "analytics_update",
		"metric": metric,
		"data":   data,
		"time":   time.Now().Unix(),
	}

	// Track failed connections for cleanup
	var failedConnections []*websocket.Conn

	for ws, active := range connections {
		if !active {
			continue
		}

		if subs, ok := subscriptions[ws]; ok {
			if subs[metric] {
				if err := ws.WriteJSON(message); err != nil {
					log.Printf("Failed to send analytics update to client: %v", err)
					failedConnections = append(failedConnections, ws)
				}
			}
		}
	}

	// Clean up failed connections
	if len(failedConnections) > 0 {
		go func() {
			connMutex.Lock()
			defer connMutex.Unlock()

			subMutex.Lock()
			defer subMutex.Unlock()

			for _, conn := range failedConnections {
				// Close connection gracefully
				conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Failed to send message"), time.Now().Add(5*time.Second))
				conn.Close()
				delete(connections, conn)
				delete(subscriptions, conn)
			}
		}()
	}
}

// BroadcastSystemHealth broadcasts system health updates to all connected clients
func BroadcastSystemHealth(data SystemHealth) {
	connMutex.RLock()
	defer connMutex.RUnlock()

	message := gin.H{
		"type": "system_health",
		"data": data,
		"time": time.Now().Unix(),
	}

	// Track failed connections for cleanup
	var failedConnections []*websocket.Conn

	for ws, active := range connections {
		if active {
			if err := ws.WriteJSON(message); err != nil {
				log.Printf("Failed to send system health update to client: %v", err)
				failedConnections = append(failedConnections, ws)
			}
		}
	}

	// Clean up failed connections
	if len(failedConnections) > 0 {
		go func() {
			connMutex.Lock()
			defer connMutex.Unlock()

			subMutex.Lock()
			defer subMutex.Unlock()

			for _, conn := range failedConnections {
				// Close connection gracefully
				conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Failed to send message"), time.Now().Add(5*time.Second))
				conn.Close()
				delete(connections, conn)
				delete(subscriptions, conn)
			}
		}()
	}
}
