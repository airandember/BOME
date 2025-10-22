package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Event types for real-time admin updates
const (
	EventSubscriberCreated    = "subscriber.created"
	EventSubscriberUpdated    = "subscriber.updated"
	EventSubscriptionCreated  = "subscription.created"
	EventSubscriptionUpdated  = "subscription.updated"
	EventSubscriptionCanceled = "subscription.canceled"
	EventPaymentReceived      = "payment.received"
	EventPaymentFailed        = "payment.failed"
	EventKPIUpdate            = "kpi.update"
	EventPlanCreated          = "plan.created"
	EventPlanUpdated          = "plan.updated"
	EventPlanDeleted          = "plan.deleted"
	EventSystemAlert          = "system.alert"
	EventConnected            = "connected"
	EventPing                 = "ping"
	EventPong                 = "pong"
)

// AdminEvent represents a real-time event for admin dashboard
type AdminEvent struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Message   string                 `json:"message,omitempty"`
}

// Client represents a connected WebSocket client
type Client struct {
	conn     *websocket.Conn
	send     chan AdminEvent
	hub      *AdminHub
	userID   int
	username string
}

// AdminHub maintains active WebSocket connections for admin users
type AdminHub struct {
	// Registered clients
	clients map[*Client]bool

	// Broadcast channel
	broadcast chan AdminEvent

	// Register client
	register chan *Client

	// Unregister client
	unregister chan *Client

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Statistics
	totalConnections    int64
	totalMessages       int64
	totalBroadcasts     int64
	connectionStartTime time.Time
}

// NewAdminHub creates a new admin hub
func NewAdminHub() *AdminHub {
	return &AdminHub{
		clients:             make(map[*Client]bool),
		broadcast:           make(chan AdminEvent, 256),
		register:            make(chan *Client),
		unregister:          make(chan *Client),
		connectionStartTime: time.Now(),
	}
}

// Run starts the hub (should be run in a goroutine)
func (h *AdminHub) Run() {
	log.Println("🌐 WebSocket AdminHub started")
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.totalConnections++
			clientCount := len(h.clients)
			h.mu.Unlock()

			log.Printf("✅ Admin WebSocket connected: %s (total: %d)", client.username, clientCount)

			// Send welcome message
			welcomeEvent := AdminEvent{
				Type:      EventConnected,
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"user_id":       client.userID,
					"username":      client.username,
					"client_count":  clientCount,
					"server_uptime": time.Since(h.connectionStartTime).Seconds(),
				},
				Message: "Connected to admin real-time updates",
			}
			select {
			case client.send <- welcomeEvent:
			default:
				// Client send buffer is full, skip welcome message
			}

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				clientCount := len(h.clients)
				h.mu.Unlock()
				log.Printf("❌ Admin WebSocket disconnected: %s (total: %d)", client.username, clientCount)
			} else {
				h.mu.Unlock()
			}

		case event := <-h.broadcast:
			h.mu.RLock()
			h.totalBroadcasts++
			clientCount := len(h.clients)
			h.mu.RUnlock()

			log.Printf("📡 Broadcasting event: %s to %d clients", event.Type, clientCount)

			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- event:
					h.mu.RUnlock()
					h.mu.Lock()
					h.totalMessages++
					h.mu.Unlock()
					h.mu.RLock()
				default:
					// Client's send buffer is full, disconnect them
					h.mu.RUnlock()
					go func(c *Client) {
						log.Printf("⚠️  Client %s send buffer full, disconnecting", c.username)
						h.unregister <- c
					}(client)
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastEvent sends an event to all connected admin clients
func (h *AdminHub) BroadcastEvent(eventType string, data map[string]interface{}, message string) {
	event := AdminEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
		Message:   message,
	}

	select {
	case h.broadcast <- event:
		// Event queued successfully
	default:
		log.Printf("⚠️  Broadcast channel full, dropping event: %s", eventType)
	}
}

// BroadcastSubscriberCreated broadcasts a new subscriber event
func (h *AdminHub) BroadcastSubscriberCreated(subscriber interface{}) {
	h.BroadcastEvent(EventSubscriberCreated, map[string]interface{}{
		"subscriber": subscriber,
	}, "New subscriber signed up!")
}

