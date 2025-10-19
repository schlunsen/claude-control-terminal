// Package server provides the Fiber-based HTTP server and REST API for CCT analytics.
// It serves the analytics dashboard, WebSocket connections, and API endpoints
// for conversation data, process monitoring, and command history.
package server

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/schlunsen/claude-control-terminal/internal/analytics"
	"github.com/schlunsen/claude-control-terminal/internal/database"
	"github.com/schlunsen/claude-control-terminal/internal/logging"
	"github.com/schlunsen/claude-control-terminal/internal/server/agents"
	"github.com/schlunsen/claude-control-terminal/internal/version"
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
	modelProviderLookup  *analytics.ModelProviderLookup
	db                   *database.Database
	repo                 *database.Repository
	config               *Config
	tlsConfig            *TLSConfig
	authMiddleware       *AuthMiddleware
	agentHandler         *agents.AgentHandler
	agentConfig          *agents.Config
	claudeDir            string
	port                 int
	quiet                bool // Suppress output when running in TUI
	verbose              bool // Enable verbose/debug logging
}

// NewServer creates a new Fiber server instance
func NewServer(claudeDir string, port int) *Server {
	return NewServerWithOptions(claudeDir, port, false, false)
}

// NewServerWithOptions creates a new Fiber server instance with options
func NewServerWithOptions(claudeDir string, port int, quiet bool, verbose bool) *Server {
	app := fiber.New(fiber.Config{
		AppName: "Claude Code Analytics",
		ServerHeader: "go-claude-templates",
		DisableStartupMessage: quiet, // Suppress Fiber startup banner in quiet mode
	})

	return &Server{
		app:       app,
		claudeDir: claudeDir,
		port:      port,
		quiet:     quiet,
		verbose:   verbose,
	}
}

