package agents

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	fiberws "github.com/gofiber/websocket/v2"
	"github.com/schlunsen/claude-agent-sdk-go/types"
	"github.com/schlunsen/claude-control-terminal/internal/logging"
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

// NewAgentHandler creates a new agent handler with the given config and database
func NewAgentHandler(config *Config, db *sql.DB) (*AgentHandler, error) {
	sessionManager, err := NewSessionManager(config, db)
	if err != nil {
		return nil, fmt.Errorf("failed to create session manager: %w", err)
	}

	return &AgentHandler{
		Config:         config,
		SessionManager: sessionManager,
		Active:         0,
	}, nil
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

		log.Printf("📥 WS INCOMING: type=%s, sessionID=%v, data=%+v", msgType, rawMsg["session_id"], rawMsg)

		// Route message to appropriate handler
		if err := h.routeMessage(ws, MessageType(msgType), rawMsg); err != nil {
			log.Printf("ERROR: Failed to handle message type %s: %v", msgType, err)
			h.sendError(ws, fmt.Sprintf("message handling failed: %v", err))
		}
	}
}

// HandleFiberWebSocket returns a Fiber WebSocket handler function
// This is compatible with Fiber's WebSocket middleware
func (h *AgentHandler) HandleFiberWebSocket(c *fiberws.Conn) {
	log.Printf("HandleFiberWebSocket: New WebSocket connection from %s", c.RemoteAddr())

	// Check concurrent session limit
	h.Mu.Lock()
	if h.Active >= h.Config.MaxConcurrentSessions {
		h.Mu.Unlock()
		logging.Warning("Max concurrent sessions reached: %d/%d", h.Active, h.Config.MaxConcurrentSessions)
		c.WriteJSON(map[string]interface{}{
			"type":    "error",
			"message": "max concurrent sessions reached",
		})
		return
	}
	h.Active++
	log.Printf("HandleFiberWebSocket: Active connections: %d/%d", h.Active, h.Config.MaxConcurrentSessions)
	h.Mu.Unlock()

	// Track which sessions are connected via this WebSocket
	connectedSessions := make(map[uuid.UUID]bool)
	var connectedSessionsMu sync.Mutex

	// Helper function to register a session with this WebSocket
	registerSession := func(sessionID uuid.UUID) {
		connectedSessionsMu.Lock()
		defer connectedSessionsMu.Unlock()

		// If session was previously connected via another WebSocket, clean it up first
		if session, err := h.SessionManager.GetSession(sessionID); err == nil {
			if session.IsWebSocketConnected() {
				logging.Warning("Session %s reconnecting - cleaning up old WebSocket state", sessionID)
				session.CleanupPendingPermissions()
				session.StopPermissionForwarder()
			}
			session.SetWebSocketConnected(true)
		}

		connectedSessions[sessionID] = true
		logging.Info("Session %s registered with WebSocket connection", sessionID)
	}

	defer func() {
		h.Mu.Lock()
		h.Active--
		h.Mu.Unlock()
		logging.Debug("WebSocket connection closed, active connections: %d", h.Active)

		// Clean up all sessions connected via this WebSocket
		connectedSessionsMu.Lock()
		for sessionID := range connectedSessions {
			if session, err := h.SessionManager.GetSession(sessionID); err == nil {
				logging.Info("Disconnecting session %s due to WebSocket close", sessionID)
				session.SetWebSocketConnected(false)
				session.CleanupPendingPermissions()
				session.StopPermissionForwarder()
			}
		}
		connectedSessionsMu.Unlock()
	}()

	log.Printf("Fiber WebSocket connection established from %s", c.RemoteAddr().String())
	logging.Info("WebSocket connection established from %s (active: %d)", c.RemoteAddr().String(), h.Active)

	// Main message loop
	for {
		var rawMsg map[string]interface{}
		if err := c.ReadJSON(&rawMsg); err != nil {
			if fiberws.IsUnexpectedCloseError(err, fiberws.CloseGoingAway, fiberws.CloseAbnormalClosure) {
				log.Printf("Error receiving message: %v", err)
			}
			return
		}

		msgType, ok := rawMsg["type"].(string)
		if !ok {
			log.Printf("ERROR: Missing or invalid message type in: %+v", rawMsg)
			h.sendFiberError(c, "missing or invalid message type")
			continue
		}

		log.Printf("📥 WS INCOMING: type=%s, sessionID=%v, data=%+v", msgType, rawMsg["session_id"], rawMsg)

		// Route message to appropriate handler
		if err := h.routeFiberMessage(c, MessageType(msgType), rawMsg, registerSession); err != nil {
			log.Printf("ERROR: Failed to handle message type %s: %v", msgType, err)
			h.sendFiberError(c, fmt.Sprintf("message handling failed: %v", err))
		}
	}
}

// sendFiberError sends an error message via Fiber WebSocket
func (h *AgentHandler) sendFiberError(c *fiberws.Conn, errMsg string) {
	err := c.WriteJSON(map[string]interface{}{
		"type":    "error",
		"message": errMsg,
	})
	if err != nil {
		log.Printf("Failed to send error message: %v", err)
	}
}

