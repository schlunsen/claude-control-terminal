package websocket

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// Hub manages WebSocket connections
type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			fmt.Printf("WebSocket client connected. Total: %d\n", len(h.clients))

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mutex.Unlock()
			fmt.Printf("WebSocket client disconnected. Total: %d\n", len(h.clients))

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
					h.unregister <- client
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

// ClientCount returns the number of connected clients
func (h *Hub) ClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// HandleWebSocket is the WebSocket handler for Fiber
func (h *Hub) HandleWebSocket() func(*websocket.Conn) {
	return func(c *websocket.Conn) {
		defer func() {
			h.unregister <- c
			c.Close()
		}()

		h.register <- c

		// Send initial welcome message
		c.WriteJSON(fiber.Map{
			"type":    "connected",
			"message": "WebSocket connected",
			"time":    time.Now(),
		})

		// Read messages from client (mostly for keepalive)
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				break
			}
		}
	}
}

// SendUpdate sends an update notification to all clients
func (h *Hub) SendUpdate(updateType string, data interface{}) {
	message := map[string]interface{}{
		"type": updateType,
		"data": data,
		"time": time.Now(),
	}

	// Convert to JSON bytes
	// Note: In production, use proper JSON marshaling
	h.Broadcast([]byte(fmt.Sprintf(`{"type":"%s","time":"%s"}`, updateType, time.Now().Format(time.RFC3339))))
}
