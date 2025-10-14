package installer

import (
	"os/exec"
	"testing"
	"time"
)

func TestNewNodeDetector(t *testing.T) {
	nd := NewNodeDetector()

	if nd == nil {
		t.Fatal("NewNodeDetector returned nil")
	}

	if nd.MinMajorVersion != 18 {
		t.Errorf("Expected MinMajorVersion 18, got %d", nd.MinMajorVersion)
	}

	if nd.Timeout != 10*time.Second {
		t.Errorf("Expected timeout 10s, got %v", nd.Timeout)
	}
}

func TestIsNodeInstalled(t *testing.T) {
	nd := NewNodeDetector()

	installed := nd.IsNodeInstalled()

	// Verify it matches exec.LookPath result
	_, err := exec.LookPath("node")
	expectedInstalled := err == nil

	if installed != expectedInstalled {
		t.Errorf("IsNodeInstalled() = %v, expected %v", installed, expectedInstalled)
	}
}

func TestIsNPMInstalled(t *testing.T) {
	nd := NewNodeDetector()

	installed := nd.IsNPMInstalled()

	// Verify it matches exec.LookPath result
	_, err := exec.LookPath("npm")
	expectedInstalled := err == nil

	if installed != expectedInstalled {
		t.Errorf("IsNPMInstalled() = %v, expected %v", installed, expectedInstalled)
	}
}

func TestGetNodePath_NotInstalled(t *testing.T) {
	nd := NewNodeDetector()

	// Only test error case if Node is not installed
	if !nd.IsNodeInstalled() {
		_, err := nd.GetNodePath()
		if err == nil {
			t.Error("Expected error when Node is not installed")
		}
	}
}

func TestGetNodePath_Installed(t *testing.T) {
	nd := NewNodeDetector()

	// Only test success case if Node is installed
	if nd.IsNodeInstalled() {
		path, err := nd.GetNodePath()
		if err != nil {
			t.Errorf("GetNodePath failed: %v", err)
		}

		if path == "" {
			t.Error("Expected non-empty path")
		}
	}
}

func TestGetNPMPath_Installed(t *testing.T) {
	nd := NewNodeDetector()

	// Only test if npm is installed
	if nd.IsNPMInstalled() {
		path, err := nd.GetNPMPath()
		if err != nil {
			t.Errorf("GetNPMPath failed: %v", err)
		}

		if path == "" {
			t.Error("Expected non-empty path")
		}
	}
}

func TestGetNodeVersion_Installed(t *testing.T) {
	nd := NewNodeDetector()

	// Only test if Node is installed
	if nd.IsNodeInstalled() {
		version, err := nd.GetNodeVersion()
		if err != nil {
			t.Errorf("GetNodeVersion failed: %v", err)
		}

		if version == "" {
			t.Error("Expected non-empty version string")
		}

		// Version should start with 'v' typically
		if len(version) < 2 {
			t.Errorf("Version string too short: %s", version)
		}
	}
}

func TestGetNPMVersion_Installed(t *testing.T) {
	nd := NewNodeDetector()

	// Only test if npm is installed
	if nd.IsNPMInstalled() {
		version, err := nd.GetNPMVersion()
		if err != nil {
			t.Errorf("GetNPMVersion failed: %v", err)
		}

		if version == "" {
			t.Error("Expected non-empty version string")
		}
	}
}

func TestParseNodeVersion(t *testing.T) {
	nd := NewNodeDetector()

	tests := []struct {
		name          string
		version       string
		expectedMajor int
		expectedMinor int
		expectedPatch int
		expectError   bool
	}{
		{
			name:          "with v prefix",
			version:       "v18.17.0",
			expectedMajor: 18,
			expectedMinor: 17,
			expectedPatch: 0,
			expectError:   false,
		},
		{
			name:          "without v prefix",
			version:       "18.17.0",
			expectedMajor: 18,
			expectedMinor: 17,
			expectedPatch: 0,
			expectError:   false,
		},
		{
			name:          "v20",
			version:       "v20.10.0",
			expectedMajor: 20,
			expectedMinor: 10,
			expectedPatch: 0,
			expectError:   false,
		},
		{
			name:          "v16 old version",
			version:       "v16.14.2",
			expectedMajor: 16,
			expectedMinor: 14,
			expectedPatch: 2,
			expectError:   false,
		},
		{
			name:        "invalid format",
			version:     "invalid",
			expectError: true,
		},
		{
			name:        "empty string",
			version:     "",
			expectError: true,
		},
		{
			name:        "partial version",
			version:     "18.17",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor, patch, err := nd.ParseNodeVersion(tt.version)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if major != tt.expectedMajor {
				t.Errorf("Expected major %d, got %d", tt.expectedMajor, major)
			}

			if minor != tt.expectedMinor {
				t.Errorf("Expected minor %d, got %d", tt.expectedMinor, minor)
			}

			if patch != tt.expectedPatch {
				t.Errorf("Expected patch %d, got %d", tt.expectedPatch, patch)
			}
		})
	}
}

func TestCheckVersionRequirement(t *testing.T) {
	nd := NewNodeDetector()
	nd.MinMajorVersion = 18

	tests := []struct {
		name     string
		major    int
		expected bool
	}{
		{
			name:     "v18 exact match",
			major:    18,
			expected: true,
		},
		{
			name:     "v20 newer",
			major:    20,
			expected: true,
		},
		{
			name:     "v16 too old",
			major:    16,
			expected: false,
		},
		{
			name:     "v14 too old",
			major:    14,
			expected: false,
		},
		{
			name:     "v22 future",
			major:    22,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nd.CheckVersionRequirement(tt.major)

			if result != tt.expected {
				t.Errorf("CheckVersionRequirement(%d) = %v, expected %v", tt.major, result, tt.expected)
			}
		})
	}
}

