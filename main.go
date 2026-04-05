package main

import (
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type model struct {
	name       string
	temp       float64
	maxTemp    float64
	power      float64
	powerLimit float64
	util       float64
	memTotal   float64
	memUsed    float64
	fanSpeed   float64 // %
	width      int
	height     int
	isDark     bool
	lastError  string
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 3).
			MarginBottom(1)

	gpuNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A0A0A0")).
			Italic(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)
)

// Dynamic styles based on theme
func getStyles(m model) struct {
	card   lipgloss.Style
	label  lipgloss.Style
	value  lipgloss.Style
	accent color.Color
} {
	lightDark := lipgloss.LightDark(m.isDark)

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lightDark(lipgloss.Color("#4A4A4A"), lipgloss.Color("#555555"))).
		Padding(1, 2).
		MarginBottom(1).
		Width(m.width - 4)

	label := lipgloss.NewStyle().
		Foreground(lightDark(lipgloss.Color("#BBBBBB"), lipgloss.Color("#999999"))).
		Width(12).            // More compact
		Align(lipgloss.Right) // Right-aligned for cleaner look

	value := lipgloss.NewStyle().
		Foreground(lightDark(lipgloss.Color("#FFFFFF"), lipgloss.Color("#EEEEEE"))).
		Bold(true)

	accent := lightDark(lipgloss.Color("#7D56F4"), lipgloss.Color("#A78BFA"))

	return struct {
		card   lipgloss.Style
		label  lipgloss.Style
		value  lipgloss.Style
		accent color.Color
	}{card, label, value, accent}
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func initialModel() model {
	name, _ := getGPUName()
	temp, _ := getGPUTemp()
	power, _ := getGPUPower()
	powerLimit, _ := getGPUPowerLimit()
	util, _ := getGPUUtil()
	memTotal, _ := getGPUMemTotal()
	memUsed, _ := getGPUMemUsed()
	fan, _ := getGPUFanSpeed()

	hasDark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)

	return model{
		name:       name,
		temp:       temp,
		maxTemp:    100,
		power:      power,
		powerLimit: powerLimit,
		util:       util,
		memTotal:   memTotal,
		memUsed:    memUsed,
		fanSpeed:   fan,
		width:      90,
		isDark:     hasDark,
		lastError:  "",
	}
}

// ====================== GPU Fetchers ======================

func getGPUName() (string, error) {
	out, err := exec.Command("nvidia-smi", "--query-gpu=name", "--format=csv,noheader,nounits").Output()
	if err != nil {
		return "Unknown GPU", fmt.Errorf("failed to get gpu name: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func queryGPU(field string) (float64, error) {
	out, err := exec.Command("nvidia-smi", "--query-gpu="+field, "--format=csv,noheader,nounits").Output()
	if err != nil {
		return 0, fmt.Errorf("nvidia-smi query failed for %s: %w", field, err)
	}
	s := strings.TrimSpace(string(out))
	if s == "" {
		return 0, fmt.Errorf("empty response for field %s", field)
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s value '%s': %w", field, s, err)
	}
	return f, nil
}

func getGPUTemp() (float64, error)       { return queryGPU("temperature.gpu") }
func getGPUPower() (float64, error)      { return queryGPU("power.draw") }
func getGPUPowerLimit() (float64, error) { return queryGPU("power.limit") }
func getGPUUtil() (float64, error)       { return queryGPU("utilization.gpu") }
func getGPUMemTotal() (float64, error)   { return queryGPU("memory.total") }
func getGPUMemUsed() (float64, error)    { return queryGPU("memory.used") }
func getGPUFanSpeed() (float64, error)   { return queryGPU("fan.speed") }

// ====================== Bubble Tea ======================

type updateMsg struct{}

func (m model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return updateMsg{} })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q", "Q":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		if m.width > 110 {
			m.width = 110
		}
		m.height = msg.Height

	case updateMsg:
		var hasError bool

		if v, err := getGPUTemp(); err == nil {
			m.temp = v
		} else {
			hasError = true
		}

		if v, err := getGPUPower(); err == nil {
			m.power = v
		} else {
			hasError = true
		}

		if v, err := getGPUPowerLimit(); err == nil && v > 0 {
			m.powerLimit = v
		} else if m.powerLimit == 0 {
			m.powerLimit = 400
			hasError = true
		}

		if v, err := getGPUUtil(); err == nil {
			m.util = v
		} else {
			hasError = true
		}

		if v, err := getGPUMemUsed(); err == nil {
			m.memUsed = v
		} else {
			hasError = true
		}

		if v, err := getGPUMemTotal(); err == nil {
			m.memTotal = v
		} else {
			hasError = true
		}

		if v, err := getGPUFanSpeed(); err == nil {
			m.fanSpeed = v
		} else {
			hasError = true
		}

		if hasError {
			m.lastError = "Warning: Could not read some GPU data (nvidia-smi failed)"
		} else {
			m.lastError = ""
		}

		return m, tea.Tick(time.Second, func(time.Time) tea.Msg { return updateMsg{} })
	}
	return m, nil
}

