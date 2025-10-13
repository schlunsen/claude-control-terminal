// Package server provides static file serving for the analytics dashboard.
// This file embeds the HTML dashboard and serves it via Fiber.
package server

import (
	_ "embed"

	"github.com/gofiber/fiber/v2"
)

//go:embed static/index.html
var indexHTML []byte

// ServeStaticFiles adds static file serving to the server
func (s *Server) ServeStaticFiles() {
	// Serve index.html at root
	s.app.Get("/", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.Send(indexHTML)
	})
}
