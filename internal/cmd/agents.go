package cmd

import (
	"fmt"

	agentspkg "github.com/schlunsen/claude-control-terminal/internal/agents"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	agentsQuiet  bool
	agentsFollow bool
	agentsLines  int
)

// agentsCmd represents the agents command
var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Manage the Claude agent WebSocket server",
	Long: `Manage the Claude agent WebSocket server for running agent conversations.

The agent server is a Go-based WebSocket server that integrates with the
Claude Agent SDK to provide real-time agent conversations with tool support.

Examples:
  cct agents start           Start the agent server
  cct agents stop            Stop the agent server
  cct agents restart         Restart the agent server
  cct agents status          Show server status
  cct agents logs            Show recent logs
  cct agents logs --follow   Follow logs in real-time`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand, show status or help
		config := agentspkg.DefaultConfig()
		launcher := agentspkg.NewLauncher(config, agentsQuiet, false)

		status, err := launcher.Status()
		if err != nil {
			pterm.Error.Println("Failed to get status:", err)
			return
		}

		pterm.Info.Println(status)
		pterm.Info.Println("\nUse 'cct agents --help' to see available commands")
	},
}

// agentsStartCmd starts the agent server
var agentsStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the agent server",
	Long:  `Start the Claude agent WebSocket server. Automatically installs dependencies if needed.`,
	Run: func(cmd *cobra.Command, args []string) {
		config := agentspkg.DefaultConfig()
		launcher := agentspkg.NewLauncher(config, agentsQuiet, false)

		if err := launcher.Start(); err != nil {
			pterm.Error.Println("Failed to start agent server:", err)
			return
		}
	},
}

// agentsStopCmd stops the agent server
var agentsStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the agent server",
	Long:  `Stop the Claude agent WebSocket server gracefully.`,
	Run: func(cmd *cobra.Command, args []string) {
		config := agentspkg.DefaultConfig()
		launcher := agentspkg.NewLauncher(config, agentsQuiet, false)

		if err := launcher.Stop(); err != nil {
			pterm.Error.Println("Failed to stop agent server:", err)
			return
		}
	},
}

// agentsRestartCmd restarts the agent server
var agentsRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the agent server",
	Long:  `Restart the Claude agent WebSocket server (stop then start).`,
	Run: func(cmd *cobra.Command, args []string) {
		config := agentspkg.DefaultConfig()
		launcher := agentspkg.NewLauncher(config, agentsQuiet, false)

		if err := launcher.Restart(); err != nil {
			pterm.Error.Println("Failed to restart agent server:", err)
			return
		}
	},
}

// agentsStatusCmd shows the agent server status
var agentsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show agent server status",
	Long:  `Show the current status of the Claude agent WebSocket server.`,
	Run: func(cmd *cobra.Command, args []string) {
		config := agentspkg.DefaultConfig()
		launcher := agentspkg.NewLauncher(config, agentsQuiet, false)

		status, err := launcher.Status()
		if err != nil {
			pterm.Error.Println("Failed to get status:", err)
			return
		}

		pterm.Info.Println(status)
	},
}

// agentsLogsCmd shows the agent server logs
var agentsLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show agent server logs",
	Long:  `Show the agent server logs. Use --follow to tail logs in real-time.`,
	Run: func(cmd *cobra.Command, args []string) {
		config := agentspkg.DefaultConfig()
		launcher := agentspkg.NewLauncher(config, true, false) // Always quiet for logs

		if err := launcher.Logs(agentsLines, agentsFollow); err != nil {
			pterm.Error.Println("Failed to show logs:", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(agentsCmd)

	// Add subcommands
	agentsCmd.AddCommand(agentsStartCmd)
	agentsCmd.AddCommand(agentsStopCmd)
	agentsCmd.AddCommand(agentsRestartCmd)
	agentsCmd.AddCommand(agentsStatusCmd)
	agentsCmd.AddCommand(agentsLogsCmd)

	// Flags for agents command
	agentsCmd.PersistentFlags().BoolVarP(&agentsQuiet, "quiet", "q", false, "Suppress output")

	// Flags for logs command
	agentsLogsCmd.Flags().BoolVarP(&agentsFollow, "follow", "f", false, "Follow log output")
	agentsLogsCmd.Flags().IntVarP(&agentsLines, "lines", "n", 50, "Number of lines to show")
}

// LaunchAgentServer is called by the --agents flag for backward compatibility
func LaunchAgentServer() {
	config := agentspkg.DefaultConfig()
	launcher := agentspkg.NewLauncher(config, false, false)

	// Check if already running
	running, _, _ := launcher.IsRunning()
	if running {
		pterm.Info.Println("Agent server is already running")
		status, _ := launcher.Status()
		pterm.Info.Println(status)
		return
	}

	// Start the server
	if err := launcher.Start(); err != nil {
		pterm.Error.Println("Failed to start agent server:", err)
		return
	}

	// Show help message
	fmt.Println()
	pterm.Info.Println("Agent server commands:")
	pterm.Info.Println("  cct agents status    - Show server status")
	pterm.Info.Println("  cct agents stop      - Stop the server")
	pterm.Info.Println("  cct agents logs      - View server logs")
	pterm.Info.Println("  cct agents logs -f   - Follow server logs")
}
