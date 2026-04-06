package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/MrRostron/rtx-mon/internal/gpu"
	"github.com/MrRostron/rtx-mon/internal/ui"
)

func main() {
	defer gpu.Shutdown()

	p := tea.NewProgram(ui.InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
