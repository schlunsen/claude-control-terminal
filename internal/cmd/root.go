package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davila7/go-claude-templates/internal/components"
	"github.com/davila7/go-claude-templates/internal/server"
	"github.com/davila7/go-claude-templates/internal/tui"
	"github.com/spf13/cobra"
)

const (
	Version = "0.0.6"
	Name    = "go-claude-templates"
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

	// Other flags
	template   string
	language   string
	framework  string
	prompt     string
	studio     bool
	sandbox    string
	e2bAPIKey  string
	anthropicAPIKey string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "cct",
	Short: "Claude Code Templates - Go Edition",
	Long: `Component templates and tracking system for Claude Code.

🚀 Setup Claude Code for any project language
🌐 Templates: https://aitmpl.com
📖 Documentation: https://docs.aitmpl.com`,
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		// Show banner for interactive mode
		isInteractive := !analytics && !chats && !agents && !chatsMobile && !plugins &&
			!healthCheck && !commandStats && !hookStats && !mcpStats &&
			!listAgents && createAgent == "" && removeAgent == "" && updateAgent == "" &&
			agent == "" && command == "" && mcp == "" && setting == "" && hook == "" &&
			workflow == "" && !studio && sandbox == ""

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

	// Other flags
	rootCmd.Flags().StringVar(&prompt, "prompt", "", "execute prompt after installation")
	rootCmd.Flags().BoolVar(&studio, "studio", false, "launch Claude Code Studio")
	rootCmd.Flags().StringVar(&sandbox, "sandbox", "", "execute in sandbox (e.g., e2b)")
	rootCmd.Flags().StringVar(&e2bAPIKey, "e2b-api-key", "", "E2B API key")
	rootCmd.Flags().StringVar(&anthropicAPIKey, "anthropic-api-key", "", "Anthropic API key")
}

func handleCommand(cmd *cobra.Command, args []string) {
	// Analytics dashboard
	if analytics {
		spinner := ShowSpinner("Launching Analytics Dashboard...")

		// Import server package
		server := createAnalyticsServer(directory)

		spinner.Success("Analytics Dashboard starting!")
		ShowInfo(fmt.Sprintf("Dashboard: http://localhost:3333"))
		ShowInfo(fmt.Sprintf("API: http://localhost:3333/api/data"))
		ShowInfo("Press Ctrl+C to stop")

		if err := server.Setup(); err != nil {
			ShowError(fmt.Sprintf("Failed to setup server: %v", err))
			return
		}

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
		fmt.Println("\n🔧 Hooks installation coming soon...")
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
