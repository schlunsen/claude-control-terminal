// Package main is the entry point for the Claude Control Terminal (CCT) application.
// CCT is a high-performance Go port of the claude-code-templates CLI tool,
// providing component templates, analytics dashboards, and real-time monitoring
// for Claude Code projects.
package main

import (
	"fmt"
	"os"

	"github.com/schlunsen/claude-control-terminal/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
