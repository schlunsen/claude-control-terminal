package tui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pterm/pterm"
	"github.com/schlunsen/claude-control-terminal/internal/server"
	"golang.org/x/term"
)

// CheckAndSetupAuthBeforeOpen checks if user auth is enabled but no users exist
// and prompts to set up authentication before opening the dashboard
func CheckAndSetupAuthBeforeOpen(claudeDir string) error {
	analyticsDir := filepath.Join(claudeDir, "analytics")
	configManager := server.NewConfigManager(claudeDir)

	// Load current config
	config, err := configManager.LoadOrCreateConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// If user auth is not enabled, no action needed
	if !config.Auth.UserAuthEnabled {
		return nil
	}

	// Check if we have users
	userStore := server.NewUserStore(analyticsDir)
	if err := userStore.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize user store: %w", err)
	}

	// If users exist, we're good to go
	if userStore.HasUsers() {
		return nil
	}

	// No users exist - prompt to create admin
	pterm.Warning.Println("")
	pterm.Warning.Println("⚠️  User authentication is enabled but no admin user exists!")
	pterm.Warning.Println("   You need to create an admin account before accessing the dashboard.")
	pterm.Warning.Println("")

	// Ask if they want to create one now
	createNow, err := promptYesNo("Create admin account now?", true)
	if err != nil {
		return err
	}

	if !createNow {
		pterm.Info.Println("You can create an admin account later by pressing 'U' in the TUI")
		return fmt.Errorf("admin account required")
	}

	// Create first admin user
	return promptCreateFirstUser(claudeDir, userStore, configManager)
}

// SetupUserAuth prompts the user to set up authentication
func SetupUserAuth(claudeDir string) error {
	analyticsDir := filepath.Join(claudeDir, "analytics")
	configManager := server.NewConfigManager(claudeDir)

	// Load current config
	config, err := configManager.LoadOrCreateConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if user auth is already enabled
	if config.Auth.UserAuthEnabled {
		pterm.Info.Println("User authentication is already enabled")

		// Check if we have users
		userStore := server.NewUserStore(analyticsDir)
		if err := userStore.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize user store: %w", err)
		}

		if userStore.HasUsers() {
			pterm.Success.Printf("You have %d user(s) configured\n", len(userStore.ListUsers()))
			return promptUserAuthMenu(claudeDir, configManager, userStore, config)
		} else {
			pterm.Warning.Println("No users configured yet")
			return promptCreateFirstUser(claudeDir, userStore, configManager)
		}
	}

	// User auth not enabled, ask if they want to enable it
	pterm.Info.Println("User authentication is currently disabled")
	pterm.Info.Println("Would you like to enable username/password authentication?")
	pterm.Info.Println("")

	enable, err := promptYesNo("Enable user authentication?", true)
	if err != nil {
		return err
	}

	if !enable {
		pterm.Info.Println("User authentication will remain disabled")
		return nil
	}

	// Ask if login should be required for all pages
	requireLogin, err := promptYesNo("Require login for all pages? (If no, only API writes require auth)", false)
	if err != nil {
		return err
	}

	// Enable user auth in config
	if err := configManager.EnableUserAuth(requireLogin); err != nil {
		return fmt.Errorf("failed to enable user auth: %w", err)
	}

	pterm.Success.Println("User authentication enabled!")

	// Create first admin user
	userStore := server.NewUserStore(analyticsDir)
	if err := userStore.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize user store: %w", err)
	}

	return promptCreateFirstUser(claudeDir, userStore, configManager)
}

