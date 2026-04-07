// Package ui contains the Bubble Tea model, update logic, and view rendering for the RTX monitoring TUI.
package ui

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/MrRostron/rtx-mon/internal/config"
	"github.com/MrRostron/rtx-mon/internal/gpu"
)

// Model represents the complete state of the terminal UI.
// It holds both the current GPU metrics and configuration, as well as UI-related state.
type Model struct {
	Config config.Config

	// GPU Metrics
	Name       string
	Temp       float64
	MaxTemp    float64 // Used for temperature progress bar scaling (default 100°C)
	Power      float64
	PowerLimit float64
	Util       float64
	MemTotal   float64
	MemUsed    float64
	FanSpeed   float64

	// UI State
	Width     int
	Height    int
	IsDark    bool
	LastError string

	// Modal state
	ShowFanModal bool
}

// InitialModel creates and returns the initial application state.
// It loads configuration, initializes the GPU backend, and fetches the first set of metrics.
func InitialModel() Model {
	// Load configuration from file (falls back to defaults on error)
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Warning loading config: %v\n", err)
		cfg = config.Config{}
	}

	var lastErr string

	// Initialize NVML (NVIDIA GPU monitoring library)
	if err := gpu.InitNVML(); err != nil {
		lastErr = err.Error()
		fmt.Println("Warning:", lastErr)
	}

	// Fetch initial GPU metrics (ignore error - we'll show empty values or error message)
	data, _ := gpu.GetAllMetrics()

	return Model{
		Config:     cfg,
		Name:       data.Name,
		Temp:       data.Temp,
		MaxTemp:    100, // Reasonable default max temperature for progress bar
		Power:      data.Power,
		PowerLimit: data.PowerLimit,
		Util:       data.Util,
		MemTotal:   data.MemTotal,
		MemUsed:    data.MemUsed,
		FanSpeed:   data.FanSpeed,
		Width:      cfg.Width,
		IsDark:     lipgloss.HasDarkBackground(os.Stdin, os.Stdout),
		LastError:  lastErr,
	}
}

// updateMsg is an internal message used to trigger periodic GPU metric updates.
type updateMsg struct{}

// Init is called once when the program starts.
// It returns a command that schedules the first update tick.
func (m Model) Init() tea.Cmd {
	return tea.Tick(
		time.Duration(m.Config.General.UpdateIntervalMs)*time.Millisecond,
		func(time.Time) tea.Msg { return updateMsg{} },
	)
}

// Update handles all incoming messages and updates the model accordingly.
// This is the core of the Bubble Tea event loop.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {

		case "ctrl+c", "q", "Q":
			// Quit the application
					return m, tea.Quit

		case "r", "R":
			// Reload configuration on demand
			if newCfg, err := config.Load(); err == nil {
				m.Config = newCfg
				m.Width = newCfg.Width
				m.LastError = "Config reloaded successfully ✓"
			} else {
				m.LastError = "Failed to reload config: " + err.Error()
			}
			return m, nil
		}

	case updateMsg:
		// Fetch fresh GPU metrics
		if newData, err := gpu.GetAllMetrics(); err == nil {
			m.Name = newData.Name
			m.Temp = newData.Temp
			m.Power = newData.Power
			m.PowerLimit = newData.PowerLimit
			m.Util = newData.Util
			m.MemTotal = newData.MemTotal
			m.MemUsed = newData.MemUsed
			m.FanSpeed = newData.FanSpeed
		} else {
			m.LastError = "Failed to read GPU data"
		}

		// Schedule the next update
		return m, tea.Tick(
			time.Duration(m.Config.General.UpdateIntervalMs)*time.Millisecond,
			func(time.Time) tea.Msg { return updateMsg{} },
		)
	}

	// No change for unhandled messages
	return m, nil
}