// Setup initializes analytics components and routes
func (s *Server) Setup() error {
	// Disable standard log output when in quiet mode (TUI)
	// This prevents log.Printf calls from writing to stdout
	if s.quiet {
		log.SetOutput(io.Discard)
	}

	// Initialize configuration
	configManager := NewConfigManager(s.claudeDir)
	config, err := configManager.LoadOrCreateConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	s.config = config

	// Override port from config if not set
	if s.port == 0 {
		s.port = config.Server.Port
	}

	// Use verbose from config if not set explicitly
	if !s.verbose && config.Server.Verbose {
		s.verbose = config.Server.Verbose
	}

	// Initialize logging if verbose is enabled
	if s.verbose {
		logDir := filepath.Join(s.claudeDir, "analytics", "logs")
		logger, err := logging.Initialize(logDir, s.verbose)
		if err != nil {
			return fmt.Errorf("failed to initialize logging: %w", err)
		}

		// Redirect stderr to capture SDK debug logs
		if err := logger.RedirectStderr(); err != nil {
			return fmt.Errorf("failed to redirect stderr: %w", err)
		}

		if !s.quiet {
			fmt.Printf("📝 Verbose logging enabled\n")
			fmt.Printf("   Main log: %s\n", logger.GetLogFilePath())
			fmt.Printf("   SDK log:  %s\n", logger.GetStderrFilePath())
		}

		logging.Info("Server setup starting (verbose mode enabled)")
	}

	// Initialize TLS certificates if enabled
	if config.TLS.Enabled {
		certManager := NewCertificateManager(s.claudeDir)
		tlsConfig, err := certManager.EnsureCertificates()
		if err != nil {
			return fmt.Errorf("failed to initialize TLS: %w", err)
		}
		s.tlsConfig = tlsConfig
	}

	// Initialize authentication
	if config.Auth.Enabled {
		apiKey, err := configManager.EnsureAPIKey()
		if err != nil {
			return fmt.Errorf("failed to initialize API key: %w", err)
		}
		s.authMiddleware = NewAuthMiddleware(apiKey, true)
	}

	// Initialize agent configuration
	agentAPIKey := os.Getenv("ANTHROPIC_API_KEY")
	if agentAPIKey == "" {
		agentAPIKey = os.Getenv("CLAUDE_API_KEY")
	}

	// Log API key status
	if agentAPIKey == "" {
		if !s.quiet {
			fmt.Printf("⚠️  WARNING: No API key found in environment variables (ANTHROPIC_API_KEY or CLAUDE_API_KEY)\n")
		}
		if s.verbose {
			logging.Warning("No API key found in environment variables")
		}
	} else {
		if s.verbose {
			logging.Info("API key loaded from environment (length: %d characters)", len(agentAPIKey))
		}
	}

	// Set retention defaults if not specified
	retentionDays := config.Agent.SessionRetentionDays
	if retentionDays == 0 {
		retentionDays = 30 // Default: 30 days
	}

	cleanupEnabled := config.Agent.CleanupEnabled
	// If not explicitly set in config, default to true

	cleanupInterval := config.Agent.CleanupIntervalHours
	if cleanupInterval == 0 {
		cleanupInterval = 24 // Default: 24 hours
	}

	agentConfig := &agents.Config{
		Model:                 config.Agent.Model,
		APIKey:                agentAPIKey,
		MaxConcurrentSessions: config.Agent.MaxConcurrentSessions,
		Verbose:               s.verbose,
		SessionRetentionDays:  retentionDays,
		CleanupEnabled:        cleanupEnabled,
		CleanupIntervalHours:  cleanupInterval,
	}
	s.agentConfig = agentConfig

	// Note: Agent handler will be initialized after database is ready

	// Configure CORS middleware
	corsConfig := cors.Config{
		AllowOrigins: strings.Join(config.CORS.AllowedOrigins, ","),
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}
	s.app.Use(cors.New(corsConfig))

	// Only add logger middleware if not in quiet mode
	if !s.quiet {
		s.app.Use(logger.New())
	}

	// Apply authentication middleware globally if enabled
	if s.authMiddleware != nil {
		s.app.Use(s.authMiddleware.Handler())
	}

	// Initialize database
	dataDir := filepath.Join(s.claudeDir, "analytics_data")
	db, err := database.Initialize(dataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	s.db = db
	s.repo = database.NewRepository(db)

	// Initialize agent handler (requires database)
	agentHandler, err := agents.NewAgentHandler(agentConfig, db.GetDB())
	if err != nil {
		return fmt.Errorf("failed to initialize agent handler: %w", err)
	}
	s.agentHandler = agentHandler

	if !s.quiet {
		fmt.Printf("🤖 Agent handler initialized (model: %s, max sessions: %d, verbose: %v)\n",
			agentConfig.Model, agentConfig.MaxConcurrentSessions, agentConfig.Verbose)
	}

	if s.verbose {
		logging.Info("Agent handler initialized: model=%s, maxSessions=%d, verbose=%v, apiKeySet=%v",
			agentConfig.Model, agentConfig.MaxConcurrentSessions, agentConfig.Verbose, agentAPIKey != "")
	}

	// Start session cleanup job
	s.agentHandler.SessionManager.StartCleanupJob()

	// Initialize analytics components
	s.conversationAnalyzer = analytics.NewConversationAnalyzer(s.claudeDir)
	s.conversationParser = analytics.NewConversationParser(s.repo)
	s.stateCalculator = analytics.NewStateCalculator()
	s.processDetector = analytics.NewProcessDetector()
	s.shellDetector = analytics.NewShellDetector()
	s.resetTracker = analytics.NewResetTracker(s.claudeDir)
	s.modelProviderLookup = analytics.NewModelProviderLookup()

	// Initialize WebSocket hub
	s.wsHub = ws.NewHub()
	go s.wsHub.Run()

	// File watcher removed - WebSocket updates triggered by database operations only

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

	// Version info
	api.Get("/version", s.handleGetVersion)

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
	api.Get("/history/all", s.handleGetAllHistory)
	api.Get("/history/shell", s.handleGetShellHistory)
	api.Get("/history/claude", s.handleGetClaudeHistory)
	api.Get("/history/stats", s.handleGetCommandStats)
	api.Post("/commands/shell", s.handleRecordShellCommand)
	api.Post("/commands/claude", s.handleRecordClaudeCommand)
	api.Delete("/history", s.handleClearAllHistory)
	api.Get("/db/stats", s.handleGetDBStats)

	// User prompts endpoints
	api.Get("/prompts", s.handleGetUserPrompts)
	api.Get("/prompts/stats", s.handleGetPromptStats)
	api.Get("/prompts/sessions", s.handleGetUniqueSessions)
	api.Post("/prompts", s.handleRecordUserPrompt)
	api.Delete("/prompts", s.handleClearAllHistory) // Alias for backward compatibility

	// Notification endpoints
	api.Post("/notifications", s.handleRecordNotification)
	api.Get("/notifications", s.handleGetNotifications)
	api.Get("/notifications/stats", s.handleGetNotificationStats)
	api.Delete("/notifications", s.handleClearNotifications)

	// Session resume endpoint
	api.Get("/sessions/:conversation_id/resume-data", s.handleGetSessionResumeData)

	// WebSocket endpoint
	s.app.Get("/ws", websocket.New(s.wsHub.HandleWebSocket()))

	// Config endpoints (for frontend to get API key securely)
	api.Get("/config/api-key", s.handleGetAPIKey)
	api.Get("/config/cwd", s.handleGetCWD)

	// Agent endpoints (serve agents from project directory)
	api.Get("/agents", s.handleListAgents)
	api.Get("/agents/:name", s.handleGetAgentDetail)

	// Agent session endpoints (for persistence)
	api.Get("/agent/sessions", s.handleGetAgentSessions)
	api.Get("/agent/sessions/:id/messages", s.handleGetAgentMessages)

	// Agent WebSocket endpoint (direct, not proxied)
	// Use Fiber's WebSocket middleware with our Fiber-compatible handler
	s.app.Get("/agent/ws", websocket.New(s.agentHandler.HandleFiberWebSocket))
}

// Handler: Health check
func (s *Server) handleHealth(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
		"time":   time.Now(),
	})
}

