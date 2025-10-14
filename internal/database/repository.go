// Package database provides data access methods for querying and persisting command history.
// This file implements the Repository pattern for database operations including
// recording commands, retrieving history, and updating statistics.
package database

import (
	"fmt"
	"strings"
)

// Repository provides data access methods for command history
type Repository struct {
	db *Database
}

// NewRepository creates a new repository instance
func NewRepository(db *Database) *Repository {
	return &Repository{db: db}
}

// RecordShellCommand saves a shell command execution
func (r *Repository) RecordShellCommand(cmd *ShellCommand) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	query := `
		INSERT INTO shell_commands (
			conversation_id, command, description, working_directory, git_branch,
			exit_code, stdout, stderr, duration_ms, executed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.db.Exec(
		query,
		cmd.ConversationID,
		cmd.Command,
		cmd.Description,
		cmd.WorkingDirectory,
		cmd.GitBranch,
		cmd.ExitCode,
		cmd.Stdout,
		cmd.Stderr,
		cmd.DurationMs,
		cmd.ExecutedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to record shell command: %w", err)
	}

	id, _ := result.LastInsertId()
	cmd.ID = id

	// Update conversation and stats
	go r.updateConversationStats(cmd.ConversationID)
	go r.updateCommandStats("shell", extractCommandName(cmd.Command), cmd.ExitCode == nil || *cmd.ExitCode == 0, cmd.DurationMs)

	return nil
}

// RecordClaudeCommand saves a Claude Code tool invocation
func (r *Repository) RecordClaudeCommand(cmd *ClaudeCommand) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	query := `
		INSERT INTO claude_commands (
			conversation_id, tool_name, parameters, result, working_directory, git_branch,
			success, error_message, duration_ms, executed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.db.Exec(
		query,
		cmd.ConversationID,
		cmd.ToolName,
		cmd.Parameters,
		cmd.Result,
		cmd.WorkingDirectory,
		cmd.GitBranch,
		cmd.Success,
		cmd.ErrorMessage,
		cmd.DurationMs,
		cmd.ExecutedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to record claude command: %w", err)
	}

	id, _ := result.LastInsertId()
	cmd.ID = id

	// Update conversation and stats
	go r.updateConversationStats(cmd.ConversationID)
	go r.updateCommandStats("claude", cmd.ToolName, cmd.Success, cmd.DurationMs)

	return nil
}

// GetShellCommands retrieves shell commands with optional filters
func (r *Repository) GetShellCommands(query *CommandHistoryQuery) ([]*ShellCommand, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	sql, args := r.buildShellCommandQuery(query)
	rows, err := r.db.db.Query(sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query shell commands: %w", err)
	}
	defer rows.Close()

	var commands []*ShellCommand
	for rows.Next() {
		cmd := &ShellCommand{}
		err := rows.Scan(
			&cmd.ID,
			&cmd.ConversationID,
			&cmd.Command,
			&cmd.Description,
			&cmd.WorkingDirectory,
			&cmd.GitBranch,
			&cmd.ExitCode,
			&cmd.Stdout,
			&cmd.Stderr,
			&cmd.DurationMs,
			&cmd.ExecutedAt,
			&cmd.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shell command: %w", err)
		}
		commands = append(commands, cmd)
	}

	return commands, nil
}

// GetClaudeCommands retrieves Claude commands with optional filters
func (r *Repository) GetClaudeCommands(query *CommandHistoryQuery) ([]*ClaudeCommand, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	sql, args := r.buildClaudeCommandQuery(query)
	rows, err := r.db.db.Query(sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query claude commands: %w", err)
	}
	defer rows.Close()

	var commands []*ClaudeCommand
	for rows.Next() {
		cmd := &ClaudeCommand{}
		err := rows.Scan(
			&cmd.ID,
			&cmd.ConversationID,
			&cmd.ToolName,
			&cmd.Parameters,
			&cmd.Result,
			&cmd.WorkingDirectory,
			&cmd.GitBranch,
			&cmd.Success,
			&cmd.ErrorMessage,
			&cmd.DurationMs,
			&cmd.ExecutedAt,
			&cmd.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan claude command: %w", err)
		}
		commands = append(commands, cmd)
	}

	return commands, nil
}

