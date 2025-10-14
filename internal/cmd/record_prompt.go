package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/schlunsen/claude-control-terminal/internal/database"
	"github.com/spf13/cobra"
)

var (
	recordSession string
	recordPrompt  string
	recordCwd     string
	recordBranch  string
)

// recordPromptCmd is a hidden command used by hooks to record user prompts
var recordPromptCmd = &cobra.Command{
	Use:    "record-prompt",
	Hidden: true,
	Short:  "Record a user prompt to the database (internal use)",
	Long:   "This command is used internally by hooks to record user prompts. Not intended for direct use.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := recordPromptToDatabase(); err != nil {
			// Silent failure - don't block Claude Code
			if verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to record prompt: %v\n", err)
			}
			os.Exit(0)
		}
		os.Exit(0)
	},
}

func init() {
	recordPromptCmd.Flags().StringVar(&recordSession, "session", "", "session/conversation ID")
	recordPromptCmd.Flags().StringVar(&recordPrompt, "prompt", "", "user prompt text")
	recordPromptCmd.Flags().StringVar(&recordCwd, "cwd", "", "working directory")
	recordPromptCmd.Flags().StringVar(&recordBranch, "branch", "", "git branch")

	recordPromptCmd.MarkFlagRequired("session")
	recordPromptCmd.MarkFlagRequired("prompt")
	recordPromptCmd.MarkFlagRequired("cwd")

	rootCmd.AddCommand(recordPromptCmd)
}

// recordPromptToDatabase saves a user prompt to the SQLite database
func recordPromptToDatabase() error {
	// Validate inputs
	if recordSession == "" || recordPrompt == "" || recordCwd == "" {
		return fmt.Errorf("missing required fields")
	}

	// Initialize database
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	dataDir := filepath.Join(claudeDir, "analytics_data")

	db, err := database.Initialize(dataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	repo := database.NewRepository(db)

	// Create user message record
	userMsg := &database.UserMessage{
		ConversationID:   recordSession,
		Message:          recordPrompt,
		WorkingDirectory: recordCwd,
		GitBranch:        recordBranch,
		MessageLength:    len(recordPrompt),
		SubmittedAt:      time.Now(),
	}

	// Save to database
	if err := repo.RecordUserMessage(userMsg); err != nil {
		return fmt.Errorf("failed to record message: %w", err)
	}

	if verbose {
		fmt.Printf("✓ Recorded prompt (session: %s, length: %d)\n", recordSession, len(recordPrompt))
	}

	return nil
}
