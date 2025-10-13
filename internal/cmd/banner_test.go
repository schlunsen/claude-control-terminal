package cmd

import (
	"strings"
	"testing"
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
	spinner := ShowSpinner("Testing spinner...")

	if spinner == nil {
		t.Error("ShowSpinner returned nil")
	}

	// Stop the spinner
	spinner.Stop()
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
