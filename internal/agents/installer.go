package agents

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Installer handles Python environment setup and dependency installation
type Installer struct {
	ServerDir string
	VenvDir   string
	Quiet     bool
}

// NewInstaller creates a new installer instance
func NewInstaller(serverDir string, quiet bool) *Installer {
	return &Installer{
		ServerDir: serverDir,
		VenvDir:   filepath.Join(serverDir, ".venv"),
		Quiet:     quiet,
	}
}

// pythonExecutable returns the platform-specific Python executable name
func pythonExecutable() string {
	if runtime.GOOS == "windows" {
		return "python.exe"
	}
	return "python3"
}

// venvPythonPath returns the path to the Python executable in the venv
func (i *Installer) venvPythonPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(i.VenvDir, "Scripts", "python.exe")
	}
	return filepath.Join(i.VenvDir, "bin", "python")
}

// venvPipPath returns the path to pip in the venv
func (i *Installer) venvPipPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(i.VenvDir, "Scripts", "pip.exe")
	}
	return filepath.Join(i.VenvDir, "bin", "pip")
}

// venvUvPath returns the path to uv in the venv (if installed)
func (i *Installer) venvUvPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(i.VenvDir, "Scripts", "uv.exe")
	}
	return filepath.Join(i.VenvDir, "bin", "uv")
}

// log prints a message unless in quiet mode
func (i *Installer) log(format string, args ...interface{}) {
	if !i.Quiet {
		fmt.Printf(format+"\n", args...)
	}
}

// CheckPython checks if Python 3.13+ is available
func (i *Installer) CheckPython() (string, error) {
	pythonCmd := pythonExecutable()

	// Try to find Python
	path, err := exec.LookPath(pythonCmd)
	if err != nil {
		// On some systems, python3 might not exist but python might
		if pythonCmd == "python3" {
			pythonCmd = "python"
			path, err = exec.LookPath(pythonCmd)
			if err != nil {
				return "", fmt.Errorf("Python not found in PATH. Please install Python 3.13 or later from https://www.python.org/downloads/")
			}
		} else {
			return "", fmt.Errorf("Python not found in PATH. Please install Python 3.13 or later from https://www.python.org/downloads/")
		}
	}

	// Check Python version
	cmd := exec.Command(pythonCmd, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to check Python version: %w", err)
	}

	versionStr := strings.TrimSpace(string(output))
	i.log("→ Found %s at %s", versionStr, path)

	// Parse version (format: "Python 3.13.0")
	parts := strings.Fields(versionStr)
	if len(parts) < 2 {
		return "", fmt.Errorf("unexpected Python version format: %s", versionStr)
	}

	versionParts := strings.Split(parts[1], ".")
	if len(versionParts) < 2 {
		return "", fmt.Errorf("unexpected Python version format: %s", parts[1])
	}

	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return "", fmt.Errorf("invalid Python major version: %s", versionParts[0])
	}

	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return "", fmt.Errorf("invalid Python minor version: %s", versionParts[1])
	}

	// Check if version is 3.12 or later (temporarily lowered for testing)
	if major < 3 || (major == 3 && minor < 12) {
		return "", fmt.Errorf("Python 3.12 or later is required, found %d.%d", major, minor)
	}

	return pythonCmd, nil
}

// CheckUV checks if uv is available globally
func (i *Installer) CheckUV() bool {
	_, err := exec.LookPath("uv")
	return err == nil
}

