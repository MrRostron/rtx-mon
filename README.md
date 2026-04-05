# RTX GPU Usage

A clean, real-time **Terminal User Interface (TUI)** for monitoring NVIDIA RTX GPU metrics on Linux.

Built with Go, [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) and [Lipgloss v2](https://github.com/charmbracelet/lipgloss).

## Features

- Real-time updates every second
- GPU Name
- Temperature
- Power Draw + Power Limit
- GPU & Memory Utilization
- Fan Speed
- Clean, modern look with automatic dark/light mode detection
- Simple keyboard controls (`q` to quit)
- Lightweight and dependency-minimal

## Screenshots

*(Add a screenshot here later — highly recommended! You can run the program and take a terminal screenshot)*

## Requirements

- Linux
- NVIDIA GPU with `nvidia-smi` installed and accessible
- Go 1.23+ (to build or install)

## Installation

### Option 1: Install directly with Go (Recommended)

```bash
go install github.com/MrRostron/rtx-gpu-usage@latest
```
Then run :
```bash
rtx-gpu-usage
```

### Option 2: Build from source
```bash
git clone https://github.com/MrRostron/rtx-gpu-usage.git
cd rtx-gpu-usage
go build -o rtx-gpu-usage .
./rtx-gpu-usage
```

## Usage
Simply run the binary:
```bash
rtx-gpu-usage
```
