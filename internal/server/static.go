// Package server provides static file serving for the analytics dashboard.
// This file embeds static HTML for the analytics dashboard.
package server

import (
	_ "embed"

	"github.com/gofiber/fiber/v2"
)

//go:embed static/index.html
var indexHTML []byte

// ServeStaticFiles adds static file serving to the server
func (s *Server) ServeStaticFiles() {
	// Serve the embedded analytics dashboard HTML
	s.app.Get("/*", func(c *fiber.Ctx) error {
		// Skip for API routes and WebSocket
		path := c.Path()
		if (len(path) >= 4 && path[:4] == "/api") || (len(path) >= 3 && path[:3] == "/ws") {
			return c.Next()
		}

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Send(indexHTML)
	})
}