// Handler: Get version info
func (s *Server) handleGetVersion(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"version": version.Version,
		"name":    version.Name,
		"time":    time.Now(),
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
	// Get CLI conversation stats
	conversations, err := s.conversationAnalyzer.LoadConversations(s.stateCalculator)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	cliTotalTokens := 0
	cliActiveCount := 0

	for _, conv := range conversations {
		cliTotalTokens += conv.Tokens
		if conv.Status == "active" {
			cliActiveCount++
		}
	}

	// Get agent session stats
	var agentSessions []agents.Session
	var agentTotalTokens int64
	var agentActiveCount int
	var agentTotalCost float64

	if s.agentHandler != nil {
		allSessions, err := s.agentHandler.SessionManager.ListAllSessions("all")
		if err == nil {
			agentSessions = allSessions
			for _, session := range allSessions {
				// Estimate tokens from message count (rough approximation)
				// TODO: Track actual tokens in messages
				agentTotalTokens += int64(session.MessageCount * 100)
				agentTotalCost += session.CostUSD

				// Count active sessions (not ended)
				if session.Status != agents.SessionStatusEnded {
					agentActiveCount++
				}
			}
		}
	}

	// Combine stats
	totalTokens := cliTotalTokens + int(agentTotalTokens)
	totalConversations := len(conversations) + len(agentSessions)
	activeCount := cliActiveCount + agentActiveCount

	// Apply soft reset delta if present
	adjustedTokens, adjustedConversations := s.resetTracker.ApplyDelta(totalTokens, totalConversations)

	avgTokens := 0
	if adjustedConversations > 0 {
		avgTokens = adjustedTokens / adjustedConversations
	}

	response := fiber.Map{
		// Combined stats
		"totalConversations":  adjustedConversations,
		"activeConversations": activeCount,
		"totalTokens":         adjustedTokens,
		"avgTokens":           avgTokens,
		"timestamp":           time.Now(),

		// Breakdown by type
		"cliConversations":   len(conversations),
		"cliActive":          cliActiveCount,
		"cliTokens":          cliTotalTokens,
		"agentSessions":      len(agentSessions),
		"agentActive":        agentActiveCount,
		"agentTokens":        agentTotalTokens,
		"agentTotalCost":     agentTotalCost,
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

	// Broadcast update to WebSocket clients with data
	s.wsHub.BroadcastData("reset_archive", fiber.Map{
		"action": "archive",
		"message": "All conversations have been archived",
	})

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

	// Broadcast update to WebSocket clients with data
	s.wsHub.BroadcastData("reset_clear", fiber.Map{
		"action": "clear",
		"message": "All conversations have been permanently deleted",
	})

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

	// Broadcast update to WebSocket clients with data
	s.wsHub.BroadcastData("reset_soft", fiber.Map{
		"action": "soft",
		"message": "Soft reset applied",
		"previousTokens": totalTokens,
		"previousConversations": len(conversations),
	})

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

	// Broadcast update to WebSocket clients with data
	s.wsHub.BroadcastData("reset_cleared", fiber.Map{
		"action": "cleared",
		"message": "Reset point cleared - showing original counts",
	})

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
	// Determine protocol and address
	protocol := "http"
	if s.tlsConfig != nil && s.tlsConfig.Enabled {
		protocol = "https"
	}

	// Bind address - use from config or default to 127.0.0.1
	bindHost := "127.0.0.1"
	if s.config != nil && s.config.Server.Host != "" {
		bindHost = s.config.Server.Host
	}
	addr := fmt.Sprintf("%s:%d", bindHost, s.port)

	if !s.quiet {
		fmt.Printf("🚀 Starting server on %s://%s\n", protocol, addr)
		fmt.Printf("📊 Analytics dashboard: %s://localhost:%d/\n", protocol, s.port)
		fmt.Printf("🔗 API endpoint: %s://localhost:%d/api/data\n", protocol, s.port)

		if s.tlsConfig != nil && s.tlsConfig.Enabled {
			fmt.Printf("🔒 TLS enabled (self-signed certificate)\n")
		}

		if s.authMiddleware != nil {
			configManager := NewConfigManager(s.claudeDir)
			fmt.Printf("🔑 Authentication enabled (API key in %s)\n", configManager.GetSecretPath())
		}
	}

	// Start server with TLS if enabled
	if s.tlsConfig != nil && s.tlsConfig.Enabled {
		return s.app.ListenTLS(addr, s.tlsConfig.CertPath, s.tlsConfig.KeyPath)
	}

	// Start without TLS
	return s.app.Listen(addr)
}

// Shutdown gracefully shuts down the server and all its components.
// It stops the file watcher, WebSocket hub, and closes the database.
func (s *Server) Shutdown() error {
	if !s.quiet {
		fmt.Println("🛑 Shutting down server...")
	}

	// Cleanup agent sessions
	if s.agentHandler != nil {
		if err := s.agentHandler.Cleanup(); err != nil && !s.quiet {
			fmt.Printf("⚠️  Error cleaning up agent sessions: %v\n", err)
		}
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

// formatBytes converts bytes to human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Handler: Get database statistics
func (s *Server) handleGetDBStats(c *fiber.Ctx) error {
	stats, err := s.db.Stats()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Add human-readable size if db_size_bytes exists
	if sizeBytes, ok := stats["db_size_bytes"].(int64); ok {
		stats["db_size_human"] = formatBytes(sizeBytes)
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
		ModelProvider    string `json:"model_provider"`
		ModelName        string `json:"model_name"`
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

	// Use model info from request, fallback to Unknown if not provided
	modelProvider := req.ModelProvider
	modelName := req.ModelName
	if modelProvider == "" {
		modelProvider = "Unknown"
	}
	if modelName == "" {
		modelName = "Unknown"
	}

	// Translate URL-based provider to human-readable name
	if s.modelProviderLookup != nil {
		modelProvider = s.modelProviderLookup.GetProviderNameFromModelInfo(modelProvider, modelName)
	}

	// Create user message record
	msg := &database.UserMessage{
		ConversationID:   req.SessionID,
		SessionName:      req.SessionName,
		Message:          req.Prompt,
		WorkingDirectory: req.WorkingDirectory,
		GitBranch:        req.GitBranch,
		ModelProvider:    modelProvider,
		ModelName:        modelName,
		MessageLength:    len(req.Prompt),
		SubmittedAt:      time.Now(),
	}

	// Record the message
	if err := s.repo.RecordUserMessage(msg); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to record prompt: %v", err),
		})
	}

	// Broadcast update to WebSocket clients with data
	s.wsHub.BroadcastData("prompt_recorded", msg)

	return c.JSON(fiber.Map{
		"status":  "recorded",
		"id":      msg.ID,
		"length":  msg.MessageLength,
		"time":    msg.SubmittedAt,
	})
}

// Handler: Clear all history (user prompts, shell commands, and claude commands)
func (s *Server) handleClearAllHistory(c *fiber.Ctx) error {
	// Get database size before clearing
	var sizeBefore int64
	if stats, err := s.db.Stats(); err == nil {
		if size, ok := stats["db_size_bytes"].(int64); ok {
			sizeBefore = size
		}
	}

	err := s.repo.DeleteAllHistory()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  err.Error(),
			"status": "failed",
		})
	}

	// Vacuum database to reclaim disk space
	if !s.quiet {
		fmt.Printf("🗑️  Vacuuming database to reclaim disk space (size before: %s)...\n", formatBytes(sizeBefore))
	}
	if err := s.db.Vacuum(); err != nil {
		// Log the error but don't fail the request since data was deleted successfully
		if !s.quiet {
			fmt.Printf("⚠️  Warning: Failed to vacuum database after clearing history: %v\n", err)
		}
	} else {
		// Get database size after vacuum
		var sizeAfter int64
		if stats, err := s.db.Stats(); err == nil {
			if size, ok := stats["db_size_bytes"].(int64); ok {
				sizeAfter = size
			}
		}
		if !s.quiet {
			fmt.Printf("✅ Database vacuum completed successfully (size after: %s, reduced by: %s)\n", 
				formatBytes(sizeAfter), formatBytes(sizeBefore-sizeAfter))
		}
	}

	// Broadcast update to WebSocket clients with data
	s.wsHub.BroadcastData("history_cleared", fiber.Map{
		"message": "All history deleted and database vacuumed",
	})

	return c.JSON(fiber.Map{
		"status":  "cleared",
		"message": "All history (prompts, shell commands, and Claude commands) have been deleted and database vacuumed",
		"time":    time.Now(),
	})
}

