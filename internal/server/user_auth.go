package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents an authenticated user
type User struct {
	Username     string    `json:"username"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsAdmin      bool      `json:"is_admin"`
}

// Session represents a user session
type Session struct {
	Token     string    `json:"token"`
	Username  string    `json:"username"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// UserStore manages user accounts
type UserStore struct {
	usersFile    string
	sessionsFile string
	users        map[string]*User
	sessions     map[string]*Session
}

const (
	// Session duration (24 hours)
	SessionDuration = 24 * time.Hour
	// Password hash cost
	PasswordHashCost = 12
)

// NewUserStore creates a new user store
func NewUserStore(configDir string) *UserStore {
	usersDir := filepath.Join(configDir, "users")
	os.MkdirAll(usersDir, 0700)

	return &UserStore{
		usersFile:    filepath.Join(usersDir, "users.json"),
		sessionsFile: filepath.Join(usersDir, "sessions.json"),
		users:        make(map[string]*User),
		sessions:     make(map[string]*Session),
	}
}

// Initialize loads users and sessions from disk
func (us *UserStore) Initialize() error {
	// Load users
	if _, err := os.Stat(us.usersFile); err == nil {
		data, err := os.ReadFile(us.usersFile)
		if err != nil {
			return fmt.Errorf("failed to read users file: %w", err)
		}

		if err := json.Unmarshal(data, &us.users); err != nil {
			return fmt.Errorf("failed to parse users file: %w", err)
		}
	}

	// Load sessions
	if _, err := os.Stat(us.sessionsFile); err == nil {
		data, err := os.ReadFile(us.sessionsFile)
		if err != nil {
			return fmt.Errorf("failed to read sessions file: %w", err)
		}

		if err := json.Unmarshal(data, &us.sessions); err != nil {
			return fmt.Errorf("failed to parse sessions file: %w", err)
		}

		// Clean up expired sessions
		us.cleanupExpiredSessions()
	}

	return nil
}

// HasUsers returns true if any users exist
func (us *UserStore) HasUsers() bool {
	return len(us.users) > 0
}

// CreateUser creates a new user account
func (us *UserStore) CreateUser(username, password string, isAdmin bool) error {
	// Validate username
	if username == "" {
		return errors.New("username cannot be empty")
	}

	// Check if user already exists
	if _, exists := us.users[username]; exists {
		return errors.New("user already exists")
	}

	// Validate password strength
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), PasswordHashCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &User{
		Username:     username,
		PasswordHash: string(passwordHash),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsAdmin:      isAdmin,
	}

	us.users[username] = user

	// Save to disk
	return us.saveUsers()
}

// Authenticate verifies username and password, returns session token on success
func (us *UserStore) Authenticate(username, password string) (string, error) {
	// Get user
	user, exists := us.users[username]
	if !exists {
		return "", errors.New("invalid username or password")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid username or password")
	}

	// Generate session token
	token, err := us.generateSessionToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}

	// Create session
	session := &Session{
		Token:     token,
		Username:  username,
		ExpiresAt: time.Now().Add(SessionDuration),
		CreatedAt: time.Now(),
	}

	us.sessions[token] = session

	// Save sessions
	if err := us.saveSessions(); err != nil {
		return "", fmt.Errorf("failed to save session: %w", err)
	}

	return token, nil
}

// ValidateSession checks if a session token is valid
func (us *UserStore) ValidateSession(token string) (*User, error) {
	// Get session
	session, exists := us.sessions[token]
	if !exists {
		return nil, errors.New("invalid session")
	}

	// Check expiration
	if time.Now().After(session.ExpiresAt) {
		delete(us.sessions, token)
		us.saveSessions()
		return nil, errors.New("session expired")
	}

	// Get user
	user, exists := us.users[session.Username]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// RevokeSession removes a session
func (us *UserStore) RevokeSession(token string) error {
	delete(us.sessions, token)
	return us.saveSessions()
}

// UpdatePassword changes a user's password
func (us *UserStore) UpdatePassword(username, oldPassword, newPassword string) error {
	// Get user
	user, exists := us.users[username]
	if !exists {
		return errors.New("user not found")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("invalid current password")
	}

	// Validate new password
	if len(newPassword) < 8 {
		return errors.New("new password must be at least 8 characters")
	}

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), PasswordHashCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user
	user.PasswordHash = string(passwordHash)
	user.UpdatedAt = time.Now()

	// Save to disk
	return us.saveUsers()
}

// DeleteUser removes a user account
func (us *UserStore) DeleteUser(username string) error {
	delete(us.users, username)
	return us.saveUsers()
}

// ListUsers returns all usernames
func (us *UserStore) ListUsers() []string {
	usernames := make([]string, 0, len(us.users))
	for username := range us.users {
		usernames = append(usernames, username)
	}
	return usernames
}

// cleanupExpiredSessions removes expired sessions
func (us *UserStore) cleanupExpiredSessions() {
	now := time.Now()
	for token, session := range us.sessions {
		if now.After(session.ExpiresAt) {
			delete(us.sessions, token)
		}
	}
}

// generateSessionToken generates a random session token
func (us *UserStore) generateSessionToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// saveUsers saves users to disk
func (us *UserStore) saveUsers() error {
	data, err := json.MarshalIndent(us.users, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal users: %w", err)
	}

	if err := os.WriteFile(us.usersFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write users file: %w", err)
	}

	return nil
}

// saveSessions saves sessions to disk
func (us *UserStore) saveSessions() error {
	data, err := json.MarshalIndent(us.sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sessions: %w", err)
	}

	if err := os.WriteFile(us.sessionsFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write sessions file: %w", err)
	}

	return nil
}
