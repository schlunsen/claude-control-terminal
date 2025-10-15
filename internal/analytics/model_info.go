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
//   "claude-haiku-3-5-20241022" -> "Haiku 3.5"
func parseHumanReadableModelName(modelID string) string {
	// Remove "claude-" prefix if present
	modelID = strings.TrimPrefix(modelID, "claude-")

	// Try pattern 1: model-major-minor (e.g., sonnet-4-5)
	re := regexp.MustCompile(`^(sonnet|opus|haiku)-(\d+)-(\d+)-`)
	if matches := re.FindStringSubmatch(modelID); len(matches) >= 4 {
		family := strings.Title(matches[1])
		major := matches[2]
		minor := matches[3]
		return family + " " + major + "." + minor
	}

	// Try pattern 2: model-major (e.g., opus-4)
	re = regexp.MustCompile(`^(sonnet|opus|haiku)-(\d+)-`)
	if matches := re.FindStringSubmatch(modelID); len(matches) >= 3 {
		family := strings.Title(matches[1])
		major := matches[2]
		return family + " " + major
	}

	// Try pattern 3: major-minor-model (e.g., 3-5-sonnet)
	re = regexp.MustCompile(`^(\d+)-(\d+)-(sonnet|opus|haiku)-`)
	if matches := re.FindStringSubmatch(modelID); len(matches) >= 4 {
		major := matches[1]
		minor := matches[2]
		family := strings.Title(matches[3])
		return family + " " + major + "." + minor
	}

	// Try pattern 4: major-model (e.g., 3-sonnet)
	re = regexp.MustCompile(`^(\d+)-(sonnet|opus|haiku)-`)
	if matches := re.FindStringSubmatch(modelID); len(matches) >= 3 {
		major := matches[1]
		family := strings.Title(matches[2])
		return family + " " + major
	}

	// Fallback: capitalize and clean up the model ID
	// Remove date suffix (e.g., -20250929)
	re = regexp.MustCompile(`-\d{8}$`)
	cleaned := re.ReplaceAllString(modelID, "")

	// Replace hyphens with spaces and capitalize
	parts := strings.Split(cleaned, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.Title(part)
		}
	}

	return strings.Join(parts, " ")
}