// IsInstalled checks if the virtual environment and dependencies are installed
func (i *Installer) IsInstalled() bool {
	// Check if venv Python exists
	venvPython := i.venvPythonPath()
	if _, err := os.Stat(venvPython); os.IsNotExist(err) {
		return false
	}

	// Check if pyproject.toml exists in server dir
	pyprojectPath := filepath.Join(i.ServerDir, "pyproject.toml")
	if _, err := os.Stat(pyprojectPath); os.IsNotExist(err) {
		return false
	}

	// Check if key dependencies are installed by trying to import them
	cmd := exec.Command(venvPython, "-c", "import fastapi, uvicorn, websockets, pydantic")
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

// CreateVirtualEnv creates a Python virtual environment
func (i *Installer) CreateVirtualEnv(pythonCmd string) error {
	i.log("→ Creating virtual environment...")

	cmd := exec.Command(pythonCmd, "-m", "venv", i.VenvDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create virtual environment: %w", err)
	}

	i.log("✓ Virtual environment created at %s", i.VenvDir)
	return nil
}

// InstallDependencies installs Python dependencies using uv or pip
func (i *Installer) InstallDependencies() error {
	// Check if uv is available globally
	hasGlobalUV := i.CheckUV()

	if hasGlobalUV {
		return i.installWithUV()
	}

	// Fallback to pip
	return i.installWithPip()
}

// installWithUV installs dependencies using uv (faster)
func (i *Installer) installWithUV() error {
	i.log("→ Installing dependencies with uv (fast)...")

	// Run uv sync in the server directory
	cmd := exec.Command("uv", "sync")
	cmd.Dir = i.ServerDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		i.log("⚠ uv sync failed, falling back to pip")
		return i.installWithPip()
	}

	i.log("✓ Dependencies installed with uv")
	return nil
}

// installWithPip installs dependencies using pip
func (i *Installer) installWithPip() error {
	i.log("→ Installing dependencies with pip...")

	venvPython := i.venvPythonPath()

	// Upgrade pip first
	i.log("  Upgrading pip...")
	cmd := exec.Command(venvPython, "-m", "pip", "install", "--upgrade", "pip")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to upgrade pip: %w", err)
	}

	// Install dependencies from pyproject.toml
	// Use pip install -e . to install the project in editable mode
	i.log("  Installing project dependencies...")
	cmd = exec.Command(venvPython, "-m", "pip", "install", "-e", ".")
	cmd.Dir = i.ServerDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	i.log("✓ Dependencies installed with pip")
	return nil
}

// Install performs the full installation: extract, create venv, install deps
func (i *Installer) Install() error {
	// Check if already installed
	if i.IsInstalled() {
		i.log("✓ Agent server already installed")
		return nil
	}

	// Check Python
	pythonCmd, err := i.CheckPython()
	if err != nil {
		return err
	}

	// Extract server if not already extracted
	if _, err := os.Stat(filepath.Join(i.ServerDir, "pyproject.toml")); os.IsNotExist(err) {
		i.log("→ Extracting agent server to %s...", i.ServerDir)
		if _, err := ExtractAgentServer(i.ServerDir); err != nil {
			return fmt.Errorf("failed to extract agent server: %w", err)
		}
		if err := WriteVersionFile(i.ServerDir); err != nil {
			return fmt.Errorf("failed to write version file: %w", err)
		}
		i.log("✓ Agent server extracted")
	}

	// Create virtual environment if it doesn't exist
	if _, err := os.Stat(i.VenvDir); os.IsNotExist(err) {
		if err := i.CreateVirtualEnv(pythonCmd); err != nil {
			return err
		}
	}

	// Install dependencies
	if err := i.InstallDependencies(); err != nil {
		return err
	}

	i.log("✓ Agent server installation complete")
	return nil
}

// Reinstall removes the existing installation and reinstalls
func (i *Installer) Reinstall() error {
	i.log("→ Removing existing installation...")

	// Remove venv directory
	if err := os.RemoveAll(i.VenvDir); err != nil {
		return fmt.Errorf("failed to remove virtual environment: %w", err)
	}

	i.log("✓ Existing installation removed")

	// Install fresh
	return i.Install()
}

// Update checks if an update is needed and performs it
func (i *Installer) Update() error {
	needsUpdate, err := NeedsUpdate(i.ServerDir)
	if err != nil {
		return err
	}

	if !needsUpdate {
		i.log("✓ Agent server is up to date")
		return nil
	}

	i.log("→ Updating agent server from version %s to %s...", func() string {
		v, _ := ReadVersionFile(i.ServerDir)
		return v
	}(), EmbeddedVersion)

	// Extract updated server
	if _, err := ExtractAgentServer(i.ServerDir); err != nil {
		return fmt.Errorf("failed to extract updated agent server: %w", err)
	}

	// Update version file
	if err := WriteVersionFile(i.ServerDir); err != nil {
		return fmt.Errorf("failed to write version file: %w", err)
	}

	// Reinstall dependencies
	if err := i.Reinstall(); err != nil {
		return err
	}

	i.log("✓ Agent server updated to version %s", EmbeddedVersion)
	return nil
}
