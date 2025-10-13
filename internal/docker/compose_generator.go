// Package docker provides docker-compose.yml generation for multi-container setups.
// This file generates compose configurations for simple, analytics, database, and full deployments
// with services like PostgreSQL, Redis, and the CCT analytics dashboard.
package docker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ComposeGenerator generates docker-compose.yml files
type ComposeGenerator struct {
	ProjectDir  string
	Services    []ComposeService
	ProjectName string
}

// ComposeService represents a service in docker-compose
type ComposeService struct {
	Name        string
	Image       string
	Build       string
	Ports       []string
	Volumes     []string
	Environment map[string]string
	DependsOn   []string
	Command     string
}

// NewComposeGenerator creates a new compose generator
func NewComposeGenerator(projectDir string) *ComposeGenerator {
	projectName := filepath.Base(projectDir)
	if projectName == "." || projectName == "/" {
		projectName = "cct-project"
	}

	return &ComposeGenerator{
		ProjectDir:  projectDir,
		Services:    []ComposeService{},
		ProjectName: projectName,
	}
}

// ComposeTemplate represents different compose templates
type ComposeTemplate string

const (
	ComposeSimple    ComposeTemplate = "simple"    // Just CCT + Claude
	ComposeAnalytics ComposeTemplate = "analytics" // CCT + Analytics dashboard
	ComposeFull      ComposeTemplate = "full"      // All services
	ComposeDatabase  ComposeTemplate = "database"  // Claude + PostgreSQL
)

// GenerateCompose generates a docker-compose.yml file
func (cg *ComposeGenerator) GenerateCompose(template ComposeTemplate, outputPath string, mcps []string) error {
	switch template {
	case ComposeSimple:
		cg.generateSimpleCompose(mcps)
	case ComposeAnalytics:
		cg.generateAnalyticsCompose(mcps)
	case ComposeFull:
		cg.generateFullCompose(mcps)
	case ComposeDatabase:
		cg.generateDatabaseCompose(mcps)
	default:
		return fmt.Errorf("unknown compose template: %s", template)
	}

	content, err := cg.renderCompose()
	if err != nil {
		return fmt.Errorf("failed to render compose file: %w", err)
	}

	// Create output directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write compose file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write compose file: %w", err)
	}

	fmt.Printf("✅ Generated docker-compose.yml: %s\n", outputPath)
	return nil
}

// generateSimpleCompose creates a simple compose with just Claude
func (cg *ComposeGenerator) generateSimpleCompose(mcps []string) {
	cg.Services = []ComposeService{
		{
			Name:  "claude",
			Build: ".",
			Ports: []string{"3333:3333"},
			Volumes: []string{
				"./:/workspace",
				"~/.claude:/root/.claude",
			},
			Environment: map[string]string{
				"CLAUDE_API_KEY": "${CLAUDE_API_KEY}",
			},
			Command: "cct",
		},
	}
}

// generateAnalyticsCompose creates compose with analytics dashboard
func (cg *ComposeGenerator) generateAnalyticsCompose(mcps []string) {
	cg.Services = []ComposeService{
		{
			Name:  "claude",
			Build: ".",
			Ports: []string{},
			Volumes: []string{
				"./:/workspace",
				"claude_data:/root/.claude",
			},
			Environment: map[string]string{
				"CLAUDE_API_KEY": "${CLAUDE_API_KEY}",
			},
			Command: "claude",
		},
		{
			Name:  "analytics",
			Build: ".",
			Ports: []string{"3333:3333"},
			Volumes: []string{
				"claude_data:/root/.claude:ro",
			},
			DependsOn: []string{"claude"},
			Command:   "cct --analytics",
		},
	}
}

// generateDatabaseCompose creates compose with PostgreSQL
func (cg *ComposeGenerator) generateDatabaseCompose(mcps []string) {
	cg.Services = []ComposeService{
		{
			Name:  "postgres",
			Image: "postgres:16-alpine",
			Ports: []string{"5432:5432"},
			Environment: map[string]string{
				"POSTGRES_USER":     "claude",
				"POSTGRES_PASSWORD": "claude_password",
				"POSTGRES_DB":       "claude_db",
			},
			Volumes: []string{
				"postgres_data:/var/lib/postgresql/data",
			},
		},
		{
			Name:  "claude",
			Build: ".",
			Ports: []string{"3333:3333"},
			Volumes: []string{
				"./:/workspace",
				"~/.claude:/root/.claude",
			},
			Environment: map[string]string{
				"CLAUDE_API_KEY": "${CLAUDE_API_KEY}",
				"DATABASE_URL":   "postgresql://claude:claude_password@postgres:5432/claude_db",
			},
			DependsOn: []string{"postgres"},
			Command:   "cct",
		},
	}
}