// Handler: Record a shell command
func (s *Server) handleRecordShellCommand(c *fiber.Ctx) error {
	type RecordShellCommandRequest struct {
		SessionID        string `json:"session_id"`
		SessionName      string `json:"session_name"`
		Command          string `json:"command"`
		Description      string `json:"description"`
		WorkingDirectory string `json:"cwd"`
		GitBranch        string `json:"branch"`
		ModelProvider    string `json:"model_provider"`
		ModelName        string `json:"model_name"`
		ExitCode         *int   `json:"exit_code"`
		Stdout           string `json:"stdout"`
		Stderr           string `json:"stderr"`
		DurationMs       *int   `json:"duration_ms"`
	}

	var req RecordShellCommandRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate required fields
	if req.SessionID == "" || req.Command == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "session_id and command are required",
		})
	}

	// Use model info from request, fallback to Unknown if not provided
	modelProvider := req.ModelProvider
	modelName := req.ModelName
	if modelProvider == "" {
		modelProvider = "Unknown"
	}
	if modelName == "" {
		modelName = "Unknown"
	}

	// Translate URL-based provider to human-readable name
	if s.modelProviderLookup != nil {
		modelProvider = s.modelProviderLookup.GetProviderNameFromModelInfo(modelProvider, modelName)
	}

	// Create shell command record
	cmd := &database.ShellCommand{
		ConversationID:   req.SessionID,
		SessionName:      req.SessionName,
		Command:          req.Command,
		Description:      req.Description,
		WorkingDirectory: req.WorkingDirectory,
		GitBranch:        req.GitBranch,
		ModelProvider:    modelProvider,
		ModelName:        modelName,
		ExitCode:         req.ExitCode,
		Stdout:           req.Stdout,
		Stderr:           req.Stderr,
		DurationMs:       req.DurationMs,
		ExecutedAt:       time.Now(),
	}

	// Record the command
	if err := s.repo.RecordShellCommand(cmd); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to record shell command: %v", err),
		})
	}

	// Broadcast update to WebSocket clients with data
	s.wsHub.BroadcastData("command_recorded", fiber.Map{
		"type": "shell",
		"data": cmd,
	})

	return c.JSON(fiber.Map{
		"status": "recorded",
		"id":     cmd.ID,
		"time":   cmd.ExecutedAt,
	})
}

