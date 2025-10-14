package analytics

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher handles file system watching for real-time updates.
// It monitors .jsonl files in the Claude directory and triggers callbacks on changes.
// Safe for concurrent use.
type FileWatcher struct {
	watcher              *fsnotify.Watcher
	claudeDir            string
	dataRefreshCallback  func() error
	isActive             bool
	isActiveMu           sync.RWMutex
	ctx                  context.Context
	cancel               context.CancelFunc
	wg                   sync.WaitGroup
	quiet                bool // Suppress output when running in TUI
}

// NewFileWatcher creates a new FileWatcher with default options.
func NewFileWatcher(claudeDir string, refreshCallback func() error) (*FileWatcher, error) {
	return NewFileWatcherWithOptions(claudeDir, refreshCallback, false)
}

// NewFileWatcherWithOptions creates a new FileWatcher with custom options.
// The quiet parameter suppresses console output when running in TUI mode.
func NewFileWatcherWithOptions(claudeDir string, refreshCallback func() error, quiet bool) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	fw := &FileWatcher{
		watcher:             watcher,
		claudeDir:           claudeDir,
		dataRefreshCallback: refreshCallback,
		isActive:            false,
		ctx:                 ctx,
		cancel:              cancel,
		quiet:               quiet,
	}

	return fw, nil
}

// Start begins watching for file changes.
// It spawns a goroutine for event watching on .jsonl files.
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

	fw.setActive(true)

	// Start watching in goroutines
	fw.wg.Add(1) // Changed from 2 to 1 - removed periodic refresh
	go fw.watchLoop()
	// Disabled periodic refresh to prevent database duplication
	// go fw.periodicRefresh()

	return nil
}

// watchLoop handles file system events until context is cancelled.
func (fw *FileWatcher) watchLoop() {
	defer fw.wg.Done()

	for {
		select {
		case <-fw.ctx.Done():
			return

		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			// Only trigger on .jsonl file changes
			if filepath.Ext(event.Name) == ".jsonl" {
				if event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Create == fsnotify.Create {
					// Debounce: wait a bit to avoid multiple rapid refreshes
					select {
					case <-time.After(100 * time.Millisecond):
						fw.triggerRefresh()
					case <-fw.ctx.Done():
						return
					}
				}
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			if !fw.quiet {
				fmt.Printf("⚠️  File watcher error: %v\n", err)
			}
		}
	}
}

// periodicRefresh triggers periodic data refreshes every 2 minutes.
func (fw *FileWatcher) periodicRefresh() {
	defer fw.wg.Done()

	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-fw.ctx.Done():
			return
		case <-ticker.C:
			fw.triggerRefresh()
		}
	}
}

// triggerRefresh calls the refresh callback if the watcher is active.
func (fw *FileWatcher) triggerRefresh() {
	if fw.dataRefreshCallback != nil && fw.IsActive() {
		if err := fw.dataRefreshCallback(); err != nil {
			if !fw.quiet {
				fmt.Printf("⚠️  Error during refresh: %v\n", err)
			}
		}
	}
}

// Stop gracefully stops the file watcher and waits for goroutines to exit.
// It is safe to call Stop multiple times.
func (fw *FileWatcher) Stop() error {
	if !fw.quiet {
		fmt.Println("🛑 Stopping file watcher...")
	}

	fw.setActive(false)
	fw.cancel()

	// Wait for goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		fw.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Goroutines finished cleanly
	case <-time.After(5 * time.Second):
		// Timeout waiting for goroutines
		if !fw.quiet {
			fmt.Println("⚠️  Warning: file watcher goroutines did not exit cleanly")
		}
	}

	if fw.watcher != nil {
		if err := fw.watcher.Close(); err != nil {
			return fmt.Errorf("failed to close watcher: %w", err)
		}
	}

	return nil
}

// IsActive returns whether the watcher is currently active.
// Thread-safe for concurrent access.
func (fw *FileWatcher) IsActive() bool {
	fw.isActiveMu.RLock()
	defer fw.isActiveMu.RUnlock()
	return fw.isActive
}

// setActive sets the active state with proper synchronization.
func (fw *FileWatcher) setActive(active bool) {
	fw.isActiveMu.Lock()
	fw.isActive = active
	fw.isActiveMu.Unlock()
}
