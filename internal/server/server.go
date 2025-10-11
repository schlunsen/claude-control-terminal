package server

import (
	"fmt"
	"time"

	"github.com/davila7/go-claude-templates/internal/analytics"
	ws "github.com/davila7/go-claude-templates/internal/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
)

// Server wraps the Fiber app and analytics components
type Server struct {
	app                  *fiber.App
	conversationAnalyzer *analytics.ConversationAnalyzer
	stateCalculator      *analytics.StateCalculator
	processDetector      *analytics.ProcessDetector
	fileWatcher          *analytics.FileWatcher
	wsHub                *ws.Hub
	claudeDir            string
	port                 int
}

// NewServer creates a new Fiber server instance
func NewServer(claudeDir string, port int) *Server {
	app := fiber.New(fiber.Config{
		AppName: "Claude Code Analytics",
		ServerHeader: "go-claude-templates",
	})

	// Middleware
	app.Use(cors.New())
	app.Use(logger.New())

	return &Server{
		app:       app,
		claudeDir: claudeDir,
		port:      port,
	}
}

// Setup initializes analytics components and routes
func (s *Server) Setup() error {
	// Initialize analytics components
	s.conversationAnalyzer = analytics.NewConversationAnalyzer(s.claudeDir)
	s.stateCalculator = analytics.NewStateCalculator()
	s.processDetector = analytics.NewProcessDetector()

	// Initialize WebSocket hub
	s.wsHub = ws.NewHub()
	go s.wsHub.Run()

	// Setup API routes
	s.setupRoutes()

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
	api.Get("/stats", s.handleGetStats)

	// Refresh endpoint
	api.Post("/refresh", s.handleRefresh)

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

	avgTokens := 0
	if len(conversations) > 0 {
		avgTokens = totalTokens / len(conversations)
	}

	return c.JSON(fiber.Map{
		"totalConversations": len(conversations),
		"activeConversations": activeCount,
		"totalTokens":        totalTokens,
		"avgTokens":          avgTokens,
		"timestamp":          time.Now(),
	})
}

// Handler: Refresh data
func (s *Server) handleRefresh(c *fiber.Ctx) error {
	// Clear caches
	s.stateCalculator.ClearCache()
	s.processDetector.ClearCache()

	return c.JSON(fiber.Map{
		"status": "refreshed",
		"time":   time.Now(),
	})
}

// Start starts the server
func (s *Server) Start() error {
	fmt.Printf("🚀 Starting server on http://localhost:%d\n", s.port)
	fmt.Printf("📊 Analytics dashboard: http://localhost:%d/\n", s.port)
	fmt.Printf("🔗 API endpoint: http://localhost:%d/api/data\n", s.port)

	return s.app.Listen(fmt.Sprintf(":%d", s.port))
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	fmt.Println("🛑 Shutting down server...")

	if s.fileWatcher != nil {
		s.fileWatcher.Stop()
	}

	return s.app.Shutdown()
}