// Handler: Record a Claude command
func (s *Server) handleRecordClaudeCommand(c *fiber.Ctx) error {
	type RecordClaudeCommandRequest struct {
		SessionID        string `json:"session_id"`
		SessionName      string `json:"session_name"`
		ToolName         string `json:"tool_name"`
		Parameters       string `json:"parameters"`
		Result           string `json:"result"`
		WorkingDirectory string `json:"cwd"`
		GitBranch        string `json:"branch"`
		ModelProvider    string `json:"model_provider"`
		ModelName        string `json:"model_name"`
		Success          bool   `json:"success"`
		ErrorMessage     string `json:"error_message"`
		DurationMs       *int   `json:"duration_ms"`
	}

	var req RecordClaudeCommandRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate required fields
	if req.SessionID == "" || req.ToolName == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "session_id and tool_name are required",
		})
	}

	// Use model info from request, fallback to Unknown if not provided
	modelProvider := req.ModelProvider
	modelName := req.ModelName
	if modelProvider == "" {
		modelProvider = "Unknown"
	}
	if modelName == "" {
		modelName = "Unknown"
	}

	// Translate URL-based provider to human-readable name
	if s.modelProviderLookup != nil {
		modelProvider = s.modelProviderLookup.GetProviderNameFromModelInfo(modelProvider, modelName)
	}

	// Create Claude command record
	cmd := &database.ClaudeCommand{
		ConversationID:   req.SessionID,
		SessionName:      req.SessionName,
		ToolName:         req.ToolName,
		Parameters:       req.Parameters,
		Result:           req.Result,
		WorkingDirectory: req.WorkingDirectory,
		GitBranch:        req.GitBranch,
		ModelProvider:    modelProvider,
		ModelName:        modelName,
		Success:          req.Success,
		ErrorMessage:     req.ErrorMessage,
		DurationMs:       req.DurationMs,
		ExecutedAt:       time.Now(),
	}

	// Record the command
	if err := s.repo.RecordClaudeCommand(cmd); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to record claude command: %v", err),
		})
	}

	// Broadcast update to WebSocket clients with data
	s.wsHub.BroadcastData("command_recorded", fiber.Map{
		"type": "claude",
		"data": cmd,
	})

	return c.JSON(fiber.Map{
		"status": "recorded",
		"id":     cmd.ID,
		"time":   cmd.ExecutedAt,
	})
}

