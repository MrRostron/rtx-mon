package main

import (
	"fmt"
	"image/color"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
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
	fanSpeed   float64
	fanManual  bool
	targetFan  float64

	width     int
	height    int
	isDark    bool
	lastError string

	showFanModal bool
	isDragging   bool
}

var (
	nvmlInitialized bool
	device          nvml.Device

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

	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(2, 4).
			Width(72)
)

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
		Width(14).
		Align(lipgloss.Right)

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
	defer func() {
		if nvmlInitialized {
			nvml.Shutdown()
		}
	}()

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func initNVML() error {
	if ret := nvml.Init(); ret != nvml.SUCCESS {
		return fmt.Errorf("NVML init: %s", nvml.ErrorString(ret))
	}
	dev, ret := nvml.DeviceGetHandleByIndex(0)
	if ret != nvml.SUCCESS {
		nvml.Shutdown()
		return fmt.Errorf("get GPU 0: %s", nvml.ErrorString(ret))
	}
	device = dev
	nvmlInitialized = true
	return nil
}

// Updated: controls ALL fans (0, 1, 2) so all 3 fans on your GPU respond
func setFanSpeed(percent int) error {
	if !nvmlInitialized {
		return fmt.Errorf("NVML not initialized")
	}

	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	applied := false
	for fan := 0; fan < 3; fan++ {
		if ret := device.SetFanSpeed_v2(fan, percent); ret == nvml.SUCCESS {
			applied = true
		}
		// ignore errors for non-existent fan indices
	}

	if !applied {
		return fmt.Errorf("failed to set any fan")
	}
	return nil
}

// Updated: restores auto mode on ALL fans
func restoreAutoFan() error {
	if !nvmlInitialized {
		return fmt.Errorf("NVML not initialized")
	}
	for fan := 0; fan < 3; fan++ {
		_ = device.SetDefaultFanSpeed_v2(fan)
	}
	return nil
}

func getAllMetrics() (model, error) {
	if !nvmlInitialized {
		return model{}, fmt.Errorf("NVML not initialized")
	}
	m := model{}

	if name, ret := device.GetName(); ret == nvml.SUCCESS {
		m.name = name
	}
	if temp, ret := device.GetTemperature(nvml.TEMPERATURE_GPU); ret == nvml.SUCCESS {
		m.temp = float64(temp)
	}
	if power, ret := device.GetPowerUsage(); ret == nvml.SUCCESS {
		m.power = float64(power) / 1000
	}
	if limit, ret := device.GetPowerManagementLimit(); ret == nvml.SUCCESS {
		m.powerLimit = float64(limit) / 1000
	}
	if util, ret := device.GetUtilizationRates(); ret == nvml.SUCCESS {
		m.util = float64(util.Gpu)
	}
	if mem, ret := device.GetMemoryInfo(); ret == nvml.SUCCESS {
		m.memTotal = float64(mem.Total) / 1024 / 1024
		m.memUsed = float64(mem.Used) / 1024 / 1024
	}
	// Use fan 0 for display (standard)
	if fan, ret := device.GetFanSpeed(); ret == nvml.SUCCESS {
		m.fanSpeed = float64(fan)
	}

	return m, nil
}

func initialModel() model {
	var lastErr string
	if err := initNVML(); err != nil {
		lastErr = err.Error()
		fmt.Println("Warning:", lastErr)
	}

	data, _ := getAllMetrics()

	return model{
		name:       data.name,
		temp:       data.temp,
		maxTemp:    100,
		power:      data.power,
		powerLimit: data.powerLimit,
		util:       data.util,
		memTotal:   data.memTotal,
		memUsed:    data.memUsed,
		fanSpeed:   data.fanSpeed,
		targetFan:  data.fanSpeed,
		fanManual:  false,
		width:      90,
		isDark:     lipgloss.HasDarkBackground(os.Stdin, os.Stdout),
		lastError:  lastErr,
	}
}

