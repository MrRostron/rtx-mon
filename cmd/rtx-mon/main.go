// Package main is the entry point for the rtx-mon application.
// This is a lightweight terminal-based monitor for NVIDIA RTX GPUs,
// built with the Bubble Tea framework for a responsive TUI.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/MrRostron/rtx-mon/internal/gpu"
	"github.com/MrRostron/rtx-mon/internal/ui"
)

// main initializes the application, sets up the GPU monitoring backend,
// launches the Bubble Tea TUI, and handles graceful shutdown.
func main() {
	// Ensure GPU resources (e.g. NVML context or any open handles) are properly
	// released when the program exits, even if it panics.
	defer gpu.Shutdown()

	// Create a new Bubble Tea program with the initial UI model.
	// The model (defined in internal/ui) manages state, updates, and rendering.
	p := tea.NewProgram(ui.InitialModel())

	// Run the program. This blocks until the user quits (e.g. via Ctrl+C or q).
	// Any runtime error from the TUI loop is captured here.
	if _, err := p.Run(); err != nil {
		// Print a user-friendly error message and exit with non-zero code
		// so the shell knows something went wrong.
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
