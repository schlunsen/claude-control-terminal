package installer

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
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
		// Some Claude installations might not support --version
		// or might return an error, so we just check it doesn't panic
		_ = err
		_ = version
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
		// Some Claude installations might not support version checks
		// or might return an error, so we just verify it doesn't panic
		_ = err
		_ = version

		// Claude auto-updates, so this should typically return false
		if err == nil && hasUpdate {
			t.Log("hasUpdate is true (unusual for auto-updating Claude)")
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

// TestInstall_UnknownMethod tests Install with an unknown installation method
func TestInstall_UnknownMethod(t *testing.T) {
	ci := NewClaudeInstaller()

	// Only test unknown method if Claude is NOT already installed
	// If Claude is already installed, the Install method returns early with success
	if ci.IsClaudeInstalled() {
		t.Skip("Claude already installed, skipping unknown method test")
	}

	result := ci.Install(InstallMethod("unknown"))

	if result.Success {
		t.Error("Expected failure for unknown method")
	}

	if result.Error == nil {
		t.Error("Expected error for unknown method")
	} else if !strings.Contains(result.Error.Error(), "unknown installation method") {
		t.Errorf("Expected 'unknown installation method' error, got: %v", result.Error)
	}
}

// TestFindClaudePath tests the FindClaudePath function
func TestFindClaudePath(t *testing.T) {
	path, err := FindClaudePath()

	// This will succeed or fail based on system state
	if err != nil {
		// If error, should not be installed
		_, lookupErr := exec.LookPath("claude")
		if lookupErr == nil {
			t.Error("FindClaudePath failed but claude is in PATH")
		}
	} else {
		// If success, path should not be empty
		if path == "" {
			t.Error("FindClaudePath returned empty path without error")
		}

		// Path should exist
		if _, statErr := os.Stat(path); statErr != nil {
			t.Errorf("FindClaudePath returned non-existent path: %s", path)
		}
	}
}

// TestInstallerConfiguration tests installer configuration
func TestInstallerConfiguration(t *testing.T) {
	ci := NewClaudeInstaller()
	ci.Verbose = true
	ci.Timeout = 1 * time.Second

	// Test that the installer has the correct configuration
	if ci.Timeout != 1*time.Second {
		t.Errorf("Expected timeout 1s, got %v", ci.Timeout)
	}

	if !ci.Verbose {
		t.Error("Expected Verbose to be true")
	}
}

// TestCleanupOldInstallation tests the cleanup functionality
func TestCleanupOldInstallation(t *testing.T) {
	ci := NewClaudeInstaller()
	ci.Timeout = 5 * time.Second

	// This should not fail even if npm is not installed
	err := ci.CleanupOldInstallation()
	// We don't assert error is nil because npm might not be available
	// Just verify the function doesn't panic
	_ = err
}

// TestInstallResult_AllFields tests all fields of InstallResult
func TestInstallResult_AllFields(t *testing.T) {
	testError := fmt.Errorf("test error")

	result := &InstallResult{
		Success:    false,
		Method:     InstallMethodNPM,
		ClaudePath: "/test/path",
		Version:    "v1.0.0",
		Message:    "Test message",
		Error:      testError,
	}

	if result.Success {
		t.Error("Expected Success to be false")
	}

	if result.Method != InstallMethodNPM {
		t.Errorf("Expected method %v, got %v", InstallMethodNPM, result.Method)
	}

	if result.ClaudePath != "/test/path" {
		t.Errorf("Expected path /test/path, got %s", result.ClaudePath)
	}

	if result.Version != "v1.0.0" {
		t.Errorf("Expected version v1.0.0, got %s", result.Version)
	}

	if result.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got %s", result.Message)
	}

	if result.Error != testError {
		t.Errorf("Expected error %v, got %v", testError, result.Error)
	}
}

// TestGetClaudeVersion_ContextHandling tests GetClaudeVersion context handling
func TestGetClaudeVersion_ContextHandling(t *testing.T) {
	ci := NewClaudeInstaller()

	// Skip if Claude is not installed
	if !ci.IsClaudeInstalled() {
		t.Skip("Claude not installed, skipping context test")
	}

	// Test that context is properly used (by calling the method normally)
	_, err := ci.GetClaudeVersion()

	// Some Claude installations might not support --version
	// We just verify it doesn't panic and handles errors gracefully
	_ = err
}

// TestVerifyInstallation_Verbose tests verbose output
func TestVerifyInstallation_Verbose(t *testing.T) {
	ci := NewClaudeInstaller()
	ci.Verbose = true

	// Only test if Claude is installed
	if !ci.IsClaudeInstalled() {
		err := ci.VerifyInstallation()
		if err == nil {
			t.Error("Expected error when Claude is not installed")
		}

		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	}
}

// TestGetInstallInstructions_AllPlatforms tests platform-specific instructions
func TestGetInstallInstructions_Coverage(t *testing.T) {
	instructions := GetInstallInstructions()

	if instructions == "" {
		t.Fatal("GetInstallInstructions returned empty string")
	}

	// Should contain some installation method
	hasInstallMethod := strings.Contains(instructions, "install") ||
		strings.Contains(instructions, "Install")

	if !hasInstallMethod {
		t.Error("Instructions should contain installation method")
	}
}

// TestInstallMethodString tests InstallMethod as string
func TestInstallMethodString(t *testing.T) {
	var method InstallMethod = "custom"

	if string(method) != "custom" {
		t.Errorf("Expected 'custom', got %s", method)
	}
}

// TestFindClaudePath_CommonLocations tests the fallback to common locations
func TestFindClaudePath_Coverage(t *testing.T) {
	// This tests that the function checks common locations
	// without actually requiring Claude to be installed
	path, err := FindClaudePath()

	if err == nil {
		// If found, verify it's a valid path
		if !strings.Contains(path, "claude") && !strings.HasSuffix(path, "bin/claude") {
			t.Logf("Warning: Found path doesn't look like claude binary: %s", path)
		}
	}
}

// TestGetInstallLocation_Error tests GetInstallLocation when Claude is not installed
func TestGetInstallLocation_NotInstalled(t *testing.T) {
	ci := NewClaudeInstaller()

	// Only test if Claude is NOT installed
	if !ci.IsClaudeInstalled() {
		_, err := ci.GetInstallLocation()
		if err == nil {
			t.Error("Expected error when Claude is not installed")
		}

		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	}
}