type updateMsg struct{}

func (m model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return updateMsg{} })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q", "Q":
			if m.showFanModal {
				m.showFanModal = false
				return m, nil
			}
			return m, tea.Quit

		case "f", "F":
			m.showFanModal = true
			m.fanManual = true
			if m.targetFan == 0 {
				m.targetFan = m.fanSpeed
			}

		case "a", "A":
			if m.showFanModal {
				if err := restoreAutoFan(); err != nil {
					m.lastError = "Restore auto failed: " + err.Error()
				} else {
					m.fanManual = false
					m.showFanModal = false
					m.lastError = "Fan restored to automatic"
				}
			}

		case "+", "=":
			if m.fanManual {
				m.targetFan = min(100, m.targetFan+5)
				_ = setFanSpeed(int(m.targetFan))
			}

		case "-":
			if m.fanManual {
				m.targetFan = max(0, m.targetFan-5)
				_ = setFanSpeed(int(m.targetFan))
			}

		case "left", "h":
			if m.fanManual && m.showFanModal {
				m.targetFan = max(0, m.targetFan-2)
				_ = setFanSpeed(int(m.targetFan))
			}

		case "right", "l":
			if m.fanManual && m.showFanModal {
				m.targetFan = min(100, m.targetFan+2)
				_ = setFanSpeed(int(m.targetFan))
			}

		case "esc":
			if m.showFanModal {
				m.showFanModal = false
			}
		}

	case tea.MouseClickMsg:
		if m.showFanModal {
			if msg.Button == tea.MouseLeft {
				if msg.Y >= 10 && msg.Y <= 12 {
					if err := restoreAutoFan(); err != nil {
						m.lastError = err.Error()
					} else {
						m.fanManual = false
						m.showFanModal = false
						m.lastError = "Fan restored to automatic"
					}
					return m, nil
				}

				barStartY := 7
				if msg.Y == barStartY || msg.Y == barStartY+1 {
					barWidth := 60
					relativeX := msg.X - 10
					if relativeX < 0 {
						relativeX = 0
					}
					percent := float64(relativeX) / float64(barWidth) * 100
					percent = max(0, min(100, percent))

					m.targetFan = percent
					m.fanManual = true
					if err := setFanSpeed(int(percent)); err != nil {
						handleFanError(&m, err)
					}
				}
			}
			return m, nil
		}

		if msg.Button == tea.MouseLeft {
			m.showFanModal = true
			m.fanManual = true
			if m.targetFan == 0 {
				m.targetFan = m.fanSpeed
			}
		}

	case tea.MouseMotionMsg:
		if m.showFanModal && m.isDragging && m.fanManual {
			barWidth := 60
			relativeX := msg.X - 10
			if relativeX < 0 {
				relativeX = 0
			}
			percent := float64(relativeX) / float64(barWidth) * 100
			percent = max(0, min(100, percent))

			m.targetFan = percent
			_ = setFanSpeed(int(percent))
		}

	case tea.MouseReleaseMsg:
		m.isDragging = false

	case updateMsg:
		if newData, err := getAllMetrics(); err == nil {
			m.name = newData.name
			m.temp = newData.temp
			m.power = newData.power
			m.powerLimit = newData.powerLimit
			m.util = newData.util
			m.memTotal = newData.memTotal
			m.memUsed = newData.memUsed
			m.fanSpeed = newData.fanSpeed
		} else {
			m.lastError = "Failed to read GPU data"
		}

		if m.fanManual {
			_ = setFanSpeed(int(m.targetFan))
		}

		return m, tea.Tick(time.Second, func(time.Time) tea.Msg { return updateMsg{} })
	}

	return m, nil
}

func handleFanError(m *model, err error) {
	errStr := err.Error()
	m.lastError = "Fan control failed: " + errStr

	if strings.Contains(errStr, "Insufficient") || strings.Contains(errStr, "Permission") || strings.Contains(errStr, "Not Supported") {
		m.lastError = "Fan control requires root on Wayland.\n\nRun with: sudo ./rtx-mon"
	}
}

