package agents

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/schlunsen/claude-agent-sdk-go/types"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// AgentHandler manages WebSocket connections and Claude Agent SDK integration
type AgentHandler struct {
	Config         *Config         // Exported for server access
	SessionManager *SessionManager // Exported for server access
	Mu             sync.Mutex      // Exported for server access
	Active         int             // Exported for server access
}

// NewAgentHandler creates a new agent handler with the given config
func NewAgentHandler(config *Config) *AgentHandler {
	return &AgentHandler{
		Config:         config,
		SessionManager: NewSessionManager(config),
		Active:         0,
	}
}

// HandleWebSocket handles WebSocket connections for Claude queries
func (h *AgentHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %v", err)
		return
	}
	defer func() {
		_ = ws.Close()
	}()

	// Check concurrent session limit
	h.Mu.Lock()
	if h.Active >= h.Config.MaxConcurrentSessions {
		h.Mu.Unlock()
		h.sendError(ws, "max concurrent sessions reached")
		return
	}
	h.Active++
	h.Mu.Unlock()

	defer func() {
		h.Mu.Lock()
		h.Active--
		h.Mu.Unlock()
	}()

	log.Printf("WebSocket connection established from %s", r.RemoteAddr)

	// Main message loop
	for {
		var rawMsg map[string]interface{}
		if err := ws.ReadJSON(&rawMsg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error receiving message: %v", err)
			}
			return
		}

		msgType, ok := rawMsg["type"].(string)
		if !ok {
			log.Printf("ERROR: Missing or invalid message type in: %+v", rawMsg)
			h.sendError(ws, "missing or invalid message type")
			continue
		}

		log.Printf("Received message type: %s with data: %+v", msgType, rawMsg)

		// Route message to appropriate handler
		if err := h.routeMessage(ws, MessageType(msgType), rawMsg); err != nil {
			log.Printf("ERROR: Failed to handle message type %s: %v", msgType, err)
			h.sendError(ws, fmt.Sprintf("message handling failed: %v", err))
		}
	}
}

// routeMessage routes messages to appropriate handlers
func (h *AgentHandler) routeMessage(ws *websocket.Conn, msgType MessageType, rawMsg map[string]interface{}) error {
	switch msgType {
	case MessageTypeAuth:
		// Authentication handled by proxy, skip
		return nil

	case MessageTypeCreateSession:
		return h.handleCreateSession(ws, rawMsg)

	case MessageTypeSendPrompt:
		return h.handleSendPrompt(ws, rawMsg)

	case MessageTypeEndSession:
		return h.handleEndSession(ws, rawMsg)

	case MessageTypeListSessions:
		return h.handleListSessions(ws)

	case MessageTypeKillAllAgents:
		return h.handleKillAllAgents(ws)

	case MessageTypePing:
		return h.handlePing(ws)

	default:
		return fmt.Errorf("unknown message type: %s", msgType)
	}
}

// handleCreateSession creates a new agent session
func (h *AgentHandler) handleCreateSession(ws *websocket.Conn, rawMsg map[string]interface{}) error {
	var msg CreateSessionMessage
	msgBytes, _ := json.Marshal(rawMsg)
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		return fmt.Errorf("invalid create_session message: %w", err)
	}

	log.Printf("Creating session: %s", msg.SessionID)

	// Create session
	session, err := h.SessionManager.CreateSession(msg.SessionID, msg.Options)
	if err != nil {
		log.Printf("ERROR: Failed to create session: %v", err)
		return err
	}

	log.Printf("Session created successfully: %s", session.ID)

	// Send session created response
	response := SessionCreatedMessage{
		BaseMessage: BaseMessage{Type: MessageTypeSessionCreated},
		SessionID:   session.ID,
		Session:     *session,
		Status:      "created",
	}

	log.Printf("Sending session_created response: %+v", response)
	if err := ws.WriteJSON(response); err != nil {
		log.Printf("ERROR: Failed to send session_created response: %v", err)
		return err
	}

	log.Printf("session_created response sent successfully")
	return nil
}

// handleSendPrompt sends a prompt to an agent session
func (h *AgentHandler) handleSendPrompt(ws *websocket.Conn, rawMsg map[string]interface{}) error {
	var msg SendPromptMessage
	msgBytes, _ := json.Marshal(rawMsg)
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		return fmt.Errorf("invalid send_prompt message: %w", err)
	}

	if msg.Prompt == "" {
		return fmt.Errorf("prompt cannot be empty")
	}

	log.Printf("Sending prompt to session %s: %s", msg.SessionID, msg.Prompt)

	// Send prompt to session
	if err := h.SessionManager.SendPrompt(msg.SessionID, msg.Prompt); err != nil {
		return err
	}

	// Get response channel
	responseChan, err := h.SessionManager.GetResponseChannel(msg.SessionID)
	if err != nil {
		return err
	}

	// Stream responses back to client
	go h.streamResponses(ws, msg.SessionID, responseChan)

	return nil
}

// streamResponses streams Claude responses back to the WebSocket client
func (h *AgentHandler) streamResponses(ws *websocket.Conn, sessionID uuid.UUID, responseChan chan types.Message) {
	for msg := range responseChan {
		if err := h.sendAgentMessage(ws, sessionID, msg); err != nil {
			log.Printf("Error sending agent message: %v", err)
			return
		}
	}
}

