package docker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewDockerfileGenerator(t *testing.T) {
	dg := NewDockerfileGenerator("/test/project")

	if dg == nil {
		t.Fatal("NewDockerfileGenerator returned nil")
	}

	if dg.ProjectDir != "/test/project" {
		t.Errorf("expected project dir '/test/project', got %q", dg.ProjectDir)
	}

	if dg.MCPs == nil {
		t.Error("MCPs should be initialized")
	}

	if len(dg.MCPs) != 0 {
		t.Errorf("expected empty MCPs list, got %d items", len(dg.MCPs))
	}
}

func TestDockerfileType(t *testing.T) {
	tests := []struct {
		name         string
		dockerfileType DockerfileType
		expected     string
	}{
		{"base type", DockerfileBase, "base"},
		{"claude type", DockerfileClaude, "claude"},
		{"analytics type", DockerfileAnalytics, "analytics"},
		{"full type", DockerfileFull, "full"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.dockerfileType) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.dockerfileType))
			}
		})
	}
}

func TestGenerateBaseDockerfile(t *testing.T) {
	dg := NewDockerfileGenerator("/test/project")

	content, err := dg.generateBaseDockerfile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if content == "" {
		t.Error("generated Dockerfile is empty")
	}

	// Check for expected content
	if !strings.Contains(content, "FROM alpine:latest") {
		t.Error("Dockerfile should contain 'FROM alpine:latest'")
	}

	if !strings.Contains(content, "COPY cct") {
		t.Error("Dockerfile should contain 'COPY cct'")
	}

	if !strings.Contains(content, "ENTRYPOINT") {
		t.Error("Dockerfile should contain 'ENTRYPOINT'")
	}
}

func TestGenerateClaudeDockerfile(t *testing.T) {
	dg := NewDockerfileGenerator("/test/project")

	content, err := dg.generateClaudeDockerfile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if content == "" {
		t.Error("generated Dockerfile is empty")
	}

	// Check for expected content
	if !strings.Contains(content, "FROM node:20-alpine") {
		t.Error("Dockerfile should contain 'FROM node:20-alpine'")
	}

	if !strings.Contains(content, "npm install -g @anthropic-ai/claude-code") {
		t.Error("Dockerfile should contain Claude CLI installation")
	}

	if !strings.Contains(content, "EXPOSE 3333") {
		t.Error("Dockerfile should expose port 3333")
	}
}

func TestGenerateClaudeDockerfileWithMCPs(t *testing.T) {
	dg := NewDockerfileGenerator("/test/project")
	dg.MCPs = []string{"github", "filesystem", "postgres"}

	content, err := dg.generateClaudeDockerfile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check for MCP installation
	if !strings.Contains(content, "github,filesystem,postgres") {
		t.Error("Dockerfile should contain MCP installation")
	}
}

func TestGenerateAnalyticsDockerfile(t *testing.T) {
	dg := NewDockerfileGenerator("/test/project")

	content, err := dg.generateAnalyticsDockerfile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if content == "" {
		t.Error("generated Dockerfile is empty")
	}

	// Check for expected content
	if !strings.Contains(content, "FROM alpine:latest") {
		t.Error("Dockerfile should contain 'FROM alpine:latest'")
	}

	if !strings.Contains(content, "--analytics") {
		t.Error("Dockerfile should contain '--analytics'")
	}

	if !strings.Contains(content, "EXPOSE 3333") {
		t.Error("Dockerfile should expose port 3333")
	}

	if !strings.Contains(content, "VOLUME") {
		t.Error("Dockerfile should contain VOLUME directive")
	}
}

func TestGenerateFullDockerfile(t *testing.T) {
	dg := NewDockerfileGenerator("/test/project")

	content, err := dg.generateFullDockerfile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if content == "" {
		t.Error("generated Dockerfile is empty")
	}

	// Check for expected content
	if !strings.Contains(content, "FROM node:20-alpine") {
		t.Error("Dockerfile should contain 'FROM node:20-alpine'")
	}

	if !strings.Contains(content, "docker-cli") {
		t.Error("Dockerfile should include docker-cli")
	}

	if !strings.Contains(content, "typescript") {
		t.Error("Dockerfile should include TypeScript")
	}

	if !strings.Contains(content, "EXPOSE 3333 8080") {
		t.Error("Dockerfile should expose multiple ports")
	}
}

func TestGenerateFullDockerfileWithMCPs(t *testing.T) {
	dg := NewDockerfileGenerator("/test/project")
	dg.MCPs = []string{"github", "supabase"}

	content, err := dg.generateFullDockerfile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check for MCP installation
	if !strings.Contains(content, "github,supabase") {
		t.Error("Dockerfile should contain MCP installation")
	}
}

