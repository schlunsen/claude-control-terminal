// Package server provides the Fiber-based HTTP server and REST API for CCT analytics.
// It serves the analytics dashboard, WebSocket connections, and API endpoints
// for conversation data, process monitoring, and command history.
package server

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/schlunsen/claude-control-terminal/internal/analytics"
	"github.com/schlunsen/claude-control-terminal/internal/database"
	ws "github.com/schlunsen/claude-control-terminal/internal/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
)

// Server wraps the Fiber app and analytics components.
// It provides a complete analytics backend with WebSocket support for real-time updates.
type Server struct {
	app                  *fiber.App
	conversationAnalyzer *analytics.ConversationAnalyzer
	conversationParser   *analytics.ConversationParser
	stateCalculator      *analytics.StateCalculator
	processDetector      *analytics.ProcessDetector
	shellDetector        *analytics.ShellDetector
	fileWatcher          *analytics.FileWatcher
	wsHub                *ws.Hub
	resetTracker         *analytics.ResetTracker
	db                   *database.Database
	repo                 *database.Repository
	claudeDir            string
	port                 int
	quiet                bool // Suppress output when running in TUI
}

// NewServer creates a new Fiber server instance
func NewServer(claudeDir string, port int) *Server {
	return NewServerWithOptions(claudeDir, port, false)
}

// NewServerWithOptions creates a new Fiber server instance with options
func NewServerWithOptions(claudeDir string, port int, quiet bool) *Server {
	app := fiber.New(fiber.Config{
		AppName: "Claude Code Analytics",
		ServerHeader: "go-claude-templates",
		DisableStartupMessage: quiet, // Suppress Fiber startup banner in quiet mode
	})

	// Middleware
	app.Use(cors.New())

	// Only add logger middleware if not in quiet mode
	if !quiet {
		app.Use(logger.New())
	}

	return &Server{
		app:       app,
		claudeDir: claudeDir,
		port:      port,
		quiet:     quiet,
	}
}

// Setup initializes analytics components and routes
func (s *Server) Setup() error {
	// Initialize database
	dataDir := filepath.Join(s.claudeDir, "analytics_data")
	db, err := database.Initialize(dataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	s.db = db
	s.repo = database.NewRepository(db)

	// Initialize analytics components
	s.conversationAnalyzer = analytics.NewConversationAnalyzer(s.claudeDir)
	s.conversationParser = analytics.NewConversationParser(s.repo)
	s.stateCalculator = analytics.NewStateCalculator()
	s.processDetector = analytics.NewProcessDetector()
	s.shellDetector = analytics.NewShellDetector()
	s.resetTracker = analytics.NewResetTracker(s.claudeDir)

	// Initialize WebSocket hub
	s.wsHub = ws.NewHub()
	go s.wsHub.Run()

	// Setup API routes
	s.setupRoutes()

	// Serve static files
	s.ServeStaticFiles()

	return nil
}

// setupRoutes configures all API endpoints
func (s *Server) setupRoutes() {
	api := s.app.Group("/api")

	// Health check
	api.Get("/health", s.handleHealth)

	// Data endpoints
	api.Get("/data", s.handleGetData)
	api.Get("/conversations", s.handleGetConversations)
	api.Get("/processes", s.handleGetProcesses)
	api.Get("/shells", s.handleGetShells)
	api.Get("/stats", s.handleGetStats)

	// Refresh endpoint
	api.Post("/refresh", s.handleRefresh)

	// Reset endpoints
	api.Post("/reset/archive", s.handleResetArchive)
	api.Post("/reset/clear", s.handleResetClear)
	api.Post("/reset/soft", s.handleResetSoft)
	api.Delete("/reset", s.handleClearReset)
	api.Get("/reset/status", s.handleResetStatus)

	// Command history endpoints
	api.Get("/history/shell", s.handleGetShellHistory)
	api.Get("/history/claude", s.handleGetClaudeHistory)
	api.Get("/history/stats", s.handleGetCommandStats)
	api.Get("/db/stats", s.handleGetDBStats)

	// User prompts endpoints
	api.Get("/prompts", s.handleGetUserPrompts)
	api.Get("/prompts/stats", s.handleGetPromptStats)
	api.Get("/prompts/sessions", s.handleGetUniqueSessions)
	api.Post("/prompts", s.handleRecordUserPrompt)
	api.Delete("/prompts", s.handleClearAllPrompts)

	// WebSocket endpoint
	s.app.Get("/ws", websocket.New(s.wsHub.HandleWebSocket()))
}

// Handler: Health check
func (s *Server) handleHealth(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
		"time":   time.Now(),
	})
}