// routeFiberMessage routes messages to appropriate handlers for Fiber WebSocket
func (h *AgentHandler) routeFiberMessage(c *fiberws.Conn, msgType MessageType, rawMsg map[string]interface{}, registerSession func(uuid.UUID)) error {
	switch msgType {
	case MessageTypeAuth:
		// Authentication handled by server middleware, skip
		return nil

	case MessageTypeCreateSession:
		return h.handleFiberCreateSession(c, rawMsg, registerSession)

	case MessageTypeSendPrompt:
		return h.handleFiberSendPrompt(c, rawMsg, registerSession)

	case MessageTypeEndSession:
		return h.handleFiberEndSession(c, rawMsg)

	case MessageTypeInterruptSession:
		return h.handleFiberInterruptSession(c, rawMsg)

	case MessageTypeDeleteSession:
		return h.handleFiberDeleteSession(c, rawMsg)

	case MessageTypeListSessions:
		return h.handleFiberListSessions(c, registerSession)

	case MessageTypeLoadMessages:
		return h.handleFiberLoadMessages(c, rawMsg)

	case MessageTypeKillAllAgents:
		return h.handleFiberKillAllAgents(c)

	case MessageTypeDeleteAllSessions:
		return h.handleFiberDeleteAllSessions(c)

	case MessageTypePing:
		return h.handleFiberPing(c)

	case MessageTypePermissionResponse:
		return h.handleFiberPermissionResponse(c, rawMsg)

	case MessageTypeAddAlwaysAllowRule:
		return h.handleFiberAddAlwaysAllowRule(c, rawMsg)

	case MessageTypeRemoveAlwaysAllowRule:
		return h.handleFiberRemoveAlwaysAllowRule(c, rawMsg)

	case MessageTypeListAlwaysAllowRules:
		return h.handleFiberListAlwaysAllowRules(c, rawMsg)

	default:
		return fmt.Errorf("unknown message type: %s", msgType)
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

		// Stop after result message (completion signal)
		if msg.GetMessageType() == "result" {
			log.Printf("Session %s: Streaming complete (received result message)", sessionID)
			return
		}
	}
}

// streamFiberResponses streams Claude responses back to the Fiber WebSocket client
func (h *AgentHandler) streamFiberResponses(c *fiberws.Conn, sessionID uuid.UUID, responseChan chan types.Message) {
	for msg := range responseChan {
		if err := h.sendFiberAgentMessage(c, sessionID, msg); err != nil {
			log.Printf("Error sending agent message: %v", err)
			return
		}

		// Stop after result message (completion signal)
		if msg.GetMessageType() == "result" {
			log.Printf("Session %s: Streaming complete (received result message)", sessionID)
			return
		}
	}
}

