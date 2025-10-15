package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/schlunsen/claude-control-terminal/internal/components"
	"github.com/schlunsen/claude-control-terminal/internal/docker"
	"github.com/schlunsen/claude-control-terminal/internal/server"
	"github.com/schlunsen/claude-control-terminal/internal/tui"
	"github.com/schlunsen/claude-control-terminal/internal/version"
	"github.com/spf13/cobra"
)

// Use version constants from version package
var (
	Version = version.Version
	Name    = version.Name
)

var (
	// Global flags
	verbose   bool
	directory string
	yesFlag   bool
	dryRun    bool
	preview   bool

	// Component flags
	agent    string
	command  string
	mcp      string
	setting  string
	hook     string
	workflow string
	scope    string // MCP installation scope: "project" or "user"

	// Service flags
	analytics    bool
	chats        bool
	agents       bool
	chatsMobile  bool
	plugins      bool
	tunnel       bool
	healthCheck  bool
	commandStats bool
	hookStats    bool
	mcpStats     bool

	// Agent management flags
	createAgent string
	listAgents  bool
	removeAgent string
	updateAgent string

	// Docker flags
	dockerCommand string
	dockerInit    bool
	dockerBuild   bool
	dockerRun     bool
	dockerStop    bool
	dockerLogs    bool
	dockerCompose bool
	dockerType    string
	dockerMCPs    string
	dockerPorts   string
	dockerVolumes string

	// Hook management flags
	installUserPromptHook     bool
	uninstallUserPromptHook   bool
	installToolHook           bool
	uninstallToolHook         bool
	installNotificationHook   bool
	uninstallNotificationHook bool
	installAllHooks           bool
	uninstallAllHooks         bool

	// Other flags
	template   string
	language   string
	framework  string
	prompt     string
	studio     bool
	sandbox    string
	e2bAPIKey  string
	anthropicAPIKey string

	// Claude installer flag
	installClaude bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "cct",
	Short: "Claude Control Terminal - Control center for Claude Code",
	Long: `Claude Control Terminal (CCT) - A powerful wrapper and control center for Claude Code.

🎮 Manage components, launch Claude, run analytics, and deploy with Docker
🚀 Component installer: 600+ agents, 200+ commands, MCPs
📊 Real-time analytics dashboard with WebSocket monitoring
🐳 Docker support for containerized Claude environments
🌐 Templates: https://aitmpl.com
📖 Documentation: https://docs.aitmpl.com`,
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle Claude installation first
		if installClaude {
			if err := tui.InstallClaude(); err != nil {
				ShowError(fmt.Sprintf("Installation failed: %v", err))
				os.Exit(1)
			}
			return
		}

		// Show banner for interactive mode
		isInteractive := !analytics && !chats && !agents && !chatsMobile && !plugins &&
			!healthCheck && !commandStats && !hookStats && !mcpStats &&
			!listAgents && createAgent == "" && removeAgent == "" && updateAgent == "" &&
			agent == "" && command == "" && mcp == "" && setting == "" && hook == "" &&
			workflow == "" && !studio && sandbox == "" &&
			!dockerInit && !dockerBuild && !dockerRun && !dockerStop && !dockerLogs && !dockerCompose &&
			!installClaude &&
			!installUserPromptHook && !uninstallUserPromptHook &&
			!installToolHook && !uninstallToolHook &&
			!installNotificationHook && !uninstallNotificationHook &&
			!installAllHooks && !uninstallAllHooks

		// If no flags provided, launch TUI
		if isInteractive {
			if err := tui.Launch(directory); err != nil {
				ShowError(fmt.Sprintf("Failed to launch TUI: %v", err))
				os.Exit(1)
			}
			return
		}

		// Handle different modes
		handleCommand(cmd, args)
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	rootCmd.PersistentFlags().StringVarP(&directory, "directory", "d", ".", "target directory")
	rootCmd.PersistentFlags().BoolVarP(&yesFlag, "yes", "y", false, "skip prompts and use defaults")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be copied without copying")
	rootCmd.Flags().BoolVarP(&preview, "preview", "p", false, "preview component content without installing")

	// Template selection flags
	rootCmd.Flags().StringVarP(&template, "template", "t", "", "specify template (e.g., common, javascript-typescript, python)")
	rootCmd.Flags().StringVarP(&language, "language", "l", "", "specify programming language (deprecated, use --template)")
	rootCmd.Flags().StringVarP(&framework, "framework", "f", "", "specify framework")

	// Component installation flags
	rootCmd.Flags().StringVar(&agent, "agent", "", "install specific agent component")
	rootCmd.Flags().StringVar(&command, "command", "", "install specific command component")
	rootCmd.Flags().StringVar(&mcp, "mcp", "", "install specific MCP component")
	rootCmd.Flags().StringVar(&setting, "setting", "", "install specific setting component")
	rootCmd.Flags().StringVar(&hook, "hook", "", "install specific hook component")
	rootCmd.Flags().StringVar(&workflow, "workflow", "", "install workflow from hash or YAML")
	rootCmd.Flags().StringVar(&scope, "scope", "project", "MCP installation scope: 'project' (default) or 'user' (global)")

	// Service flags
	rootCmd.Flags().BoolVar(&analytics, "analytics", false, "launch analytics dashboard")
	rootCmd.Flags().BoolVar(&chats, "chats", false, "launch mobile-first chats interface")
	rootCmd.Flags().BoolVar(&agents, "agents", false, "launch agents dashboard")
	rootCmd.Flags().BoolVar(&chatsMobile, "chats-mobile", false, "launch mobile chats interface")
	rootCmd.Flags().BoolVar(&plugins, "plugins", false, "launch plugin dashboard")
	rootCmd.Flags().BoolVar(&tunnel, "tunnel", false, "enable Cloudflare Tunnel for remote access")

	// Analysis flags
	rootCmd.Flags().BoolVar(&healthCheck, "health-check", false, "run health check")
	rootCmd.Flags().BoolVar(&commandStats, "command-stats", false, "analyze commands")
	rootCmd.Flags().BoolVar(&hookStats, "hook-stats", false, "analyze hooks")
	rootCmd.Flags().BoolVar(&mcpStats, "mcp-stats", false, "analyze MCPs")

	// Agent management flags
	rootCmd.Flags().StringVar(&createAgent, "create-agent", "", "create a global agent")
	rootCmd.Flags().BoolVar(&listAgents, "list-agents", false, "list all installed global agents")
	rootCmd.Flags().StringVar(&removeAgent, "remove-agent", "", "remove a global agent")
	rootCmd.Flags().StringVar(&updateAgent, "update-agent", "", "update a global agent")

	// Docker flags
	rootCmd.Flags().BoolVar(&dockerInit, "docker-init", false, "initialize Docker files (Dockerfile, .dockerignore)")
	rootCmd.Flags().BoolVar(&dockerBuild, "docker-build", false, "build Docker image")
	rootCmd.Flags().BoolVar(&dockerRun, "docker-run", false, "run Docker container")
	rootCmd.Flags().BoolVar(&dockerStop, "docker-stop", false, "stop Docker container")
	rootCmd.Flags().BoolVar(&dockerLogs, "docker-logs", false, "view Docker container logs")
	rootCmd.Flags().BoolVar(&dockerCompose, "docker-compose", false, "generate docker-compose.yml")
	rootCmd.Flags().StringVar(&dockerType, "docker-type", "claude", "Docker type: base, claude, analytics, full")
	rootCmd.Flags().StringVar(&dockerMCPs, "docker-mcps", "", "MCPs to include (comma-separated)")
	rootCmd.Flags().StringVar(&dockerCommand, "docker-command", "", "command to run in container")

	// Other flags
	rootCmd.Flags().StringVar(&prompt, "prompt", "", "execute prompt after installation")
	rootCmd.Flags().BoolVar(&studio, "studio", false, "launch Claude Code Studio")
	rootCmd.Flags().StringVar(&sandbox, "sandbox", "", "execute in sandbox (e.g., e2b)")
	rootCmd.Flags().StringVar(&e2bAPIKey, "e2b-api-key", "", "E2B API key")
	rootCmd.Flags().StringVar(&anthropicAPIKey, "anthropic-api-key", "", "Anthropic API key")

	// Hook management flags
	rootCmd.Flags().BoolVar(&installUserPromptHook, "install-user-prompt-hook", false, "install user prompt logger hook (project-only)")
	rootCmd.Flags().BoolVar(&uninstallUserPromptHook, "uninstall-user-prompt-hook", false, "uninstall user prompt logger hook")
	rootCmd.Flags().BoolVar(&installToolHook, "install-tool-hook", false, "install tool logger hook (project-only)")
	rootCmd.Flags().BoolVar(&uninstallToolHook, "uninstall-tool-hook", false, "uninstall tool logger hook")
	rootCmd.Flags().BoolVar(&installNotificationHook, "install-notification-hook", false, "install notification logger hook (project-only)")
	rootCmd.Flags().BoolVar(&uninstallNotificationHook, "uninstall-notification-hook", false, "uninstall notification logger hook")
	rootCmd.Flags().BoolVar(&installAllHooks, "install-all-hooks", false, "install all hooks (user-prompt + tool + notification loggers, project-only)")
	rootCmd.Flags().BoolVar(&uninstallAllHooks, "uninstall-all-hooks", false, "uninstall all hooks")

	// Claude installer flag
	rootCmd.Flags().BoolVar(&installClaude, "install-claude", false, "install Claude CLI automatically")
}

