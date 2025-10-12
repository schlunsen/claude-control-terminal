package docker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// DockerfileGenerator generates Dockerfiles for different use cases
type DockerfileGenerator struct {
	ProjectDir string
	MCPs       []string
}

// NewDockerfileGenerator creates a new Dockerfile generator
func NewDockerfileGenerator(projectDir string) *DockerfileGenerator {
	return &DockerfileGenerator{
		ProjectDir: projectDir,
		MCPs:       []string{},
	}
}

// DockerfileType represents different Dockerfile templates
type DockerfileType string

const (
	DockerfileBase      DockerfileType = "base"
	DockerfileClaude    DockerfileType = "claude"
	DockerfileAnalytics DockerfileType = "analytics"
	DockerfileFull      DockerfileType = "full"
)

// GenerateDockerfile generates a Dockerfile based on the specified type
func (dg *DockerfileGenerator) GenerateDockerfile(dockerfileType DockerfileType, outputPath string, mcps []string) error {
	dg.MCPs = mcps

	var content string
	var err error

	switch dockerfileType {
	case DockerfileBase:
		content, err = dg.generateBaseDockerfile()
	case DockerfileClaude:
		content, err = dg.generateClaudeDockerfile()
	case DockerfileAnalytics:
		content, err = dg.generateAnalyticsDockerfile()
	case DockerfileFull:
		content, err = dg.generateFullDockerfile()
	default:
		return fmt.Errorf("unknown Dockerfile type: %s", dockerfileType)
	}

	if err != nil {
		return fmt.Errorf("failed to generate Dockerfile: %w", err)
	}

	// Create output directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write Dockerfile
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	fmt.Printf("✅ Generated Dockerfile: %s\n", outputPath)
	return nil
}

// generateBaseDockerfile creates a minimal Dockerfile with just CCT
func (dg *DockerfileGenerator) generateBaseDockerfile() (string, error) {
	tmpl := `# Claude Control Terminal - Base Image
FROM alpine:latest

LABEL maintainer="CCT"
LABEL description="Base Claude Control Terminal image"

# Install dependencies
RUN apk add --no-cache \
    ca-certificates \
    git \
    curl

# Copy CCT binary
COPY cct /usr/local/bin/cct
RUN chmod +x /usr/local/bin/cct

WORKDIR /workspace

# Set up .claude directory
RUN mkdir -p /root/.claude

ENTRYPOINT ["cct"]
CMD ["--help"]
`
	return tmpl, nil
}

// generateClaudeDockerfile creates a Dockerfile with CCT + Claude CLI + MCPs
func (dg *DockerfileGenerator) generateClaudeDockerfile() (string, error) {
	mcpInstalls := ""
	if len(dg.MCPs) > 0 {
		mcpInstalls = fmt.Sprintf("RUN cct --mcp \"%s\" --directory /root\n", strings.Join(dg.MCPs, ","))
	}

	tmplStr := `# Claude Control Terminal - Full Claude Environment
FROM node:20-alpine

LABEL maintainer="CCT"
LABEL description="Claude Code environment with CCT, Claude CLI, and MCPs"

# Install system dependencies
RUN apk add --no-cache \
    ca-certificates \
    git \
    curl \
    bash \
    python3 \
    py3-pip

# Install Claude CLI
RUN npm install -g @anthropic-ai/claude-code

# Copy CCT binary
COPY cct /usr/local/bin/cct
RUN chmod +x /usr/local/bin/cct

WORKDIR /workspace

# Set up .claude directory
RUN mkdir -p /root/.claude

# Install MCPs if specified
{{ .MCPInstalls }}

# Expose default port for analytics
EXPOSE 3333

ENTRYPOINT ["claude"]
CMD ["--help"]
`

	data := map[string]string{
		"MCPInstalls": mcpInstalls,
	}

	tmpl, err := template.New("claude").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// generateAnalyticsDockerfile creates a Dockerfile optimized for analytics dashboard
func (dg *DockerfileGenerator) generateAnalyticsDockerfile() (string, error) {
	tmpl := `# Claude Control Terminal - Analytics Dashboard
FROM alpine:latest

LABEL maintainer="CCT"
LABEL description="CCT Analytics Dashboard"

# Install dependencies
RUN apk add --no-cache \
    ca-certificates \
    curl

# Copy CCT binary
COPY cct /usr/local/bin/cct
RUN chmod +x /usr/local/bin/cct

WORKDIR /workspace

# Set up .claude directory for conversation monitoring
RUN mkdir -p /root/.claude

# Expose analytics port
EXPOSE 3333

# Volume for conversation data
VOLUME ["/root/.claude"]

ENTRYPOINT ["cct"]
CMD ["--analytics"]
`
	return tmpl, nil
}

// generateFullDockerfile creates a comprehensive Dockerfile with all features
func (dg *DockerfileGenerator) generateFullDockerfile() (string, error) {
	mcpInstalls := ""
	if len(dg.MCPs) > 0 {
		mcpInstalls = fmt.Sprintf("RUN cct --mcp \"%s\" --directory /root\n", strings.Join(dg.MCPs, ","))
	}

	tmplStr := `# Claude Control Terminal - Full Featured Image
FROM node:20-alpine

LABEL maintainer="CCT"
LABEL description="Complete Claude Control Terminal environment"

# Install system dependencies
RUN apk add --no-cache \
    ca-certificates \
    git \
    curl \
    bash \
    python3 \
    py3-pip \
    docker-cli \
    make \
    gcc \
    g++ \
    libc-dev

# Install Claude CLI
RUN npm install -g @anthropic-ai/claude-code

# Copy CCT binary
COPY cct /usr/local/bin/cct
RUN chmod +x /usr/local/bin/cct

WORKDIR /workspace

# Set up .claude directory
RUN mkdir -p /root/.claude

# Install MCPs if specified
{{ .MCPInstalls }}

# Install common development tools
RUN npm install -g \
    typescript \
    ts-node \
    prettier \
    eslint

# Expose ports
EXPOSE 3333 8080

# Volume mounts
VOLUME ["/workspace", "/root/.claude"]

ENTRYPOINT ["cct"]
CMD ["--help"]
`

	data := map[string]string{
		"MCPInstalls": mcpInstalls,
	}

	tmpl, err := template.New("full").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GenerateDockerIgnore creates a .dockerignore file
func (dg *DockerfileGenerator) GenerateDockerIgnore(outputPath string) error {
	content := `# Git
.git
.gitignore
.gitattributes

# CI/CD
.github
.gitlab-ci.yml

# Documentation
*.md
docs/

# Tests
*_test.go
test/
tests/
TEST*.sh

# Build artifacts
dist/
*.exe
*.dll
*.so
*.dylib

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Node
node_modules/
npm-debug.log

# Go
vendor/
*.out
coverage.*

# Temporary files
tmp/
temp/
*.tmp
`

	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write .dockerignore: %w", err)
	}

	fmt.Printf("✅ Generated .dockerignore: %s\n", outputPath)
	return nil
}
