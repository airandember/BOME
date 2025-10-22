package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrader is the shared WebSocket upgrader configuration
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Add proper origin check for production
		return true
	},
}