// BroadcastSubscriberUpdated broadcasts a subscriber update event
func (h *AdminHub) BroadcastSubscriberUpdated(subscriber interface{}) {
	h.BroadcastEvent(EventSubscriberUpdated, map[string]interface{}{
		"subscriber": subscriber,
	}, "Subscriber updated")
}

// BroadcastSubscriptionCreated broadcasts a new subscription event
func (h *AdminHub) BroadcastSubscriptionCreated(subscription interface{}) {
	h.BroadcastEvent(EventSubscriptionCreated, map[string]interface{}{
		"subscription": subscription,
	}, "New subscription created!")
}

// BroadcastSubscriptionUpdated broadcasts a subscription update event
func (h *AdminHub) BroadcastSubscriptionUpdated(subscription interface{}) {
	h.BroadcastEvent(EventSubscriptionUpdated, map[string]interface{}{
		"subscription": subscription,
	}, "Subscription updated")
}

// BroadcastSubscriptionCanceled broadcasts a subscription cancellation event
func (h *AdminHub) BroadcastSubscriptionCanceled(subscription interface{}) {
	h.BroadcastEvent(EventSubscriptionCanceled, map[string]interface{}{
		"subscription": subscription,
	}, "Subscription canceled")
}

// BroadcastPaymentReceived broadcasts a payment received event
func (h *AdminHub) BroadcastPaymentReceived(payment interface{}) {
	h.BroadcastEvent(EventPaymentReceived, map[string]interface{}{
		"payment": payment,
	}, "Payment received!")
}

// BroadcastPaymentFailed broadcasts a payment failed event
func (h *AdminHub) BroadcastPaymentFailed(payment interface{}) {
	h.BroadcastEvent(EventPaymentFailed, map[string]interface{}{
		"payment": payment,
	}, "Payment failed")
}

// BroadcastKPIUpdate broadcasts KPI updates
func (h *AdminHub) BroadcastKPIUpdate(kpis interface{}) {
	h.BroadcastEvent(EventKPIUpdate, map[string]interface{}{
		"kpis": kpis,
	}, "KPIs updated")
}

// BroadcastPlanCreated broadcasts a plan created event
func (h *AdminHub) BroadcastPlanCreated(plan interface{}) {
	h.BroadcastEvent(EventPlanCreated, map[string]interface{}{
		"plan": plan,
	}, "Subscription plan created")
}

// BroadcastPlanUpdated broadcasts a plan updated event
func (h *AdminHub) BroadcastPlanUpdated(plan interface{}) {
	h.BroadcastEvent(EventPlanUpdated, map[string]interface{}{
		"plan": plan,
	}, "Subscription plan updated")
}

// BroadcastPlanDeleted broadcasts a plan deleted event
func (h *AdminHub) BroadcastPlanDeleted(planID int) {
	h.BroadcastEvent(EventPlanDeleted, map[string]interface{}{
		"plan_id": planID,
	}, "Subscription plan deleted")
}

// BroadcastSystemAlert broadcasts a system alert
func (h *AdminHub) BroadcastSystemAlert(alertType, message string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["alert_type"] = alertType
	h.BroadcastEvent(EventSystemAlert, data, message)
}

// GetStats returns hub statistics
func (h *AdminHub) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return map[string]interface{}{
		"active_connections":  len(h.clients),
		"total_connections":   h.totalConnections,
		"total_messages":      h.totalMessages,
		"total_broadcasts":    h.totalBroadcasts,
		"uptime_seconds":      time.Since(h.connectionStartTime).Seconds(),
		"broadcast_queue_len": len(h.broadcast),
	}
}

// GetClientCount returns the number of active clients
func (h *AdminHub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// readPump reads messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for user %s: %v", c.username, err)
			}
			break
		}

		// Handle incoming messages (e.g., ping requests)
		var msg struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(message, &msg); err == nil {
			if msg.Type == EventPing {
				// Respond with pong
				pongEvent := AdminEvent{
					Type:      EventPong,
					Timestamp: time.Now(),
					Data:      map[string]interface{}{},
				}
				select {
				case c.send <- pongEvent:
				default:
				}
			}
		}
	}
}

// writePump writes messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case event, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(event); err != nil {
				log.Printf("Error writing to WebSocket for user %s: %v", c.username, err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
