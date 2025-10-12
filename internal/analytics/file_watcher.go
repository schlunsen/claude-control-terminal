package analytics

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher handles file system watching for real-time updates
type FileWatcher struct {
	watcher              *fsnotify.Watcher
	claudeDir            string
	dataRefreshCallback  func() error
	isActive             bool
	stopChan             chan bool
	quiet                bool // Suppress output when running in TUI
}

// NewFileWatcher creates a new FileWatcher
func NewFileWatcher(claudeDir string, refreshCallback func() error) (*FileWatcher, error) {
	return NewFileWatcherWithOptions(claudeDir, refreshCallback, false)
}

// NewFileWatcherWithOptions creates a new FileWatcher with options
func NewFileWatcherWithOptions(claudeDir string, refreshCallback func() error, quiet bool) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	fw := &FileWatcher{
		watcher:             watcher,
		claudeDir:           claudeDir,
		dataRefreshCallback: refreshCallback,
		isActive:            false,
		stopChan:            make(chan bool),
		quiet:               quiet,
	}

	return fw, nil
}

// Start begins watching for file changes
func (fw *FileWatcher) Start() error {
	// Add the Claude directory to watch
	err := fw.watcher.Add(fw.claudeDir)
	if err != nil {
		return fmt.Errorf("failed to watch directory: %w", err)
	}

	// Watch subdirectories for .jsonl files
	if !fw.quiet {
		pattern := filepath.Join(fw.claudeDir, "**/*.jsonl")
		fmt.Printf("👀 Watching for changes in: %s\n", pattern)
	}

	fw.isActive = true

	// Start watching in a goroutine
	go fw.watchLoop()

	// Start periodic refresh
	go fw.periodicRefresh()

	return nil
}

// watchLoop handles file system events
func (fw *FileWatcher) watchLoop() {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			// Only trigger on .jsonl file changes
			if filepath.Ext(event.Name) == ".jsonl" {
				if event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Create == fsnotify.Create {
					// Debounce: wait a bit to avoid multiple rapid refreshes
					time.Sleep(100 * time.Millisecond)
					fw.triggerRefresh()
				}
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			if !fw.quiet {
				fmt.Printf("⚠️  File watcher error: %v\n", err)
			}

		case <-fw.stopChan:
			return
		}
	}
}

// periodicRefresh triggers periodic data refreshes
func (fw *FileWatcher) periodicRefresh() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fw.triggerRefresh()
		case <-fw.stopChan:
			return
		}
	}
}

// triggerRefresh calls the refresh callback
func (fw *FileWatcher) triggerRefresh() {
	if fw.dataRefreshCallback != nil && fw.isActive {
		if err := fw.dataRefreshCallback(); err != nil {
			if !fw.quiet {
				fmt.Printf("⚠️  Error during refresh: %v\n", err)
			}
		}
	}
}

// Stop stops the file watcher
func (fw *FileWatcher) Stop() error {
	if !fw.quiet {
		fmt.Println("🛑 Stopping file watcher...")
	}

	fw.isActive = false
	close(fw.stopChan)

	if fw.watcher != nil {
		return fw.watcher.Close()
	}

	return nil
}

// IsActive returns whether the watcher is active
func (fw *FileWatcher) IsActive() bool {
	return fw.isActive
}
