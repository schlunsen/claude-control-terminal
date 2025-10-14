package analytics

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestNewFileWatcher(t *testing.T) {
	tmpDir := t.TempDir()

	callback := func() error {
		return nil
	}

	fw, err := NewFileWatcher(tmpDir, callback)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}

	if fw == nil {
		t.Fatal("NewFileWatcher returned nil")
	}

	if fw.claudeDir != tmpDir {
		t.Errorf("Expected claudeDir %s, got %s", tmpDir, fw.claudeDir)
	}

	if fw.dataRefreshCallback == nil {
		t.Error("dataRefreshCallback should not be nil")
	}

	if fw.watcher == nil {
		t.Error("watcher should not be nil")
	}

	if fw.IsActive() {
		t.Error("watcher should not be active initially")
	}

	// Cleanup
	fw.Stop()
}

func TestNewFileWatcherWithOptions(t *testing.T) {
	tmpDir := t.TempDir()

	callback := func() error { return nil }

	tests := []struct {
		name  string
		quiet bool
	}{
		{"quiet mode", true},
		{"verbose mode", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fw, err := NewFileWatcherWithOptions(tmpDir, callback, tt.quiet)
			if err != nil {
				t.Fatalf("NewFileWatcherWithOptions failed: %v", err)
			}

			if fw.quiet != tt.quiet {
				t.Errorf("Expected quiet=%v, got %v", tt.quiet, fw.quiet)
			}

			fw.Stop()
		})
	}
}

func TestFileWatcher_Start(t *testing.T) {
	tmpDir := t.TempDir()

	callback := func() error { return nil }
	fw, err := NewFileWatcher(tmpDir, callback)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer fw.Stop()

	err = fw.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !fw.IsActive() {
		t.Error("watcher should be active after Start")
	}
}

func TestFileWatcher_Stop(t *testing.T) {
	tmpDir := t.TempDir()

	callback := func() error { return nil }
	fw, err := NewFileWatcherWithOptions(tmpDir, callback, true)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}

	fw.Start()

	err = fw.Stop()
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}

	if fw.IsActive() {
		t.Error("watcher should not be active after Stop")
	}
}

func TestFileWatcher_MultipleStops(t *testing.T) {
	tmpDir := t.TempDir()

	callback := func() error { return nil }
	fw, err := NewFileWatcherWithOptions(tmpDir, callback, true)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}

	fw.Start()

	// First stop
	err = fw.Stop()
	if err != nil {
		t.Errorf("First Stop failed: %v", err)
	}

	// Second stop should not error
	err = fw.Stop()
	if err != nil {
		t.Errorf("Second Stop failed: %v", err)
	}
}

func TestFileWatcher_IsActiveConcurrent(t *testing.T) {
	tmpDir := t.TempDir()

	callback := func() error { return nil }
	fw, err := NewFileWatcherWithOptions(tmpDir, callback, true)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer fw.Stop()

	fw.Start()

	// Test concurrent access to IsActive
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = fw.IsActive()
			}
		}()
	}

	wg.Wait()
}

func TestFileWatcher_TriggerRefresh(t *testing.T) {
	tmpDir := t.TempDir()

	callCount := 0
	var mu sync.Mutex

	callback := func() error {
		mu.Lock()
		callCount++
		mu.Unlock()
		return nil
	}

	fw, err := NewFileWatcherWithOptions(tmpDir, callback, true)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer fw.Stop()

	fw.Start()

	// Trigger refresh manually
	fw.triggerRefresh()

	// Wait a bit for callback to execute
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	count := callCount
	mu.Unlock()

	if count != 1 {
		t.Errorf("Expected callback to be called 1 time, got %d", count)
	}
}

func TestFileWatcher_CallbackError(t *testing.T) {
	tmpDir := t.TempDir()

	expectedError := errors.New("callback error")
	callback := func() error {
		return expectedError
	}

	fw, err := NewFileWatcherWithOptions(tmpDir, callback, true)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer fw.Stop()

	fw.Start()

	// Trigger refresh with error - should not panic
	fw.triggerRefresh()

	// Wait a bit
	time.Sleep(50 * time.Millisecond)
}

func TestFileWatcher_NilCallback(t *testing.T) {
	tmpDir := t.TempDir()

	fw, err := NewFileWatcher(tmpDir, nil)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer fw.Stop()

	fw.Start()

	// Trigger refresh with nil callback - should not panic
	fw.triggerRefresh()

	// Wait a bit
	time.Sleep(50 * time.Millisecond)
}

func TestFileWatcher_FileChange(t *testing.T) {
	tmpDir := t.TempDir()

	callCount := 0
	var mu sync.Mutex

	callback := func() error {
		mu.Lock()
		callCount++
		mu.Unlock()
		return nil
	}

	fw, err := NewFileWatcherWithOptions(tmpDir, callback, true)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer fw.Stop()

	fw.Start()

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Create a .jsonl file
	testFile := filepath.Join(tmpDir, "test.jsonl")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Wait for debounce and callback
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	count := callCount
	mu.Unlock()

	if count < 1 {
		t.Errorf("Expected callback to be called at least once, got %d", count)
	}
}

