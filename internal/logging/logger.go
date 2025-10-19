package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger provides application-wide logging with file output
type Logger struct {
	file       *os.File
	logger     *log.Logger
	verbose    bool
	mu         sync.Mutex
	logFile    string
	stderrFile string
}

var (
	globalLogger *Logger
	once         sync.Once
)

// Initialize creates the global logger instance
// logDir: directory where log files will be written
// verbose: enable debug/verbose logging
func Initialize(logDir string, verbose bool) (*Logger, error) {
	var err error
	once.Do(func() {
		globalLogger, err = newLogger(logDir, verbose)
	})
	return globalLogger, err
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	return globalLogger
}

// newLogger creates a new logger instance
func newLogger(logDir string, verbose bool) (*Logger, error) {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Generate log filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile := filepath.Join(logDir, fmt.Sprintf("cct_%s.log", timestamp))
	stderrFile := filepath.Join(logDir, fmt.Sprintf("sdk_stderr_%s.log", timestamp))

	// Create log file
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	// Create logger
	logger := log.New(file, "", log.LstdFlags|log.Lshortfile)

	l := &Logger{
		file:       file,
		logger:     logger,
		verbose:    verbose,
		logFile:    logFile,
		stderrFile: stderrFile,
	}

	// Log initialization
	l.Info("Logger initialized (verbose=%v)", verbose)
	l.Info("Log file: %s", logFile)
	l.Info("SDK stderr file: %s", stderrFile)

	return l, nil
}

// Debug logs a debug message (only when verbose is enabled)
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.verbose {
		l.log("DEBUG", format, args...)
	}
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log("INFO", format, args...)
}

// Warning logs a warning message
func (l *Logger) Warning(format string, args ...interface{}) {
	l.log("WARNING", format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log("ERROR", format, args...)
}

// log is the internal logging method
func (l *Logger) log(level string, format string, args ...interface{}) {
	if l == nil || l.logger == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	msg := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] %s", level, msg)
}

// GetLogFilePath returns the path to the main log file
func (l *Logger) GetLogFilePath() string {
	if l == nil {
		return ""
	}
	return l.logFile
}

// GetStderrFilePath returns the path to the stderr log file
func (l *Logger) GetStderrFilePath() string {
	if l == nil {
		return ""
	}
	return l.stderrFile
}

// RedirectStderr redirects os.Stderr to the stderr log file
// This captures all stderr output (including from claude-agent-sdk-go) to a file
func (l *Logger) RedirectStderr() error {
	if l == nil {
		return fmt.Errorf("logger not initialized")
	}

	// Create stderr log file
	stderrFile, err := os.OpenFile(l.stderrFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create stderr log file: %w", err)
	}

	// Save original stderr
	originalStderr := os.Stderr

	// Create a multi-writer that writes to both the file and original stderr
	// This way we can still see errors in the console if needed
	multiWriter := io.MultiWriter(stderrFile, originalStderr)

	// Create a pipe
	reader, writer, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %w", err)
	}

	// Set os.Stderr to the writer end of the pipe
	os.Stderr = writer

	// Start goroutine to copy from pipe to both file and original stderr
	go func() {
		_, _ = io.Copy(multiWriter, reader)
	}()

	l.Info("Stderr redirected to: %s", l.stderrFile)
	return nil
}

// Close closes the log file
func (l *Logger) Close() error {
	if l == nil || l.file == nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.Info("Logger shutting down")
	return l.file.Close()
}

// IsVerbose returns whether verbose logging is enabled
func (l *Logger) IsVerbose() bool {
	if l == nil {
		return false
	}
	return l.verbose
}

// Helper functions for global logger

// Debug logs a debug message to the global logger
func Debug(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(format, args...)
	}
}

// Info logs an info message to the global logger
func Info(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Info(format, args...)
	}
}

// Warning logs a warning message to the global logger
func Warning(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Warning(format, args...)
	}
}

// Error logs an error message to the global logger
func Error(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Error(format, args...)
	}
}