func handleCommand(cmd *cobra.Command, args []string) {
	// Hook management commands
	if installUserPromptHook || uninstallUserPromptHook || installToolHook || uninstallToolHook || installAllHooks || uninstallAllHooks {
		handleHookManagement()
		return
	}

	// Docker commands
	if dockerInit || dockerBuild || dockerRun || dockerStop || dockerLogs || dockerCompose {
		handleDockerCommands(directory)
		return
	}

	// Analytics dashboard
	if analytics {
		spinner := ShowSpinner("Launching Analytics Dashboard...")

		// Import server package
		server := createAnalyticsServer(directory)

		spinner.Success("Analytics Dashboard starting!")
		ShowInfo("Press Ctrl+C to stop")

		if err := server.Setup(); err != nil {
			ShowError(fmt.Sprintf("Failed to setup server: %v", err))
			return
		}

		// Server prints its own startup messages with correct protocol and ports
		if err := server.Start(); err != nil {
			ShowError(fmt.Sprintf("Failed to start server: %v", err))
		}
		return
	}

	// Chats interface
	if chats || chatsMobile {
		fmt.Println("💬 Launching Chats Interface...")
		fmt.Println("(Implementation coming soon)")
		return
	}

	// Agents dashboard
	if agents {
		fmt.Println("🤖 Launching Agents Dashboard...")
		fmt.Println("(Implementation coming soon)")
		return
	}

	// Health check
	if healthCheck {
		fmt.Println("🔍 Running Health Check...")
		fmt.Println("(Implementation coming soon)")
		return
	}

	// Stats analysis
	if commandStats {
		fmt.Println("📊 Analyzing Commands...")
		fmt.Println("(Implementation coming soon)")
		return
	}

	if hookStats {
		fmt.Println("🔧 Analyzing Hooks...")
		fmt.Println("(Implementation coming soon)")
		return
	}

	if mcpStats {
		fmt.Println("🔌 Analyzing MCPs...")
		fmt.Println("(Implementation coming soon)")
		return
	}

	// Agent management
	if listAgents {
		fmt.Println("📋 Listing Global Agents...")
		fmt.Println("(Implementation coming soon)")
		return
	}

	if createAgent != "" {
		fmt.Printf("🤖 Creating Global Agent: %s\n", createAgent)
		fmt.Println("(Implementation coming soon)")
		return
	}

	if removeAgent != "" {
		fmt.Printf("🗑️  Removing Global Agent: %s\n", removeAgent)
		fmt.Println("(Implementation coming soon)")
		return
	}

	if updateAgent != "" {
		fmt.Printf("🔄 Updating Global Agent: %s\n", updateAgent)
		fmt.Println("(Implementation coming soon)")
		return
	}

	// Component installation
	if agent != "" || command != "" || mcp != "" || setting != "" || hook != "" {
		handleComponentInstallation(directory)
		return
	}

	// Default: Project setup
	fmt.Println("⚙️  Project Setup")
	fmt.Println("(Implementation coming soon)")
}