// Handler: Get all data
func (s *Server) handleGetData(c *fiber.Ctx) error {
	conversations, err := s.conversationAnalyzer.LoadConversations(s.stateCalculator)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	processes, _ := s.processDetector.DetectRunningClaudeProcesses()

	return c.JSON(fiber.Map{
		"conversations":      conversations,
		"activeProcessCount": len(processes),
		"claudeDir":          s.claudeDir,
		"timestamp":          time.Now(),
	})
}

// Handler: Get conversations
func (s *Server) handleGetConversations(c *fiber.Ctx) error {
	conversations, err := s.conversationAnalyzer.LoadConversations(s.stateCalculator)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(conversations)
}

// Handler: Get running processes
func (s *Server) handleGetProcesses(c *fiber.Ctx) error {
	processes, err := s.processDetector.DetectRunningClaudeProcesses()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	stats, _ := s.processDetector.GetProcessStats()

	return c.JSON(fiber.Map{
		"processes": processes,
		"stats":     stats,
	})
}

// Handler: Get statistics
func (s *Server) handleGetStats(c *fiber.Ctx) error {
	conversations, err := s.conversationAnalyzer.LoadConversations(s.stateCalculator)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	totalTokens := 0
	activeCount := 0

	for _, conv := range conversations {
		totalTokens += conv.Tokens
		if conv.Status == "active" {
			activeCount++
		}
	}

	// Apply soft reset delta if present
	adjustedTokens, adjustedConversations := s.resetTracker.ApplyDelta(totalTokens, len(conversations))

	avgTokens := 0
	if adjustedConversations > 0 {
		avgTokens = adjustedTokens / adjustedConversations
	}

	response := fiber.Map{
		"totalConversations":  adjustedConversations,
		"activeConversations": activeCount,
		"totalTokens":         adjustedTokens,
		"avgTokens":           avgTokens,
		"timestamp":           time.Now(),
	}

	// Include reset info if present
	if resetPoint := s.resetTracker.GetResetPoint(); resetPoint != nil {
		response["resetActive"] = true
		response["resetTimestamp"] = resetPoint.Timestamp
		response["resetReason"] = resetPoint.Reason
	} else {
		response["resetActive"] = false
	}

	return c.JSON(response)
}

// Handler: Get background shells
func (s *Server) handleGetShells(c *fiber.Ctx) error {
	shells, err := s.shellDetector.DetectBackgroundShells()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	stats, _ := s.shellDetector.GetShellStats()

	return c.JSON(fiber.Map{
		"shells": shells,
		"stats":  stats,
	})
}

// Handler: Refresh data
func (s *Server) handleRefresh(c *fiber.Ctx) error {
	// Clear caches
	s.stateCalculator.ClearCache()
	s.processDetector.ClearCache()
	s.shellDetector.ClearCache()

	return c.JSON(fiber.Map{
		"status": "refreshed",
		"time":   time.Now(),
	})
}

// Handler: Reset analytics by archiving conversations
func (s *Server) handleResetArchive(c *fiber.Ctx) error {
	err := s.conversationAnalyzer.ArchiveConversations()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   err.Error(),
			"status":  "failed",
		})
	}

	// Clear caches after reset
	s.stateCalculator.ClearCache()
	s.processDetector.ClearCache()
	s.shellDetector.ClearCache()

	// Broadcast update to WebSocket clients
	s.wsHub.Broadcast([]byte(`{"event":"reset","action":"archive"}`))

	return c.JSON(fiber.Map{
		"status":  "archived",
		"message": "All conversations have been archived",
		"time":    time.Now(),
	})
}

// Handler: Reset analytics by clearing conversations (permanent delete)
func (s *Server) handleResetClear(c *fiber.Ctx) error {
	err := s.conversationAnalyzer.ClearConversations()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   err.Error(),
			"status":  "failed",
		})
	}

	// Clear caches after reset
	s.stateCalculator.ClearCache()
	s.processDetector.ClearCache()
	s.shellDetector.ClearCache()

	// Clear any soft reset
	s.resetTracker.ClearResetPoint()

	// Broadcast update to WebSocket clients
	s.wsHub.Broadcast([]byte(`{"event":"reset","action":"clear"}`))

	return c.JSON(fiber.Map{
		"status":  "cleared",
		"message": "All conversations have been permanently deleted",
		"time":    time.Now(),
	})
}

