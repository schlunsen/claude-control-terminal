package agents

// Config holds configuration for the agent handler
type Config struct {
	Model                 string
	APIKey                string
	MaxConcurrentSessions int
	Verbose               bool
}