// handleComponentInstallation handles installation of individual components
func handleComponentInstallation(targetDir string) {
	if preview {
		fmt.Println("👁️  Previewing Components...")
		handleComponentPreview(targetDir)
		return
	}

	fmt.Println("📦 Installing Components...")

	hasErrors := false

	// Install agents
	if agent != "" {
		agents := parseComponentList(agent)
		if len(agents) > 0 {
			fmt.Printf("\n🤖 Installing %d agent(s)...\n", len(agents))
			installer := components.NewAgentInstaller()
			if err := installer.InstallMultipleAgents(agents, targetDir, false); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				hasErrors = true
			}
		}
	}

	// Install commands
	if command != "" {
		commands := parseComponentList(command)
		if len(commands) > 0 {
			fmt.Printf("\n⚡ Installing %d command(s)...\n", len(commands))
			installer := components.NewCommandInstaller()
			if err := installer.InstallMultipleCommands(commands, targetDir, false); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				hasErrors = true
			}
		}
	}

	// Install MCPs
	if mcp != "" {
		mcps := parseComponentList(mcp)
		if len(mcps) > 0 {
			fmt.Printf("\n🔌 Installing %d MCP(s)...\n", len(mcps))

			// Parse scope
			mcpScope := components.ParseMCPScope(scope)
			installer := components.NewMCPInstaller(mcpScope)

			if err := installer.InstallMultipleMCPs(mcps, targetDir, false); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				hasErrors = true
			}
		}
	}

	// Settings and hooks
	if setting != "" {
		fmt.Println("\n⚙️  Settings installation coming soon...")
	}

	if hook != "" {
		handleHookInstallation(hook)
	}

	if !hasErrors {
		fmt.Println("\n✅ All components installed successfully!")
	}
}