// Handler: Soft reset (delta-based, doesn't delete data)
func (s *Server) handleResetSoft(c *fiber.Ctx) error {
	// Get current totals
	conversations, err := s.conversationAnalyzer.LoadConversations(s.stateCalculator)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	totalTokens := 0
	for _, conv := range conversations {
		totalTokens += conv.Tokens
	}

	// Set reset point with current totals
	reason := "Manual soft reset"
	if err := s.resetTracker.SetResetPoint(totalTokens, len(conversations), reason); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  err.Error(),
			"status": "failed",
		})
	}

	// Broadcast update to WebSocket clients
	s.wsHub.Broadcast([]byte(`{"event":"reset","action":"soft"}`))

	return c.JSON(fiber.Map{
		"status":         "reset",
		"message":        "Soft reset applied - counts will now start from zero",
		"previousTokens": totalTokens,
		"previousConversations": len(conversations),
		"time":           time.Now(),
	})
}

// Handler: Clear soft reset (restore original counts)
func (s *Server) handleClearReset(c *fiber.Ctx) error {
	if !s.resetTracker.HasResetPoint() {
		return c.JSON(fiber.Map{
			"status":  "no_reset",
			"message": "No active reset to clear",
		})
	}

	if err := s.resetTracker.ClearResetPoint(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Broadcast update to WebSocket clients
	s.wsHub.Broadcast([]byte(`{"event":"reset","action":"cleared"}`))

	return c.JSON(fiber.Map{
		"status":  "cleared",
		"message": "Reset point cleared - showing original counts",
		"time":    time.Now(),
	})
}

// Handler: Get reset status
func (s *Server) handleResetStatus(c *fiber.Ctx) error {
	resetPoint := s.resetTracker.GetResetPoint()

	if resetPoint == nil {
		return c.JSON(fiber.Map{
			"active": false,
		})
	}

	return c.JSON(fiber.Map{
		"active":              true,
		"timestamp":           resetPoint.Timestamp,
		"reason":              resetPoint.Reason,
		"tokenDelta":          resetPoint.TokenDelta,
		"conversationDelta":   resetPoint.ConversationDelta,
	})
}

// Start starts the server
func (s *Server) Start() error {
	if !s.quiet {
		fmt.Printf("🚀 Starting server on http://localhost:%d\n", s.port)
		fmt.Printf("📊 Analytics dashboard: http://localhost:%d/\n", s.port)
		fmt.Printf("🔗 API endpoint: http://localhost:%d/api/data\n", s.port)
	}

	return s.app.Listen(fmt.Sprintf(":%d", s.port))
}

// Shutdown gracefully shuts down the server and all its components.
// It stops the file watcher, WebSocket hub, and closes the database.
func (s *Server) Shutdown() error {
	if !s.quiet {
		fmt.Println("🛑 Shutting down server...")
	}

	// Stop file watcher
	if s.fileWatcher != nil {
		if err := s.fileWatcher.Stop(); err != nil && !s.quiet {
			fmt.Printf("⚠️  Error stopping file watcher: %v\n", err)
		}
	}

	// Shutdown WebSocket hub
	if s.wsHub != nil {
		if err := s.wsHub.Shutdown(); err != nil && !s.quiet {
			fmt.Printf("⚠️  Error shutting down WebSocket hub: %v\n", err)
		}
	}

	// Close database
	if s.db != nil {
		if err := s.db.Close(); err != nil && !s.quiet {
			fmt.Printf("⚠️  Error closing database: %v\n", err)
		}
	}

	// Shutdown Fiber app
	return s.app.Shutdown()
}

// Handler: Get shell command history
func (s *Server) handleGetShellHistory(c *fiber.Ctx) error {
	query := &database.CommandHistoryQuery{
		ConversationID: c.Query("conversation_id"),
		Limit:          c.QueryInt("limit", 100),
		Offset:         c.QueryInt("offset", 0),
	}

	commands, err := s.repo.GetShellCommands(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"commands": commands,
		"count":    len(commands),
		"query":    query,
	})
}

// Handler: Get Claude command history
func (s *Server) handleGetClaudeHistory(c *fiber.Ctx) error {
	query := &database.CommandHistoryQuery{
		ConversationID: c.Query("conversation_id"),
		ToolName:       c.Query("tool_name"),
		Limit:          c.QueryInt("limit", 100),
		Offset:         c.QueryInt("offset", 0),
	}

	commands, err := s.repo.GetClaudeCommands(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"commands": commands,
		"count":    len(commands),
		"query":    query,
	})
}