// generateFullCompose creates comprehensive compose with all services
func (cg *ComposeGenerator) generateFullCompose(mcps []string) {
	cg.Services = []ComposeService{
		{
			Name:  "postgres",
			Image: "postgres:16-alpine",
			Ports: []string{"5432:5432"},
			Environment: map[string]string{
				"POSTGRES_USER":     "claude",
				"POSTGRES_PASSWORD": "claude_password",
				"POSTGRES_DB":       "claude_db",
			},
			Volumes: []string{
				"postgres_data:/var/lib/postgresql/data",
			},
		},
		{
			Name:  "redis",
			Image: "redis:7-alpine",
			Ports: []string{"6379:6379"},
			Volumes: []string{
				"redis_data:/data",
			},
		},
		{
			Name:  "claude",
			Build: ".",
			Ports: []string{"8080:8080"},
			Volumes: []string{
				"./:/workspace",
				"claude_data:/root/.claude",
			},
			Environment: map[string]string{
				"CLAUDE_API_KEY": "${CLAUDE_API_KEY}",
				"DATABASE_URL":   "postgresql://claude:claude_password@postgres:5432/claude_db",
				"REDIS_URL":      "redis://redis:6379",
			},
			DependsOn: []string{"postgres", "redis"},
			Command:   "claude",
		},
		{
			Name:  "analytics",
			Build: ".",
			Ports: []string{"3333:3333"},
			Volumes: []string{
				"claude_data:/root/.claude:ro",
			},
			DependsOn: []string{"claude"},
			Command:   "cct --analytics",
		},
	}
}

// renderCompose renders the compose file content
func (cg *ComposeGenerator) renderCompose() (string, error) {
	tmplStr := `version: '3.8'

services:
{{- range .Services }}
  {{ .Name }}:
{{- if .Image }}
    image: {{ .Image }}
{{- else if .Build }}
    build: {{ .Build }}
{{- end }}
{{- if .Ports }}
    ports:
{{- range .Ports }}
      - "{{ . }}"
{{- end }}
{{- end }}
{{- if .Volumes }}
    volumes:
{{- range .Volumes }}
      - {{ . }}
{{- end }}
{{- end }}
{{- if .Environment }}
    environment:
{{- range $key, $value := .Environment }}
      {{ $key }}: {{ $value }}
{{- end }}
{{- end }}
{{- if .DependsOn }}
    depends_on:
{{- range .DependsOn }}
      - {{ . }}
{{- end }}
{{- end }}
{{- if .Command }}
    command: {{ .Command }}
{{- end }}
    restart: unless-stopped
{{ end }}

volumes:
{{- if .HasVolumes }}
{{- range .VolumeNames }}
  {{ . }}:
{{- end }}
{{- else }}
  claude_data:
{{- end }}
`

	tmpl, err := template.New("compose").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	data := map[string]interface{}{
		"Services":    cg.Services,
		"HasVolumes":  cg.hasVolumes(),
		"VolumeNames": cg.extractVolumeNames(),
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// hasVolumes checks if any services use named volumes
func (cg *ComposeGenerator) hasVolumes() bool {
	for _, service := range cg.Services {
		for _, volume := range service.Volumes {
			if !strings.Contains(volume, "/") && !strings.HasPrefix(volume, "~") && !strings.HasPrefix(volume, ".") {
				return true
			}
		}
	}
	return false
}

// extractVolumeNames extracts named volumes from services
func (cg *ComposeGenerator) extractVolumeNames() []string {
	volumes := make(map[string]bool)

	for _, service := range cg.Services {
		for _, volume := range service.Volumes {
			parts := strings.Split(volume, ":")
			if len(parts) > 0 {
				volumeName := parts[0]
				// Check if it's a named volume (not a path)
				if !strings.Contains(volumeName, "/") && !strings.HasPrefix(volumeName, "~") && !strings.HasPrefix(volumeName, ".") {
					volumes[volumeName] = true
				}
			}
		}
	}

	result := make([]string, 0, len(volumes))
	for volume := range volumes {
		result = append(result, volume)
	}

	return result
}

// GenerateEnvFile generates a .env.example file for docker-compose
func (cg *ComposeGenerator) GenerateEnvFile(outputPath string) error {
	content := `# Claude API Configuration
CLAUDE_API_KEY=your_claude_api_key_here

# Database Configuration (if using PostgreSQL)
POSTGRES_USER=claude
POSTGRES_PASSWORD=claude_password
POSTGRES_DB=claude_db
DATABASE_URL=postgresql://claude:claude_password@postgres:5432/claude_db

# Redis Configuration (if using Redis)
REDIS_URL=redis://redis:6379

# CCT Configuration
CCT_PORT=3333
CCT_LOG_LEVEL=info
`

	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write .env.example: %w", err)
	}

	fmt.Printf("✅ Generated .env.example: %s\n", outputPath)
	return nil
}
