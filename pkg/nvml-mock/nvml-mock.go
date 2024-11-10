package nvmlmock

import (
	"golang.zabbix.com/plugin/nvidia/pkg/nvml"
)

var (
	_ nvml.Device = (*MockDevice)(nil)
	_ nvml.Runner = (*MockRunner)(nil)
)

type MockRunner struct {
	IsInit         bool
	DriverVersion  string
	NvmlVersion    string
	DevicesByIndex []*nvml.NVMLDevice
	DevicesByUUID  map[string]*nvml.NVMLDevice
	WantedErr      error
}

type MockDevice struct {
	Temperature    int
	Name           string
	MemoryInfo     *nvml.MemoryInfo
	MemoryInfoV2   *nvml.MemoryInfoV2
	PcieThroughput uint
	FanSpeed       uint
	PowerUsage     uint
	WantedErr      error
}

func (m *MockRunner) InitNVML() error {
	if m.WantedErr != nil {
		return m.WantedErr
	}

	if m.IsInit {
		return nvml.ErrAlreadyInitialized
	}

	m.IsInit = true
	return nil
}

func (m *MockRunner) InitNVMLv2() error {
	if m.WantedErr != nil {
		return m.WantedErr
	}

	if m.IsInit {
		return nvml.ErrAlreadyInitialized
	}

	m.IsInit = true
	return nil
}

// Initialize NVML and any necessary resources
func (m *MockRunner) GetDeviceCount() (uint, error) {
	return uint(len(m.DevicesByIndex)), nil
}

// Initialize NVML and any necessary resources
func (m *MockRunner) GetDeviceCountV2() (uint, error) {
	return uint(len(m.DevicesByIndex)), nil
}

// Get a device by index
func (m *MockRunner) GetDeviceByIndexV2(index uint) (*nvml.NVMLDevice, error) {
	device := m.DevicesByIndex[index]
	if device == nil {
		return nil, nvml.ErrNotFound
	}
	return device, nil
}

func (m *MockRunner) GetDeviceByUUID(uuid string) (*nvml.NVMLDevice, error) {
	if m.WantedErr != nil {
		return nil, m.WantedErr
	}

	device, exists := m.DevicesByUUID[uuid]
	if !exists {
		return nil, nvml.ErrNotFound
	}

	return device, nil
}

// Get NVML version
func (m *MockRunner) GetNVMLVersion() (string, error) {
	if m.WantedErr != nil {
		return "", m.WantedErr
	}

	return m.NvmlVersion, nil
}

// Get Driver version
func (m *MockRunner) GetDriverVersion() (string, error) {
	if m.WantedErr != nil {
		return "", m.WantedErr
	}

	return m.DriverVersion, nil
}

// Shutdown NVML and clean up resources
func (m *MockRunner) ShutdownNVML() error {
	if m.WantedErr != nil {
		return m.WantedErr
	}

	if !m.IsInit {
		return nvml.ErrUninitialized
	}

	m.IsInit = false

	return nil
}

func (m *MockRunner) Close() error {
	if m.WantedErr != nil {
		return m.WantedErr
	}

	return nil
}

// Get the temperature of the device.
func (m *MockDevice) GetTemperature() (int, error) {
	if m.WantedErr != nil {
		return 0, m.WantedErr
	}

	return m.Temperature, nil
}

// Get the memory information for the device.
func (m *MockDevice) GetMemoryInfo() (*nvml.MemoryInfo, error) {
	if m.WantedErr != nil {
		return nil, m.WantedErr
	}

	return m.MemoryInfo, nil
}

func (m *MockDevice) GetBAR1MemoryInfo() (*nvml.MemoryInfo, error) {
	if m.WantedErr != nil {
		return nil, m.WantedErr
	}

	return m.MemoryInfo, nil
}

// Get the memory information for the device.
func (m *MockDevice) GetMemoryInfoV2() (*nvml.MemoryInfoV2, error) {
	if m.WantedErr != nil {
		return nil, m.WantedErr
	}

	return m.MemoryInfoV2, nil
}

// Get the fan speed of the device.
func (m *MockDevice) GetFanSpeed() (uint, error) {
	if m.WantedErr != nil {
		return 0, m.WantedErr
	}

	return m.FanSpeed, nil
}

// Get PCIe throughput (TX or RX).
func (m *MockDevice) GetPcieThroughput(metricType nvml.PcieMetricType) (uint, error) {
	if m.WantedErr != nil {
		return 0, m.WantedErr
	}

	return m.PcieThroughput, nil
}

// GetUtilizationRates returns the GPU and memory utilization in that order (GPU, Memory).
func (m *MockDevice) GetUtilizationRates() (uint, uint, error) {
	if m.WantedErr != nil {
		return 0, 0, m.WantedErr
	}

}

// Get the UUID of the device.
func (m *MockDevice) GetUUID() (string, error) {
	if m.WantedErr != nil {
		return "", m.WantedErr
	}
}

// Get the name of the device.
func (m *MockDevice) GetName() (string, error) {
	if m.WantedErr != nil {
		return "", m.WantedErr
	}
}

// Get the serial number of the device.
func (m *MockDevice) GetSerial() (string, error) {
	if m.WantedErr != nil {
		return "", m.WantedErr
	}
}

func (m *MockDevice) GetPowerUsage() (uint, error) {
	if m.WantedErr != nil {
		return 0, m.WantedErr
	}
}

func (m *MockDevice) GetPerformanceState() (uint, error) {
	if m.WantedErr != nil {
		return 0, m.WantedErr
	}
}

func (m *MockDevice) GetClockInfo(clockType nvml.ClockType) (uint, error) {
	if m.WantedErr != nil {
		return 0, m.WantedErr
	}
}

func (m *MockDevice) GetPowerManagementLimit() (uint, error) {
	if m.WantedErr != nil {
		return 0, m.WantedErr
	}
}

func (m *MockDevice) GetTotalEnergyConsumption() (uint64, error) {
	if m.WantedErr != nil {
		return 0, m.WantedErr
	}
}

func (m *MockDevice) GetEncoderStats() (uint, uint, uint, error) {
	if m.WantedErr != nil {
		return 0, 0, 0, m.WantedErr
	}
}

func (m *MockDevice) GetEncoderUtilization() (uint, uint, error) {
	if m.WantedErr != nil {
		return 0, 0, m.WantedErr
	}
}

func (m *MockDevice) GetDecoderUtilization() (uint, uint, error) {
	if m.WantedErr != nil {
		return 0, 0, m.WantedErr
	}
}

func (m *MockDevice) GetMemoryErrorCounter(
	errorType nvml.MemoryErrorType,
	memoryLocation nvml.MemoryLocation,
	counterType nvml.EccCounterType,
) (uint64, error) {
	if m.WantedErr != nil {
		return 0, m.WantedErr
	}
}

func (m *MockDevice) GetEccMode() (bool, bool, error) {
	if m.WantedErr != nil {
		return false, false, m.WantedErr
	}
}