func progressBar(width int, percent float64, accent color.Color) string {
	if percent > 100 {
		percent = 100
	}
	if percent < 0 {
		percent = 0
	}
	filled := int(float64(width) * percent / 100)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	col := accent
	if percent > 80 {
		col = lipgloss.Color("#EF4444")
	} else if percent > 60 {
		col = lipgloss.Color("#EAB308")
	}

	return lipgloss.NewStyle().Foreground(col).Render("["+bar+"] ") +
		lipgloss.NewStyle().Foreground(col).Bold(true).Render(fmt.Sprintf("%5.1f%%", percent))
}

func statusDot(p float64) string {
	c := lipgloss.Color("#22C55E")
	if p > 80 {
		c = lipgloss.Color("#EF4444")
	} else if p > 60 {
		c = lipgloss.Color("#EAB308")
	}
	return lipgloss.NewStyle().Foreground(c).Render("●")
}

func (m model) View() tea.View {
	styles := getStyles(m)
	barWidth := m.width - 22
	if barWidth < 40 {
		barWidth = 40
	}

	tempP := (m.temp / m.maxTemp) * 100
	powerP := 0.0
	if m.powerLimit > 0 {
		powerP = (m.power / m.powerLimit) * 100
	}
	memP := 0.0
	if m.memTotal > 0 {
		memP = (m.memUsed / m.memTotal) * 100
	}

	tempCard := styles.card.Render(lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, styles.label.Render("Temperature"), statusDot(tempP)),
		styles.value.Render(fmt.Sprintf("%.0f °C", m.temp)),
		progressBar(barWidth, tempP, styles.accent),
	))

	utilCard := styles.card.Render(lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, styles.label.Render("Utilization"), statusDot(m.util)),
		styles.value.Render(fmt.Sprintf("%.0f %%", m.util)),
		progressBar(barWidth, m.util, styles.accent),
	))

	powerCard := styles.card.Render(lipgloss.JoinVertical(lipgloss.Left,
		styles.label.Render("Power Draw"),
		styles.value.Render(fmt.Sprintf("%.1f / %.0f W", m.power, m.powerLimit)),
		progressBar(barWidth, powerP, styles.accent),
	))

	memCard := styles.card.Render(lipgloss.JoinVertical(lipgloss.Left,
		styles.label.Render("Memory"),
		styles.value.Render(fmt.Sprintf("%.1f / %.0f GB", m.memUsed/1024, m.memTotal/1024)),
		progressBar(barWidth, memP, styles.accent),
	))

	fanCard := styles.card.Render(lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, styles.label.Render("Fan Speed"), statusDot(m.fanSpeed)),
		styles.value.Render(fmt.Sprintf("%.0f %%", m.fanSpeed)),
		progressBar(barWidth, m.fanSpeed, styles.accent),
		helpStyle.Render("  Press f to open fan control"),
	))

	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Width(m.width).AlignHorizontal(lipgloss.Center).Render(" RTX Monitor "),
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
		helpStyle.Render("  q: quit • f: fan control"),
	)

	if m.showFanModal {
		modalContent := lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render("   Fan Speed Control"),
			"",
			styles.value.Render(fmt.Sprintf("Current: %.0f%%     Target: %.0f%%", m.fanSpeed, m.targetFan)),
			"",
			progressBar(60, m.targetFan, styles.accent),
			"",
			helpStyle.Render("   Click / drag bar to set speed"),
			helpStyle.Render("   +/- or ← → : fine adjust     a : Restore Automatic"),
			helpStyle.Render("   Esc or q : close"),
		)

		modal := modalStyle.Render(modalContent)
		centered := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)

		v := tea.NewView(centered)
		v.AltScreen = true
		v.MouseMode = tea.MouseModeCellMotion
		return v
	}

	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

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