// ====================== UI Helpers ======================

func progressBar(width int, percent float64, accent color.Color) string {
	if percent > 100 {
		percent = 100
	}
	if percent < 0 {
		percent = 0
	}

	filled := int(float64(width) * percent / 100)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	barColor := accent
	if percent > 80 {
		barColor = lipgloss.Color("#EF4444")
	} else if percent > 60 {
		barColor = lipgloss.Color("#EAB308")
	}

	barStyle := lipgloss.NewStyle().Foreground(barColor)
	percentStyle := lipgloss.NewStyle().Foreground(barColor).Bold(true)

	return barStyle.Render("["+bar+"]") + " " + percentStyle.Render(fmt.Sprintf("%5.1f%%", percent))
}

func statusDot(percent float64) string {
	col := lipgloss.Color("#22C55E")
	if percent > 80 {
		col = lipgloss.Color("#EF4444")
	} else if percent > 60 {
		col = lipgloss.Color("#EAB308")
	}
	return lipgloss.NewStyle().Foreground(col).Render("●")
}

// ====================== Main View ======================

func (m model) View() tea.View {
	// Safe calculations
	tempPercent := (m.temp / m.maxTemp) * 100
	powerPercent := 0.0
	if m.powerLimit > 0 {
		powerPercent = (m.power / m.powerLimit) * 100
	}
	utilPercent := m.util
	memPercent := 0.0
	if m.memTotal > 0 {
		memPercent = (m.memUsed / m.memTotal) * 100
	}
	fanPercent := m.fanSpeed

	usedGB := m.memUsed / 1024
	totalGB := m.memTotal / 1024
	if totalGB == 0 {
		totalGB = 1
	}

	// Optimized bar width for compact full-width cards
	barWidth := m.width - 18
	if barWidth < 30 {
		barWidth = 30
	}

	styles := getStyles(m)

	// Cards using full width style
	tempCard := styles.card.Render(lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, styles.label.Render("Temperature"), statusDot(tempPercent)),
		styles.value.Render(fmt.Sprintf("%.0f °C", m.temp)),
		progressBar(barWidth, tempPercent, styles.accent),
	))

	utilCard := styles.card.Render(lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, styles.label.Render("Utilization"), statusDot(utilPercent)),
		styles.value.Render(fmt.Sprintf("%.0f %%", m.util)),
		progressBar(barWidth, utilPercent, styles.accent),
	))

	powerCard := styles.card.Render(lipgloss.JoinVertical(lipgloss.Left,
		styles.label.Render("Power Draw"),
		styles.value.Render(fmt.Sprintf("%.2f / %.0f W", m.power, m.powerLimit)),
		progressBar(barWidth, powerPercent, styles.accent),
	))

	memCard := styles.card.Render(lipgloss.JoinVertical(lipgloss.Left,
		styles.label.Render("Memory"),
		styles.value.Render(fmt.Sprintf("%.1f / %.0f GB", usedGB, totalGB)),
		progressBar(barWidth, memPercent, styles.accent),
	))

	fanCard := styles.card.Render(lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, styles.label.Render("Fan Speed"), statusDot(fanPercent)),
		styles.value.Render(fmt.Sprintf("%.0f %%", m.fanSpeed)),
		progressBar(barWidth, fanPercent, styles.accent),
	))

	// Final layout - all cards stacked vertically
	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Width(m.width).AlignHorizontal(lipgloss.Center).Render(" NVIDIA GPU Monitor "),
		gpuNameStyle.Render("  "+m.name),
		"",

		tempCard,
		utilCard,
		powerCard,
		memCard,
		fanCard,

		"",
	)

	if m.lastError != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, content, errorStyle.Render("  ⚠ "+m.lastError))
	}

	content = lipgloss.JoinVertical(lipgloss.Left, content,
		helpStyle.Render("  Press q to quit • Updates every second"),
	)

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
