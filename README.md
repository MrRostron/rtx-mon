# RTX-MON

A clean, lightweight, and highly customizable **Terminal User Interface (TUI)** for real-time NVIDIA RTX GPU monitoring on Linux.

Built with **Go**, [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) and [Lipgloss v2](https://github.com/charmbracelet/lipgloss), using the official **NVIDIA NVML** library for fast and accurate metrics.

![RTX-MON Screenshot](https://github.com/user-attachments/assets/fb17e947-4ec7-4e7e-8525-e6ffcc020386)

## Features

- Real-time GPU monitoring with configurable refresh rate
- Beautiful card-based layout with color-coded progress bars and status indicators
- Fully customizable via `config.toml` (colors, layout, visibility, temperature unit, etc.)
- Live config reloading (`r` key)
- Automatic dark/light mode detection
- Graceful error handling and NVML resource management
- Supports Celsius or Fahrenheit
- Very lightweight and responsive

### Displayed Metrics
- GPU Name
- Temperature (with color-coded status)
- Power Draw + Power Limit (Watts)
- GPU Utilization (%)
- VRAM Usage (GB)
- Fan Speed (%)

## Screenshots

<img width="3834" height="2155" alt="2026-04-05-161806_hyprshot" src="https://github.com/user-attachments/assets/55c49821-54f7-4698-8ca7-99647eaf5d8a" />


## Requirements

- Linux (tested on Arch Linux)
- NVIDIA GPU with proprietary drivers installed (`nvidia-utils`)
- `nvidia-smi` should work (confirms drivers are properly installed)
  
## Installation

### Option 1: Install directly with Go (Recommended)

```bash
go install github.com/MrRostron/rtx-mon@latest
```
Then run :
```bash
rtx-mon
```

### Option 2: Build from source
```bash
git clone https://github.com/MrRostron/rtx-mon.git
cd rtx-mon
go build -o rtx-mon .
./rtx-mon
```

## Usage
Simply run the binary:
```bash
rtx-mon
```

## Configuration
On first run, rtx-mon automatically creates a config file at:
```bash
~/.config/rtx-mon/config.toml
```
You can customize:

*Title, width, update interval
*Which metrics to show/hide
*Temperature unit (°C or °F)
*Border style, card padding/height
*Full color theme (accent, title, high/medium thresholds, etc.)
After editing, press r in the app to apply changes instantly.

## Roadmap

### Planned Features

- [ ] **Multi-GPU support** — Switch between or display multiple GPUs
- [ ] **Command-line flags** — `--interval`, `--gpu <index>`, `--help`, `--version`
- [ ] **Configurable refresh rate** — Allow users to set update speed
- [ ] **Sparkline graphs** — Small inline graphs showing utilization and temperature trends over time
- [ ] **Per-process GPU usage** — Show which processes are using the GPU (memory + compute)
- [ ] **Help screen** — Press `?` to show all keyboard shortcuts
- [ ] **Better memory display** — Option for GB or MB
- [ ] **Export metrics** — Log data to CSV/JSON for analysis

### Future Ideas

- Compact "mini-mode" or status bar overlay
- Integration with `tmux`, `waybar`, or desktop status bars
- Color themes / custom styling
- Support for additional metrics (clock speeds, encoder/decoder usage, energy consumption)

### Done ✓

- Initial TUI with Bubble Tea + Lipgloss
- NVML integration (fast & reliable GPU data)
- Automatic dark/light mode
- Progress bars with color coding
- Graceful error handling
