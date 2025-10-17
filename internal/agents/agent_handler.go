package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	claude "github.com/schlunsen/claude-agent-sdk-go"
	"golang.org/x/net/websocket"
)

// QueryRequest represents a query message from the WebSocket client
type QueryRequest struct {
	Prompt string `json:"prompt"`
}

// ResponseMessage represents a response message sent to the WebSocket client
type ResponseMessage struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
	Error   string      `json:"error,omitempty"`
}

// AgentHandler manages WebSocket connections and Claude Agent SDK integration
type AgentHandler struct {
	config *Config
	mu     sync.Mutex
	active int
}

// NewAgentHandler creates a new agent handler with the given config
func NewAgentHandler(config *Config) *AgentHandler {
	return &AgentHandler{
		config: config,
		active: 0,
	}
}

// HandleWebSocket handles WebSocket connections for Claude queries
func (h *AgentHandler) HandleWebSocket(ws *websocket.Conn) {
	defer func() {
		_ = ws.Close()
	}()

	// Check concurrent session limit
	h.mu.Lock()
	if h.active >= h.config.MaxConcurrentSessions {
		h.mu.Unlock()
		h.sendError(ws, "max concurrent sessions reached")
		return
	}
	h.active++
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		h.active--
		h.mu.Unlock()
	}()

	log.Printf("WebSocket connection established from %s", ws.Request().RemoteAddr)

	// Main message loop
	for {
		var req QueryRequest
		if err := websocket.JSON.Receive(ws, &req); err != nil {
			if err.Error() != "EOF" {
				log.Printf("Error receiving message: %v", err)
			}
			return
		}

		if req.Prompt == "" {
			h.sendError(ws, "prompt cannot be empty")
			continue
		}

		log.Printf("Received query: %s", req.Prompt)

		// Process the query
		if err := h.processQuery(ws, req.Prompt); err != nil {
			log.Printf("Error processing query: %v", err)
			h.sendError(ws, fmt.Sprintf("query failed: %v", err))
		}
	}
}

// processQuery executes a Claude query and streams responses back
func (h *AgentHandler) processQuery(ws *websocket.Conn, prompt string) error {
	ctx := context.Background()

	// Execute query using the SDK's public Query function
	// Note: We pass nil for options to use defaults, as the SDK's options contain internal types
	messages, err := claude.Query(ctx, prompt, nil)
	if err != nil {
		return fmt.Errorf("failed to create query: %w", err)
	}

	// Stream responses back to client
	for msg := range messages {
		if err := h.sendMessage(ws, msg); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	}

	return nil
}

// sendMessage sends a Claude message to the WebSocket client
func (h *AgentHandler) sendMessage(ws *websocket.Conn, msg interface{}) error {
	// Since the SDK's types are internal, we work with interface{} and JSON marshaling
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Unmarshal as a generic map to extract type information
	var msgMap map[string]interface{}
	if err := json.Unmarshal(msgJSON, &msgMap); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Extract message type
	msgType, ok := msgMap["type"].(string)
	if !ok {
		msgType = "unknown"
	}

	// Build response
	resp := ResponseMessage{
		Type:    msgType,
		Content: msgMap,
	}

	return websocket.JSON.Send(ws, resp)
}

// sendError sends an error message to the WebSocket client
func (h *AgentHandler) sendError(ws *websocket.Conn, errMsg string) {
	resp := ResponseMessage{
		Type:  "error",
		Error: errMsg,
	}
	if err := websocket.JSON.Send(ws, resp); err != nil {
		log.Printf("Failed to send error message: %v", err)
	}
}

// GetStats returns current handler statistics
func (h *AgentHandler) GetStats() map[string]interface{} {
	h.mu.Lock()
	defer h.mu.Unlock()

	return map[string]interface{}{
		"active_sessions": h.active,
		"max_sessions":    h.config.MaxConcurrentSessions,
	}
}

// HealthCheck endpoint handler
func (h *AgentHandler) HealthCheck(ws *websocket.Conn) {
	defer func() {
		_ = ws.Close()
	}()

	stats := h.GetStats()
	statsJSON, _ := json.Marshal(stats)

	_ = websocket.Message.Send(ws, string(statsJSON))
}