// View renders the current state of the model into a styled terminal view.
// This is called on every frame.
func (m Model) View() tea.View {
	styles := GetStyles(m) // Get dynamic styles based on current theme and config
	barWidth := m.Width - 22
	if barWidth < 40 {
		barWidth = 40 // Minimum reasonable width for progress bars
	}

	// Handle temperature unit conversion (Celsius ↔ Fahrenheit)
	displayTemp := m.Temp
	tempUnit := "°C"
	if m.Config.General.TempUnit == "F" {
		displayTemp = m.Temp*9/5 + 32
		tempUnit = "°F"
	}

	// Calculate percentages for progress bars
	tempP := (m.Temp / m.MaxTemp) * 100
	powerP := 0.0
	if m.PowerLimit > 0 {
		powerP = (m.Power / m.PowerLimit) * 100
	}
	memP := 0.0
	if m.MemTotal > 0 {
		memP = (m.MemUsed / m.MemTotal) * 100
	}

	// Build the UI content vertically
	content := lipgloss.JoinVertical(lipgloss.Left,
		styles.Title.Width(m.Width).AlignHorizontal(lipgloss.Center).Render(" "+m.Config.Title+" "),
		gpuNameStyle.Render("  "+m.Name),
		"",
	)

	// Conditionally render metric cards based on user configuration
	if m.Config.General.ShowTemp {
		tempCard := styles.Card.Render(lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Left,
				styles.Label.Render("Temperature"),
				StatusDot(tempP, styles.High, styles.Medium),
			),
			styles.Value.Render(fmt.Sprintf("%.0f %s", displayTemp, tempUnit)),
			ProgressBar(barWidth, tempP, styles.Accent, styles.High, styles.Medium),
		))
		content = lipgloss.JoinVertical(lipgloss.Left, content, tempCard)
	}

	if m.Config.General.ShowUtil {
		utilCard := styles.Card.Render(lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Left,
				styles.Label.Render("Utilization"),
				StatusDot(m.Util, styles.High, styles.Medium),
			),
			styles.Value.Render(fmt.Sprintf("%.0f %%", m.Util)),
			ProgressBar(barWidth, m.Util, styles.Accent, styles.High, styles.Medium),
		))
		content = lipgloss.JoinVertical(lipgloss.Left, content, utilCard)
	}

	if m.Config.General.ShowPower {
		powerCard := styles.Card.Render(lipgloss.JoinVertical(lipgloss.Left,
			styles.Label.Render("Power Draw"),
			styles.Value.Render(fmt.Sprintf("%.1f / %.0f W", m.Power, m.PowerLimit)),
			ProgressBar(barWidth, powerP, styles.Accent, styles.High, styles.Medium),
		))
		content = lipgloss.JoinVertical(lipgloss.Left, content, powerCard)
	}

	if m.Config.General.ShowMemory {
		memCard := styles.Card.Render(lipgloss.JoinVertical(lipgloss.Left,
			styles.Label.Render("Memory"),
			styles.Value.Render(fmt.Sprintf("%.1f / %.0f GB", m.MemUsed/1024, m.MemTotal/1024)),
			ProgressBar(barWidth, memP, styles.Accent, styles.High, styles.Medium),
		))
		content = lipgloss.JoinVertical(lipgloss.Left, content, memCard)
	}

	if m.Config.General.ShowFan {
		fanCard := styles.Card.Render(lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Left,
				styles.Label.Render("Fan Speed"),
				StatusDot(m.FanSpeed, styles.High, styles.Medium),
			),
			styles.Value.Render(fmt.Sprintf("%.0f %%", m.FanSpeed)),
			ProgressBar(barWidth, m.FanSpeed, styles.Accent, styles.High, styles.Medium),
		))
		content = lipgloss.JoinVertical(lipgloss.Left, content, fanCard)
	}

	content = lipgloss.JoinVertical(lipgloss.Left, content, "")

	// Show error message if present
	if m.LastError != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, content,
			errorStyleBase.Render("  ⚠ "+m.LastError),
		)
	}

	// Help text at the bottom
	content = lipgloss.JoinVertical(lipgloss.Left, content,
		helpStyle.Render("  q: quit • r: reload config"),
	)

	return tea.NewView(content)
}
