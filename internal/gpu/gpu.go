package gpu

import (
	"fmt"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type Metrics struct {
	Name       string
	Temp       float64
	Power      float64
	PowerLimit float64
	Util       float64
	MemTotal   float64
	MemUsed    float64
	FanSpeed   float64
}

var (
	nvmlInitialized bool
	device          nvml.Device
)

func InitNVML() error {
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

func Shutdown() {
	if nvmlInitialized {
		nvml.Shutdown()
	}
}

func GetAllMetrics() (Metrics, error) {
	if !nvmlInitialized {
		return Metrics{}, fmt.Errorf("NVML not initialized")
	}

	m := Metrics{}

	if name, ret := device.GetName(); ret == nvml.SUCCESS {
		m.Name = name
	}
	if temp, ret := device.GetTemperature(nvml.TEMPERATURE_GPU); ret == nvml.SUCCESS {
		m.Temp = float64(temp)
	}
	if power, ret := device.GetPowerUsage(); ret == nvml.SUCCESS {
		m.Power = float64(power) / 1000
	}
	if limit, ret := device.GetPowerManagementLimit(); ret == nvml.SUCCESS {
		m.PowerLimit = float64(limit) / 1000
	}
	if util, ret := device.GetUtilizationRates(); ret == nvml.SUCCESS {
		m.Util = float64(util.Gpu)
	}
	if mem, ret := device.GetMemoryInfo(); ret == nvml.SUCCESS {
		m.MemTotal = float64(mem.Total) / 1024 / 1024
		m.MemUsed = float64(mem.Used) / 1024 / 1024
	}
	if fan, ret := device.GetFanSpeed(); ret == nvml.SUCCESS {
		m.FanSpeed = float64(fan)
	}

	return m, nil
}