// sendAgentMessage sends a Claude message to the WebSocket client
func (h *AgentHandler) sendAgentMessage(ws *websocket.Conn, sessionID uuid.UUID, msg types.Message) error {
	msgType := msg.GetMessageType()
	log.Printf("sendAgentMessage: msgType=%s, msg=%+v", msgType, msg)

	var response AgentMessageResponse
	response.Type = MessageTypeAgentMessage
	response.SessionID = sessionID

	switch msgType {
	case "assistant":
		if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
			log.Printf("Assistant message type assertion succeeded, content blocks: %d", len(assistantMsg.Content))
			var textContent []string
			var toolUses []map[string]interface{}

			for i, block := range assistantMsg.Content {
				log.Printf("Block %d: type=%s, block=%+v", i, block.GetType(), block)

				if textBlock, ok := block.(*types.TextBlock); ok {
					log.Printf("TextBlock found with text: %s", textBlock.Text)
					textContent = append(textContent, textBlock.Text)
				} else if toolUseBlock, ok := block.(*types.ToolUseBlock); ok {
					log.Printf("ToolUseBlock found: name=%s, id=%s", toolUseBlock.Name, toolUseBlock.ID)
					toolUses = append(toolUses, map[string]interface{}{
						"id":     toolUseBlock.ID,
						"name":   toolUseBlock.Name,
						"input":  toolUseBlock.Input,
						"status": "running",
					})

					// Broadcast agent_tool_use event for metrics tracking
					toolUseEvent := map[string]interface{}{
						"type":       string(MessageTypeAgentToolUse),
						"session_id": sessionID.String(),
						"tool":       toolUseBlock.Name,
						"parameters": toolUseBlock.Input,
					}
					if err := ws.WriteJSON(toolUseEvent); err != nil {
						log.Printf("Failed to send agent_tool_use event: %v", err)
					}
				} else {
					log.Printf("Block %d is not a TextBlock or ToolUseBlock (type=%T)", i, block)
				}
			}
			log.Printf("Extracted %d text blocks and %d tool uses", len(textContent), len(toolUses))

			response.Content = map[string]interface{}{
				"type":  "assistant",
				"text":  textContent,
				"tools": toolUses,
			}
		} else {
			log.Printf("Failed to assert message as AssistantMessage (type=%T)", msg)
		}

	case "user":
		if userMsg, ok := msg.(*types.UserMessage); ok {
			var toolResults []map[string]interface{}

			// Check if user message content is a slice of ContentBlocks (tool results)
			if contentBlocks, ok := userMsg.Content.([]types.ContentBlock); ok {
				for _, block := range contentBlocks {
					if toolResultBlock, ok := block.(*types.ToolResultBlock); ok {
						log.Printf("ToolResultBlock found: tool_use_id=%s", toolResultBlock.ToolUseID)
						toolResults = append(toolResults, map[string]interface{}{
							"tool_use_id": toolResultBlock.ToolUseID,
							"content":     toolResultBlock.Content,
							"is_error":    toolResultBlock.IsError,
							"status":      "completed",
						})
					}
				}
			}

			response.Content = map[string]interface{}{
				"type":         "user",
				"content":      userMsg.Content,
				"tool_results": toolResults,
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

	case "system", "control_request":
		if systemMsg, ok := msg.(*types.SystemMessage); ok {
			// Check if this is a permission request (control_request)
			if msg.GetMessageType() == "control_request" && systemMsg.Request != nil {
				// This is a permission request - forward to frontend as permission_request
				response.Type = MessageTypePermissionRequest
				response.Content = map[string]interface{}{
					"type":          "permission_request",
					"permission_id": systemMsg.Request["permission_id"],
					"tool":          systemMsg.Request["tool"],
					"action":        systemMsg.Request["action"],
					"details":       systemMsg.Request,
				}
			} else {
				// Regular system message
				response.Content = map[string]interface{}{
					"type":    "system",
					"subtype": systemMsg.Subtype,
					"data":    systemMsg.Data,
				}
			}
		}

	default:
		response.Content = map[string]interface{}{
			"type": "unknown",
			"raw":  msg,
		}
	}

	// Add git branch to metadata
	session, err := h.SessionManager.GetSession(sessionID)
	if err == nil && session.GitBranch != "" {
		response.Metadata = map[string]interface{}{
			"git_branch": session.GitBranch,
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

// handleListSessions lists all sessions from database
func (h *AgentHandler) handleListSessions(ws *websocket.Conn) error {
	sessions, err := h.SessionManager.ListAllSessions("all")
	if err != nil {
		h.sendError(ws, fmt.Sprintf("failed to list sessions: %v", err))
		return fmt.Errorf("failed to list sessions: %w", err)
	}

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

// Fiber WebSocket Handler Methods (duplicates of above for Fiber compatibility)

// handleFiberCreateSession creates a new agent session (Fiber version)
func (h *AgentHandler) handleFiberCreateSession(c *fiberws.Conn, rawMsg map[string]interface{}, registerSession func(uuid.UUID)) error {
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

	// Register session with this WebSocket connection
	registerSession(msg.SessionID)

	log.Printf("Session created successfully: %s", session.ID)

	// Send session created response
	response := SessionCreatedMessage{
		BaseMessage: BaseMessage{Type: MessageTypeSessionCreated},
		SessionID:   session.ID,
		Session:     *session,
		Status:      "created",
	}

	log.Printf("Sending session_created response: %+v", response)
	if err := c.WriteJSON(response); err != nil {
		log.Printf("ERROR: Failed to send session_created response: %v", err)
		return err
	}

	log.Printf("session_created response sent successfully")
	return nil
}

// handleFiberSendPrompt sends a prompt to an agent session (Fiber version)
// Note: This returns a response channel that must be monitored by the main handler
func (h *AgentHandler) handleFiberSendPrompt(c *fiberws.Conn, rawMsg map[string]interface{}, registerSession func(uuid.UUID)) error {
	var msg SendPromptMessage
	msgBytes, _ := json.Marshal(rawMsg)
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		return fmt.Errorf("invalid send_prompt message: %w", err)
	}

	// Check if we have content array (new format) or prompt string (legacy format)
	hasContent := len(msg.Content) > 0
	hasPrompt := msg.Prompt != ""

	if !hasContent && !hasPrompt {
		return fmt.Errorf("either prompt or content must be provided")
	}

	// Get the session first
	session, err := h.SessionManager.GetSession(msg.SessionID)
	if err != nil {
		return err
	}

	// Register session with this WebSocket connection (handles reconnections)
	registerSession(msg.SessionID)

	// Start monitoring for permission requests BEFORE sending the prompt
	// This prevents a race condition where the SDK requests permission
	// before the goroutine is ready to receive it
	// Only start if not already running to prevent multiple goroutines
	if session.StartPermissionForwarder() {
		go h.forwardPermissionRequests(c, msg.SessionID, session)
	}

	// Send prompt or content to session
	if hasContent {
		// New format: structured content with images
		log.Printf("Sending structured content to session %s (%d blocks)", msg.SessionID, len(msg.Content))
		if err := h.SessionManager.SendPromptWithContent(msg.SessionID, msg.Content); err != nil {
			return err
		}
	} else {
		// Legacy format: plain text prompt
		log.Printf("Sending prompt to session %s: %s", msg.SessionID, msg.Prompt)
		if err := h.SessionManager.SendPrompt(msg.SessionID, msg.Prompt); err != nil {
			return err
		}
	}

	// Get response channel
	responseChan, err := h.SessionManager.GetResponseChannel(msg.SessionID)
	if err != nil {
		return err
	}

	// Stream responses back to client in a goroutine
	// This allows the handler to process subsequent prompts
	go h.streamFiberResponses(c, msg.SessionID, responseChan)

	return nil
}

// sendFiberAgentMessage sends a Claude message to the WebSocket client (Fiber version)
func (h *AgentHandler) sendFiberAgentMessage(c *fiberws.Conn, sessionID uuid.UUID, msg types.Message) error {
	msgType := msg.GetMessageType()
	log.Printf("sendFiberAgentMessage: msgType=%s, msg=%+v", msgType, msg)

	var response AgentMessageResponse
	response.Type = MessageTypeAgentMessage
	response.SessionID = sessionID

	switch msgType {
	case "assistant":
		if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
			log.Printf("Assistant message type assertion succeeded, content blocks: %d", len(assistantMsg.Content))
			var textContent []string
			var toolUses []map[string]interface{}

			for i, block := range assistantMsg.Content {
				log.Printf("Block %d: type=%s, block=%+v", i, block.GetType(), block)

				if textBlock, ok := block.(*types.TextBlock); ok {
					log.Printf("TextBlock found with text: %s", textBlock.Text)
					textContent = append(textContent, textBlock.Text)
				} else if toolUseBlock, ok := block.(*types.ToolUseBlock); ok {
					log.Printf("ToolUseBlock found: name=%s, id=%s", toolUseBlock.Name, toolUseBlock.ID)
					toolUses = append(toolUses, map[string]interface{}{
						"id":     toolUseBlock.ID,
						"name":   toolUseBlock.Name,
						"input":  toolUseBlock.Input,
						"status": "running",
					})

					// Broadcast agent_tool_use event for metrics tracking
					toolUseEvent := map[string]interface{}{
						"type":       string(MessageTypeAgentToolUse),
						"session_id": sessionID.String(),
						"tool":       toolUseBlock.Name,
						"parameters": toolUseBlock.Input,
					}
					if err := c.WriteJSON(toolUseEvent); err != nil {
						log.Printf("Failed to send agent_tool_use event: %v", err)
					}
				} else {
					log.Printf("Block %d is not a TextBlock or ToolUseBlock (type=%T)", i, block)
				}
			}
			log.Printf("Extracted %d text blocks and %d tool uses", len(textContent), len(toolUses))

			response.Content = map[string]interface{}{
				"type":  "assistant",
				"text":  textContent,
				"tools": toolUses,
			}
		} else {
			log.Printf("Failed to assert message as AssistantMessage (type=%T)", msg)
		}

	case "user":
		if userMsg, ok := msg.(*types.UserMessage); ok {
			var toolResults []map[string]interface{}

			// Check if user message content is a slice of ContentBlocks (tool results)
			if contentBlocks, ok := userMsg.Content.([]types.ContentBlock); ok {
				for _, block := range contentBlocks {
					if toolResultBlock, ok := block.(*types.ToolResultBlock); ok {
						log.Printf("ToolResultBlock found: tool_use_id=%s", toolResultBlock.ToolUseID)
						toolResults = append(toolResults, map[string]interface{}{
							"tool_use_id": toolResultBlock.ToolUseID,
							"content":     toolResultBlock.Content,
							"is_error":    toolResultBlock.IsError,
							"status":      "completed",
						})
					}
				}
			}

			response.Content = map[string]interface{}{
				"type":         "user",
				"content":      userMsg.Content,
				"tool_results": toolResults,
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

	case "system", "control_request":
		if systemMsg, ok := msg.(*types.SystemMessage); ok {
			// Check if this is a permission request (control_request)
			if msg.GetMessageType() == "control_request" && systemMsg.Request != nil {
				// This is a permission request - forward to frontend as permission_request
				log.Printf("🔐 Permission request detected: tool=%v, action=%v", systemMsg.Request["tool"], systemMsg.Request["action"])
				response.Type = MessageTypePermissionRequest
				response.Content = map[string]interface{}{
					"type":          "permission_request",
					"permission_id": systemMsg.Request["permission_id"],
					"tool":          systemMsg.Request["tool"],
					"action":        systemMsg.Request["action"],
					"details":       systemMsg.Request,
				}
			} else {
				// Regular system message
				response.Content = map[string]interface{}{
					"type":    "system",
					"subtype": systemMsg.Subtype,
					"data":    systemMsg.Data,
				}
			}
		}

	default:
		log.Printf("Unknown message type: %s", msgType)
		return fmt.Errorf("unknown message type: %s", msgType)
	}

	// Add git branch to metadata
	session, err := h.SessionManager.GetSession(sessionID)
	if err == nil && session.GitBranch != "" {
		response.Metadata = map[string]interface{}{
			"git_branch": session.GitBranch,
		}
	}

	log.Printf("📤 WS OUTGOING: type=%s, sessionID=%s, response=%+v", response.Type, response.SessionID, response)
	if err := c.WriteJSON(response); err != nil {
		log.Printf("ERROR: Failed to send agent message: %v", err)
		return err
	}

	log.Printf("✅ Message sent to WebSocket client")
	return nil
}

// handleFiberEndSession ends an agent session (Fiber version)
func (h *AgentHandler) handleFiberEndSession(c *fiberws.Conn, rawMsg map[string]interface{}) error {
	var msg EndSessionMessage
	msgBytes, _ := json.Marshal(rawMsg)
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		return fmt.Errorf("invalid end_session message: %w", err)
	}

	// End session
	if err := h.SessionManager.EndSession(msg.SessionID); err != nil {
		return err
	}

	// Send session ended response
	response := BaseMessage{Type: MessageTypeSessionEnded}
	return c.WriteJSON(response)
}

// handleFiberInterruptSession interrupts an agent session (Fiber version)
func (h *AgentHandler) handleFiberInterruptSession(c *fiberws.Conn, rawMsg map[string]interface{}) error {
	var msg InterruptSessionMessage
	msgBytes, _ := json.Marshal(rawMsg)
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		return fmt.Errorf("invalid interrupt_session message: %w", err)
	}

	logging.Info("Interrupting session: %s", msg.SessionID)

	// Interrupt session (cancels context but keeps session alive)
	if err := h.SessionManager.InterruptSession(msg.SessionID); err != nil {
		logging.Error("Failed to interrupt session %s: %v", msg.SessionID, err)
		h.sendFiberError(c, fmt.Sprintf("failed to interrupt session: %v", err))
		return err
	}

	// Send session interrupted response
	response := SessionInterruptedMessage{
		BaseMessage: BaseMessage{Type: MessageTypeSessionInterrupted},
		SessionID:   msg.SessionID,
		Status:      "interrupted",
	}
	return c.WriteJSON(response)
}

// handleFiberDeleteSession deletes an agent session (Fiber version)
func (h *AgentHandler) handleFiberDeleteSession(c *fiberws.Conn, rawMsg map[string]interface{}) error {
	var msg DeleteSessionMessage
	msgBytes, _ := json.Marshal(rawMsg)
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		return fmt.Errorf("invalid delete_session message: %w", err)
	}

	// Delete session from database
	if err := h.SessionManager.DeleteSession(msg.SessionID); err != nil {
		h.sendFiberError(c, fmt.Sprintf("failed to delete session: %v", err))
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Send session deleted response
	response := SessionDeletedMessage{
		BaseMessage: BaseMessage{Type: MessageTypeSessionDeleted},
		SessionID:   msg.SessionID,
		Status:      "deleted",
	}
	return c.WriteJSON(response)
}

// handleFiberListSessions lists all sessions from database (Fiber version)
func (h *AgentHandler) handleFiberListSessions(c *fiberws.Conn, registerSession func(uuid.UUID)) error {
	log.Printf("handleFiberListSessions: Fetching all sessions from database")
	sessions, err := h.SessionManager.ListAllSessions("all")
	if err != nil {
		log.Printf("ERROR: Failed to list sessions: %v", err)
		h.sendFiberError(c, fmt.Sprintf("failed to list sessions: %v", err))
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	log.Printf("handleFiberListSessions: Found %d sessions in database", len(sessions))
	for i, session := range sessions {
		log.Printf("  Session %d: ID=%s, Status=%s, Created=%s", i+1, session.ID, session.Status, session.CreatedAt)

		// Register active sessions with this WebSocket connection
		// This allows reconnection after page reload
		if session.Status == SessionStatusActive || session.Status == SessionStatusIdle || session.Status == SessionStatusProcessing {
			registerSession(session.ID)
			logging.Info("Registered active session %s with reconnected WebSocket", session.ID)
		}
	}

	response := SessionsListMessage{
		BaseMessage: BaseMessage{Type: MessageTypeSessionsList},
		Sessions:    sessions,
	}

	log.Printf("handleFiberListSessions: Sending response with %d sessions", len(sessions))
	return c.WriteJSON(response)
}

// handleFiberLoadMessages loads messages for a session with pagination (Fiber version)
func (h *AgentHandler) handleFiberLoadMessages(c *fiberws.Conn, rawMsg map[string]interface{}) error {
	// Parse session ID
	sessionIDStr, ok := rawMsg["session_id"].(string)
	if !ok {
		h.sendFiberError(c, "missing or invalid session_id")
		return fmt.Errorf("missing or invalid session_id")
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		h.sendFiberError(c, "invalid session ID format")
		return fmt.Errorf("invalid session ID format")
	}

	// Parse pagination params with defaults
	limit := 50
	offset := 0

	if limitVal, ok := rawMsg["limit"].(float64); ok {
		limit = int(limitVal)
	}
	if offsetVal, ok := rawMsg["offset"].(float64); ok {
		offset = int(offsetVal)
	}

	// Validate pagination params
	// Increased max limit to 1000 to support long conversations
	if limit < 1 || limit > 1000 {
		limit = 100 // Default to 100 instead of 50
	}
	if offset < 0 {
		offset = 0
	}

	// Get messages from storage
	messagesPtr, hasMore, err := h.SessionManager.GetMessages(sessionID, limit, offset)
	if err != nil {
		h.sendFiberError(c, fmt.Sprintf("failed to load messages: %v", err))
		return fmt.Errorf("failed to load messages: %w", err)
	}

	// Convert []*MessageRecord to []MessageRecord
	messages := make([]MessageRecord, len(messagesPtr))
	for i, msgPtr := range messagesPtr {
		messages[i] = *msgPtr
	}

	// Send response
	response := MessagesLoadedMessage{
		BaseMessage: BaseMessage{Type: MessageTypeMessagesLoaded},
		SessionID:   sessionID,
		Messages:    messages,
		HasMore:     hasMore,
		Count:       len(messages),
		Limit:       limit,
		Offset:      offset,
	}

	return c.WriteJSON(response)
}

// handleFiberKillAllAgents kills all active agent sessions (Fiber version)
func (h *AgentHandler) handleFiberKillAllAgents(c *fiberws.Conn) error {
	count := h.SessionManager.EndAllSessions()
	response := map[string]interface{}{
		"type":    "kill_all_agents_response",
		"count":   count,
		"message": fmt.Sprintf("Killed %d agent sessions", count),
	}
	return c.WriteJSON(response)
}

// handleFiberDeleteAllSessions deletes all sessions from database (Fiber version)
func (h *AgentHandler) handleFiberDeleteAllSessions(c *fiberws.Conn) error {
	count, err := h.SessionManager.DeleteAllSessions()
	if err != nil {
		h.sendFiberError(c, fmt.Sprintf("failed to delete all sessions: %v", err))
		return fmt.Errorf("failed to delete all sessions: %w", err)
	}

	response := AllSessionsDeletedMessage{
		BaseMessage: BaseMessage{Type: MessageTypeAllSessionsDeleted},
		Count:       count,
	}
	return c.WriteJSON(response)
}

// forwardPermissionRequests monitors the session's permission request channel
// and forwards requests to the WebSocket client
func (h *AgentHandler) forwardPermissionRequests(c *fiberws.Conn, sessionID uuid.UUID, session *AgentSession) {
	logging.Info("🚀 Permission forwarder started for session %s", sessionID)

	defer func() {
		session.StopPermissionForwarder()
		logging.Info("🛑 Permission forwarder stopped for session %s", sessionID)
	}()

	// Create a ticker to periodically check WebSocket connection state
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case permReq, ok := <-session.permissionReqChan:
			if !ok {
				logging.Info("Permission request channel closed for session %s", sessionID)
				return
			}

			// Check if WebSocket is still connected before forwarding
			if !session.IsWebSocketConnected() {
				logging.Warning("⚠️ WebSocket disconnected, denying permission request: %s", permReq.RequestID)
				select {
				case permReq.ResponseChan <- PermissionResponse{
					Approved:    false,
					DenyMessage: "WebSocket connection lost",
				}:
				default:
				}
				continue
			}

			logging.Info("🔐 PERMISSION REQUEST RECEIVED FROM CHANNEL: tool=%s, requestID=%s, input=%+v", permReq.ToolName, permReq.RequestID, permReq.Input)

			// Generate human-readable description
			description := formatPermissionDescription(permReq.ToolName, permReq.Input)

			// Send permission request to frontend
			response := PermissionRequestMessage{
				BaseMessage:    BaseMessage{Type: MessageTypePermissionRequest},
				SessionID:      sessionID,
				PermissionID:   permReq.RequestID,
				Tool:           permReq.ToolName,
				Action:         "use_tool",
				Details:        permReq.Input,
				Description:    description,
			}

			logging.Info("📤 WS SENDING PERMISSION REQUEST TO FRONTEND: permissionID=%s, tool=%s, description=%s", permReq.RequestID, permReq.ToolName, description)

			if err := c.WriteJSON(response); err != nil {
				logging.Error("❌ Failed to send permission request to WebSocket: %v", err)

				// Mark session as disconnected
				session.SetWebSocketConnected(false)

				// Send error response back to callback
				select {
				case permReq.ResponseChan <- PermissionResponse{
					Approved:    false,
					DenyMessage: "Failed to send permission request to frontend (WebSocket error)",
				}:
				default:
				}

				// Clean up any other pending permissions
				session.CleanupPendingPermissions()
				return
			}

			logging.Info("✅ Permission request sent to WebSocket successfully: %s", permReq.RequestID)

		case <-ticker.C:
			// Periodically check if WebSocket is still connected
			if !session.IsWebSocketConnected() {
				logging.Info("⏱️ WebSocket connection lost, stopping permission forwarder for session %s", sessionID)
				session.CleanupPendingPermissions()
				return
			}

		case <-session.ctx.Done():
			logging.Info("Session %s context cancelled, stopping permission request forwarding", sessionID)
			session.CleanupPendingPermissions()
			return
		}
	}
}

// handleFiberPermissionResponse handles permission responses from the frontend
func (h *AgentHandler) handleFiberPermissionResponse(c *fiberws.Conn, rawMsg map[string]interface{}) error {
	logging.Info("📥 RAW PERMISSION RESPONSE from frontend: %+v", rawMsg)

	var msg PermissionResponseMessage
	msgBytes, _ := json.Marshal(rawMsg)
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		return fmt.Errorf("invalid permission_response message: %w", err)
	}

	logging.Info("📥 PARSED permission response: sessionID=%s, permissionID='%s', approved=%v",
		msg.SessionID, msg.PermissionID, msg.Approved)

	// Get the session
	session, err := h.SessionManager.GetSession(msg.SessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Find the pending permission request
	// NOTE: Frontend doesn't send permission_id, so we look for any pending permission in this session
	session.permMu.Lock()
	var responseChan chan PermissionResponse
	var exists bool

	// Try to find by permission ID first (if frontend sends it)
	if msg.PermissionID != "" {
		responseChan, exists = session.pendingPermissions[msg.PermissionID]
	}

	// If not found or no ID provided, get the first (and should be only) pending permission
	if !exists {
		for _, ch := range session.pendingPermissions {
			responseChan = ch
			exists = true
			break
		}
	}
	session.permMu.Unlock()

	if !exists {
		logging.Warning("No pending permission request found for session %s (permission_id='%s')", msg.SessionID, msg.PermissionID)
		return fmt.Errorf("no pending permission request found for ID: %s", msg.PermissionID)
	}

	logging.Info("✅ Found pending permission, sending response to callback")

	// Send response to the callback
	response := PermissionResponse{
		Approved:    msg.Approved,
		DenyMessage: "User denied permission",
	}

	select {
	case responseChan <- response:
		log.Printf("Permission response delivered to callback: %s", msg.PermissionID)
	case <-time.After(5 * time.Second):
		log.Printf("ERROR: Timeout delivering permission response to callback")
		return fmt.Errorf("timeout delivering permission response")
	}

	// Send acknowledgement to frontend
	ack := BaseMessage{Type: MessageTypePermissionAcknowledged}
	return c.WriteJSON(ack)
}

// handleFiberPing responds to ping with pong (Fiber version)
func (h *AgentHandler) handleFiberPing(c *fiberws.Conn) error {
	response := BaseMessage{Type: MessageTypePong}
	return c.WriteJSON(response)
}

// handleFiberAddAlwaysAllowRule adds an always-allow rule to a session
func (h *AgentHandler) handleFiberAddAlwaysAllowRule(c *fiberws.Conn, rawMsg map[string]interface{}) error {
	var msg AddAlwaysAllowRuleMessage
	msgBytes, _ := json.Marshal(rawMsg)
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		return fmt.Errorf("invalid add_always_allow_rule message: %w", err)
	}

	logging.Info("Adding always-allow rule to session %s: %s (mode: %s)", msg.SessionID, msg.Rule.Description, msg.Rule.MatchMode)

	// Get session to find working directory
	session, err := h.SessionManager.GetSession(msg.SessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Get working directory for this session
	workingDir := "."
	if session.Options.WorkingDirectory != nil {
		workingDir = *session.Options.WorkingDirectory
	}

	// Create settings manager
	settingsManager := NewClaudeSettingsManager(workingDir)

	// Format permission string in Claude Desktop format
	permissionStr := FormatPermissionString(msg.Rule.Tool, msg.Rule.Pattern)
	logging.Info("📝 Adding permission to settings.local.json: %s", permissionStr)

	// Add to settings.local.json
	if err := settingsManager.AddPermission(permissionStr); err != nil {
		logging.Error("Failed to add permission to settings: %v", err)
		return fmt.Errorf("failed to add permission: %w", err)
	}

	// Generate ID and timestamp for response
	if msg.Rule.ID == "" {
		msg.Rule.ID = uuid.New().String()
	}
	msg.Rule.CreatedAt = time.Now()

	// Add rule to in-memory session for immediate effect
	h.SessionManager.mu.Lock()
	session.Options.AlwaysAllowRules = append(session.Options.AlwaysAllowRules, msg.Rule)
	totalRules := len(session.Options.AlwaysAllowRules)
	h.SessionManager.mu.Unlock()

	logging.Info("✅ Rule added to session in-memory cache (total rules: %d)", totalRules)
	logging.Info("   Rule details: tool=%s, mode=%s, pattern=%v", msg.Rule.Tool, msg.Rule.MatchMode, msg.Rule.Pattern)

	// IMPORTANT: If there's a pending permission request (from the UI that triggered this),
	// we need to approve it now so the SDK can continue
	// The frontend will optimistically remove the permission from the UI
	if msg.PermissionID != "" {
		logging.Info("📤 Approving pending permission request: %s", msg.PermissionID)

		// Find the pending permission request
		session.permMu.Lock()
		responseChan, exists := session.pendingPermissions[msg.PermissionID]
		session.permMu.Unlock()

		if exists {
			// Create permission update to tell SDK about the new rule
			// This allows the SDK to apply the rule immediately without restarting
			allowBehavior := types.PermissionBehaviorAllow
			sessionDest := types.DestinationSession
			permUpdate := types.PermissionUpdate{
				Type:     "addRules",
				Behavior: &allowBehavior,
				Rules: []types.PermissionRuleValue{
					{
						ToolName:    msg.Rule.Tool,
						RuleContent: &permissionStr,
					},
				},
				Destination: &sessionDest,
			}

			// Send approval response to the SDK's callback with the updated permissions
			// This allows the SDK to continue execution with the newly added permission
			select {
			case responseChan <- PermissionResponse{
				Approved:           true,
				UpdatedPermissions: []types.PermissionUpdate{permUpdate},
				DenyMessage:        "",
			}:
				logging.Info("✅ Permission approval sent to SDK for request: %s (with permission update)", msg.PermissionID)
				// Clean up the pending permission immediately after approval
				session.permMu.Lock()
				delete(session.pendingPermissions, msg.PermissionID)
				session.permMu.Unlock()
			case <-time.After(2 * time.Second):
				logging.Warning("⚠️ Timeout sending permission approval to SDK")
			}
		} else {
			logging.Warning("⚠️ No pending permission found for ID: %s", msg.PermissionID)
		}
	}

	// Reload session settings to pick up the new rule from settings.local.json
	// This closes and recreates the client, forcing the CLI to reload settings
	logging.Info("🔄 Reloading session settings to apply new always-allow rule")
	if err := h.SessionManager.ReloadSessionSettings(msg.SessionID); err != nil {
		logging.Error("Failed to reload session settings: %v", err)
		// Don't fail the request - the rule is still saved and will work after restart
	} else {
		// After reloading (which interrupts the session), automatically send "continue"
		// to resume execution with the new settings applied
		logging.Info("▶️  Auto-resuming session with 'continue' command")
		go func() {
			// Small delay to ensure the response is sent first
			time.Sleep(100 * time.Millisecond)
			if err := h.SessionManager.SendPrompt(msg.SessionID, "continue"); err != nil {
				logging.Error("Failed to auto-continue session: %v", err)
			}
		}()
	}

	// Send confirmation with the full rule (including generated ID)
	response := AlwaysAllowRulesListMessage{
		BaseMessage: BaseMessage{Type: MessageTypeAlwaysAllowRulesList},
		SessionID:   msg.SessionID,
		Rules:       session.Options.AlwaysAllowRules,
	}

	return c.WriteJSON(response)
}

// handleFiberRemoveAlwaysAllowRule removes an always-allow rule from a session
func (h *AgentHandler) handleFiberRemoveAlwaysAllowRule(c *fiberws.Conn, rawMsg map[string]interface{}) error {
	var msg RemoveAlwaysAllowRuleMessage
	msgBytes, _ := json.Marshal(rawMsg)
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		return fmt.Errorf("invalid remove_always_allow_rule message: %w", err)
	}

	logging.Info("Removing always-allow rule %s from session %s", msg.RuleID, msg.SessionID)

	// Get session
	session, err := h.SessionManager.GetSession(msg.SessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Get working directory for this session
	workingDir := "."
	if session.Options.WorkingDirectory != nil {
		workingDir = *session.Options.WorkingDirectory
	}

	// Create settings manager
	settingsManager := NewClaudeSettingsManager(workingDir)

	// Find the rule to remove and format its permission string
	h.SessionManager.mu.Lock()
	var ruleToRemove *AlwaysAllowRule
	newRules := []AlwaysAllowRule{}
	for _, rule := range session.Options.AlwaysAllowRules {
		if rule.ID != msg.RuleID {
			newRules = append(newRules, rule)
		} else {
			ruleToRemove = &rule
		}
	}
	session.Options.AlwaysAllowRules = newRules
	h.SessionManager.mu.Unlock()

	// Remove from settings.local.json
	if ruleToRemove != nil {
		permissionStr := FormatPermissionString(ruleToRemove.Tool, ruleToRemove.Pattern)
		logging.Info("🗑️ Removing permission from settings.local.json: %s", permissionStr)
		if err := settingsManager.RemovePermission(permissionStr); err != nil {
			logging.Error("Failed to remove permission from settings: %v", err)
		}
	}

	// Send updated rules list
	response := AlwaysAllowRulesListMessage{
		BaseMessage: BaseMessage{Type: MessageTypeAlwaysAllowRulesList},
		SessionID:   msg.SessionID,
		Rules:       session.Options.AlwaysAllowRules,
	}

	return c.WriteJSON(response)
}

// handleFiberListAlwaysAllowRules lists all always-allow rules for a session
func (h *AgentHandler) handleFiberListAlwaysAllowRules(c *fiberws.Conn, rawMsg map[string]interface{}) error {
	var msg ListAlwaysAllowRulesMessage
	msgBytes, _ := json.Marshal(rawMsg)
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		return fmt.Errorf("invalid list_always_allow_rules message: %w", err)
	}

	// Get session
	session, err := h.SessionManager.GetSession(msg.SessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Send rules list
	response := AlwaysAllowRulesListMessage{
		BaseMessage: BaseMessage{Type: MessageTypeAlwaysAllowRulesList},
		SessionID:   msg.SessionID,
		Rules:       session.Options.AlwaysAllowRules,
	}

	return c.WriteJSON(response)
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

// formatPermissionDescription generates a human-readable description for a permission request
func formatPermissionDescription(toolName string, input map[string]interface{}) string {
	switch toolName {
	case "Bash":
		if cmd, ok := input["command"].(string); ok {
			return fmt.Sprintf("Execute command: %s", cmd)
		}
		return "Execute a bash command"

	case "Read":
		if path, ok := input["file_path"].(string); ok {
			return fmt.Sprintf("Read file: %s", path)
		}
		return "Read a file"

	case "Write":
		if path, ok := input["file_path"].(string); ok {
			return fmt.Sprintf("Write to file: %s", path)
		}
		return "Write to a file"

	case "Edit":
		if path, ok := input["file_path"].(string); ok {
			return fmt.Sprintf("Edit file: %s", path)
		}
		return "Edit a file"

	case "Glob":
		if pattern, ok := input["pattern"].(string); ok {
			return fmt.Sprintf("Search files matching: %s", pattern)
		}
		return "Search for files"

	case "Grep":
		if pattern, ok := input["pattern"].(string); ok {
			return fmt.Sprintf("Search content matching: %s", pattern)
		}
		return "Search file contents"

	case "WebSearch":
		if query, ok := input["query"].(string); ok {
			return fmt.Sprintf("Web search: %s", query)
		}
		return "Perform a web search"

	case "WebFetch":
		if url, ok := input["url"].(string); ok {
			return fmt.Sprintf("Fetch URL: %s", url)
		}
		return "Fetch a web page"

	default:
		return fmt.Sprintf("Use %s tool", toolName)
	}
}