// handleComponentPreview displays component content without installing
func handleComponentPreview(targetDir string) {
	hasErrors := false

	// Preview agents
	if agent != "" {
		agents := parseComponentList(agent)
		if len(agents) > 0 {
			fmt.Printf("\n🤖 Previewing %d agent(s)...\n", len(agents))
			installer := components.NewAgentInstaller()
			if err := installer.PreviewMultipleAgents(agents); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				hasErrors = true
			}
		}
	}

	// Preview commands
	if command != "" {
		commands := parseComponentList(command)
		if len(commands) > 0 {
			fmt.Printf("\n⚡ Previewing %d command(s)...\n", len(commands))
			installer := components.NewCommandInstaller()
			if err := installer.PreviewMultipleCommands(commands); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				hasErrors = true
			}
		}
	}

	// Preview MCPs
	if mcp != "" {
		mcps := parseComponentList(mcp)
		if len(mcps) > 0 {
			fmt.Printf("\n🔌 Previewing %d MCP(s)...\n", len(mcps))
			mcpScope := components.ParseMCPScope(scope)
			installer := components.NewMCPInstaller(mcpScope)
			if err := installer.PreviewMultipleMCPs(mcps); err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				hasErrors = true
			}
		}
	}

	if !hasErrors {
		fmt.Println("\n✅ Preview completed successfully!")
	}
}

