package server

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pterm/pterm"
)

// AuthMiddleware creates a middleware for API key authentication
type AuthMiddleware struct {
	apiKey  string
	enabled bool
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(apiKey string, enabled bool) *AuthMiddleware {
	return &AuthMiddleware{
		apiKey:  apiKey,
		enabled: enabled,
	}
}

// Handler returns the Fiber middleware handler
func (am *AuthMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip if authentication is disabled
		if !am.enabled {
			return c.Next()
		}

		// Allow GET requests without authentication (for browser access)
		// Only require auth for POST, PUT, DELETE, PATCH
		method := c.Method()
		if method == "GET" || method == "OPTIONS" {
			return c.Next()
		}

		// Check Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing Authorization header",
				"message": "API key required. Include 'Authorization: Bearer <api-key>' header",
			})
		}

		// Parse Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid Authorization header format",
				"message": "Use 'Authorization: Bearer <api-key>' format",
			})
		}

		token := parts[1]

		// Validate token
		if token != am.apiKey {
			pterm.Warning.Printf("Unauthorized API request from %s\n", c.IP())
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid API key",
				"message": "The provided API key is not valid",
			})
		}

		// Token is valid, continue
		return c.Next()
	}
}

// ProtectEndpoint creates a one-off middleware to protect specific endpoints
func ProtectEndpoint(apiKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" || parts[1] != apiKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		return c.Next()
	}
}
