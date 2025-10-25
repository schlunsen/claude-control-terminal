// Package docker provides Docker containerization support for Claude Code environments.
// It handles Docker image building, container management, and provides utilities
// for running Claude Code in containerized environments.
package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DockerManager handles Docker operations for CCT
type DockerManager struct {
	ProjectDir  string
	ImageName   string
	ImageTag    string
	ContainerName string
}

// NewDockerManager creates a new Docker manager instance
func NewDockerManager(projectDir string) *DockerManager {
	return &DockerManager{
		ProjectDir:    projectDir,
		ImageName:     "cct-claude",
		ImageTag:      "latest",
		ContainerName: "cct-claude-container",
	}
}

// IsDockerAvailable checks if Docker is installed and running
func (dm *DockerManager) IsDockerAvailable() bool {
	cmd := exec.Command("docker", "version")
	err := cmd.Run()
	return err == nil
}

// BuildImage builds a Docker image with the current configuration
func (dm *DockerManager) BuildImage(dockerfilePath string) error {
	if !dm.IsDockerAvailable() {
		return fmt.Errorf("docker is not installed or not running")
	}

	fmt.Printf("🐳 Building Docker image: %s:%s\n", dm.ImageName, dm.ImageTag)

	cmd := exec.Command("docker", "build",
		"-t", fmt.Sprintf("%s:%s", dm.ImageName, dm.ImageTag),
		"-f", dockerfilePath,
		dm.ProjectDir,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build Docker image: %w", err)
	}

	fmt.Println("✅ Docker image built successfully!")
	return nil
}

// RunContainer runs a Docker container with the specified configuration
func (dm *DockerManager) RunContainer(opts RunOptions) error {
	if !dm.IsDockerAvailable() {
		return fmt.Errorf("docker is not installed or not running")
	}

	// Stop existing container if it exists
	_ = dm.StopContainer()

	fmt.Printf("🐳 Running Docker container: %s\n", dm.ContainerName)

	args := []string{"run", "-d", "--name", dm.ContainerName}

	// Add port mappings
	for host, container := range opts.Ports {
		args = append(args, "-p", fmt.Sprintf("%d:%d", host, container))
	}

	// Add volume mounts
	for host, container := range opts.Volumes {
		args = append(args, "-v", fmt.Sprintf("%s:%s", host, container))
	}

	// Add environment variables
	for key, value := range opts.Environment {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Add image
	args = append(args, fmt.Sprintf("%s:%s", dm.ImageName, dm.ImageTag))

	// Add command if specified
	if opts.Command != "" {
		args = append(args, strings.Split(opts.Command, " ")...)
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run Docker container: %w", err)
	}

	fmt.Println("✅ Docker container started successfully!")
	fmt.Printf("📦 Container name: %s\n", dm.ContainerName)
	return nil
}

// StopContainer stops and removes the Docker container
func (dm *DockerManager) StopContainer() error {
	if !dm.IsDockerAvailable() {
		return fmt.Errorf("docker is not installed or not running")
	}

	fmt.Printf("🛑 Stopping container: %s\n", dm.ContainerName)

	// Stop container
	stopCmd := exec.Command("docker", "stop", dm.ContainerName)
	_ = stopCmd.Run() // Ignore error if container doesn't exist

	// Remove container
	rmCmd := exec.Command("docker", "rm", dm.ContainerName)
	_ = rmCmd.Run() // Ignore error if container doesn't exist

	return nil
}

// GetContainerLogs retrieves logs from the running container
func (dm *DockerManager) GetContainerLogs(follow bool) error {
	if !dm.IsDockerAvailable() {
		return fmt.Errorf("docker is not installed or not running")
	}

	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, dm.ContainerName)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ExecInContainer executes a command inside the running container
func (dm *DockerManager) ExecInContainer(command string) error {
	if !dm.IsDockerAvailable() {
		return fmt.Errorf("docker is not installed or not running")
	}

	args := []string{"exec", "-it", dm.ContainerName}
	args = append(args, strings.Split(command, " ")...)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// GetDockerfileDir returns the directory for Dockerfile templates
func GetDockerfileDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".claude", "docker")
}

// RunOptions contains configuration for running a Docker container
type RunOptions struct {
	Ports       map[int]int       // Host port to container port mapping
	Volumes     map[string]string // Host path to container path mapping
	Environment map[string]string // Environment variables
	Command     string            // Command to run in container
}

// NewRunOptions creates default run options
func NewRunOptions() RunOptions {
	return RunOptions{
		Ports:       make(map[int]int),
		Volumes:     make(map[string]string),
		Environment: make(map[string]string),
	}
}
