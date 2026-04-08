// Package ui contains the Bubble Tea model, update logic, and view rendering for the RTX monitoring TUI.
//
// styles.go is responsible for all visual styling using Lip Gloss. It supports full customization
// through the config file (colors, borders, padding, etc.) and provides dynamic theming.
package ui

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// Predefined global styles used across the UI
var (
	// helpStyle is used for the bottom help text (keyboard shortcuts)
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)

	// errorStyleBase is used to display warning/error messages in red
	errorStyleBase = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	// gpuNameStyle styles the GPU model name displayed below the title
	gpuNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A0A0A0")).
			Italic(true)
)

// GetStyles returns a set of Lip Gloss styles and colors based on the current configuration and model state.
// This allows the UI to be fully themeable and responsive to user settings.
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

	// Determine border style from config (rounded, single, double, or none)
	cardBorder := getBorderStyle(cfg.Appearance.BorderStyle)

	// Main card style - used for each metric section (Temperature, Utilization, etc.)
	card := lipgloss.NewStyle().
		Border(cardBorder).
		BorderForeground(lipgloss.Color(cfg.Colors.Border)).
		Padding(cfg.Appearance.CardPadding, 1).
		MarginBottom(1).
		Width(m.Width - 4)

	// Optional fixed height for cards (useful for consistent layout)
	if cfg.Appearance.CardHeight > 0 {
		card = card.Height(cfg.Appearance.CardHeight)
	}

	// Label style (e.g. "Temperature", "Utilization")
	label := lipgloss.NewStyle().
		Foreground(lipgloss.Color(cfg.Colors.Label)).
		Width(15).
		Align(lipgloss.Left)

	// Value style (the large numeric values)
	value := lipgloss.NewStyle().
		Foreground(lipgloss.Color(cfg.Colors.Text)).
		Bold(true)

	// Accent color used primarily for progress bars
	accent := lipgloss.Color(cfg.Colors.Accent)

	// Title style (the big header at the top)
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

// getBorderStyle converts a string config value into the corresponding Lip Gloss border type.
// Defaults to rounded borders for a modern look.
func getBorderStyle(style string) lipgloss.Border {
	switch style {
	case "double":
		return lipgloss.DoubleBorder()
	case "single":
		return lipgloss.NormalBorder()
	case "none":
		return lipgloss.HiddenBorder()
	default: // "rounded" is the default
		return lipgloss.RoundedBorder()
	}
}

// ProgressBar renders a visual progress bar with color-coded fill level.
// The bar changes color based on percentage: accent (normal) → medium → high (warning).
func ProgressBar(width int, percent float64, accent, high, medium color.Color) string {
	// Clamp percentage between 0 and 100
	if percent > 100 {
		percent = 100
	}
	if percent < 0 {
		percent = 0
	}

	filled := int(float64(width) * percent / 100)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	// Choose color based on severity
	col := accent
	if percent > 80 {
		col = high
	} else if percent > 60 {
		col = medium
	}

	// Render the bar and percentage
	return lipgloss.NewStyle().Foreground(col).Render("["+bar+"] ") +
		lipgloss.NewStyle().Foreground(col).Bold(true).Render(fmt.Sprintf("%5.1f%%", percent))
}

// StatusDot returns a colored status indicator (●) next to labels.
// Green by default, changes to medium/high when values are elevated.
func StatusDot(p float64, high, medium color.Color) string {
	c := lipgloss.Color("#22C55E") // Default green (good/low)
	if p > 80 {
		c = high
	} else if p > 60 {
		c = medium
	}
	return lipgloss.NewStyle().Foreground(c).Render(" ● ")
}
