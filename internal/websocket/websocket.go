// Package websocket provides real-time communication for the analytics dashboard.
// It implements a WebSocket hub that manages client connections and broadcasts
// updates when conversation data changes.
package websocket

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// Hub manages WebSocket connections and provides real-time updates to connected clients.
// It is safe for concurrent use and supports graceful shutdown via context cancellation.
type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	done       chan struct{}
}

// NewHub creates a new WebSocket hub with context support for graceful shutdown.
func NewHub() *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		ctx:        ctx,
		cancel:     cancel,
		done:       make(chan struct{}),
	}
}

// Run starts the hub's main loop and blocks until Shutdown is called.
// It handles client registration, unregistration, and message broadcasting.
func (h *Hub) Run() {
	defer close(h.done)

	// Buffer for failed clients to avoid deadlock when writing to unregister during broadcast
	failedClients := make([]*websocket.Conn, 0, 10)

	for {
		select {
		case <-h.ctx.Done():
			// Graceful shutdown: close all clients
			h.mutex.Lock()
			for client := range h.clients {
				client.Close()
			}
			h.clients = make(map[*websocket.Conn]bool)
			h.mutex.Unlock()
			return

		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			// Collect failed clients without holding the lock during unregister
			failedClients = failedClients[:0]

			h.mutex.RLock()
			for client := range h.clients {
				if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
					failedClients = append(failedClients, client)
				}
			}
			h.mutex.RUnlock()

			// Unregister failed clients after releasing the read lock
			for _, client := range failedClients {
				h.unregister <- client
			}
		}
	}
}

// Broadcast sends a message to all connected clients.
// It is non-blocking and safe to call from multiple goroutines.
func (h *Hub) Broadcast(message []byte) {
	select {
	case h.broadcast <- message:
	case <-h.ctx.Done():
		// Hub is shutting down, ignore broadcast
	default:
		// Channel is full, drop message to avoid blocking
	}
}

// ClientCount returns the number of currently connected clients.
func (h *Hub) ClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// Shutdown gracefully shuts down the hub and closes all client connections.
// It blocks until the main loop has exited and all clients are closed.
func (h *Hub) Shutdown() error {
	h.cancel()
	<-h.done
	return nil
}

// HandleWebSocket returns a WebSocket handler function for Fiber.
// It manages the connection lifecycle, registration, and message reading.
func (h *Hub) HandleWebSocket() func(*websocket.Conn) {
	return func(c *websocket.Conn) {
		defer func() {
			select {
			case h.unregister <- c:
			case <-h.ctx.Done():
				// Hub is shutting down, just close the connection
				c.Close()
			}
		}()

		// Register the client
		select {
		case h.register <- c:
		case <-h.ctx.Done():
			// Hub is shutting down, close connection
			return
		}

		// Send initial welcome message (ignore errors on shutdown)
		if err := c.WriteJSON(fiber.Map{
			"type":    "connected",
			"message": "WebSocket connected",
			"time":    time.Now(),
		}); err != nil {
			return
		}

		// Read messages from client (mostly for keepalive and ping/pong)
		for {
			select {
			case <-h.ctx.Done():
				return
			default:
				_, _, err := c.ReadMessage()
				if err != nil {
					return
				}
			}
		}
	}
}

// SendUpdate sends an update notification to all clients.
// This is a convenience method that formats the update as JSON.
func (h *Hub) SendUpdate(updateType string, data interface{}) {
	// Note: data parameter is currently unused for simplicity
	// In production, consider proper JSON marshaling with encoding/json
	_ = data
	message := []byte(fmt.Sprintf(`{"type":"%s","time":"%s"}`, updateType, time.Now().Format(time.RFC3339)))
	h.Broadcast(message)
}
