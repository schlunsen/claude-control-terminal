package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
)

// Config holds the analytics server configuration
type Config struct {
	TLS     TLSSettings     `json:"tls"`
	Auth    AuthSettings    `json:"auth"`
	Server  ServerSettings  `json:"server"`
	CORS    CORSSettings    `json:"cors"`
	Agent   AgentSettings   `json:"agent"`
}

// TLSSettings holds TLS configuration
type TLSSettings struct {
	Enabled  bool   `json:"enabled"`
	CertPath string `json:"cert_path,omitempty"`
	KeyPath  string `json:"key_path,omitempty"`
}

// AuthSettings holds authentication configuration
type AuthSettings struct {
	Enabled    bool   `json:"enabled"`
	APIKeyPath string `json:"api_key_path,omitempty"`
}

// ServerSettings holds server configuration
type ServerSettings struct {
	Port      int    `json:"port"`
	Host      string `json:"host"`
	Quiet     bool   `json:"quiet"`
	Verbose   bool   `json:"verbose"`
}

// CORSSettings holds CORS configuration
type CORSSettings struct {
	AllowedOrigins []string `json:"allowed_origins"`
}

// AgentSettings holds agent configuration
type AgentSettings struct {
	Model                 string `json:"model"`
	MaxConcurrentSessions int    `json:"max_concurrent_sessions"`
	SessionRetentionDays  int    `json:"session_retention_days"`
	CleanupEnabled        bool   `json:"cleanup_enabled"`
	CleanupIntervalHours  int    `json:"cleanup_interval_hours"`
}

// ConfigManager handles configuration loading and saving
type ConfigManager struct {
	configDir  string
	configFile string
	secretFile string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(claudeDir string) *ConfigManager {
	configDir := filepath.Join(claudeDir, "analytics")
	return &ConfigManager{
		configDir:  configDir,
		configFile: filepath.Join(configDir, "config.json"),
		secretFile: filepath.Join(configDir, ".secret"),
	}
}

// LoadOrCreateConfig loads existing config or creates default
func (cm *ConfigManager) LoadOrCreateConfig() (*Config, error) {
	// Create analytics directory if it doesn't exist
	if err := os.MkdirAll(cm.configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create analytics directory: %w", err)
	}

	// Check if config file exists
	if _, err := os.Stat(cm.configFile); os.IsNotExist(err) {
		// Create default config
		config := cm.getDefaultConfig()
		if err := cm.SaveConfig(config); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
		pterm.Info.Println("Created default analytics configuration")
		return config, nil
	}

	// Load existing config
	data, err := os.ReadFile(cm.configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to file
func (cm *ConfigManager) SaveConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cm.configFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// EnsureAPIKey ensures an API key exists, generates one if needed
func (cm *ConfigManager) EnsureAPIKey() (string, error) {
	// Check if secret file exists
	if data, err := os.ReadFile(cm.secretFile); err == nil {
		apiKey := string(data)
		if len(apiKey) > 0 {
			return apiKey, nil
		}
	}

	// Generate new API key
	apiKey, err := cm.generateAPIKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate API key: %w", err)
	}

	// Save to file
	if err := os.WriteFile(cm.secretFile, []byte(apiKey), 0600); err != nil {
		return "", fmt.Errorf("failed to write API key: %w", err)
	}

	pterm.Success.Println("Generated new API key for authentication")
	pterm.Info.Printf("API key saved to: %s\n", cm.secretFile)

	return apiKey, nil
}

// GetAPIKey returns the current API key
func (cm *ConfigManager) GetAPIKey() (string, error) {
	data, err := os.ReadFile(cm.secretFile)
	if err != nil {
		return "", fmt.Errorf("failed to read API key: %w", err)
	}
	return string(data), nil
}

// generateAPIKey generates a random API key
func (cm *ConfigManager) generateAPIKey() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// getDefaultConfig returns the default configuration
func (cm *ConfigManager) getDefaultConfig() *Config {
	return &Config{
		TLS: TLSSettings{
			Enabled: true,
		},
		Auth: AuthSettings{
			Enabled:    true,
			APIKeyPath: cm.secretFile,
		},
		Server: ServerSettings{
			Port:  3333,
			Host:  "127.0.0.1", // Localhost only by default for security
			Quiet: false,
		},
		CORS: CORSSettings{
			AllowedOrigins: []string{
				"http://localhost:3333",
				"https://localhost:3333",
				"http://127.0.0.1:3333",
				"https://127.0.0.1:3333",
			},
		},
		Agent: AgentSettings{
			Model:                 "claude-3-5-sonnet-latest",
			MaxConcurrentSessions: 10,
		},
	}
}

// GetConfigPath returns the path to the config file
func (cm *ConfigManager) GetConfigPath() string {
	return cm.configFile
}

// GetSecretPath returns the path to the secret file
func (cm *ConfigManager) GetSecretPath() string {
	return cm.secretFile
}
