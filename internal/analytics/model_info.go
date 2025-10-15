package analytics

import (
	"os"
	"regexp"
	"strings"
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

// parseHumanReadableModelName converts model ID to display name
// Examples:
//   "claude-sonnet-4-5-20250929" -> "Sonnet 4.5"
//   "claude-opus-4-20250514" -> "Opus 4"
//   "claude-3-5-sonnet-20241022" -> "Sonnet 3.5"
func parseHumanReadableModelName(modelID string) string {
	// Remove "claude-" prefix and date suffix
	modelID = strings.TrimPrefix(modelID, "claude-")
	re := regexp.MustCompile(`-\d{8}$`)
	modelID = re.ReplaceAllString(modelID, "")

	// Extract family and version: "sonnet-4-5" or "3-5-sonnet"
	re = regexp.MustCompile(`(sonnet|opus|haiku)|(\d+)`)
	matches := re.FindAllString(modelID, -1)

	var family string
	var version []string

	for _, match := range matches {
		if match == "sonnet" || match == "opus" || match == "haiku" {
			family = strings.Title(match)
		} else {
			version = append(version, match)
		}
	}

	if family == "" {
		return modelID // Fallback to original
	}

	if len(version) > 0 {
		return family + " " + strings.Join(version, ".")
	}

	return family
}
