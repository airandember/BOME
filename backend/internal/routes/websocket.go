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

		// Upgrade HTTP connection to WebSocket
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			return
		}

		log.Printf("WebSocket connection established from %s", c.ClientIP())

		// Initialize connection
		connMutex.Lock()
		connections[ws] = true
		connMutex.Unlock()

		// Initialize subscriptions for this connection
		subMutex.Lock()
		subscriptions[ws] = make(map[string]bool)
		subMutex.Unlock()

		// Set up ping/pong handlers
		ws.SetPingHandler(func(data string) error {
			return ws.WriteControl(websocket.PongMessage, []byte(data), time.Now().Add(10*time.Second))
		})

		ws.SetPongHandler(func(data string) error {
			// Reset read deadline on pong
			return ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		})

		// Start ping ticker
		ticker := time.NewTicker(30 * time.Second)
		go func() {
			defer ticker.Stop()
			for range ticker.C {
				if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
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

			ws.Close()
			ticker.Stop()
		}()

		// Set initial read deadline
		ws.SetReadDeadline(time.Now().Add(60 * time.Second))

		// Handle incoming messages
		for {
			messageType, message, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				break
			}

			// Reset read deadline
			ws.SetReadDeadline(time.Now().Add(60 * time.Second))

			if messageType == websocket.TextMessage {
				var msg struct {
					Type    string   `json:"type"`
					Metrics []string `json:"metrics,omitempty"`
				}

				if err := json.Unmarshal(message, &msg); err != nil {
					sendError(ws, "Invalid message format")
					continue
				}

				switch msg.Type {
				case "subscribe":
					handleSubscribe(ws, msg.Metrics)
				case "unsubscribe":
					handleUnsubscribe(ws, msg.Metrics)
				default:
					sendError(ws, "Unknown message type")
				}
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
	}

	for ws, active := range connections {
		if !active {
			continue
		}

		if subs, ok := subscriptions[ws]; ok {
			if subs[metric] {
				ws.WriteJSON(message)
			}
		}
	}
}

// BroadcastSystemHealth broadcasts system health updates to all connected clients
func BroadcastSystemHealth(data SystemHealth) {
	connMutex.RLock()
	defer connMutex.RUnlock()

	message := gin.H{
		"type": "system_health",
		"data": data,
	}

	for ws, active := range connections {
		if active {
			ws.WriteJSON(message)
		}
	}
}