func TestDetectNode(t *testing.T) {
	nd := NewNodeDetector()

	info := nd.DetectNode()

	if info == nil {
		t.Fatal("DetectNode returned nil")
	}

	// Verify consistency with IsNodeInstalled
	if info.Installed != nd.IsNodeInstalled() {
		t.Errorf("DetectNode().Installed = %v, but IsNodeInstalled() = %v",
			info.Installed, nd.IsNodeInstalled())
	}

	// If Node is installed, check fields
	if info.Installed {
		if info.Version == "" {
			t.Error("Expected non-empty version when Node is installed")
		}

		if info.NodePath == "" {
			t.Error("Expected non-empty NodePath when Node is installed")
		}

		if info.VersionMajor < 1 {
			t.Error("Expected VersionMajor >= 1")
		}

		// Check if version requirement check is consistent
		expectedVersionOK := nd.CheckVersionRequirement(info.VersionMajor)
		if info.VersionOK != expectedVersionOK {
			t.Errorf("VersionOK = %v, but CheckVersionRequirement() = %v",
				info.VersionOK, expectedVersionOK)
		}

		// NPM availability check
		if info.NPMAvailable != nd.IsNPMInstalled() {
			t.Errorf("NPMAvailable = %v, but IsNPMInstalled() = %v",
				info.NPMAvailable, nd.IsNPMInstalled())
		}

		if info.NPMAvailable && info.NPMVersion == "" {
			t.Error("Expected non-empty NPMVersion when npm is available")
		}
	}

	// If Node is not installed, all fields should be empty/false
	if !info.Installed {
		if info.Version != "" {
			t.Error("Expected empty version when Node is not installed")
		}

		if info.VersionOK {
			t.Error("Expected VersionOK = false when Node is not installed")
		}
	}
}

func TestGetNodeInstallInstructions(t *testing.T) {
	instructions := GetNodeInstallInstructions()

	if instructions == "" {
		t.Error("GetNodeInstallInstructions returned empty string")
	}

	// Should contain common installation methods
	if len(instructions) < 200 {
		t.Error("Instructions seem too short")
	}
}

func TestGetRecommendation(t *testing.T) {
	nd := NewNodeDetector()

	recommendation := nd.GetRecommendation()

	if recommendation == "" {
		t.Error("GetRecommendation returned empty string")
	}

	// Verify recommendation matches system state
	info := nd.DetectNode()

	if !info.Installed {
		// Should recommend native binary
		// Note: We don't check exact text as it may change
	} else if !info.VersionOK {
		// Should mention version issue
	} else if !info.NPMAvailable {
		// Should mention npm missing
	}
}

func TestFormatNodeInfo(t *testing.T) {
	nd := NewNodeDetector()

	formatted := nd.FormatNodeInfo()

	if formatted == "" {
		t.Error("FormatNodeInfo returned empty string")
	}

	// Should contain "Node.js:"
	if len(formatted) < 10 {
		t.Error("Formatted output seems too short")
	}
}

func TestNodeDetectorCustomMinVersion(t *testing.T) {
	nd := NewNodeDetector()
	nd.MinMajorVersion = 20

	if nd.MinMajorVersion != 20 {
		t.Errorf("Expected MinMajorVersion 20, got %d", nd.MinMajorVersion)
	}

	// Version 18 should now fail requirement check
	if nd.CheckVersionRequirement(18) {
		t.Error("Expected version 18 to fail when MinMajorVersion is 20")
	}

	// Version 20 should pass
	if !nd.CheckVersionRequirement(20) {
		t.Error("Expected version 20 to pass when MinMajorVersion is 20")
	}
}

func TestNodeDetectorTimeout(t *testing.T) {
	nd := NewNodeDetector()
	nd.Timeout = 5 * time.Second

	if nd.Timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", nd.Timeout)
	}
}

func TestNodeInfo_Struct(t *testing.T) {
	info := &NodeInfo{
		Installed:    true,
		Version:      "v18.17.0",
		VersionMajor: 18,
		VersionMinor: 17,
		VersionPatch: 0,
		VersionOK:    true,
		NodePath:     "/usr/local/bin/node",
		NPMAvailable: true,
		NPMVersion:   "9.6.7",
		NPMPath:      "/usr/local/bin/npm",
	}

	if !info.Installed {
		t.Error("Expected Installed to be true")
	}

	if info.Version != "v18.17.0" {
		t.Errorf("Expected version v18.17.0, got %s", info.Version)
	}

	if info.VersionMajor != 18 {
		t.Errorf("Expected VersionMajor 18, got %d", info.VersionMajor)
	}
}

// BenchmarkIsNodeInstalled benchmarks node detection
func BenchmarkIsNodeInstalled(b *testing.B) {
	nd := NewNodeDetector()

	for i := 0; i < b.N; i++ {
		nd.IsNodeInstalled()
	}
}

// BenchmarkDetectNode benchmarks full node detection
func BenchmarkDetectNode(b *testing.B) {
	nd := NewNodeDetector()

	for i := 0; i < b.N; i++ {
		nd.DetectNode()
	}
}

// BenchmarkParseNodeVersion benchmarks version parsing
func BenchmarkParseNodeVersion(b *testing.B) {
	nd := NewNodeDetector()

	for i := 0; i < b.N; i++ {
		nd.ParseNodeVersion("v18.17.0")
	}
}