// parseComponentList parses comma-separated component names
func parseComponentList(input string) []string {
	if input == "" {
		return nil
	}

	parts := strings.Split(input, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// createAnalyticsServer creates an analytics server instance
func createAnalyticsServer(targetDir string) *server.Server {
	// Get Claude directory (default to ~/.claude)
	claudeDir := filepath.Join(os.Getenv("HOME"), ".claude")

	// Check if custom directory specified
	if targetDir != "." && targetDir != "" {
		claudeDir = filepath.Join(targetDir, ".claude")
	}

	return server.NewServer(claudeDir, 3333)
}

// handleHookInstallation handles installation of hooks (legacy via --hook flag)
func handleHookInstallation(hookName string) {
	fmt.Printf("\n🔧 Installing Hook: %s\n", hookName)

	hookInstaller := components.NewHookInstaller()

	switch hookName {
	case "user-prompt-logger":
		if err := hookInstaller.InstallUserPromptLogger(); err != nil {
			ShowError(fmt.Sprintf("Failed to install hook: %v", err))
			return
		}
	case "tool-logger":
		if err := hookInstaller.InstallToolLogger(); err != nil {
			ShowError(fmt.Sprintf("Failed to install hook: %v", err))
			return
		}
	case "notification-logger":
		if err := hookInstaller.InstallNotificationLogger(); err != nil {
			ShowError(fmt.Sprintf("Failed to install hook: %v", err))
			return
		}
	case "all":
		if err := hookInstaller.InstallAllHooks(); err != nil {
			ShowError(fmt.Sprintf("Failed to install hooks: %v", err))
			return
		}
	default:
		ShowError(fmt.Sprintf("Unknown hook: %s", hookName))
		ShowInfo("Available hooks:")
		ShowInfo("  - user-prompt-logger: Capture user prompts for analytics")
		ShowInfo("  - tool-logger: Capture all tool usage (Bash, Read, Edit, etc.)")
		ShowInfo("  - notification-logger: Capture permission requests and idle alerts")
		ShowInfo("  - all: Install all hooks")
		return
	}
}

// handleHookManagement handles hook installation and removal via dedicated flags
func handleHookManagement() {
	hookInstaller := components.NewHookInstaller()

	// Install all hooks
	if installAllHooks {
		if err := hookInstaller.InstallAllHooks(); err != nil {
			ShowError(fmt.Sprintf("Failed to install hooks: %v", err))
			os.Exit(1)
		}
		return
	}

	// Uninstall all hooks
	if uninstallAllHooks {
		if err := hookInstaller.UninstallAllHooks(); err != nil {
			ShowError(fmt.Sprintf("Failed to uninstall hooks: %v", err))
			os.Exit(1)
		}
		return
	}

	// Install user prompt hook
	if installUserPromptHook {
		if err := hookInstaller.InstallUserPromptLogger(); err != nil {
			ShowError(fmt.Sprintf("Failed to install user prompt hook: %v", err))
			os.Exit(1)
		}
		return
	}

	// Uninstall user prompt hook
	if uninstallUserPromptHook {
		if err := hookInstaller.UninstallUserPromptLogger(); err != nil {
			ShowError(fmt.Sprintf("Failed to uninstall user prompt hook: %v", err))
			os.Exit(1)
		}
		return
	}

	// Install tool hook
	if installToolHook {
		if err := hookInstaller.InstallToolLogger(); err != nil {
			ShowError(fmt.Sprintf("Failed to install tool hook: %v", err))
			os.Exit(1)
		}
		return
	}

	// Uninstall tool hook
	if uninstallToolHook {
		if err := hookInstaller.UninstallToolLogger(); err != nil {
			ShowError(fmt.Sprintf("Failed to uninstall tool hook: %v", err))
			os.Exit(1)
		}
		return
	}

	// Install notification hook
	if installNotificationHook {
		if err := hookInstaller.InstallNotificationLogger(); err != nil {
			ShowError(fmt.Sprintf("Failed to install notification hook: %v", err))
			os.Exit(1)
		}
		return
	}

	// Uninstall notification hook
	if uninstallNotificationHook {
		if err := hookInstaller.UninstallNotificationLogger(); err != nil {
			ShowError(fmt.Sprintf("Failed to uninstall notification hook: %v", err))
			os.Exit(1)
		}
		return
	}
}

// handleDockerCommands handles all Docker-related operations
func handleDockerCommands(targetDir string) {
	dm := docker.NewDockerManager(targetDir)

	// Check if Docker is available
	if !dm.IsDockerAvailable() {
		ShowError("Docker is not installed or not running")
		ShowInfo("Please install Docker: https://docs.docker.com/get-docker/")
		return
	}

	// Parse MCPs list if provided
	mcpsList := parseComponentList(dockerMCPs)

	// Docker init - generate Dockerfile and .dockerignore
	if dockerInit {
		fmt.Println("🐳 Initializing Docker files...")

		generator := docker.NewDockerfileGenerator(targetDir)

		// Parse docker type
		var dockerfileType docker.DockerfileType
		switch dockerType {
		case "base":
			dockerfileType = docker.DockerfileBase
		case "claude":
			dockerfileType = docker.DockerfileClaude
		case "analytics":
			dockerfileType = docker.DockerfileAnalytics
		case "full":
			dockerfileType = docker.DockerfileFull
		default:
			dockerfileType = docker.DockerfileClaude
		}

		// Generate Dockerfile
		dockerfilePath := filepath.Join(targetDir, "Dockerfile")
		if err := generator.GenerateDockerfile(dockerfileType, dockerfilePath, mcpsList); err != nil {
			ShowError(fmt.Sprintf("Failed to generate Dockerfile: %v", err))
			return
		}

		// Generate .dockerignore
		dockerignorePath := filepath.Join(targetDir, ".dockerignore")
		if err := generator.GenerateDockerIgnore(dockerignorePath); err != nil {
			ShowError(fmt.Sprintf("Failed to generate .dockerignore: %v", err))
			return
		}

		ShowSuccess("Docker files generated successfully!")
		ShowInfo(fmt.Sprintf("Dockerfile: %s", dockerfilePath))
		ShowInfo(fmt.Sprintf(".dockerignore: %s", dockerignorePath))
		return
	}

	// Docker compose - generate docker-compose.yml
	if dockerCompose {
		fmt.Println("🐳 Generating docker-compose.yml...")

		generator := docker.NewComposeGenerator(targetDir)

		// Parse compose template
		var composeTemplate docker.ComposeTemplate
		switch dockerType {
		case "simple":
			composeTemplate = docker.ComposeSimple
		case "analytics":
			composeTemplate = docker.ComposeAnalytics
		case "database":
			composeTemplate = docker.ComposeDatabase
		case "full":
			composeTemplate = docker.ComposeFull
		default:
			composeTemplate = docker.ComposeSimple
		}

		// Generate docker-compose.yml
		composePath := filepath.Join(targetDir, "docker-compose.yml")
		if err := generator.GenerateCompose(composeTemplate, composePath, mcpsList); err != nil {
			ShowError(fmt.Sprintf("Failed to generate docker-compose.yml: %v", err))
			return
		}

		// Generate .env.example
		envPath := filepath.Join(targetDir, ".env.example")
		if err := generator.GenerateEnvFile(envPath); err != nil {
			ShowError(fmt.Sprintf("Failed to generate .env.example: %v", err))
			return
		}

		ShowSuccess("Docker Compose files generated successfully!")
		ShowInfo(fmt.Sprintf("docker-compose.yml: %s", composePath))
		ShowInfo(fmt.Sprintf(".env.example: %s", envPath))
		ShowInfo("Copy .env.example to .env and configure your environment variables")
		return
	}

	// Docker build
	if dockerBuild {
		dockerfilePath := filepath.Join(targetDir, "Dockerfile")
		if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
			ShowError("Dockerfile not found. Run with --docker-init first")
			return
		}

		// Build the cct binary first
		ShowInfo("Building cct binary for Docker image...")
		buildCmd := exec.Command("make", "build")
		buildCmd.Dir = targetDir
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr

		if err := buildCmd.Run(); err != nil {
			ShowError(fmt.Sprintf("Failed to build cct binary: %v", err))
			ShowInfo("Please ensure you have Go installed and run 'make build' manually")
			return
		}

		// Check if binary exists in target directory
		binaryPath := filepath.Join(targetDir, "cct")
		if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
			ShowError("cct binary not found after build. Please run 'make build' manually")
			return
		}

		ShowSuccess("cct binary built successfully!")

		if err := dm.BuildImage(dockerfilePath); err != nil {
			ShowError(fmt.Sprintf("Failed to build Docker image: %v", err))
			return
		}
		return
	}

	// Docker run
	if dockerRun {
		opts := docker.NewRunOptions()

		// Default port mapping for analytics
		opts.Ports[3333] = 3333

		// Mount current directory
		absPath, _ := filepath.Abs(targetDir)
		opts.Volumes[absPath] = "/workspace"

		// Mount .claude directory
		claudeDir := filepath.Join(os.Getenv("HOME"), ".claude")
		opts.Volumes[claudeDir] = "/root/.claude"

		// Set command if provided
		if dockerCommand != "" {
			opts.Command = dockerCommand
		}

		if err := dm.RunContainer(opts); err != nil {
			ShowError(fmt.Sprintf("Failed to run Docker container: %v", err))
			return
		}

		ShowInfo("To view logs: cct --docker-logs")
		ShowInfo("To stop container: cct --docker-stop")
		return
	}

	// Docker stop
	if dockerStop {
		if err := dm.StopContainer(); err != nil {
			ShowError(fmt.Sprintf("Failed to stop Docker container: %v", err))
			return
		}
		ShowSuccess("Docker container stopped successfully!")
		return
	}

	// Docker logs
	if dockerLogs {
		fmt.Println("📋 Docker container logs:")
		if err := dm.GetContainerLogs(false); err != nil {
			ShowError(fmt.Sprintf("Failed to get logs: %v", err))
			return
		}
		return
	}
}
