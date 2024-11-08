package nvmlmock

import (
	"golang.zabbix.com/plugin/nvidia/pkg/nvml"
)

var (
	// _ nvml.Device = (*MockDevice)(nil)
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
