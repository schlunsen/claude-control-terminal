// Package server provides static file serving for the analytics dashboard.
// This file embeds the Nuxt dashboard and serves it via Fiber.
package server

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

//go:embed frontend/.output/public/*
var frontendAssets embed.FS

// ServeStaticFiles adds static file serving to the server
func (s *Server) ServeStaticFiles() {

	// Get the embedded filesystem for the public directory
	publicFS, err := fs.Sub(frontendAssets, "frontend/.output/public")
	if err != nil {
		panic("failed to load embedded frontend assets: " + err.Error())
	}

	// Serve static assets (must come before catch-all)
	// This will handle /_nuxt/*, favicon.ico, robots.txt, etc.
	s.app.Use("/", filesystem.New(filesystem.Config{
		Root:   http.FS(publicFS),
		Browse: false,
		Next: func(c *fiber.Ctx) bool {
			// Skip for API routes and WebSocket
			path := c.Path()
			return (len(path) >= 4 && path[:4] == "/api") || (len(path) >= 3 && path[:3] == "/ws")
		},
	}))
	
	// Catch-all route for SPA - serve index.html for any unmatched routes
	// This must come after API routes and static files
	s.app.Get("/*", func(c *fiber.Ctx) error {
		// Read index.html from embedded filesystem
		indexData, err := fs.ReadFile(publicFS, "index.html")
		if err != nil {
			return c.Status(500).SendString("Failed to load application")
		}
		
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Send(indexData)
	})
}
