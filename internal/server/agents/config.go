package agents

// Config holds configuration for the agent handler
type Config struct {
	Model                 string
	APIKey                string
	MaxConcurrentSessions int
	Verbose               bool
	// Session retention configuration
	SessionRetentionDays  int  // Days to keep ended sessions (default: 30)
	CleanupEnabled        bool // Enable automatic cleanup (default: true)
	CleanupIntervalHours  int  // Cleanup interval in hours (default: 24)
}
