package docker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDockerManager(t *testing.T) {
	dm := NewDockerManager("/test/project")

	if dm == nil {
		t.Fatal("NewDockerManager returned nil")
	}

	if dm.ProjectDir != "/test/project" {
		t.Errorf("expected project dir '/test/project', got %q", dm.ProjectDir)
	}

	if dm.ImageName != "cct-claude" {
		t.Errorf("expected image name 'cct-claude', got %q", dm.ImageName)
	}

	if dm.ImageTag != "latest" {
		t.Errorf("expected image tag 'latest', got %q", dm.ImageTag)
	}

	if dm.ContainerName != "cct-claude-container" {
		t.Errorf("expected container name 'cct-claude-container', got %q", dm.ContainerName)
	}
}

func TestIsDockerAvailable(t *testing.T) {
	dm := NewDockerManager("/test/project")

	// Just test that it doesn't panic
	// Result depends on whether Docker is installed
	available := dm.IsDockerAvailable()

	if available {
		t.Log("Docker is available")
	} else {
		t.Log("Docker is not available")
	}
}

func TestDockerManagerStruct(t *testing.T) {
	dm := DockerManager{
		ProjectDir:    "/test/dir",
		ImageName:     "test-image",
		ImageTag:      "v1.0",
		ContainerName: "test-container",
	}

	if dm.ProjectDir != "/test/dir" {
		t.Errorf("expected project dir '/test/dir', got %q", dm.ProjectDir)
	}

	if dm.ImageName != "test-image" {
		t.Errorf("expected image name 'test-image', got %q", dm.ImageName)
	}

	if dm.ImageTag != "v1.0" {
		t.Errorf("expected image tag 'v1.0', got %q", dm.ImageTag)
	}

	if dm.ContainerName != "test-container" {
		t.Errorf("expected container name 'test-container', got %q", dm.ContainerName)
	}
}

func TestGetDockerfileDir(t *testing.T) {
	dir := GetDockerfileDir()

	if dir == "" {
		t.Error("GetDockerfileDir returned empty string")
	}

	if !filepath.IsAbs(dir) {
		t.Errorf("expected absolute path, got %q", dir)
	}

	if !contains(dir, ".claude") {
		t.Errorf("expected path to contain '.claude', got %q", dir)
	}

	if !contains(dir, "docker") {
		t.Errorf("expected path to contain 'docker', got %q", dir)
	}
}

func TestRunOptions(t *testing.T) {
	opts := NewRunOptions()

	if opts.Ports == nil {
		t.Error("Ports should be initialized")
	}

	if opts.Volumes == nil {
		t.Error("Volumes should be initialized")
	}

	if opts.Environment == nil {
		t.Error("Environment should be initialized")
	}

	// Test adding configurations
	opts.Ports[8080] = 80
	opts.Volumes["/host/path"] = "/container/path"
	opts.Environment["TEST_VAR"] = "test_value"
	opts.Command = "echo hello"

	if opts.Ports[8080] != 80 {
		t.Error("Port mapping was not added correctly")
	}

	if opts.Volumes["/host/path"] != "/container/path" {
		t.Error("Volume mapping was not added correctly")
	}

	if opts.Environment["TEST_VAR"] != "test_value" {
		t.Error("Environment variable was not added correctly")
	}

	if opts.Command != "echo hello" {
		t.Errorf("expected command 'echo hello', got %q", opts.Command)
	}
}

func TestRunOptionsStruct(t *testing.T) {
	opts := RunOptions{
		Ports: map[int]int{
			8080: 80,
			3000: 3000,
		},
		Volumes: map[string]string{
			"/host": "/container",
		},
		Environment: map[string]string{
			"KEY": "value",
		},
		Command: "test command",
	}

	if len(opts.Ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(opts.Ports))
	}

	if len(opts.Volumes) != 1 {
		t.Errorf("expected 1 volume, got %d", len(opts.Volumes))
	}

	if len(opts.Environment) != 1 {
		t.Errorf("expected 1 env var, got %d", len(opts.Environment))
	}

	if opts.Command != "test command" {
		t.Errorf("expected command 'test command', got %q", opts.Command)
	}
}

func TestBuildImageWithoutDocker(t *testing.T) {
	dm := NewDockerManager("/test/project")

	// If Docker is not available, this should return an error
	// If Docker is available, it will try to build (and likely fail with missing Dockerfile)
	err := dm.BuildImage("/nonexistent/Dockerfile")

	if err == nil {
		t.Log("BuildImage succeeded (Docker is available and build worked)")
	} else {
		t.Logf("BuildImage failed as expected: %v", err)
	}
}

func TestStopContainerWithoutDocker(t *testing.T) {
	dm := NewDockerManager("/test/project")

	// This should handle the case where Docker isn't available
	// or container doesn't exist gracefully
	err := dm.StopContainer()

	// Error might be nil if Docker is not available (check is first)
	// or if container doesn't exist (errors are ignored)
	_ = err
}

func TestNewDockerManagerWithDifferentPaths(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"absolute path", "/absolute/path"},
		{"relative path", "./relative/path"},
		{"current dir", "."},
		{"empty path", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm := NewDockerManager(tt.path)
			if dm == nil {
				t.Error("NewDockerManager returned nil")
			}
			if dm.ProjectDir != tt.path {
				t.Errorf("expected project dir %q, got %q", tt.path, dm.ProjectDir)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestDockerManagerCustomization(t *testing.T) {
	dm := NewDockerManager("/test/project")

	// Test customizing the manager
	dm.ImageName = "custom-image"
	dm.ImageTag = "v2.0"
	dm.ContainerName = "custom-container"

	if dm.ImageName != "custom-image" {
		t.Errorf("expected image name 'custom-image', got %q", dm.ImageName)
	}

	if dm.ImageTag != "v2.0" {
		t.Errorf("expected image tag 'v2.0', got %q", dm.ImageTag)
	}

	if dm.ContainerName != "custom-container" {
		t.Errorf("expected container name 'custom-container', got %q", dm.ContainerName)
	}
}

func TestGetContainerLogsWithoutDocker(t *testing.T) {
	dm := NewDockerManager("/test/project")

	// Test getting logs (will fail if Docker not available or container doesn't exist)
	err := dm.GetContainerLogs(false)

	if err != nil {
		t.Logf("GetContainerLogs failed as expected: %v", err)
	}
}

func TestExecInContainerWithoutDocker(t *testing.T) {
	dm := NewDockerManager("/test/project")

	// Test exec (will fail if Docker not available or container doesn't exist)
	err := dm.ExecInContainer("echo test")

	if err != nil {
		t.Logf("ExecInContainer failed as expected: %v", err)
	}
}

func TestGetDockerfileDirExists(t *testing.T) {
	// Test with current HOME
	_ = GetDockerfileDir()

	// Test with temporary HOME
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	newDir := GetDockerfileDir()
	if newDir == "" {
		t.Error("GetDockerfileDir returned empty string")
	}
}
