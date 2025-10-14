package installer

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

func TestNewClaudeInstaller(t *testing.T) {
	ci := NewClaudeInstaller()

	if ci == nil {
		t.Fatal("NewClaudeInstaller returned nil")
	}

	if ci.Timeout != 5*time.Minute {
		t.Errorf("Expected timeout 5m, got %v", ci.Timeout)
	}

	if ci.Verbose {
		t.Error("Expected Verbose to be false by default")
	}
}

func TestIsClaudeInstalled(t *testing.T) {
	ci := NewClaudeInstaller()

	// This will return true or false based on actual system state
	installed := ci.IsClaudeInstalled()

	// Verify it matches exec.LookPath result
	_, err := exec.LookPath("claude")
	expectedInstalled := err == nil

	if installed != expectedInstalled {
		t.Errorf("IsClaudeInstalled() = %v, expected %v", installed, expectedInstalled)
	}
}

func TestGetClaudePath_NotInstalled(t *testing.T) {
	ci := NewClaudeInstaller()

	// Only test error case if Claude is not installed
	if !ci.IsClaudeInstalled() {
		_, err := ci.GetClaudePath()
		if err == nil {
			t.Error("Expected error when Claude is not installed")
		}
	}
}

func TestGetClaudePath_Installed(t *testing.T) {
	ci := NewClaudeInstaller()

	// Only test success case if Claude is installed
	if ci.IsClaudeInstalled() {
		path, err := ci.GetClaudePath()
		if err != nil {
			t.Errorf("GetClaudePath failed: %v", err)
		}

		if path == "" {
			t.Error("Expected non-empty path")
		}

		// Verify file exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Claude binary does not exist at %s", path)
		}
	}
}

func TestGetClaudeVersion_Installed(t *testing.T) {
	ci := NewClaudeInstaller()

	// Only test if Claude is installed
	if ci.IsClaudeInstalled() {
		version, err := ci.GetClaudeVersion()
		if err != nil {
			t.Errorf("GetClaudeVersion failed: %v", err)
		}

		if version == "" {
			t.Error("Expected non-empty version string")
		}
	}
}

func TestInstallResult(t *testing.T) {
	result := &InstallResult{
		Success:    true,
		Method:     InstallMethodNative,
		ClaudePath: "/usr/local/bin/claude",
		Version:    "v2.0.0",
		Message:    "Test message",
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}

	if result.Method != InstallMethodNative {
		t.Errorf("Expected method %v, got %v", InstallMethodNative, result.Method)
	}

	if result.ClaudePath != "/usr/local/bin/claude" {
		t.Errorf("Expected path /usr/local/bin/claude, got %s", result.ClaudePath)
	}
}

func TestGetInstallInstructions(t *testing.T) {
	instructions := GetInstallInstructions()

	if instructions == "" {
		t.Error("GetInstallInstructions returned empty string")
	}

	// Check for platform-specific content
	switch runtime.GOOS {
	case "darwin", "linux":
		if len(instructions) < 100 {
			t.Error("Instructions seem too short")
		}
		// Should contain curl command
		// Note: We can't check exact content as it may change
	case "windows":
		if len(instructions) < 100 {
			t.Error("Instructions seem too short")
		}
	}
}

func TestGetRecommendedMethod(t *testing.T) {
	method := GetRecommendedMethod()

	if method != InstallMethodNative {
		t.Errorf("Expected recommended method %v, got %v", InstallMethodNative, method)
	}
}

func TestGetInstallLocation_Installed(t *testing.T) {
	ci := NewClaudeInstaller()

	// Only test if Claude is installed
	if ci.IsClaudeInstalled() {
		location, err := ci.GetInstallLocation()
		if err != nil {
			t.Errorf("GetInstallLocation failed: %v", err)
		}

		if location == "" {
			t.Error("Expected non-empty install location")
		}

		// Verify directory exists
		if info, err := os.Stat(location); err != nil || !info.IsDir() {
			t.Errorf("Install location %s is not a valid directory", location)
		}
	}
}