// Handler: Get command statistics
func (s *Server) handleGetCommandStats(c *fiber.Ctx) error {
	commandType := c.Query("type") // 'shell', 'claude', or empty for all
	limit := c.QueryInt("limit", 50)

	stats, err := s.repo.GetCommandStats(commandType, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"stats": stats,
		"count": len(stats),
	})
}

// Handler: Get database statistics
func (s *Server) handleGetDBStats(c *fiber.Ctx) error {
	stats, err := s.db.Stats()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"stats":     stats,
		"db_path":   s.db.Path(),
		"timestamp": time.Now(),
	})
}


// Handler: Get user prompts
func (s *Server) handleGetUserPrompts(c *fiber.Ctx) error {
	query := &database.CommandHistoryQuery{
		ConversationID: c.Query("conversation_id"),
		Limit:          c.QueryInt("limit", 100),
		Offset:         c.QueryInt("offset", 0),
	}

	messages, err := s.repo.GetUserMessages(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"prompts": messages,
		"count":   len(messages),
		"query":   query,
	})
}

// Handler: Get prompt statistics
func (s *Server) handleGetPromptStats(c *fiber.Ctx) error {
	// Get total count of prompts
	allPrompts, err := s.repo.GetUserMessages(&database.CommandHistoryQuery{
		Limit: 0, // No limit to get accurate count
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	totalPrompts := len(allPrompts)

	// Calculate average prompt length
	totalLength := 0
	if totalPrompts > 0 {
		for _, msg := range allPrompts {
			totalLength += msg.MessageLength
		}
	}

	avgLength := 0
	if totalPrompts > 0 {
		avgLength = totalLength / totalPrompts
	}

	// Get unique conversations
	conversationSet := make(map[string]bool)
	for _, msg := range allPrompts {
		if msg.ConversationID != "" {
			conversationSet[msg.ConversationID] = true
		}
	}

	// Get unique branches
	branchSet := make(map[string]bool)
	for _, msg := range allPrompts {
		if msg.GitBranch != "" {
			branchSet[msg.GitBranch] = true
		}
	}

	return c.JSON(fiber.Map{
		"total_prompts":      totalPrompts,
		"avg_prompt_length":  avgLength,
		"unique_conversations": len(conversationSet),
		"unique_branches":    len(branchSet),
		"timestamp":          time.Now(),
	})
}

// Handler: Get unique sessions from user prompts
func (s *Server) handleGetUniqueSessions(c *fiber.Ctx) error {
	sessions, err := s.repo.GetUniqueSessions()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// Handler: Record a new user prompt
func (s *Server) handleRecordUserPrompt(c *fiber.Ctx) error {
	// Parse request body
	type RecordPromptRequest struct{
		SessionID        string `json:"session_id"`
		SessionName      string `json:"session_name"`
		Prompt           string `json:"prompt"`
		WorkingDirectory string `json:"cwd"`
		GitBranch        string `json:"branch"`
	}

	var req RecordPromptRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate required fields
	if req.SessionID == "" || req.Prompt == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "session_id and prompt are required",
		})
	}

	// Create user message record
	msg := &database.UserMessage{
		ConversationID:   req.SessionID,
		SessionName:      req.SessionName,
		Message:          req.Prompt,
		WorkingDirectory: req.WorkingDirectory,
		GitBranch:        req.GitBranch,
		MessageLength:    len(req.Prompt),
		SubmittedAt:      time.Now(),
	}

	// Record the message
	if err := s.repo.RecordUserMessage(msg); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to record prompt: %v", err),
		})
	}

	// Broadcast update to WebSocket clients
	s.wsHub.Broadcast([]byte(`{"event":"prompt_recorded"}`))

	return c.JSON(fiber.Map{
		"status":  "recorded",
		"id":      msg.ID,
		"length":  msg.MessageLength,
		"time":    msg.SubmittedAt,
	})
}

// Handler: Clear all user prompts
func (s *Server) handleClearAllPrompts(c *fiber.Ctx) error {
	err := s.repo.DeleteAllUserMessages()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  err.Error(),
			"status": "failed",
		})
	}

	// Broadcast update to WebSocket clients
	s.wsHub.Broadcast([]byte(`{"event":"prompts_cleared"}`))

	return c.JSON(fiber.Map{
		"status":  "cleared",
		"message": "All user prompts have been deleted",
		"time":    time.Now(),
	})
}