// GetCommandStats retrieves aggregated command statistics
func (r *Repository) GetCommandStats(commandType string, limit int) ([]*CommandStat, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	query := `
		SELECT id, command_type, command_name, execution_count,
		       success_count, failure_count, avg_duration_ms,
		       last_executed_at, created_at, updated_at
		FROM command_stats
		WHERE 1=1
	`

	args := []interface{}{}
	if commandType != "" {
		query += " AND command_type = ?"
		args = append(args, commandType)
	}

	query += " ORDER BY execution_count DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := r.db.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query command stats: %w", err)
	}
	defer rows.Close()

	var stats []*CommandStat
	for rows.Next() {
		stat := &CommandStat{}
		err := rows.Scan(
			&stat.ID,
			&stat.CommandType,
			&stat.CommandName,
			&stat.ExecutionCount,
			&stat.SuccessCount,
			&stat.FailureCount,
			&stat.AvgDurationMs,
			&stat.LastExecutedAt,
			&stat.CreatedAt,
			&stat.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan command stat: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

// UpsertConversation creates or updates a conversation record
func (r *Repository) UpsertConversation(conv *Conversation) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	query := `
		INSERT INTO conversations (
			id, project_path, started_at, last_activity_at,
			total_commands, total_shell_commands, total_tokens, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			project_path = excluded.project_path,
			last_activity_at = excluded.last_activity_at,
			total_commands = excluded.total_commands,
			total_shell_commands = excluded.total_shell_commands,
			total_tokens = excluded.total_tokens,
			status = excluded.status,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := r.db.db.Exec(
		query,
		conv.ID,
		conv.ProjectPath,
		conv.StartedAt,
		conv.LastActivityAt,
		conv.TotalCommands,
		conv.TotalShellCommands,
		conv.TotalTokens,
		conv.Status,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert conversation: %w", err)
	}

	return nil
}

// Helper methods

func (r *Repository) buildShellCommandQuery(query *CommandHistoryQuery) (string, []interface{}) {
	sql := `
		SELECT id, conversation_id, command, description, working_directory, git_branch,
		       exit_code, stdout, stderr, duration_ms, executed_at, created_at
		FROM shell_commands
		WHERE 1=1
	`

	args := []interface{}{}

	if query.ConversationID != "" {
		sql += " AND conversation_id = ?"
		args = append(args, query.ConversationID)
	}

	if query.StartDate != nil {
		sql += " AND executed_at >= ?"
		args = append(args, query.StartDate)
	}

	if query.EndDate != nil {
		sql += " AND executed_at <= ?"
		args = append(args, query.EndDate)
	}

	sql += " ORDER BY executed_at DESC"

	if query.Limit > 0 {
		sql += " LIMIT ?"
		args = append(args, query.Limit)
	}

	if query.Offset > 0 {
		sql += " OFFSET ?"
		args = append(args, query.Offset)
	}

	return sql, args
}

func (r *Repository) buildClaudeCommandQuery(query *CommandHistoryQuery) (string, []interface{}) {
	sql := `
		SELECT id, conversation_id, tool_name, parameters, result, working_directory, git_branch,
		       success, error_message, duration_ms, executed_at, created_at
		FROM claude_commands
		WHERE 1=1
	`

	args := []interface{}{}

	if query.ConversationID != "" {
		sql += " AND conversation_id = ?"
		args = append(args, query.ConversationID)
	}

	if query.ToolName != "" {
		sql += " AND tool_name = ?"
		args = append(args, query.ToolName)
	}

	if query.StartDate != nil {
		sql += " AND executed_at >= ?"
		args = append(args, query.StartDate)
	}

	if query.EndDate != nil {
		sql += " AND executed_at <= ?"
		args = append(args, query.EndDate)
	}

	sql += " ORDER BY executed_at DESC"

	if query.Limit > 0 {
		sql += " LIMIT ?"
		args = append(args, query.Limit)
	}

	if query.Offset > 0 {
		sql += " OFFSET ?"
		args = append(args, query.Offset)
	}

	return sql, args
}

func (r *Repository) updateConversationStats(conversationID string) {
	// Update conversation totals
	query := `
		UPDATE conversations
		SET total_commands = (
			SELECT COUNT(*) FROM claude_commands WHERE conversation_id = ?
		),
		total_shell_commands = (
			SELECT COUNT(*) FROM shell_commands WHERE conversation_id = ?
		),
		last_activity_at = CURRENT_TIMESTAMP,
		updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	r.db.db.Exec(query, conversationID, conversationID, conversationID)
}

func (r *Repository) updateCommandStats(commandType, commandName string, success bool, durationMs *int) {
	duration := 0
	if durationMs != nil {
		duration = *durationMs
	}

	successCount := 0
	failureCount := 0
	if success {
		successCount = 1
	} else {
		failureCount = 1
	}

	query := `
		INSERT INTO command_stats (
			command_type, command_name, execution_count,
			success_count, failure_count, avg_duration_ms, last_executed_at
		) VALUES (?, ?, 1, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(command_type, command_name) DO UPDATE SET
			execution_count = execution_count + 1,
			success_count = success_count + ?,
			failure_count = failure_count + ?,
			avg_duration_ms = (avg_duration_ms * execution_count + ?) / (execution_count + 1),
			last_executed_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
	`

	r.db.db.Exec(
		query,
		commandType,
		commandName,
		successCount,
		failureCount,
		duration,
		successCount,
		failureCount,
		duration,
	)
}

func extractCommandName(command string) string {
	// Extract first word of command (e.g., "git status" -> "git")
	parts := strings.Fields(command)
	if len(parts) > 0 {
		return parts[0]
	}
	return command
}

// RecordUserMessage saves a user's input message
func (r *Repository) RecordUserMessage(msg *UserMessage) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	query := `
		INSERT INTO user_messages (
			conversation_id, session_name, message, working_directory, git_branch,
			message_length, submitted_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.db.Exec(
		query,
		msg.ConversationID,
		msg.SessionName,
		msg.Message,
		msg.WorkingDirectory,
		msg.GitBranch,
		msg.MessageLength,
		msg.SubmittedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to record user message: %w", err)
	}

	id, _ := result.LastInsertId()
	msg.ID = id

	return nil
}

// GetUserMessages retrieves user messages with optional filters
func (r *Repository) GetUserMessages(query *CommandHistoryQuery) ([]*UserMessage, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	sql := `
		SELECT id, conversation_id, COALESCE(session_name, '') as session_name, message, working_directory, git_branch,
		       message_length, submitted_at, created_at
		FROM user_messages
		WHERE 1=1
	`

	args := []interface{}{}

	if query.ConversationID != "" {
		sql += " AND conversation_id = ?"
		args = append(args, query.ConversationID)
	}

	if query.StartDate != nil {
		sql += " AND submitted_at >= ?"
		args = append(args, query.StartDate)
	}

	if query.EndDate != nil {
		sql += " AND submitted_at <= ?"
		args = append(args, query.EndDate)
	}

	sql += " ORDER BY submitted_at DESC"

	if query.Limit > 0 {
		sql += " LIMIT ?"
		args = append(args, query.Limit)
	}

	if query.Offset > 0 {
		sql += " OFFSET ?"
		args = append(args, query.Offset)
	}

	rows, err := r.db.db.Query(sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query user messages: %w", err)
	}
	defer rows.Close()

	var messages []*UserMessage
	for rows.Next() {
		msg := &UserMessage{}
		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SessionName,
			&msg.Message,
			&msg.WorkingDirectory,
			&msg.GitBranch,
			&msg.MessageLength,
			&msg.SubmittedAt,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// SaveProvider saves or updates a provider configuration
// It sets the provider as current and unsets all other providers
func (r *Repository) SaveProvider(provider *ProviderConfig) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	tx, err := r.db.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Unset all other providers as current
	if _, err := tx.Exec("UPDATE providers SET is_current = 0"); err != nil {
		return fmt.Errorf("failed to update current providers: %w", err)
	}

	// Insert or update the provider
	query := `
		INSERT INTO providers (provider_id, api_key, custom_url, model_name, is_current)
		VALUES (?, ?, ?, ?, 1)
		ON CONFLICT(provider_id) DO UPDATE SET
			api_key = excluded.api_key,
			custom_url = excluded.custom_url,
			model_name = excluded.model_name,
			is_current = 1,
			updated_at = CURRENT_TIMESTAMP
	`

	if _, err := tx.Exec(query, provider.ProviderID, provider.APIKey, provider.CustomURL, provider.ModelName); err != nil {
		return fmt.Errorf("failed to save provider: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetProvider retrieves a specific provider configuration
func (r *Repository) GetProvider(providerID string) (*ProviderConfig, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	query := `
		SELECT provider_id, api_key, custom_url, model_name, is_current, created_at, updated_at
		FROM providers
		WHERE provider_id = ?
	`

	provider := &ProviderConfig{}
	err := r.db.db.QueryRow(query, providerID).Scan(
		&provider.ProviderID,
		&provider.APIKey,
		&provider.CustomURL,
		&provider.ModelName,
		&provider.IsCurrent,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	return provider, nil
}

// GetCurrentProvider retrieves the currently active provider configuration
func (r *Repository) GetCurrentProvider() (*ProviderConfig, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	query := `
		SELECT provider_id, api_key, custom_url, model_name, is_current, created_at, updated_at
		FROM providers
		WHERE is_current = 1
		LIMIT 1
	`

	provider := &ProviderConfig{}
	err := r.db.db.QueryRow(query).Scan(
		&provider.ProviderID,
		&provider.APIKey,
		&provider.CustomURL,
		&provider.ModelName,
		&provider.IsCurrent,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get current provider: %w", err)
	}

	return provider, nil
}

// GetAllProviders retrieves all saved provider configurations
func (r *Repository) GetAllProviders() ([]*ProviderConfig, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	query := `
		SELECT provider_id, api_key, custom_url, model_name, is_current, created_at, updated_at
		FROM providers
		ORDER BY updated_at DESC
	`

	rows, err := r.db.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query providers: %w", err)
	}
	defer rows.Close()

	var providers []*ProviderConfig
	for rows.Next() {
		provider := &ProviderConfig{}
		err := rows.Scan(
			&provider.ProviderID,
			&provider.APIKey,
			&provider.CustomURL,
			&provider.ModelName,
			&provider.IsCurrent,
			&provider.CreatedAt,
			&provider.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider: %w", err)
		}
		providers = append(providers, provider)
	}

	return providers, nil
}

// DeleteProvider removes a provider configuration
func (r *Repository) DeleteProvider(providerID string) error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	query := "DELETE FROM providers WHERE provider_id = ?"
	_, err := r.db.db.Exec(query, providerID)
	if err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}

	return nil
}

// DeleteAllProviders removes all provider configurations
func (r *Repository) DeleteAllProviders() error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	query := "DELETE FROM providers"
	_, err := r.db.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete all providers: %w", err)
	}

	return nil
}

// DeleteAllUserMessages removes all user messages
func (r *Repository) DeleteAllUserMessages() error {
	r.db.mu.Lock()
	defer r.db.mu.Unlock()

	query := "DELETE FROM user_messages"
	_, err := r.db.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete all user messages: %w", err)
	}

	return nil
}

// GetUniqueSessions retrieves all unique session IDs and names from user messages
func (r *Repository) GetUniqueSessions() ([]map[string]string, error) {
	r.db.mu.RLock()
	defer r.db.mu.RUnlock()

	query := `
		SELECT conversation_id, COALESCE(session_name, '') as session_name, MAX(submitted_at) as last_submitted
		FROM user_messages
		WHERE conversation_id != '' AND conversation_id IS NOT NULL
		GROUP BY conversation_id, session_name
		ORDER BY last_submitted DESC
	`

	rows, err := r.db.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query unique sessions: %w", err)
	}
	defer rows.Close()

	var sessions []map[string]string
	for rows.Next() {
		var conversationID, sessionName, lastSubmitted string
		err := rows.Scan(&conversationID, &sessionName, &lastSubmitted)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, map[string]string{
			"conversation_id": conversationID,
			"session_name":    sessionName,
		})
	}

	return sessions, nil
}
