// Package gpu provides low-level access to NVIDIA GPU metrics using NVML (NVIDIA Management Library).
// It handles initialization, metric collection, and cleanup for the RTX monitoring tool.
package gpu

import (
	"fmt"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

// Metrics holds all the GPU information we want to display in the TUI.
// All numeric values are in human-friendly units (Celsius, Watts, MiB, %).
type Metrics struct {
	Name       string  // GPU model name (e.g. "NVIDIA GeForce RTX 4090")
	Temp       float64 // Current GPU temperature in Celsius
	Power      float64 // Current power draw in Watts
	PowerLimit float64 // Maximum power limit in Watts
	Util       float64 // GPU utilization percentage (0-100)
	MemTotal   float64 // Total VRAM in MiB
	MemUsed    float64 // Used VRAM in MiB
	FanSpeed   float64 // Fan speed percentage (0-100)
}

// Package-level variables to maintain NVML state.
// We only support monitoring the first GPU (index 0) for simplicity.
var (
	nvmlInitialized bool        // Tracks whether NVML has been successfully initialized
	device          nvml.Device // Handle to the GPU device (currently only GPU 0)
)

// InitNVML initializes the NVIDIA Management Library and acquires a handle to the first GPU.
// This must be called before any metrics can be retrieved.
// Returns an error if NVML fails to initialize or if no GPU is found.
func InitNVML() error {
	// Initialize NVML
	if ret := nvml.Init(); ret != nvml.SUCCESS {
		return fmt.Errorf("NVML init: %s", nvml.ErrorString(ret))
	}

	// Get handle for the first GPU (index 0)
	dev, ret := nvml.DeviceGetHandleByIndex(0)
	if ret != nvml.SUCCESS {
		// Clean up NVML if we can't get a device handle
		nvml.Shutdown()
		return fmt.Errorf("get GPU 0: %s", nvml.ErrorString(ret))
	}

	device = dev
	nvmlInitialized = true
	return nil
}

// Shutdown releases all NVML resources.
// It is safe to call multiple times and should always be called before the program exits.
// This is invoked via defer in main.go.
func Shutdown() {
	if nvmlInitialized {
		nvml.Shutdown()
		nvmlInitialized = false // Optional: reset state
	}
}

// GetAllMetrics queries the GPU for all supported metrics in a single call.
// It gracefully skips any metric that fails to retrieve (common with some GPUs or drivers).
// Returns the populated Metrics struct and nil error on success.
func GetAllMetrics() (Metrics, error) {
	if !nvmlInitialized {
		return Metrics{}, fmt.Errorf("NVML not initialized")
	}

	m := Metrics{}

	// GPU model name
	if name, ret := device.GetName(); ret == nvml.SUCCESS {
		m.Name = name
	}

	// GPU temperature (in Celsius)
	if temp, ret := device.GetTemperature(nvml.TEMPERATURE_GPU); ret == nvml.SUCCESS {
		m.Temp = float64(temp)
	}

	// Current power usage (converted from milliwatts to watts)
	if power, ret := device.GetPowerUsage(); ret == nvml.SUCCESS {
		m.Power = float64(power) / 1000
	}

	// Power management limit (converted from milliwatts to watts)
	if limit, ret := device.GetPowerManagementLimit(); ret == nvml.SUCCESS {
		m.PowerLimit = float64(limit) / 1000
	}

	// GPU core utilization percentage
	if util, ret := device.GetUtilizationRates(); ret == nvml.SUCCESS {
		m.Util = float64(util.Gpu)
	}

	// Memory information (converted from bytes to MiB)
	if mem, ret := device.GetMemoryInfo(); ret == nvml.SUCCESS {
		m.MemTotal = float64(mem.Total) / 1024 / 1024
		m.MemUsed = float64(mem.Used) / 1024 / 1024
	}

	// Fan speed percentage
	if fan, ret := device.GetFanSpeed(); ret == nvml.SUCCESS {
		m.FanSpeed = float64(fan)
	}

	return m, nil
}