// promptCreateFirstUser prompts to create the first admin user
func promptCreateFirstUser(claudeDir string, userStore *server.UserStore, configManager *server.ConfigManager) error {
	pterm.Info.Println("")
	pterm.Info.Println("Let's create your admin account")
	pterm.Info.Println("")

	// Get username
	reader := bufio.NewReader(os.Stdin)
	pterm.Print("Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read username: %w", err)
	}
	username = strings.TrimSpace(username)

	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Get password
	pterm.Print("Password (min 8 characters): ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println() // Add newline after password input

	password := string(passwordBytes)
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	// Confirm password
	pterm.Print("Confirm password: ")
	confirmBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password confirmation: %w", err)
	}
	fmt.Println() // Add newline after password input

	if string(confirmBytes) != password {
		return fmt.Errorf("passwords do not match")
	}

	// Create admin user
	if err := userStore.CreateUser(username, password, true); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	pterm.Success.Printf("✅ Admin user '%s' created successfully!\n", username)
	pterm.Info.Println("")
	pterm.Info.Println("You can now:")
	pterm.Info.Println("  • Login with your username and password")
	pterm.Info.Println("  • Create additional users via the dashboard")
	pterm.Info.Println("  • Change your password in settings")
	pterm.Info.Println("")

	// Prompt to restart server if it's running
	pterm.Warning.Println("⚠️  Please restart the analytics server for changes to take effect")
	pterm.Info.Println("   (Press 'A' to toggle server off then on again)")

	return nil
}

// promptUserAuthMenu shows options for managing user authentication
func promptUserAuthMenu(claudeDir string, configManager *server.ConfigManager, userStore *server.UserStore, config *server.Config) error {
	pterm.Info.Println("")
	pterm.Info.Println("User Authentication Options:")
	pterm.Info.Println("")
	pterm.Info.Println("  1. Create a new user")
	pterm.Info.Println("  2. List users")
	pterm.Info.Println("  3. Disable user authentication")
	pterm.Info.Println("  4. Cancel")
	pterm.Info.Println("")

	reader := bufio.NewReader(os.Stdin)
	pterm.Print("Select option (1-4): ")
	choice, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read choice: %w", err)
	}
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		return promptCreateAdditionalUser(userStore)
	case "2":
		users := userStore.ListUsers()
		pterm.Info.Printf("Configured users (%d):\n", len(users))
		for _, username := range users {
			pterm.Info.Printf("  • %s\n", username)
		}
		return nil
	case "3":
		disable, err := promptYesNo("Are you sure you want to disable user authentication?", false)
		if err != nil {
			return err
		}
		if disable {
			if err := configManager.DisableUserAuth(); err != nil {
				return fmt.Errorf("failed to disable user auth: %w", err)
			}
			pterm.Success.Println("User authentication disabled")
			pterm.Warning.Println("Please restart the analytics server for changes to take effect")
		}
		return nil
	case "4":
		return nil
	default:
		pterm.Warning.Println("Invalid choice")
		return nil
	}
}

// promptCreateAdditionalUser prompts to create an additional user
func promptCreateAdditionalUser(userStore *server.UserStore) error {
	pterm.Info.Println("")

	// Get username
	reader := bufio.NewReader(os.Stdin)
	pterm.Print("Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read username: %w", err)
	}
	username = strings.TrimSpace(username)

	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Ask if user should be admin
	isAdmin, err := promptYesNo("Make this user an admin?", false)
	if err != nil {
		return err
	}

	// Get password
	pterm.Print("Password (min 8 characters): ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println()

	password := string(passwordBytes)
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	// Confirm password
	pterm.Print("Confirm password: ")
	confirmBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password confirmation: %w", err)
	}
	fmt.Println()

	if string(confirmBytes) != password {
		return fmt.Errorf("passwords do not match")
	}

	// Create user
	if err := userStore.CreateUser(username, password, isAdmin); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	pterm.Success.Printf("✅ User '%s' created successfully!\n", username)
	return nil
}

// promptYesNo prompts for a yes/no answer
func promptYesNo(question string, defaultYes bool) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	prompt := question
	if defaultYes {
		prompt += " [Y/n]: "
	} else {
		prompt += " [y/N]: "
	}

	pterm.Print(prompt)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read answer: %w", err)
	}

	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer == "" {
		return defaultYes, nil
	}

	return answer == "y" || answer == "yes", nil
}