// Handler: Get all history (unified endpoint for shell commands, claude commands, and user prompts)
func (s *Server) handleGetAllHistory(c *fiber.Ctx) error {
	conversationID := c.Query("conversation_id")
	limit := c.QueryInt("limit", 100)
	offset := c.QueryInt("offset", 0)

	query := &database.CommandHistoryQuery{
		ConversationID: conversationID,
		Limit:          limit,
		Offset:         offset,
	}

	// Fetch all four types
	shellCommands, err := s.repo.GetShellCommands(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get shell commands: %v", err),
		})
	}

	claudeCommands, err := s.repo.GetClaudeCommands(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get claude commands: %v", err),
		})
	}

	userMessages, err := s.repo.GetUserMessages(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get user messages: %v", err),
		})
	}

	notifications, err := s.repo.GetNotifications(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get notifications: %v", err),
		})
	}

	// Combine into unified response with type field
	type HistoryItem struct {
		Type             string      `json:"type"`
		ID               int64       `json:"id"`
		ConversationID   string      `json:"conversation_id"`
		SessionName      string      `json:"session_name,omitempty"`
		Timestamp        time.Time   `json:"timestamp"`
		WorkingDirectory string      `json:"working_directory,omitempty"`
		GitBranch        string      `json:"git_branch,omitempty"`
		Content          interface{} `json:"content"`
	}

	var allHistory []HistoryItem

	// Add shell commands
	for _, cmd := range shellCommands {
		allHistory = append(allHistory, HistoryItem{
			Type:             "shell",
			ID:               cmd.ID,
			ConversationID:   cmd.ConversationID,
			SessionName:      cmd.SessionName,
			Timestamp:        cmd.ExecutedAt,
			WorkingDirectory: cmd.WorkingDirectory,
			GitBranch:        cmd.GitBranch,
			Content:          cmd,
		})
	}

	// Add claude commands
	for _, cmd := range claudeCommands {
		allHistory = append(allHistory, HistoryItem{
			Type:             "claude",
			ID:               cmd.ID,
			ConversationID:   cmd.ConversationID,
			SessionName:      cmd.SessionName,
			Timestamp:        cmd.ExecutedAt,
			WorkingDirectory: cmd.WorkingDirectory,
			GitBranch:        cmd.GitBranch,
			Content:          cmd,
		})
	}

	// Add user messages
	for _, msg := range userMessages {
		allHistory = append(allHistory, HistoryItem{
			Type:             "prompt",
			ID:               msg.ID,
			ConversationID:   msg.ConversationID,
			SessionName:      msg.SessionName,
			Timestamp:        msg.SubmittedAt,
			WorkingDirectory: msg.WorkingDirectory,
			GitBranch:        msg.GitBranch,
			Content:          msg,
		})
	}

	// Add notifications
	for _, notif := range notifications {
		allHistory = append(allHistory, HistoryItem{
			Type:             "notification",
			ID:               notif.ID,
			ConversationID:   notif.ConversationID,
			SessionName:      notif.SessionName,
			Timestamp:        notif.NotifiedAt,
			WorkingDirectory: notif.WorkingDirectory,
			GitBranch:        notif.GitBranch,
			Content:          notif,
		})
	}

	// Sort by timestamp descending
	// Simple bubble sort since we're dealing with already sorted slices
	for i := 0; i < len(allHistory)-1; i++ {
		for j := i + 1; j < len(allHistory); j++ {
			if allHistory[i].Timestamp.Before(allHistory[j].Timestamp) {
				allHistory[i], allHistory[j] = allHistory[j], allHistory[i]
			}
		}
	}

	return c.JSON(fiber.Map{
		"history": allHistory,
		"count":   len(allHistory),
		"query":   query,
	})
}

// Handler: Record a notification
func (s *Server) handleRecordNotification(c *fiber.Ctx) error {
	type RecordNotificationRequest struct {
		SessionID        string `json:"session_id"`
		SessionName      string `json:"session_name"`
		NotificationType string `json:"notification_type"`
		Message          string `json:"message"`
		ToolName         string `json:"tool_name"`
		CommandDetails   string `json:"command_details"`
		WorkingDirectory string `json:"cwd"`
		GitBranch        string `json:"branch"`
		ModelProvider    string `json:"model_provider"`
		ModelName        string `json:"model_name"`
	}

	var req RecordNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate required fields
	if req.SessionID == "" || req.Message == "" || req.NotificationType == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "session_id, message, and notification_type are required",
		})
	}

	// Use model info from request, fallback to Unknown if not provided
	modelProvider := req.ModelProvider
	modelName := req.ModelName
	if modelProvider == "" {
		modelProvider = "Unknown"
	}
	if modelName == "" {
		modelName = "Unknown"
	}

	// Translate URL-based provider to human-readable name
	if s.modelProviderLookup != nil {
		modelProvider = s.modelProviderLookup.GetProviderNameFromModelInfo(modelProvider, modelName)
	}

	// Create notification record
	notif := &database.Notification{
		ConversationID:   req.SessionID,
		SessionName:      req.SessionName,
		NotificationType: req.NotificationType,
		Message:          req.Message,
		ToolName:         req.ToolName,
		CommandDetails:   req.CommandDetails,
		WorkingDirectory: req.WorkingDirectory,
		GitBranch:        req.GitBranch,
		ModelProvider:    modelProvider,
		ModelName:        modelName,
		NotifiedAt:       time.Now(),
	}

	// Record the notification
	if err := s.repo.RecordNotification(notif); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to record notification: %v", err),
		})
	}

	// Broadcast update to WebSocket clients with data
	s.wsHub.BroadcastData("notification_recorded", notif)

	return c.JSON(fiber.Map{
		"status": "recorded",
		"id":     notif.ID,
		"time":   notif.NotifiedAt,
	})
}