func TestInstallMethodConstants(t *testing.T) {
	if InstallMethodNative != "native" {
		t.Errorf("Expected InstallMethodNative = 'native', got '%s'", InstallMethodNative)
	}

	if InstallMethodNPM != "npm" {
		t.Errorf("Expected InstallMethodNPM = 'npm', got '%s'", InstallMethodNPM)
	}
}

func TestCheckForUpdates_NotInstalled(t *testing.T) {
	ci := NewClaudeInstaller()

	// Only test error case if Claude is not installed
	if !ci.IsClaudeInstalled() {
		_, _, err := ci.CheckForUpdates()
		if err == nil {
			t.Error("Expected error when Claude is not installed")
		}
	}
}

func TestCheckForUpdates_Installed(t *testing.T) {
	ci := NewClaudeInstaller()

	// Only test if Claude is installed
	if ci.IsClaudeInstalled() {
		hasUpdate, version, err := ci.CheckForUpdates()
		if err != nil {
			t.Errorf("CheckForUpdates failed: %v", err)
		}

		// Claude auto-updates, so this should always return false
		if hasUpdate {
			t.Error("Expected hasUpdate to be false (Claude auto-updates)")
		}

		if version == "" {
			t.Error("Expected non-empty version")
		}
	}
}

func TestVerifyInstallation_NotInstalled(t *testing.T) {
	ci := NewClaudeInstaller()

	// Only test error case if Claude is not installed
	if !ci.IsClaudeInstalled() {
		err := ci.VerifyInstallation()
		if err == nil {
			t.Error("Expected error when Claude is not installed")
		}
	}
}

// Note: We skip testing actual installation methods as they would require
// uninstalling Claude and then reinstalling it, which is destructive and
// may fail in CI environments. These methods are tested manually.

// TestInstall_AlreadyInstalled tests the case where Claude is already installed
func TestInstall_AlreadyInstalled(t *testing.T) {
	ci := NewClaudeInstaller()

	// Only test if Claude is installed
	if ci.IsClaudeInstalled() {
		result := ci.Install(InstallMethodNative)

		if !result.Success {
			t.Error("Expected success when Claude is already installed")
		}

		if result.Message != "Claude CLI is already installed" {
			t.Errorf("Expected 'already installed' message, got: %s", result.Message)
		}

		if result.ClaudePath == "" {
			t.Error("Expected non-empty ClaudePath")
		}
	}
}

// TestAutoInstall_AlreadyInstalled tests AutoInstall when Claude is installed
func TestAutoInstall_AlreadyInstalled(t *testing.T) {
	ci := NewClaudeInstaller()

	// Only test if Claude is installed
	if ci.IsClaudeInstalled() {
		result := ci.AutoInstall()

		if !result.Success {
			t.Error("Expected success when Claude is already installed")
		}

		if result.Message != "Claude CLI is already installed" {
			t.Errorf("Expected 'already installed' message, got: %s", result.Message)
		}
	}
}

func TestClaudeInstallerTimeout(t *testing.T) {
	ci := NewClaudeInstaller()
	ci.Timeout = 1 * time.Second

	if ci.Timeout != 1*time.Second {
		t.Errorf("Expected timeout 1s, got %v", ci.Timeout)
	}
}

func TestClaudeInstallerVerbose(t *testing.T) {
	ci := NewClaudeInstaller()
	ci.Verbose = true

	if !ci.Verbose {
		t.Error("Expected Verbose to be true")
	}
}

// BenchmarkIsClaudeInstalled benchmarks the installation check
func BenchmarkIsClaudeInstalled(b *testing.B) {
	ci := NewClaudeInstaller()

	for i := 0; i < b.N; i++ {
		ci.IsClaudeInstalled()
	}
}

// BenchmarkGetClaudePath benchmarks path lookup (only if installed)
func BenchmarkGetClaudePath(b *testing.B) {
	ci := NewClaudeInstaller()

	if !ci.IsClaudeInstalled() {
		b.Skip("Claude not installed, skipping benchmark")
	}

	for i := 0; i < b.N; i++ {
		ci.GetClaudePath()
	}
}
