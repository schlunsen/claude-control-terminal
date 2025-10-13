package analytics

import (
	"testing"
	"time"
)

func TestNewProcessDetector(t *testing.T) {
	pd := NewProcessDetector()

	if pd == nil {
		t.Fatal("NewProcessDetector returned nil")
	}

	if pd.cache.TTL != 500*time.Millisecond {
		t.Errorf("expected TTL 500ms, got %v", pd.cache.TTL)
	}

	if pd.cache.Data != nil {
		t.Error("cache data should be nil initially")
	}
}

func TestParseProcessOutput(t *testing.T) {
	pd := NewProcessDetector()

	tests := []struct {
		name          string
		output        string
		expectedCount int
	}{
		{
			name:          "empty output",
			output:        "",
			expectedCount: 0,
		},
		{
			name: "valid claude process",
			output: `USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
user     12345  0.0  0.1  1000  2000 pts/0    S+   10:00   0:00 claude`,
			expectedCount: 1,
		},
		{
			name: "claude with arguments",
			output: `USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
user     12345  0.0  0.1  1000  2000 pts/0    S+   10:00   0:00 claude --help`,
			expectedCount: 1,
		},
		{
			name: "claude with cwd",
			output: `USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
user     12345  0.0  0.1  1000  2000 pts/0    S+   10:00   0:00 claude --cwd /test/path`,
			expectedCount: 1,
		},
		{
			name: "chrome crashpad should be filtered",
			output: `USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
user     12345  0.0  0.1  1000  2000 pts/0    S+   10:00   0:00 /Applications/Google Chrome.app/Contents/Frameworks/Google Chrome Framework.framework/Helpers/chrome_crashpad_handler --monitor-self claude`,
			expectedCount: 0,
		},
		{
			name: "analytics process should be filtered",
			output: `USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
user     12345  0.0  0.1  1000  2000 pts/0    S+   10:00   0:00 ./cct --analytics claude`,
			expectedCount: 0,
		},
		{
			name: "grep process should be filtered",
			output: `USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
user     12345  0.0  0.1  1000  2000 pts/0    S+   10:00   0:00 grep claude`,
			expectedCount: 0,
		},
		{
			name: "claude app should be filtered",
			output: `USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
user     12345  0.0  0.1  1000  2000 pts/0    S+   10:00   0:00 /Applications/Claude.app/Contents/MacOS/Claude`,
			expectedCount: 0,
		},
		{
			name: "multiple processes",
			output: `USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
user     12345  0.0  0.1  1000  2000 pts/0    S+   10:00   0:00 claude
user     67890  0.0  0.1  1000  2000 pts/1    S+   10:01   0:00 /usr/bin/claude --help`,
			expectedCount: 2,
		},
		{
			name: "insufficient fields",
			output: `USER PID COMMAND
user 123 claude`,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processes := pd.parseProcessOutput(tt.output)
			if len(processes) != tt.expectedCount {
				t.Errorf("expected %d processes, got %d", tt.expectedCount, len(processes))
			}
		})
	}
}

func TestParseProcessOutputWorkingDir(t *testing.T) {
	pd := NewProcessDetector()

	output := `USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
user     12345  0.0  0.1  1000  2000 pts/0    S+   10:00   0:00 claude --cwd=/home/user/project`

	processes := pd.parseProcessOutput(output)

	if len(processes) != 1 {
		t.Fatalf("expected 1 process, got %d", len(processes))
	}

	process := processes[0]
	if process.WorkingDir == "unknown" {
		t.Error("expected working directory to be parsed")
	}

	if process.PID != "12345" {
		t.Errorf("expected PID '12345', got %q", process.PID)
	}

	if process.User != "user" {
		t.Errorf("expected user 'user', got %q", process.User)
	}

	if process.Status != "running" {
		t.Errorf("expected status 'running', got %q", process.Status)
	}
}

func TestGetCachedProcesses(t *testing.T) {
	pd := NewProcessDetector()

	// Initially should return empty
	cached := pd.GetCachedProcesses()
	if len(cached) != 0 {
		t.Errorf("expected empty cache, got %d processes", len(cached))
	}

	// Set some cache data
	pd.cache = ProcessCache{
		Data: []Process{
			{
				PID:     "12345",
				Command: "claude",
				Status:  "running",
			},
		},
		Timestamp: time.Now(),
		TTL:       500 * time.Millisecond,
	}

	// Should return cached data
	cached = pd.GetCachedProcesses()
	if len(cached) != 1 {
		t.Errorf("expected 1 cached process, got %d", len(cached))
	}

	if cached[0].PID != "12345" {
		t.Errorf("expected PID '12345', got %q", cached[0].PID)
	}
}

