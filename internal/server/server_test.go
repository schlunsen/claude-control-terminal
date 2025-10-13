package server

import (
	"testing"
)

func TestNewServer(t *testing.T) {
	server := NewServer("/test/claude", 3333)

	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	if server.claudeDir != "/test/claude" {
		t.Errorf("expected claudeDir '/test/claude', got %q", server.claudeDir)
	}

	if server.port != 3333 {
		t.Errorf("expected port 3333, got %d", server.port)
	}

	if server.quiet {
		t.Error("expected quiet to be false by default")
	}

	if server.app == nil {
		t.Error("app should be initialized")
	}
}

func TestNewServerWithOptions(t *testing.T) {
	tests := []struct {
		name      string
		claudeDir string
		port      int
		quiet     bool
	}{
		{
			name:      "default settings",
			claudeDir: "/test/claude",
			port:      3333,
			quiet:     false,
		},
		{
			name:      "quiet mode",
			claudeDir: "/test/claude",
			port:      8080,
			quiet:     true,
		},
		{
			name:      "custom port",
			claudeDir: "/custom/path",
			port:      9999,
			quiet:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServerWithOptions(tt.claudeDir, tt.port, tt.quiet)

			if server == nil {
				t.Fatal("NewServerWithOptions returned nil")
			}

			if server.claudeDir != tt.claudeDir {
				t.Errorf("expected claudeDir %q, got %q", tt.claudeDir, server.claudeDir)
			}

			if server.port != tt.port {
				t.Errorf("expected port %d, got %d", tt.port, server.port)
			}

			if server.quiet != tt.quiet {
				t.Errorf("expected quiet %v, got %v", tt.quiet, server.quiet)
			}

			if server.app == nil {
				t.Error("app should be initialized")
			}
		})
	}
}

func TestServerStruct(t *testing.T) {
	// Test that Server struct can be created with zero values
	server := Server{
		claudeDir: "/test",
		port:      3000,
		quiet:     false,
	}

	if server.claudeDir != "/test" {
		t.Errorf("expected claudeDir '/test', got %q", server.claudeDir)
	}

	if server.port != 3000 {
		t.Errorf("expected port 3000, got %d", server.port)
	}

	if server.quiet {
		t.Error("expected quiet to be false")
	}
}

func TestServerDefaultPort(t *testing.T) {
	server := NewServer("/test/claude", 3333)

	if server.port != 3333 {
		t.Errorf("expected default port 3333, got %d", server.port)
	}
}

func TestServerQuietModeFlag(t *testing.T) {
	// Test non-quiet mode
	server1 := NewServerWithOptions("/test", 3333, false)
	if server1.quiet {
		t.Error("expected quiet to be false")
	}

	// Test quiet mode
	server2 := NewServerWithOptions("/test", 3333, true)
	if !server2.quiet {
		t.Error("expected quiet to be true")
	}
}

func TestNewServerInitializesFiberApp(t *testing.T) {
	server := NewServer("/test/claude", 3333)

	if server.app == nil {
		t.Fatal("Fiber app should be initialized")
	}

	// Verify app has basic configuration
	// We can't deeply inspect Fiber config without exposing internals,
	// but we can verify the app is not nil and usable
}

func TestServerWithDifferentPorts(t *testing.T) {
	tests := []int{3000, 3333, 8080, 9000, 10000}

	for _, port := range tests {
		server := NewServer("/test", port)

		if server.port != port {
			t.Errorf("expected port %d, got %d", port, server.port)
		}
	}
}

func TestServerWithDifferentDirectories(t *testing.T) {
	tests := []string{
		"/home/user/.claude",
		"/var/claude",
		"/tmp/test-claude",
		"./relative/path",
		".",
	}

	for _, dir := range tests {
		server := NewServer(dir, 3333)

		if server.claudeDir != dir {
			t.Errorf("expected claudeDir %q, got %q", dir, server.claudeDir)
		}
	}
}

func TestServerFieldsInitialization(t *testing.T) {
	server := NewServer("/test", 3333)

	// Check that some fields are properly initialized to nil/zero values
	// (they will be set up during Setup())
	if server.conversationAnalyzer != nil {
		t.Error("conversationAnalyzer should be nil before Setup()")
	}

	if server.conversationParser != nil {
		t.Error("conversationParser should be nil before Setup()")
	}

	if server.stateCalculator != nil {
		t.Error("stateCalculator should be nil before Setup()")
	}

	if server.processDetector != nil {
		t.Error("processDetector should be nil before Setup()")
	}

	if server.wsHub != nil {
		t.Error("wsHub should be nil before Setup()")
	}

	if server.db != nil {
		t.Error("db should be nil before Setup()")
	}

	if server.repo != nil {
		t.Error("repo should be nil before Setup()")
	}
}

func TestNewServerWithEmptyDirectory(t *testing.T) {
	server := NewServer("", 3333)

	if server == nil {
		t.Fatal("NewServer should handle empty directory")
	}

	if server.claudeDir != "" {
		t.Errorf("expected empty claudeDir, got %q", server.claudeDir)
	}
}

func TestNewServerWithZeroPort(t *testing.T) {
	server := NewServer("/test", 0)

	if server == nil {
		t.Fatal("NewServer should handle zero port")
	}

	if server.port != 0 {
		t.Errorf("expected port 0, got %d", server.port)
	}
}

func TestServerMultipleInstances(t *testing.T) {
	// Test that we can create multiple server instances
	server1 := NewServer("/test1", 3001)
	server2 := NewServer("/test2", 3002)

	if server1 == nil || server2 == nil {
		t.Fatal("should be able to create multiple server instances")
	}

	if server1.claudeDir == server2.claudeDir {
		t.Error("servers should have different directories")
	}

	if server1.port == server2.port {
		t.Error("servers should have different ports")
	}
}

func TestServerQuietModeSuppressesLogger(t *testing.T) {
	// Create server in quiet mode
	server := NewServerWithOptions("/test", 3333, true)

	if !server.quiet {
		t.Error("server should be in quiet mode")
	}

	// The Fiber app should be configured with DisableStartupMessage
	// We can't directly test this without starting the server,
	// but we can verify the flag is set correctly
	if server.app == nil {
		t.Error("app should be initialized even in quiet mode")
	}
}