// Handler: Get notifications
func (s *Server) handleGetNotifications(c *fiber.Ctx) error {
	query := &database.CommandHistoryQuery{
		ConversationID: c.Query("conversation_id"),
		Limit:          c.QueryInt("limit", 100),
		Offset:         c.QueryInt("offset", 0),
	}

	notifications, err := s.repo.GetNotifications(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"notifications": notifications,
		"count":         len(notifications),
		"query":         query,
	})
}

// Handler: Get notification statistics
func (s *Server) handleGetNotificationStats(c *fiber.Ctx) error {
	stats, err := s.repo.GetNotificationStats()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(stats)
}

// Handler: Clear all notifications
func (s *Server) handleClearNotifications(c *fiber.Ctx) error {
	err := s.repo.DeleteAllNotifications()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  err.Error(),
			"status": "failed",
		})
	}

	// Broadcast update to WebSocket clients with data
	s.wsHub.BroadcastData("notifications_cleared", fiber.Map{
		"message": "All notifications deleted",
	})

	return c.JSON(fiber.Map{
		"status":  "cleared",
		"message": "All notifications have been deleted",
		"time":    time.Now(),
	})
}

// Handler: Get session resume data
func (s *Server) handleGetSessionResumeData(c *fiber.Ctx) error {
	conversationID := c.Params("conversation_id")

	if conversationID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "conversation_id is required",
		})
	}

	// Fetch recent history (last 20 messages for context)
	query := &database.CommandHistoryQuery{
		ConversationID: conversationID,
		Limit:          20,
		Offset:         0,
	}

	// Get user messages (prompts)
	userMessages, err := s.repo.GetUserMessages(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get user messages: %v", err),
		})
	}

	if len(userMessages) == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "no conversation found with this ID",
		})
	}

	// Extract working directory (use the most recent one available)
	workingDir := ""
	sessionName := ""
	for i := len(userMessages) - 1; i >= 0; i-- {
		if userMessages[i].WorkingDirectory != "" {
			workingDir = userMessages[i].WorkingDirectory
			break
		}
	}

	// Get session name from first message
	if len(userMessages) > 0 {
		sessionName = userMessages[0].SessionName
	}

	// Format context for agent (last 10 messages)
	contextLimit := 10
	if len(userMessages) < contextLimit {
		contextLimit = len(userMessages)
	}

	contextMessages := userMessages[len(userMessages)-contextLimit:]

	// Build formatted context string
	var contextBuilder strings.Builder
	contextBuilder.WriteString("Previous conversation history:\n\n")

	for _, msg := range contextMessages {
		contextBuilder.WriteString(fmt.Sprintf("User: %s\n", msg.Message))
		contextBuilder.WriteString(fmt.Sprintf("(at %s)\n\n", msg.SubmittedAt.Format("3:04 PM")))
	}

	return c.JSON(fiber.Map{
		"conversation_id":    conversationID,
		"session_name":       sessionName,
		"working_directory":  workingDir,
		"context":            contextBuilder.String(),
		"total_messages":     len(userMessages),
		"last_activity":      userMessages[len(userMessages)-1].SubmittedAt,
		"messages":           contextMessages,
	})
}

// Handler: Get API key for frontend (secured endpoint)
func (s *Server) handleGetAPIKey(c *fiber.Ctx) error {
	// Only allow GET requests from same origin (browser)
	// This prevents external services from stealing the key
	origin := c.Get("Origin")
	if origin == "" {
		// Allow requests without Origin header (same-origin requests)
		configManager := NewConfigManager(s.claudeDir)
		apiKey, err := configManager.EnsureAPIKey()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to get API key",
			})
		}

		return c.JSON(fiber.Map{
			"apiKey": apiKey,
		})
	}

	// If Origin is present, it must match allowed origins
	allowed := false
	for _, allowedOrigin := range s.config.CORS.AllowedOrigins {
		if origin == allowedOrigin {
			allowed = true
			break
		}
	}

	if !allowed {
		return c.Status(403).JSON(fiber.Map{
			"error": "Forbidden",
		})
	}

	configManager := NewConfigManager(s.claudeDir)
	apiKey, err := configManager.EnsureAPIKey()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get API key",
		})
	}

	return c.JSON(fiber.Map{
		"apiKey": apiKey,
	})
}

// Handler: Get current working directory where cct was launched
func (s *Server) handleGetCWD(c *fiber.Ctx) error {
	cwd, err := os.Getwd()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get current working directory",
		})
	}

	return c.JSON(fiber.Map{
		"cwd": cwd,
	})
}