func TestGenerateDockerfile(t *testing.T) {
	tempDir := t.TempDir()
	dg := NewDockerfileGenerator("/test/project")

	tests := []struct {
		name           string
		dockerfileType DockerfileType
		mcps           []string
		shouldError    bool
	}{
		{
			name:           "base dockerfile",
			dockerfileType: DockerfileBase,
			mcps:           []string{},
			shouldError:    false,
		},
		{
			name:           "claude dockerfile",
			dockerfileType: DockerfileClaude,
			mcps:           []string{"github"},
			shouldError:    false,
		},
		{
			name:           "analytics dockerfile",
			dockerfileType: DockerfileAnalytics,
			mcps:           []string{},
			shouldError:    false,
		},
		{
			name:           "full dockerfile",
			dockerfileType: DockerfileFull,
			mcps:           []string{"github", "postgres"},
			shouldError:    false,
		},
		{
			name:           "unknown type",
			dockerfileType: DockerfileType("unknown"),
			mcps:           []string{},
			shouldError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(tempDir, string(tt.dockerfileType), "Dockerfile")
			err := dg.GenerateDockerfile(tt.dockerfileType, outputPath, tt.mcps)

			if tt.shouldError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.shouldError {
				// Verify file was created
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Error("Dockerfile was not created")
				}

				// Read and verify content
				content, err := os.ReadFile(outputPath)
				if err != nil {
					t.Fatalf("failed to read Dockerfile: %v", err)
				}

				if len(content) == 0 {
					t.Error("Dockerfile is empty")
				}
			}
		})
	}
}

func TestGenerateDockerIgnore(t *testing.T) {
	tempDir := t.TempDir()
	dg := NewDockerfileGenerator("/test/project")

	outputPath := filepath.Join(tempDir, ".dockerignore")
	err := dg.GenerateDockerIgnore(outputPath)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error(".dockerignore was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read .dockerignore: %v", err)
	}

	contentStr := string(content)

	// Check for expected patterns
	expectedPatterns := []string{
		".git",
		"*.md",
		"*_test.go",
		"node_modules/",
		".DS_Store",
		"*.tmp",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(contentStr, pattern) {
			t.Errorf(".dockerignore should contain pattern %q", pattern)
		}
	}
}

func TestDockerfileGeneratorStruct(t *testing.T) {
	dg := DockerfileGenerator{
		ProjectDir: "/custom/project",
		MCPs:       []string{"mcp1", "mcp2"},
	}

	if dg.ProjectDir != "/custom/project" {
		t.Errorf("expected project dir '/custom/project', got %q", dg.ProjectDir)
	}

	if len(dg.MCPs) != 2 {
		t.Errorf("expected 2 MCPs, got %d", len(dg.MCPs))
	}

	if dg.MCPs[0] != "mcp1" || dg.MCPs[1] != "mcp2" {
		t.Error("MCPs not set correctly")
	}
}

func TestGenerateDockerfileInvalidPath(t *testing.T) {
	dg := NewDockerfileGenerator("/test/project")

	// Try to write to invalid path (no permissions)
	err := dg.GenerateDockerfile(DockerfileBase, "/root/impossible/Dockerfile", []string{})

	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestGenerateDockerIgnoreInvalidPath(t *testing.T) {
	dg := NewDockerfileGenerator("/test/project")

	// Try to write to invalid path
	err := dg.GenerateDockerIgnore("/root/impossible/.dockerignore")

	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestDockerfileTypesConstants(t *testing.T) {
	types := []DockerfileType{
		DockerfileBase,
		DockerfileClaude,
		DockerfileAnalytics,
		DockerfileFull,
	}

	// Verify all types are distinct
	seen := make(map[string]bool)
	for _, dt := range types {
		s := string(dt)
		if seen[s] {
			t.Errorf("duplicate Dockerfile type: %q", s)
		}
		seen[s] = true
	}

	if len(types) != 4 {
		t.Errorf("expected 4 Dockerfile types, got %d", len(types))
	}
}

func TestGenerateDockerfileCreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	dg := NewDockerfileGenerator("/test/project")

	// Use nested directory that doesn't exist
	outputPath := filepath.Join(tempDir, "subdir", "nested", "Dockerfile")
	err := dg.GenerateDockerfile(DockerfileBase, outputPath, []string{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify directory was created
	dir := filepath.Dir(outputPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("directory was not created")
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Dockerfile was not created")
	}
}