// sendAgentMessage sends a Claude message to the WebSocket client
func (h *AgentHandler) sendAgentMessage(ws *websocket.Conn, sessionID uuid.UUID, msg types.Message) error {
	// Skip messages that shouldn't be displayed to users (e.g., system init, debug, result metadata)
	if !msg.ShouldDisplayToUser() {
		log.Printf("Skipping message not meant for user display: type=%s", msg.GetMessageType())
		return nil
	}

	msgType := msg.GetMessageType()
	log.Printf("sendAgentMessage: msgType=%s, msg=%+v", msgType, msg)

	var response AgentMessageResponse
	response.Type = MessageTypeAgentMessage
	response.SessionID = sessionID

	switch msgType {
	case "assistant":
		if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
			log.Printf("Assistant message type assertion succeeded, content blocks: %d", len(assistantMsg.Content))
			// Extract text content
			var textContent []string
			for i, block := range assistantMsg.Content {
				log.Printf("Block %d: type=%s, block=%+v", i, block.GetType(), block)
				if textBlock, ok := block.(*types.TextBlock); ok {
					log.Printf("TextBlock found with text: %s", textBlock.Text)
					textContent = append(textContent, textBlock.Text)
				} else {
					log.Printf("Block %d is not a TextBlock (type=%T)", i, block)
				}
			}
			log.Printf("Extracted %d text blocks: %v", len(textContent), textContent)
			response.Content = map[string]interface{}{
				"type": "assistant",
				"text": textContent,
			}
		} else {
			log.Printf("Failed to assert message as AssistantMessage (type=%T)", msg)
		}

	case "user":
		if userMsg, ok := msg.(*types.UserMessage); ok {
			response.Content = map[string]interface{}{
				"type":    "user",
				"content": userMsg.Content,
			}
		}

	case "result":
		if resultMsg, ok := msg.(*types.ResultMessage); ok {
			content := map[string]interface{}{
				"type":        "result",
				"success":     true,
				"num_turns":   resultMsg.NumTurns,
				"duration_ms": resultMsg.DurationMs,
				"is_error":    resultMsg.IsError,
			}
			if resultMsg.TotalCostUSD != nil {
				content["cost_usd"] = *resultMsg.TotalCostUSD
			}
			if resultMsg.Usage != nil {
				content["usage"] = resultMsg.Usage
			}
			response.Content = content
		}

	case "system":
		if systemMsg, ok := msg.(*types.SystemMessage); ok {
			response.Content = map[string]interface{}{
				"type":    "system",
				"subtype": systemMsg.Subtype,
				"data":    systemMsg.Data,
			}
		}

	default:
		response.Content = map[string]interface{}{
			"type": "unknown",
			"raw":  msg,
		}
	}

	return ws.WriteJSON(response)
}

// handleEndSession ends an agent session
func (h *AgentHandler) handleEndSession(ws *websocket.Conn, rawMsg map[string]interface{}) error {
	var msg EndSessionMessage
	msgBytes, _ := json.Marshal(rawMsg)
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		return fmt.Errorf("invalid end_session message: %w", err)
	}

	log.Printf("Ending session: %s", msg.SessionID)

	// End session
	if err := h.SessionManager.EndSession(msg.SessionID); err != nil {
		return err
	}

	// Send session ended response
	response := SessionEndedMessage{
		BaseMessage: BaseMessage{Type: MessageTypeSessionEnded},
		SessionID:   msg.SessionID,
		Status:      "ended",
	}

	return ws.WriteJSON(response)
}

// handleListSessions lists all active sessions
func (h *AgentHandler) handleListSessions(ws *websocket.Conn) error {
	sessions := h.SessionManager.ListSessions()

	response := SessionsListMessage{
		BaseMessage: BaseMessage{Type: MessageTypeSessionsList},
		Sessions:    sessions,
	}

	return ws.WriteJSON(response)
}

// handleKillAllAgents kills all active agent sessions
func (h *AgentHandler) handleKillAllAgents(ws *websocket.Conn) error {
	count := h.SessionManager.EndAllSessions()

	response := AgentsKilledMessage{
		BaseMessage: BaseMessage{Type: MessageTypeAgentsKilled},
		Count:       count,
	}

	return ws.WriteJSON(response)
}

// handlePing responds to ping with pong
func (h *AgentHandler) handlePing(ws *websocket.Conn) error {
	response := BaseMessage{Type: MessageTypePong}
	return ws.WriteJSON(response)
}

// sendError sends an error message to the WebSocket client
func (h *AgentHandler) sendError(ws *websocket.Conn, errMsg string) {
	resp := ErrorMessage{
		BaseMessage: BaseMessage{Type: MessageTypeError},
		Content:     nil,
		Message:     errMsg,
	}
	log.Printf("Sending error message: %s", errMsg)
	if err := ws.WriteJSON(resp); err != nil {
		// Don't log broken pipe errors - client already disconnected
		if !websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) &&
			err.Error() != "write: broken pipe" {
			log.Printf("Failed to send error message: %v", err)
		}
	}
}

// GetStats returns current handler statistics
func (h *AgentHandler) GetStats() map[string]interface{} {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	sessions := h.SessionManager.ListSessions()

	return map[string]interface{}{
		"active_connections": h.Active,
		"max_connections":    h.Config.MaxConcurrentSessions,
		"active_sessions":    len(sessions),
	}
}

// Cleanup ends all active sessions gracefully
func (h *AgentHandler) Cleanup() error {
	count := h.SessionManager.EndAllSessions()
	log.Printf("Cleaned up %d active agent sessions", count)
	return nil
}

// HealthCheck endpoint handler
func (h *AgentHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %v", err)
		return
	}
	defer func() {
		_ = ws.Close()
	}()

	stats := h.GetStats()
	_ = ws.WriteJSON(stats)
}
