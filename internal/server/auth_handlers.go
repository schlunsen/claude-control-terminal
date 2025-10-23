package server

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string    `json:"token"`
	Username  string    `json:"username"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// CreateUserRequest represents a user creation request
type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

// SessionAuthMiddleware creates a middleware for session authentication
type SessionAuthMiddleware struct {
	userStore *UserStore
	enabled   bool
}

// NewSessionAuthMiddleware creates a new session authentication middleware
func NewSessionAuthMiddleware(userStore *UserStore, enabled bool) *SessionAuthMiddleware {
	return &SessionAuthMiddleware{
		userStore: userStore,
		enabled:   enabled,
	}
}

// Handler returns the Fiber middleware handler
func (sam *SessionAuthMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip if authentication is disabled
		if !sam.enabled {
			return c.Next()
		}

		// Skip auth for login endpoint
		if c.Path() == "/api/auth/login" || c.Path() == "/api/auth/status" {
			return c.Next()
		}

		// Allow GET requests for static files and public endpoints
		method := c.Method()
		if method == "GET" || method == "OPTIONS" {
			// Check if this is a public endpoint (dashboard pages, assets, etc.)
			path := c.Path()
			if path == "/" || path == "/api/health" || path == "/api/version" ||
				path == "/assets/" || path == "/favicon.ico" {
				return c.Next()
			}
		}

		// Get session token from cookie or Authorization header
		token := c.Cookies("session_token")
		if token == "" {
			// Try Authorization header
			authHeader := c.Get("Authorization")
			if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				token = authHeader[7:]
			}
		}

		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		// Validate session
		user, err := sam.userStore.ValidateSession(token)
		if err != nil {
			// Clear invalid cookie
			c.ClearCookie("session_token")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired session",
			})
		}

		// Store user in context
		c.Locals("user", user)

		return c.Next()
	}
}

// handleLogin handles user login
func (s *Server) handleLogin(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Authenticate user
	token, err := s.userStore.Authenticate(req.Username, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid username or password",
		})
	}

	// Set session cookie
	c.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HTTPOnly: true,
		Secure:   s.config.TLS.Enabled, // Only set Secure flag if TLS is enabled
		SameSite: "Lax",
		Expires:  time.Now().Add(24 * time.Hour),
	})

	return c.JSON(LoginResponse{
		Token:     token,
		Username:  req.Username,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
}

// handleLogout handles user logout
func (s *Server) handleLogout(c *fiber.Ctx) error {
	// Get session token
	token := c.Cookies("session_token")
	if token == "" {
		authHeader := c.Get("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}

	if token != "" {
		// Revoke session
		s.userStore.RevokeSession(token)
	}

	// Clear cookie
	c.ClearCookie("session_token")

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

// handleAuthStatus returns the current authentication status
func (s *Server) handleAuthStatus(c *fiber.Ctx) error {
	// Get session token
	token := c.Cookies("session_token")
	if token == "" {
		authHeader := c.Get("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}

	// Check if user auth is enabled
	if !s.config.Auth.UserAuthEnabled {
		return c.JSON(fiber.Map{
			"enabled":       false,
			"authenticated": false,
			"requireLogin":  false,
		})
	}

	// Validate session
	if token != "" {
		user, err := s.userStore.ValidateSession(token)
		if err == nil {
			return c.JSON(fiber.Map{
				"enabled":       true,
				"authenticated": true,
				"requireLogin":  s.config.Auth.RequireLogin,
				"username":      user.Username,
				"isAdmin":       user.IsAdmin,
			})
		}
	}

	return c.JSON(fiber.Map{
		"enabled":       true,
		"authenticated": false,
		"requireLogin":  s.config.Auth.RequireLogin,
	})
}

// handleChangePassword handles password change
func (s *Server) handleChangePassword(c *fiber.Ctx) error {
	// Get current user from context
	user, ok := c.Locals("user").(*User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update password
	if err := s.userStore.UpdatePassword(user.Username, req.OldPassword, req.NewPassword); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password updated successfully",
	})
}

// handleCreateUser handles user creation (admin only)
func (s *Server) handleCreateUser(c *fiber.Ctx) error {
	// Get current user from context
	user, ok := c.Locals("user").(*User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	// Check if user is admin
	if !user.IsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin access required",
		})
	}

	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Create user
	if err := s.userStore.CreateUser(req.Username, req.Password, req.IsAdmin); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":  "User created successfully",
		"username": req.Username,
	})
}

// handleListUsers lists all users (admin only)
func (s *Server) handleListUsers(c *fiber.Ctx) error {
	// Get current user from context
	user, ok := c.Locals("user").(*User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	// Check if user is admin
	if !user.IsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin access required",
		})
	}

	users := s.userStore.ListUsers()

	return c.JSON(fiber.Map{
		"users": users,
		"count": len(users),
	})
}

// handleDeleteUser deletes a user (admin only)
func (s *Server) handleDeleteUser(c *fiber.Ctx) error {
	// Get current user from context
	user, ok := c.Locals("user").(*User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	// Check if user is admin
	if !user.IsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin access required",
		})
	}

	username := c.Params("username")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username is required",
		})
	}

	// Prevent deleting self
	if username == user.Username {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot delete your own account",
		})
	}

	// Delete user
	if err := s.userStore.DeleteUser(username); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}