func TestGetCachedProcessesExpired(t *testing.T) {
	pd := NewProcessDetector()

	// Set expired cache data
	pd.cache = ProcessCache{
		Data: []Process{
			{
				PID:     "12345",
				Command: "claude",
				Status:  "running",
			},
		},
		Timestamp: time.Now().Add(-1 * time.Second),
		TTL:       500 * time.Millisecond,
	}

	// Should return empty due to expiration
	cached := pd.GetCachedProcesses()
	if len(cached) != 0 {
		t.Errorf("expected empty cache (expired), got %d processes", len(cached))
	}
}

func TestClearCacheProcessDetector(t *testing.T) {
	pd := NewProcessDetector()

	// Set some cache data
	pd.cache = ProcessCache{
		Data: []Process{
			{
				PID:     "12345",
				Command: "claude",
				Status:  "running",
			},
		},
		Timestamp: time.Now(),
		TTL:       500 * time.Millisecond,
	}

	// Clear cache
	pd.ClearCache()

	// Verify cache is cleared
	if pd.cache.Data != nil {
		t.Error("cache data should be nil after clear")
	}

	if !pd.cache.Timestamp.IsZero() {
		t.Error("cache timestamp should be zero after clear")
	}
}

func TestProcessStruct(t *testing.T) {
	now := time.Now()
	p := Process{
		PID:        "12345",
		Command:    "claude --help",
		WorkingDir: "/test/path",
		StartTime:  now,
		Status:     "running",
		User:       "testuser",
	}

	if p.PID != "12345" {
		t.Errorf("expected PID '12345', got %q", p.PID)
	}

	if p.Command != "claude --help" {
		t.Errorf("expected command 'claude --help', got %q", p.Command)
	}

	if p.WorkingDir != "/test/path" {
		t.Errorf("expected working dir '/test/path', got %q", p.WorkingDir)
	}

	if p.Status != "running" {
		t.Errorf("expected status 'running', got %q", p.Status)
	}

	if p.User != "testuser" {
		t.Errorf("expected user 'testuser', got %q", p.User)
	}
}

func TestProcessCacheStruct(t *testing.T) {
	now := time.Now()
	ttl := 500 * time.Millisecond

	cache := ProcessCache{
		Data: []Process{
			{PID: "12345", Command: "claude"},
		},
		Timestamp: now,
		TTL:       ttl,
	}

	if len(cache.Data) != 1 {
		t.Errorf("expected 1 process in cache, got %d", len(cache.Data))
	}

	if cache.TTL != ttl {
		t.Errorf("expected TTL %v, got %v", ttl, cache.TTL)
	}

	if !cache.Timestamp.Equal(now) {
		t.Error("timestamp mismatch")
	}
}

func TestProcessStatsStruct(t *testing.T) {
	stats := ProcessStats{
		Total:                 5,
		WithKnownWorkingDir:   3,
		WithUnknownWorkingDir: 2,
		Processes: []Process{
			{PID: "1", WorkingDir: "/path1"},
			{PID: "2", WorkingDir: "unknown"},
			{PID: "3", WorkingDir: "/path2"},
			{PID: "4", WorkingDir: "unknown"},
			{PID: "5", WorkingDir: "/path3"},
		},
	}

	if stats.Total != 5 {
		t.Errorf("expected total 5, got %d", stats.Total)
	}

	if stats.WithKnownWorkingDir != 3 {
		t.Errorf("expected 3 with known dir, got %d", stats.WithKnownWorkingDir)
	}

	if stats.WithUnknownWorkingDir != 2 {
		t.Errorf("expected 2 with unknown dir, got %d", stats.WithUnknownWorkingDir)
	}

	if len(stats.Processes) != 5 {
		t.Errorf("expected 5 processes, got %d", len(stats.Processes))
	}
}

func TestDetectRunningClaudeProcesses(t *testing.T) {
	pd := NewProcessDetector()

	// This test will actually run ps aux on the system
	// We can't guarantee Claude is running, so we just test it doesn't error
	processes, err := pd.DetectRunningClaudeProcesses()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// processes might be empty, which is fine
	if processes == nil {
		t.Error("processes should not be nil, even if empty")
	}
}

func TestHasActiveProcesses(t *testing.T) {
	pd := NewProcessDetector()

	// Test with empty cache
	hasActive, err := pd.HasActiveProcesses()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Result depends on actual system state, just verify it doesn't crash
	_ = hasActive
}

func TestGetProcessStats(t *testing.T) {
	pd := NewProcessDetector()

	// Test getting stats
	stats, err := pd.GetProcessStats()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if stats == nil {
		t.Fatal("stats should not be nil")
	}

	// Verify structure
	if stats.Total < 0 {
		t.Error("total should not be negative")
	}

	if stats.WithKnownWorkingDir < 0 {
		t.Error("WithKnownWorkingDir should not be negative")
	}

	if stats.WithUnknownWorkingDir < 0 {
		t.Error("WithUnknownWorkingDir should not be negative")
	}

	if stats.Total != stats.WithKnownWorkingDir+stats.WithUnknownWorkingDir {
		t.Error("total should equal sum of known and unknown")
	}

	if stats.Processes == nil {
		t.Error("Processes should not be nil")
	}
}
