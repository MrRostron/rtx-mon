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

type Model struct {
	Config config.Config

	Name       string
	Temp       float64
	MaxTemp    float64
	Power      float64
	PowerLimit float64
	Util       float64
	MemTotal   float64
	MemUsed    float64
	FanSpeed   float64
	FanManual  bool
	TargetFan  float64

	Width     int
	Height    int
	IsDark    bool
	LastError string

	ShowFanModal bool
}

func InitialModel() Model {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Warning loading config: %v\n", err)
		cfg = config.Config{}
	}

	var lastErr string
	if err := gpu.InitNVML(); err != nil {
		lastErr = err.Error()
		fmt.Println("Warning:", lastErr)
	}

	data, _ := gpu.GetAllMetrics()

	return Model{
		Config:     cfg,
		Name:       data.Name,
		Temp:       data.Temp,
		MaxTemp:    100,
		Power:      data.Power,
		PowerLimit: data.PowerLimit,
		Util:       data.Util,
		MemTotal:   data.MemTotal,
		MemUsed:    data.MemUsed,
		FanSpeed:   data.FanSpeed,
		TargetFan:  data.FanSpeed,
		FanManual:  false,
		Width:      cfg.Width,
		IsDark:     lipgloss.HasDarkBackground(os.Stdin, os.Stdout),
		LastError:  lastErr,
	}
}

type updateMsg struct{}

func (m Model) Init() tea.Cmd {
	return tea.Tick(time.Duration(m.Config.General.UpdateIntervalMs)*time.Millisecond,
		func(time.Time) tea.Msg { return updateMsg{} })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q", "Q":
			if m.ShowFanModal {
				m.ShowFanModal = false
				return m, nil
			}
			return m, tea.Quit

		case "r", "R":
			if newCfg, err := config.Load(); err == nil {
				m.Config = newCfg
				m.Width = newCfg.Width
				m.LastError = "Config reloaded successfully ✓"
			} else {
				m.LastError = "Failed to reload config: " + err.Error()
			}
			return m, nil

		case "f", "F":
			if m.Config.Fan.Enabled {
				m.ShowFanModal = true
				m.FanManual = true
				if m.TargetFan == 0 {
					m.TargetFan = m.FanSpeed
				}
			}

		case "a", "A":
			if m.ShowFanModal {
				if err := gpu.RestoreAutoFan(); err != nil {
					m.LastError = "Restore auto failed: " + err.Error()
				} else {
					m.FanManual = false
					m.ShowFanModal = false
					m.LastError = "Fan restored to automatic"
				}
			}

		case "+", "=":
			if m.FanManual {
				m.TargetFan = min(100, m.TargetFan+5)
				_ = gpu.SetFanSpeed(int(m.TargetFan))
			}

		case "-":
			if m.FanManual {
				m.TargetFan = max(0, m.TargetFan-5)
				_ = gpu.SetFanSpeed(int(m.TargetFan))
			}

		case "left", "h":
			if m.FanManual && m.ShowFanModal {
				m.TargetFan = max(0, m.TargetFan-2)
				_ = gpu.SetFanSpeed(int(m.TargetFan))
			}

		case "right", "l":
			if m.FanManual && m.ShowFanModal {
				m.TargetFan = min(100, m.TargetFan+2)
				_ = gpu.SetFanSpeed(int(m.TargetFan))
			}

		case "esc":
			if m.ShowFanModal {
				m.ShowFanModal = false
			}
		}

	case updateMsg:
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

		if m.FanManual {
			_ = gpu.SetFanSpeed(int(m.TargetFan))
		}

		return m, tea.Tick(time.Duration(m.Config.General.UpdateIntervalMs)*time.Millisecond,
			func(time.Time) tea.Msg { return updateMsg{} })
	}

	return m, nil
}

func (m Model) View() tea.View {
	styles := GetStyles(m)
	barWidth := m.Width - 22
	if barWidth < 40 {
		barWidth = 40
	}

	// Temperature conversion
	displayTemp := m.Temp
	tempUnit := "°C"
	if m.Config.General.TempUnit == "F" {
		displayTemp = m.Temp*9/5 + 32
		tempUnit = "°F"
	}

	tempP := (m.Temp / m.MaxTemp) * 100
	powerP := 0.0
	if m.PowerLimit > 0 {
		powerP = (m.Power / m.PowerLimit) * 100
	}
	memP := 0.0
	if m.MemTotal > 0 {
		memP = (m.MemUsed / m.MemTotal) * 100
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		styles.Title.Width(m.Width).AlignHorizontal(lipgloss.Center).Render(" "+m.Config.Title+" "),
		gpuNameStyle.Render("  "+m.Name),
		"",
	)

	// Conditionally render cards
	if m.Config.General.ShowTemp {
		tempCard := styles.Card.Render(lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Left, styles.Label.Render("Temperature"), StatusDot(tempP, styles.High, styles.Medium)),
			styles.Value.Render(fmt.Sprintf("%.0f %s", displayTemp, tempUnit)),
			ProgressBar(barWidth, tempP, styles.Accent, styles.High, styles.Medium),
		))
		content = lipgloss.JoinVertical(lipgloss.Left, content, tempCard)
	}

	if m.Config.General.ShowUtil {
		utilCard := styles.Card.Render(lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Left, styles.Label.Render("Utilization"), StatusDot(m.Util, styles.High, styles.Medium)),
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
			lipgloss.JoinHorizontal(lipgloss.Left, styles.Label.Render("Fan Speed"), StatusDot(m.FanSpeed, styles.High, styles.Medium)),
			styles.Value.Render(fmt.Sprintf("%.0f %%", m.FanSpeed)),
			ProgressBar(barWidth, m.FanSpeed, styles.Accent, styles.High, styles.Medium),
			helpStyle.Render("  Press f to open fan control"),
		))
		content = lipgloss.JoinVertical(lipgloss.Left, content, fanCard)
	}

	content = lipgloss.JoinVertical(lipgloss.Left, content, "")

	if m.LastError != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, content, errorStyleBase.Render("  ⚠ "+m.LastError))
	}

	content = lipgloss.JoinVertical(lipgloss.Left, content,
		helpStyle.Render("  q: quit • r: reload config • f: fan control"),
	)

	if m.ShowFanModal && m.Config.Fan.Enabled {
		modalContent := lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(m.Config.Colors.Accent)).Render("   Fan Speed Control"),
			"",
			styles.Value.Render(fmt.Sprintf("Current: %.0f%%     Target: %.0f%%", m.FanSpeed, m.TargetFan)),
			"",
			ProgressBar(60, m.TargetFan, styles.Accent, styles.High, styles.Medium),
			"",
			helpStyle.Render("   +/- or ← → : fine adjust     a : Restore Automatic"),
			helpStyle.Render("   Esc or q : close"),
		)

		modal := modalStyle.Render(modalContent)
		centered := lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, modal)

		v := tea.NewView(centered)
		v.AltScreen = true
		return v
	}

	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// Helper functions
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
