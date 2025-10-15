package analytics

import (
	"os"
	"sync"
)

// ModelInfo holds parsed model information
type ModelInfo struct {
	Provider string
	Name     string
}

var (
	cachedModelInfo *ModelInfo
	modelInfoOnce   sync.Once
)

// GetModelInfo reads and parses model information from environment variables
// Returns cached result on subsequent calls
func GetModelInfo() ModelInfo {
	modelInfoOnce.Do(func() {
		cachedModelInfo = parseModelFromEnv()
	})
	if cachedModelInfo != nil {
		return *cachedModelInfo
	}
	return ModelInfo{Provider: "Unknown", Name: "Unknown"}
}

// parseModelFromEnv reads ANTHROPIC_MODEL and parses it
func parseModelFromEnv() *ModelInfo {
	modelEnv := os.Getenv("ANTHROPIC_MODEL")
	if modelEnv == "" {
		return &ModelInfo{Provider: "Unknown", Name: "Unknown"}
	}

	// Parse model string like "claude-sonnet-4-5-20250929"
	// Extract provider and human-readable name

	provider := "Anthropic"
	name := parseHumanReadableModelName(modelEnv)

	return &ModelInfo{
		Provider: provider,
		Name:     name,
	}
}

// parseHumanReadableModelName returns the model ID as-is
func parseHumanReadableModelName(modelID string) string {
	return modelID
}
