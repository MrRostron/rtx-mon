// Package config handles loading and management of the application's configuration.
// It uses TOML format and automatically creates a default config file on first run
// in the user's config directory (~/.config/rtx-mon/config.toml).
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the complete user-configurable settings for rtx-mon.
// All fields are tagged for TOML serialization and support live reloading.
type Config struct {
	Title string `toml:"title"` // Title displayed at the top of the UI
	Width int    `toml:"width"` // Target terminal width for layout calculations

	General struct {
		UpdateIntervalMs int    `toml:"update_interval_ms"` // Refresh rate in milliseconds
		TempUnit         string `toml:"temp_unit"`          // "C" or "F"
		ShowTemp         bool   `toml:"show_temp"`
		ShowUtil         bool   `toml:"show_util"`
		ShowPower        bool   `toml:"show_power"`
		ShowMemory       bool   `toml:"show_memory"`
		ShowFan          bool   `toml:"show_fan"`
	} `toml:"general"`

	Appearance struct {
		BorderStyle string `toml:"border_style"` // "rounded", "double", "single", "none"
		CardPadding int    `toml:"card_padding"`
		CardHeight  int    `toml:"card_height"` // 0 = auto, positive value = fixed height
	} `toml:"appearance"`

	Colors struct {
		Accent  string `toml:"accent"`
		TitleBG string `toml:"title_bg"`
		TitleFG string `toml:"title_fg"`
		Text    string `toml:"text"`
		Label   string `toml:"label"`
		Error   string `toml:"error"`
		Border  string `toml:"border"`
		High    string `toml:"high"`   // Used for >80% values (critical)
		Medium  string `toml:"medium"` // Used for 60-80% values (warning)
	} `toml:"colors"`

	Fan struct {
		Enabled       bool `toml:"enabled"`
		DefaultTarget int  `toml:"default_target"`
	} `toml:"fan"`
}

// defaultConfig is the embedded default configuration that gets written
// to disk the first time the application runs.
const defaultConfig = `# RTX Monitor - Default Configuration
# You can edit this file freely. Press 'r' in the app to reload live.

title = "RTX Monitor"
width = 94

[general]
update_interval_ms = 800
temp_unit = "C"          # "C" or "F"

show_temp   = true
show_util   = true
show_power  = true
show_memory = true
show_fan    = true

[appearance]
border_style = "rounded"   # rounded | double | single | none
card_padding = 1
card_height  = 0           # 0 = auto (content-based). Set to 5-8 for fixed taller cards

[colors]
accent   = "#C4A7E7"   # main highlight (soft purple)
title_bg = "#7D56F4"
title_fg = "#FAFAFA"
text     = "#E0DEF4"
label    = "#908CAA"
error    = "#EB6F92"
border   = "#524F67"
high     = "#EB6F92"   # critical (>80%)
medium   = "#F6C177"   # warning (60-80%)

[fan]
enabled        = false
default_target = 45
`

// Load reads the configuration from disk, creating the default file if it doesn't exist.
// It also applies sensible fallback defaults for any missing or invalid values.
func Load() (Config, error) {
	// Get the user's configuration directory (usually ~/.config on Linux/macOS)
	configDir, err := os.UserConfigDir()
	if err != nil {
		return Config{}, err
	}

	// Build full path: ~/.config/rtx-mon/config.toml
	configPath := filepath.Join(configDir, "rtx-mon", "config.toml")

	// If config file doesn't exist, create directory and write default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
			return Config{}, err
		}

		if err := os.WriteFile(configPath, []byte(defaultConfig), 0o644); err != nil {
			return Config{}, err
		}

		fmt.Printf("✅ Created default config at ~/.config/rtx-mon/config.toml\n")
		fmt.Println("  Customize it and press 'r' in rtx-mon for live reload!")
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	// Parse TOML into Config struct
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("invalid TOML: %w", err)
	}

	// Apply sensible defaults for critical values
	if cfg.Width < 60 {
		cfg.Width = 94
	}
	if cfg.General.UpdateIntervalMs < 100 {
		cfg.General.UpdateIntervalMs = 800
	}
	if cfg.General.TempUnit != "F" {
		cfg.General.TempUnit = "C"
	}
	if cfg.Appearance.CardPadding < 0 {
		cfg.Appearance.CardPadding = 1
	}
	if cfg.Appearance.BorderStyle == "" {
		cfg.Appearance.BorderStyle = "rounded"
	}
	if cfg.Appearance.CardHeight < 0 {
		cfg.Appearance.CardHeight = 0
	}

	return cfg, nil
}
