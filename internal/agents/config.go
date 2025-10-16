package agents

import (
	"os"
	"path/filepath"
	"strconv"
)

// Config holds configuration for the agent server
type Config struct {
	Host             string
	Port             int
	LogLevel         string
	Reload           bool
	AuthEnabled      bool
	MaxConcurrentSessions int
	ServerDir        string
	PIDFile          string
	LogFile          string
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	serverDir := filepath.Join(homeDir, ".claude", "agents_server")

	return &Config{
		Host:                  getEnvOrDefault("AGENT_SERVER_HOST", "127.0.0.1"),
		Port:                  getEnvIntOrDefault("AGENT_SERVER_PORT", 8001),
		LogLevel:              getEnvOrDefault("AGENT_SERVER_LOG_LEVEL", "INFO"),
		Reload:                getEnvBoolOrDefault("AGENT_SERVER_RELOAD", false),
		AuthEnabled:           getEnvBoolOrDefault("AGENT_SERVER_AUTH_ENABLED", true),
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

// ToEnvVars converts config to environment variables for the Python process
func (c *Config) ToEnvVars() []string {
	env := os.Environ()
	env = append(env,
		"AGENT_SERVER_HOST="+c.Host,
		"AGENT_SERVER_PORT="+strconv.Itoa(c.Port),
		"AGENT_SERVER_LOG_LEVEL="+c.LogLevel,
		"AGENT_SERVER_RELOAD="+strconv.FormatBool(c.Reload),
		"AGENT_SERVER_AUTH_ENABLED="+strconv.FormatBool(c.AuthEnabled),
		"AGENT_SERVER_MAX_CONCURRENT_SESSIONS="+strconv.Itoa(c.MaxConcurrentSessions),
	)
	return env
}
