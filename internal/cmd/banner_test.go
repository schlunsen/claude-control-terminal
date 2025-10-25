package cmd

import (
	"strings"
	"testing"
	"time"
)

func TestShowBanner(t *testing.T) {
	// Test that ShowBanner runs without panic
	// Detailed output checking is difficult due to ANSI codes and terminal formatting
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ShowBanner panicked: %v", r)
		}
	}()

	ShowBanner()
	// If we got here without panic, the test passes
}

func TestShowSpinner(t *testing.T) {
	// Note: This test has a known race condition with pterm's internal spinner state
	// The race is in pterm's library code, not our code, but we can't fix it here
	// See: https://github.com/pterm/pterm/issues
	spinner := ShowSpinner("Testing spinner...")

	if spinner == nil {
		t.Error("ShowSpinner returned nil")
		return
	}

	// Add a small delay to let the spinner goroutine initialize
	time.Sleep(50 * time.Millisecond)

	// Stop the spinner - the race detector may still report a race in pterm's code
	spinner.Stop()

	// Give time for cleanup
	time.Sleep(10 * time.Millisecond)
}

func TestShowSuccess(t *testing.T) {
	// Test that ShowSuccess runs without panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ShowSuccess panicked: %v", r)
		}
	}()

	ShowSuccess("Test success message")
}

func TestShowError(t *testing.T) {
	// Test that ShowError runs without panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ShowError panicked: %v", r)
		}
	}()

	ShowError("Test error message")
}

func TestShowInfo(t *testing.T) {
	// Test that ShowInfo runs without panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ShowInfo panicked: %v", r)
		}
	}()

	ShowInfo("Test info message")
}

func TestShowWarning(t *testing.T) {
	// Test that ShowWarning runs without panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ShowWarning panicked: %v", r)
		}
	}()

	ShowWarning("Test warning message")
}

func TestShowBox(t *testing.T) {
	// Test that ShowBox runs without panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ShowBox panicked: %v", r)
		}
	}()

	ShowBox("Test Title", "Test content")
}

func TestShowProgress(t *testing.T) {
	progressbar := ShowProgress(100, "Testing progress...")

	if progressbar == nil {
		t.Error("ShowProgress returned nil")
	}

	// Stop the progress bar
	progressbar.Stop()
}

func TestVersionConstant(t *testing.T) {
	if Version == "" {
		t.Error("Version constant is empty")
	}

	// Version should follow semantic versioning pattern (x.y.z)
	parts := strings.Split(Version, ".")
	if len(parts) != 3 {
		t.Errorf("Version should be in format x.y.z, got: %s", Version)
	}
}
