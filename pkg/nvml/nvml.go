/*
** Zabbix
** Copyright (C) 2001-2025 Zabbix SIA
**
** Licensed under the Apache License, Version 2.0 (the "License");
** you may not use this file except in compliance with the License.
** You may obtain a copy of the License at
**
**     http://www.apache.org/licenses/LICENSE-2.0
**
** Unless required by applicable law or agreed to in writing, software
** distributed under the License is distributed on an "AS IS" BASIS,
** WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
** See the License for the specific language governing permissions and
** limitations under the License.
**/

package nvml

// Runner defines the interface for an NVML runner.
//
//nolint:interfacebloat
type Runner interface {
	// InitNVML initializes the NVML library using the older NVML interface.
	Init() error
	// InitNVMLv2 initializes the NVML library using the NVML v2 interface.
	InitV2() error

	// GetDeviceCount retrieves the number of NVIDIA devices using the standard NVML interface.
	GetDeviceCount() (uint, error)

	// GetDeviceCountV2 retrieves the number of NVIDIA devices using the NVML v2 interface.
	GetDeviceCountV2() (uint, error)

	// GetDeviceByIndexV2 retrieves a handle to an NVIDIA device by its index using the NVML v2 interface.
	GetDeviceByIndexV2(index uint) (Device, error)

	// GetDeviceByUUID retrieves a handle to an NVIDIA device by its UUID.
	GetDeviceByUUID(uuid string) (Device, error)

	// GetNVMLVersion retrieves the version of the NVML library currently in use.
	GetNVMLVersion() (string, error)

	// GetDriverVersion retrieves the version of the NVIDIA driver currently in use.
	GetDriverVersion() (string, error)

	// Shutdown NVML and clean up resources
	ShutdownNVML() error

	// Close releases the resources associated with the loaded library in the Runner.
	Close() error
}

// Device defines the methods for interacting with a GPU device.
//
//nolint:interfacebloat
type Device interface {
	// GetTemperature retrieves the temperature of the NVIDIA device using the default sensor.
	GetTemperature() (int, error)

	// GetMemoryInfo retrieves memory information for the NVIDIA device.
	GetMemoryInfo() (*MemoryInfo, error)

	// GetBAR1MemoryInfo retrieves BAR1 memory information for the NVIDIA device.
	GetBAR1MemoryInfo() (*MemoryInfo, error)

	// GetMemoryInfoV2 retrieves detailed memory information for the NVIDIA device using the NVML v2 interface.
	GetMemoryInfoV2() (*MemoryInfoV2, error)

	// GetFanSpeed retrieves the current fan speed of the NVIDIA device as a percentage of its maximum speed.
	GetFanSpeed() (uint, error)

	// GetPCIeThroughput retrieves the PCIe throughput for the NVIDIA device, based on the specified metric type.
	GetPCIeThroughput(metricType PcieMetricType) (uint, error)

	// GetUtilizationRates retrieves the GPU and memory utilization rates for the device.
	// The GPU utilization represents the percentage of time over the past sampling period
	// that the GPU was actively processing, while the memory utilization indicates the
	// percentage of time the memory was being accessed.
	//
	// Returns:
	//   - gpuUtilization (uint): The GPU utilization rate as a percentage (0-100).
	//   - memoryUtilization (uint): The memory utilization rate as a percentage (0-100).
	//   - error: An error object if there is a failure in verifying the NVML symbol existence
	//     or in retrieving the utilization rates from NVML.
	GetUtilizationRates() (uint, uint, error)

	// GetUUID retrieves the UUID of the NVIDIA device.
	GetUUID() (string, error)

	// GetName retrieves the name of the NVIDIA device.
	GetName() (string, error)

	// GetSerial retrieves the serial number of the NVIDIA device.
	GetSerial() (string, error)

	// GetPowerUsage retrieves the power usage of the NVIDIA device in milliwatts.
	GetPowerUsage() (uint, error)

	// GetPerformanceState retrieves the performance state (P-state) of the NVIDIA device.
	GetPerformanceState() (uint, error)

	// GetClockInfo retrieves the clock rate for the specified clock type of the NVIDIA device.
	GetClockInfo(clockType ClockType) (uint, error)

	// GetPowerManagementLimit retrieves the power management limit of the NVIDIA device in milliwatts.
	GetPowerManagementLimit() (uint, error)

	// GetTotalEnergyConsumption retrieves the total energy consumption of the NVIDIA device in millijoules.
	GetTotalEnergyConsumption() (uint64, error)

	// GetEncoderStats retrieves statistics related to the encoder activity on the device.
	// It returns the following statistics:
	//   - sessionCount: the number of active encoder sessions.
	//   - averageFps: the average frames per second across all active encoder sessions.
	//   - averageLatency: the average latency (in milliseconds) across all active encoder sessions.
	GetEncoderStats() (uint, uint, uint, error)

	// GetEncoderUtilization retrieves the encoder utilization statistics for the device.
	// It returns the following values:
	//   - utilization: the percentage of time over the past sampling period during which the encoder was active.
	//   - samplingPeriodUs: the sampling period duration in microseconds,
	//     indicating how long the utilization metric was measured.
	GetEncoderUtilization() (uint, uint, error)

	// GetDecoderUtilization retrieves the decoder utilization statistics for the device.
	// It returns the following values:
	//   - utilization: the percentage of time over the past sampling period during which the decoder was active.
	//   - samplingPeriodUs: the sampling period duration in microseconds, indicating how long the utilization
	//     metric was measured.
	GetDecoderUtilization() (uint, uint, error)

	// GetMemoryErrorCounter retrieves the ECC memory error count for the specified error type,
	// memory location, and counter type.
	GetMemoryErrorCounter(
		errorType MemoryErrorType,
		memoryLocation MemoryLocation,
		counterType EccCounterType,
	) (uint64, error)

	// GetEccMode retrieves the current and pending ECC (Error Correction Code) modes for the device.
	// ECC mode indicates whether error correction is enabled or disabled on the device.
	//
	// Returns:
	//   - `currentEnabled` (bool): `true` if ECC is currently enabled, `false` if disabled.
	//   - `pendingEnabled` (bool): `true` if ECC will be enabled on the next reboot, `false` if it will be disabled.
	//   - `error` (error): An error if the function fails to retrieve the ECC mode, otherwise `nil`.
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
