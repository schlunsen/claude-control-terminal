package agents

import (
	"os"
	"path/filepath"
	"strconv"
)

// Config holds configuration for the agent server
type Config struct {
	Host                  string
	Port                  int
	LogLevel              string
	Model                 string
	APIKey                string
	MaxConcurrentSessions int
	ServerDir             string
	PIDFile               string
	LogFile               string
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	serverDir := filepath.Join(homeDir, ".claude", "agents_server")

	// Try ANTHROPIC_API_KEY first (standard), then fall back to CLAUDE_API_KEY
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("CLAUDE_API_KEY")
	}

	// Get model from ANTHROPIC_MODEL (set by provider config), then AGENT_SERVER_MODEL, then default
	model := os.Getenv("ANTHROPIC_MODEL")
	if model == "" {
		model = getEnvOrDefault("AGENT_SERVER_MODEL", "claude-sonnet-4.5-20250514")
	}

	return &Config{
		Host:                  getEnvOrDefault("AGENT_SERVER_HOST", "127.0.0.1"),
		Port:                  getEnvIntOrDefault("AGENT_SERVER_PORT", 8001),
		LogLevel:              getEnvOrDefault("AGENT_SERVER_LOG_LEVEL", "INFO"),
		Model:                 model,
		APIKey:                apiKey,
		MaxConcurrentSessions: getEnvIntOrDefault("AGENT_SERVER_MAX_CONCURRENT_SESSIONS", 10),
		ServerDir:             serverDir,
		PIDFile:               filepath.Join(serverDir, ".pid"),
		LogFile:               filepath.Join(serverDir, "server.log"),
	}
}

// GetServerDir returns the agent server installation directory
func GetServerDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".claude", "agents_server")
	}
	return filepath.Join(homeDir, ".claude", "agents_server")
}

// getEnvOrDefault returns the environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault returns the environment variable as int or default
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvBoolOrDefault returns the environment variable as bool or default
func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

