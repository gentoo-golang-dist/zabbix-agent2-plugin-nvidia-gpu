package nvml

// Runner defines the interface for an NVML runner.
//
//nolint:interfacebloat
type Runner interface {
	InitNVML() error

	InitNVMLv2() error
	// Initialize NVML and any necessary resources
	GetDeviceCount() (uint, error)

	// Initialize NVML and any necessary resources
	GetDeviceCountV2() (uint, error)

	// Get a device by index
	GetDeviceByIndexV2(index uint) (*NVMLDevice, error)

	GetDeviceByUUID(uuid string) (*NVMLDevice, error)

	// Get NVML version
	GetNVMLVersion() (string, error)

	// Get Driver version
	GetDriverVersion() (string, error)

	// Shutdown NVML and clean up resources
	ShutdownNVML() error

	Close() error
}

// Device defines the methods for interacting with a GPU device.
//
//nolint:interfacebloat
type Device interface {
	// Get the temperature of the device.
	GetTemperature() (int, error)

	// Get the memory information for the device.
	GetMemoryInfo() (*MemoryInfo, error)

	GetBAR1MemoryInfo() (*MemoryInfo, error)

	// Get the memory information for the device.
	GetMemoryInfoV2() (*MemoryInfoV2, error)

	// Get the fan speed of the device.
	GetFanSpeed() (uint, error)

	// Get PCIe throughput (TX or RX).
	GetPcieThroughput(metricType PcieMetricType) (uint, error)

	// GetUtilizationRates returns the GPU and memory utilization in that order (GPU, Memory).
	GetUtilizationRates() (uint, uint, error)

	// Get the UUID of the device.
	GetUUID() (string, error)

	// Get the name of the device.
	GetName() (string, error)

	// Get the serial number of the device.
	GetSerial() (string, error)

	GetPowerUsage() (uint, error)

	GetPerformanceState() (uint, error)

	GetClockInfo(clockType ClockType) (uint, error)

	GetPowerManagementLimit() (uint, error)

	GetTotalEnergyConsumption() (uint64, error)

	GetEncoderStats() (uint, uint, uint, error)

	GetEncoderUtilization() (uint, uint, error)

	GetDecoderUtilization() (uint, uint, error)

	GetMemoryErrorCounter(
		errorType MemoryErrorType,
		memoryLocation MemoryLocation,
		counterType EccCounterType,
	) (uint64, error)

	GetEccMode() (bool, bool, error)
}

// MemoryInfoV2 represents the memory information of the device (in bytes).
type MemoryInfoV2 struct {
	Total    uint64 `json:"total_memory_bytes"`    // Total memory available
	Reserved uint64 `json:"reserved_memory_bytes"` // Memory reserved by the system
	Free     uint64 `json:"free_memory_bytes"`     // Free memory available
	Used     uint64 `json:"used_memory_bytes"`     // Memory currently being used + reserved
}

// MemoryInfo represents the memory information of the device (in bytes).
type MemoryInfo struct {
	Total uint64 `json:"total_memory_bytes"` // Total memory available
	Free  uint64 `json:"free_memory_bytes"`  // Free memory available
	Used  uint64 `json:"used_memory_bytes"`  // Memory currently being used
}
