package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now; restrict in production
	},
}

// WebSocketHandler handles WebSocket connections (placeholder for future use).
func WebSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade to WebSocket"})
		return
	}
	defer conn.Close()

	// Example: Echo loop (replace with real logic as needed)
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if err := conn.WriteMessage(mt, message); err != nil {
			break
		}
	}
}
