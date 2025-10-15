// Package server provides static file serving for the analytics dashboard.
// This file embeds the Nuxt-generated static files from frontend/.output/public
package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

//go:embed frontend/.output/public/*
var frontendFiles embed.FS

// ServeStaticFiles adds static file serving to the server
func (s *Server) ServeStaticFiles() {
	// Get the public subdirectory from the embedded filesystem
	publicFS, err := fs.Sub(frontendFiles, "frontend/.output/public")
	if err != nil {
		panic("Failed to load embedded frontend files: " + err.Error())
	}

	// Serve static files from embedded filesystem using http.FS wrapper
	s.app.Use("/", filesystem.New(filesystem.Config{
		Root:         http.FS(publicFS),
		Browse:       false,
		Index:        "index.html",
		NotFoundFile: "404.html",
		MaxAge:       3600, // 1 hour cache for static assets
	}))

	// Fallback to index.html for SPA routes (must come after API/WS routes)
	s.app.Use(func(c *fiber.Ctx) error {
		// Skip for API routes and WebSocket
		path := c.Path()
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/ws") {
			return c.Next()
		}

		// Check if file exists in embedded FS
		if _, err := publicFS.Open(strings.TrimPrefix(path, "/")); err == nil {
			return c.Next()
		}

		// Serve index.html for non-existent routes (SPA fallback)
		indexContent, err := fs.ReadFile(publicFS, "index.html")
		if err != nil {
			return c.Status(404).SendString("404 Not Found")
		}

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Send(indexContent)
	})
}