func TestFileWatcher_NonJsonlFileIgnored(t *testing.T) {
	tmpDir := t.TempDir()

	callCount := 0
	var mu sync.Mutex

	callback := func() error {
		mu.Lock()
		callCount++
		mu.Unlock()
		return nil
	}

	fw, err := NewFileWatcherWithOptions(tmpDir, callback, true)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer fw.Stop()

	fw.Start()

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Create a non-.jsonl file
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Wait a bit
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	count := callCount
	mu.Unlock()

	// Non-.jsonl files should not trigger callback
	if count > 0 {
		t.Errorf("Expected callback to not be called for non-.jsonl file, but was called %d times", count)
	}
}

func TestFileWatcher_PeriodicRefresh(t *testing.T) {
	tmpDir := t.TempDir()

	callCount := 0
	var mu sync.Mutex

	callback := func() error {
		mu.Lock()
		callCount++
		mu.Unlock()
		return nil
	}

	fw, err := NewFileWatcherWithOptions(tmpDir, callback, true)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer fw.Stop()

	// Note: In production, periodic refresh is every 2 minutes
	// In tests, we just verify the function exists and can be called
	// We can't easily test the actual timer without mocking time

	fw.Start()

	// Wait a short time
	time.Sleep(100 * time.Millisecond)

	// Manually test the periodicRefresh goroutine is running
	// by verifying the watcher is active
	if !fw.IsActive() {
		t.Error("Expected watcher to be active")
	}
}

func TestFileWatcher_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()

	callback := func() error { return nil }
	fw, err := NewFileWatcherWithOptions(tmpDir, callback, true)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}

	fw.Start()

	// Verify context is active
	select {
	case <-fw.ctx.Done():
		t.Error("Context should not be cancelled initially")
	default:
		// Context is active
	}

	// Stop and verify context is cancelled
	fw.Stop()

	select {
	case <-fw.ctx.Done():
		// Context is cancelled
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should be cancelled after Stop")
	}
}

func TestFileWatcher_SetActive(t *testing.T) {
	tmpDir := t.TempDir()

	callback := func() error { return nil }
	fw, err := NewFileWatcherWithOptions(tmpDir, callback, true)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer fw.Stop()

	if fw.IsActive() {
		t.Error("Expected inactive initially")
	}

	fw.setActive(true)
	if !fw.IsActive() {
		t.Error("Expected active after setActive(true)")
	}

	fw.setActive(false)
	if fw.IsActive() {
		t.Error("Expected inactive after setActive(false)")
	}
}

func TestFileWatcher_ConcurrentSetActive(t *testing.T) {
	tmpDir := t.TempDir()

	callback := func() error { return nil }
	fw, err := NewFileWatcherWithOptions(tmpDir, callback, true)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer fw.Stop()

	// Test concurrent setActive calls
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				fw.setActive(true)
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				fw.setActive(false)
			}
		}()
	}

	wg.Wait()
}

func TestFileWatcher_StopTimeout(t *testing.T) {
	tmpDir := t.TempDir()

	// This test verifies the timeout behavior in Stop()
	// Create a callback that might take a while
	callback := func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	fw, err := NewFileWatcherWithOptions(tmpDir, callback, true)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}

	fw.Start()

	// Stop should complete even if goroutines are busy
	stopDone := make(chan error, 1)
	go func() {
		stopDone <- fw.Stop()
	}()

	select {
	case err := <-stopDone:
		if err != nil {
			t.Errorf("Stop failed: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Error("Stop did not complete within timeout")
	}
}

func TestFileWatcher_InvalidDirectory(t *testing.T) {
	// Use a directory that definitely doesn't exist
	nonExistentDir := "/this/directory/does/not/exist/at/all"

	callback := func() error { return nil }
	fw, err := NewFileWatcher(nonExistentDir, callback)
	if err != nil {
		t.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer fw.Stop()

	// Start should fail with invalid directory
	err = fw.Start()
	if err == nil {
		t.Error("Expected error when starting watcher with invalid directory")
	}
}

// BenchmarkFileWatcher_TriggerRefresh benchmarks the refresh trigger
func BenchmarkFileWatcher_TriggerRefresh(b *testing.B) {
	tmpDir := b.TempDir()

	callCount := 0
	callback := func() error {
		callCount++
		return nil
	}

	fw, err := NewFileWatcherWithOptions(tmpDir, callback, true)
	if err != nil {
		b.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer fw.Stop()

	fw.Start()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fw.triggerRefresh()
	}
}

// BenchmarkFileWatcher_IsActive benchmarks the IsActive method
func BenchmarkFileWatcher_IsActive(b *testing.B) {
	tmpDir := b.TempDir()

	callback := func() error { return nil }
	fw, err := NewFileWatcherWithOptions(tmpDir, callback, true)
	if err != nil {
		b.Fatalf("NewFileWatcher failed: %v", err)
	}
	defer fw.Stop()

	fw.Start()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fw.IsActive()
	}
}