// Handler: List all available agents from .claude/agents/ directory
func (s *Server) handleListAgents(c *fiber.Ctx) error {
	cwd, err := os.Getwd()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get current working directory",
		})
	}

	agentsDir := filepath.Join(cwd, ".claude", "agents")

	// Check if agents directory exists
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		return c.JSON(fiber.Map{
			"agents": make(map[string]interface{}),
			"count":  0,
			"dir":    agentsDir,
		})
	}

	// Read all markdown files in agents directory
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to read agents directory: %v", err),
		})
	}

	agents := make(map[string]interface{})

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process markdown files
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		agentName := strings.TrimSuffix(entry.Name(), ".md")
		filePath := filepath.Join(agentsDir, entry.Name())

		// Read file
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		// Parse frontmatter
		agentData := parseFrontmatter(string(content))
		if agentData != nil && agentData["name"] != nil {
			agents[agentName] = agentData
		}
	}

	return c.JSON(fiber.Map{
		"agents": agents,
		"count":  len(agents),
		"dir":    agentsDir,
	})
}

// Handler: Get specific agent details with full system prompt
func (s *Server) handleGetAgentDetail(c *fiber.Ctx) error {
	agentName := c.Params("name")
	if agentName == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Agent name is required",
		})
	}

	cwd, err := os.Getwd()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get current working directory",
		})
	}

	agentFile := filepath.Join(cwd, ".claude", "agents", agentName+".md")

	// Check if agent file exists
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("Agent '%s' not found", agentName),
		})
	}

	// Read file
	content, err := os.ReadFile(agentFile)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to read agent file: %v", err),
		})
	}

	// Parse frontmatter and system prompt
	agentData := parseFrontmatterWithPrompt(string(content))
	if agentData == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to parse agent file",
		})
	}

	return c.JSON(fiber.Map{
		"agent": agentData,
	})
}

// parseFrontmatter extracts YAML frontmatter from markdown content
func parseFrontmatter(content string) map[string]interface{} {
	// Look for frontmatter between --- markers
	lines := strings.Split(content, "\n")
	if len(lines) < 3 || lines[0] != "---" {
		return nil
	}

	// Find closing ---
	var endLine int
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			endLine = i
			break
		}
	}

	if endLine == 0 {
		return nil
	}

	// Extract YAML lines
	yamlLines := lines[1:endLine]
	data := make(map[string]interface{})

	for _, line := range yamlLines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Simple YAML parsing (key: value format)
		if idx := strings.Index(line, ":"); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])

			// Remove quotes if present
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = value[1 : len(value)-1]
			}

			data[key] = value
		}
	}

	return data
}

// parseFrontmatterWithPrompt extracts frontmatter AND system prompt from markdown
func parseFrontmatterWithPrompt(content string) map[string]interface{} {
	// Look for frontmatter between --- markers
	lines := strings.Split(content, "\n")
	if len(lines) < 3 || lines[0] != "---" {
		return nil
	}

	// Find closing ---
	var endLine int
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			endLine = i
			break
		}
	}

	if endLine == 0 {
		return nil
	}

	// Extract YAML lines
	yamlLines := lines[1:endLine]
	data := make(map[string]interface{})

	for _, line := range yamlLines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Simple YAML parsing (key: value format)
		if idx := strings.Index(line, ":"); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])

			// Remove quotes if present
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = value[1 : len(value)-1]
			}

			data[key] = value
		}
	}

	// Extract system prompt (everything after the closing ---)
	if endLine+1 < len(lines) {
		promptLines := lines[endLine+1:]
		// Skip empty lines at start
		for i := 0; i < len(promptLines); i++ {
			if strings.TrimSpace(promptLines[i]) != "" {
				promptLines = promptLines[i:]
				break
			}
		}
		systemPrompt := strings.Join(promptLines, "\n")
		systemPrompt = strings.TrimSpace(systemPrompt)
		data["system_prompt"] = systemPrompt
		data["system_prompt_length"] = len(systemPrompt)
	}

	return data
}

// Handler: Get agent sessions (with optional status filter)
func (s *Server) handleGetAgentSessions(c *fiber.Ctx) error {
	if s.agentHandler == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "agent handler not initialized",
		})
	}

	// Get status filter from query params (default: "all")
	statusFilter := c.Query("status", "all")

	// Get sessions from storage
	sessions, err := s.agentHandler.SessionManager.ListAllSessions(statusFilter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to list sessions: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"sessions": sessions,
		"count":    len(sessions),
		"filter":   statusFilter,
	})
}

// Handler: Get messages for an agent session (with pagination)
func (s *Server) handleGetAgentMessages(c *fiber.Ctx) error {
	if s.agentHandler == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "agent handler not initialized",
		})
	}

	// Parse session ID from URL params
	sessionIDStr := c.Params("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid session ID",
		})
	}

	// Parse pagination params
	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	// Validate pagination params
	if limit < 1 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Get messages from storage
	messages, hasMore, err := s.agentHandler.SessionManager.GetMessages(sessionID, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get messages: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"session_id": sessionID,
		"messages":   messages,
		"count":      len(messages),
		"limit":      limit,
		"offset":     offset,
		"has_more":   hasMore,
	})
}
