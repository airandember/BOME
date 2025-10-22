package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrader is the WebSocket upgrader with CORS configuration
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")

		// Allow local development
		if origin == "http://localhost:5173" ||
			origin == "http://localhost:3000" ||
			origin == "http://127.0.0.1:5173" ||
			origin == "http://127.0.0.1:3000" {
			return true
		}

		// Allow production domains (update these with your actual domains)
		if origin == "https://yourdomain.com" ||
			origin == "https://www.yourdomain.com" ||
			origin == "https://admin.yourdomain.com" {
			return true
		}

		// Reject all other origins
		return false
	},
}
