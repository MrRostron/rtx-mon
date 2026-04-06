package ui

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)

	errorStyleBase = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(2, 4).
			Width(72)

	gpuNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A0A0A0")).
			Italic(true)
)

func GetStyles(m Model) struct {
	Card   lipgloss.Style
	Label  lipgloss.Style
	Value  lipgloss.Style
	Accent color.Color
	Title  lipgloss.Style
	High   color.Color
	Medium color.Color
} {
	cfg := m.Config

	cardBorder := getBorderStyle(cfg.Appearance.BorderStyle)
	card := lipgloss.NewStyle().
		Border(cardBorder).
		BorderForeground(lipgloss.Color(cfg.Colors.Border)).
		Padding(cfg.Appearance.CardPadding, 1).
		MarginBottom(1).
		Width(m.Width - 4)

	if cfg.Appearance.CardHeight > 0 {
		card = card.Height(cfg.Appearance.CardHeight)
	}

	label := lipgloss.NewStyle().
		Foreground(lipgloss.Color(cfg.Colors.Label)).
		Width(15).
		Align(lipgloss.Left)

	value := lipgloss.NewStyle().
		Foreground(lipgloss.Color(cfg.Colors.Text)).
		Bold(true)

	accent := lipgloss.Color(cfg.Colors.Accent)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(cfg.Colors.TitleFG)).
		Background(lipgloss.Color(cfg.Colors.TitleBG)).
		Padding(0, 3).
		MarginBottom(1)

	return struct {
		Card   lipgloss.Style
		Label  lipgloss.Style
		Value  lipgloss.Style
		Accent color.Color
		Title  lipgloss.Style
		High   color.Color
		Medium color.Color
	}{
		Card:   card,
		Label:  label,
		Value:  value,
		Accent: accent,
		Title:  title,
		High:   lipgloss.Color(cfg.Colors.High),
		Medium: lipgloss.Color(cfg.Colors.Medium),
	}
}

func getBorderStyle(style string) lipgloss.Border {
	switch style {
	case "double":
		return lipgloss.DoubleBorder()
	case "single":
		return lipgloss.NormalBorder()
	case "none":
		return lipgloss.HiddenBorder()
	default: // "rounded"
		return lipgloss.RoundedBorder()
	}
}

func ProgressBar(width int, percent float64, accent, high, medium color.Color) string {
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
		col = high
	} else if percent > 60 {
		col = medium
	}

	return lipgloss.NewStyle().Foreground(col).Render("["+bar+"] ") +
		lipgloss.NewStyle().Foreground(col).Bold(true).Render(fmt.Sprintf("%5.1f%%", percent))
}

func StatusDot(p float64, high, medium color.Color) string {
	c := lipgloss.Color("#22C55E")
	if p > 80 {
		c = high
	} else if p > 60 {
		c = medium
	}
	return lipgloss.NewStyle().Foreground(c).Render(" ● ")
}
