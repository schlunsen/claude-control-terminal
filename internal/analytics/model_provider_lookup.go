// Package analytics provides conversation analysis, state calculation, and model provider lookup.
// This file contains logic for mapping ANTHROPIC_BASE_URL values to provider IDs using providers.json.
package analytics

import (
	"strings"

	"github.com/schlunsen/claude-control-terminal/internal/providers"
)

// ModelProviderLookup maps ANTHROPIC_BASE_URL values to provider IDs
type ModelProviderLookup struct {
	providers []providers.Provider
}

// NewModelProviderLookup creates a new model provider lookup instance
func NewModelProviderLookup() *ModelProviderLookup {
	return &ModelProviderLookup{
		providers: providers.GetAvailableProviders(),
	}
}

// GetProviderName returns the provider ID for a given base URL
// Uses providers.json for matching. Returns lowercase provider ID.
func (m *ModelProviderLookup) GetProviderName(baseURL string) string {
	// Handle empty or default case - return "anthropic" (Claude default)
	if baseURL == "" || baseURL == "https://api.anthropic.com" {
		return "anthropic"
	}

	// Normalize URL - remove trailing slash for comparison
	normalizedURL := strings.TrimSuffix(baseURL, "/")

	// Try exact match first
	for _, provider := range m.providers {
		if provider.BaseURL == "" {
			continue // Skip Claude (default) and Custom
		}

		providerURL := strings.TrimSuffix(provider.BaseURL, "/")
		if normalizedURL == providerURL {
			return provider.ID
		}
	}

	// Try prefix matching for URLs with paths (e.g., "https://api.deepseek.com/anthropic")
	for _, provider := range m.providers {
		if provider.BaseURL == "" {
			continue
		}

		providerURL := strings.TrimSuffix(provider.BaseURL, "/")
		if strings.HasPrefix(normalizedURL, providerURL) {
			return provider.ID
		}
	}

	// Extract domain for unknown custom URLs
	if strings.HasPrefix(baseURL, "http") {
		parts := strings.Split(baseURL, "//")
		if len(parts) >= 2 {
			domainParts := strings.Split(parts[1], "/")
			if len(domainParts) > 0 {
				domain := domainParts[0]
				// Remove port if present
				if colonIndex := strings.Index(domain, ":"); colonIndex != -1 {
					domain = domain[:colonIndex]
				}
				// Return domain-based name (lowercase)
				return "custom (" + domain + ")"
			}
		}
	}

	// Fallback for unknown URLs
	return "custom"
}

// GetProviderNameFromModelInfo is a convenience function that takes model provider and model name
// and returns the provider ID. If the provider is already a provider ID (not a URL),
// it returns it as-is. If it's a URL, it performs the lookup.
func (m *ModelProviderLookup) GetProviderNameFromModelInfo(provider, modelName string) string {
	// If provider is empty or "Unknown", try to infer from model name
	if provider == "" || provider == "Unknown" || provider == "unknown" {
		// Check if model name suggests a known provider
		if modelName != "" && modelName != "Unknown" && modelName != "unknown" {
			// Try to match model name to a provider
			modelLower := strings.ToLower(modelName)
			for _, p := range m.providers {
				for _, model := range p.Models {
					if strings.ToLower(model) == modelLower {
						return p.ID
					}
				}
			}
			// Default to anthropic for unknown but non-empty model names
			return "anthropic"
		}
		return "unknown"
	}

	// If provider is already a provider ID (not a URL), return it as-is
	if !strings.HasPrefix(provider, "http") {
		return provider
	}

	// If provider is a URL, perform lookup
	return m.GetProviderName(provider)
}