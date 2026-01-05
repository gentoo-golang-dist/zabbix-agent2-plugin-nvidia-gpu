/*
** Zabbix
** Copyright (C) 2001-2026 Zabbix SIA
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

package nvml //nolint:gocritic

/*
#cgo CFLAGS: -I${SRCDIR}/nvml-sdk/include
#cgo CFLAGS: -DNVML_NO_UNVERSIONED_FUNC_DEFS=1

#cgo linux LDFLAGS: -Wl,--export-dynamic -Wl,--unresolved-symbols=ignore-in-object-files

#include "nvml.h"

#cgo linux LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>
*/
import "C" //nolint:gocritic,gci
import (
	"sync"
	"unsafe" //nolint:gocritic

	"golang.zabbix.com/sdk/errs"
)

var (
	_ Device = (*NVMLDevice)(nil)
	_ Runner = (*NVMLRunner)(nil)
)

// NVMLRunner represents the NVML runner, responsible for managing
// the dynamic loading and initialization of the NVML library.
// It holds a pointer to the dynamically loaded library.
type NVMLRunner struct {
	dynamicLib  unsafe.Pointer // Pointer to the dynamically loaded NVML library
	procListMux *sync.Mutex
	procList    map[string]struct{}
}

// NVMLDevice represents an NVIDIA device and provides methods to interact with and retrieve information from it.
// It holds a handle to the NVML device and a reference to the Runner, which manages NVML operations and symbol loading.
type NVMLDevice struct {
	handle C.nvmlDevice_t // Handle to the NVML device
	runner *NVMLRunner    // Reference to the Runner for managing NVML operations
}

// NewNVMLRunner creates a new NVML Runner instance, loading the NVML library.
func NewNVMLRunner() (*NVMLRunner, error) {
	dynamicLib, err := loadLibrary()
	if err != nil {
		return nil, err
	}

	runner := &NVMLRunner{
		dynamicLib:  dynamicLib,
		procListMux: &sync.Mutex{},
		procList:    make(map[string]struct{}),
	}

	return runner, nil
}

// Init initializes the NVML library using the older NVML interface.
func (runner *NVMLRunner) Init() error {
	err := runner.symbolExists("nvmlInit")
	if err != nil {
		return errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	result := C.nvmlInit()
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return errs.Wrap(err, "failed to initialize NVML")
	}

	return nil
}

// InitV2 initializes the NVML library using the NVML v2 interface.
func (runner *NVMLRunner) InitV2() error {
	err := runner.symbolExists("nvmlInit_v2")
	if err != nil {
		return errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	result := C.nvmlInit_v2()
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return errs.Wrap(err, "failed to initialize NVML v2")
	}

	return nil
}

// GetDriverVersion retrieves the version of the NVIDIA driver currently in use.
func (runner *NVMLRunner) GetDriverVersion() (string, error) {
	err := runner.symbolExists("nvmlSystemGetDriverVersion")
	if err != nil {
		return "", errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var version [systemDriverVersionBufferSize]C.char

	result := C.nvmlSystemGetDriverVersion(&version[0], systemDriverVersionBufferSize)
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return "", errs.Wrap(err, "failed to get NVML driver version")
	}

	return C.GoString(&version[0]), nil
}

// GetNVMLVersion retrieves the version of the NVML library currently in use.
func (runner *NVMLRunner) GetNVMLVersion() (string, error) {
	err := runner.symbolExists("nvmlSystemGetNVMLVersion")
	if err != nil {
		return "", errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var version [systemNVMLVersionBufferSize]C.char

	result := C.nvmlSystemGetNVMLVersion(&version[0], systemNVMLVersionBufferSize)
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return "", errs.Wrap(err, "failed to get NVML version")
	}

	return C.GoString(&version[0]), nil
}

// GetDeviceCountV2 retrieves the number of NVIDIA devices using the NVML v2 interface.
func (runner *NVMLRunner) GetDeviceCountV2() (uint, error) {
	err := runner.symbolExists("nvmlDeviceGetCount_v2")
	if err != nil {
		return 0, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var deviceCount C.uint

	result := C.nvmlDeviceGetCount_v2(&deviceCount)
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "failed to get NVML device count")
	}

	return uint(deviceCount), nil
}

// GetDeviceCount retrieves the number of NVIDIA devices using the standard NVML interface.
func (runner *NVMLRunner) GetDeviceCount() (uint, error) {
	err := runner.symbolExists("nvmlDeviceGetCount")
	if err != nil {
		return 0, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var deviceCount C.uint

	result := C.nvmlDeviceGetCount(&deviceCount)
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "failed to get NVML device count")
	}

	return uint(deviceCount), nil
}

// GetDeviceByIndexV2 retrieves a handle to an NVIDIA device by its index using the NVML v2 interface.
//
//nolint:ireturn
func (runner *NVMLRunner) GetDeviceByIndexV2(index uint) (Device, error) {
	err := runner.symbolExists("nvmlDeviceGetHandleByIndex_v2")
	if err != nil {
		return nil, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var deviceHandle C.nvmlDevice_t

	result := C.nvmlDeviceGetHandleByIndex_v2(C.uint(index), &deviceHandle) //nolint:gocritic,nlreturn

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return nil, errs.Wrap(err, "failed to get NVML device handle by index")
	}

	return &NVMLDevice{
		handle: deviceHandle,
		runner: runner,
	}, nil
}

// GetDeviceByUUID retrieves a handle to an NVIDIA device by its UUID.
//
//nolint:ireturn
func (runner *NVMLRunner) GetDeviceByUUID(uuid string) (Device, error) {
	err := runner.symbolExists("nvmlDeviceGetHandleByUUID")
	if err != nil {
		return nil, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	// +1 accounts for terminator sign that is going to be added in the next step.
	if len(uuid)+1 > deviceUUIDBufferSize {
		return nil, errs.New("UUID string too long")
	}

	// Adds terminator sign inside
	cUUID := C.CString(uuid)

	defer C.free(unsafe.Pointer(cUUID)) //nolint:nlreturn

	var deviceHandle C.nvmlDevice_t

	result := C.nvmlDeviceGetHandleByUUID(cUUID, &deviceHandle) //nolint:gocritic,nlreturn
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return nil, errs.Wrap(err, "failed to get NVML device handle by UUID")
	}

	return &NVMLDevice{
		handle: deviceHandle,
		runner: runner,
	}, nil
}

// GetTemperature retrieves the temperature of the NVIDIA device using the default sensor.
func (device *NVMLDevice) GetTemperature() (int, error) {
	err := device.runner.symbolExists("nvmlDeviceGetTemperature")
	if err != nil {
		return 0, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var temperature C.uint

	// 0 is currently the only sensor available
	result := C.nvmlDeviceGetTemperature(device.handle, 0, &temperature) //nolint:nlreturn
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "failed to get NVML device temperature")
	}

	return int(temperature), nil
}

// GetFanSpeed retrieves the current fan speed of the NVIDIA device as a percentage of its maximum speed.
func (device *NVMLDevice) GetFanSpeed() (uint, error) {
	err := device.runner.symbolExists("nvmlDeviceGetFanSpeed")
	if err != nil {
		return 0, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var fanSpeed C.uint

	result := C.nvmlDeviceGetFanSpeed(device.handle, &fanSpeed) //nolint:nlreturn
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "failed to get NVML device fan speed")
	}

	return uint(fanSpeed), nil
}

// GetUUID retrieves the UUID of the NVIDIA device.
func (device *NVMLDevice) GetUUID() (string, error) {
	err := device.runner.symbolExists("nvmlDeviceGetUUID")
	if err != nil {
		return "", errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var deviceUUID [deviceUUIDBufferSize]C.char

	result := C.nvmlDeviceGetUUID(device.handle, &deviceUUID[0], deviceUUIDBufferSize) //nolint:nlreturn
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return "", errs.Wrap(err, "failed to get NVML device UUID")
	}

	return C.GoString(&deviceUUID[0]), nil
}

// GetSerial retrieves the serial number of the NVIDIA device.
func (device *NVMLDevice) GetSerial() (string, error) {
	err := device.runner.symbolExists("nvmlDeviceGetSerial")
	if err != nil {
		return "", errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var serial [deviceSerialBufferSize]C.char

	result := C.nvmlDeviceGetSerial(device.handle, &serial[0], deviceSerialBufferSize) //nolint:nlreturn
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return "", errs.Wrap(err, "failed to get NVML device serial")
	}

	return C.GoString(&serial[0]), nil
}

// GetName retrieves the name of the NVIDIA device.
func (device *NVMLDevice) GetName() (string, error) {
	err := device.runner.symbolExists("nvmlDeviceGetName")
	if err != nil {
		return "", errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var deviceName [deviceNameBufferSize]C.char

	result := C.nvmlDeviceGetName(device.handle, &deviceName[0], deviceNameBufferSize) //nolint:nlreturn
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return "", errs.Wrap(err, "failed to get NVML device name")
	}

	return C.GoString(&deviceName[0]), nil
}

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
func (device *NVMLDevice) GetUtilizationRates() (uint, uint, error) {
	err := device.runner.symbolExists("nvmlDeviceGetUtilizationRates")
	if err != nil {
		return 0, 0, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var utilization C.nvmlUtilization_t

	result := C.nvmlDeviceGetUtilizationRates(device.handle, &utilization) //nolint:nlreturn
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, 0, errs.Wrap(err, "failed to get NVML device utilization rates")
	}

	return uint(utilization.gpu), uint(utilization.memory), nil
}

// GetMemoryInfoV2 retrieves detailed memory information for the NVIDIA device using the NVML v2 interface.
func (device *NVMLDevice) GetMemoryInfoV2() (*MemoryInfoV2, error) {
	err := device.runner.symbolExists("nvmlDeviceGetMemoryInfo_v2")
	if err != nil {
		return nil, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var memory C.nvmlMemory_v2_t
	result := C.nvmlDeviceGetMemoryInfo_v2(device.handle, &memory) //nolint:nlreturn
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return nil, errs.Wrap(err, "failed to get NVML device memory info")
	}

	return &MemoryInfoV2{
		Total:    uint64(memory.total),
		Free:     uint64(memory.free),
		Used:     uint64(memory.used),
		Reserved: uint64(memory.reserved),
	}, nil
}

// GetBAR1MemoryInfo retrieves BAR1 memory information for the NVIDIA device.
func (device *NVMLDevice) GetBAR1MemoryInfo() (*MemoryInfo, error) {
	err := device.runner.symbolExists("nvmlDeviceGetBAR1MemoryInfo")
	if err != nil {
		return nil, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var memory C.nvmlBAR1Memory_t
	result := C.nvmlDeviceGetBAR1MemoryInfo(device.handle, &memory) //nolint:nlreturn
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return nil, errs.Wrap(err, "failed to get NVML BAR1 memory info")
	}

	return &MemoryInfo{
		Total: uint64(memory.bar1Total),
		Free:  uint64(memory.bar1Free),
		Used:  uint64(memory.bar1Used),
	}, nil
}

// GetMemoryInfo retrieves memory information for the NVIDIA device.
func (device *NVMLDevice) GetMemoryInfo() (*MemoryInfo, error) {
	err := device.runner.symbolExists("nvmlDeviceGetMemoryInfo")
	if err != nil {
		return nil, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var memory C.nvmlMemory_t
	result := C.nvmlDeviceGetMemoryInfo(device.handle, &memory) //nolint:nlreturn
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return nil, errs.Wrap(err, "failed to get NVML device memory info")
	}

	return &MemoryInfo{
		Total: uint64(memory.total),
		Free:  uint64(memory.free),
		Used:  uint64(memory.used),
	}, nil
}

// GetPCIeThroughput retrieves the PCIe throughput for the NVIDIA device, based on the specified metric type.
func (device *NVMLDevice) GetPCIeThroughput(metricType PcieMetricType) (uint, error) {
	err := device.runner.symbolExists("nvmlDeviceGetPcieThroughput")
	if err != nil {
		return 0, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var throughput C.uint

	metricTypeC := C.nvmlPcieUtilCounter_t(metricType)
	result := C.nvmlDeviceGetPcieThroughput(device.handle, metricTypeC, &throughput) //nolint:nlreturn

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "failed to get NVML PCIe throughput")
	}

	return uint(throughput), nil
}

// GetClockInfo retrieves the clock rate for the specified clock type of the NVIDIA device.
func (device *NVMLDevice) GetClockInfo(clockType ClockType) (uint, error) {
	err := device.runner.symbolExists("nvmlDeviceGetClockInfo")
	if err != nil {
		return 0, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var clockRate C.uint

	clockTypeC := C.nvmlClockType_t(clockType)
	result := C.nvmlDeviceGetClockInfo(device.handle, clockTypeC, &clockRate) //nolint:nlreturn

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "failed to get NVML device clock info")
	}

	return uint(clockRate), nil
}

// GetPowerUsage retrieves the power usage of the NVIDIA device in milliwatts.
func (device *NVMLDevice) GetPowerUsage() (uint, error) {
	err := device.runner.symbolExists("nvmlDeviceGetPowerUsage")
	if err != nil {
		return 0, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var power C.uint
	result := C.nvmlDeviceGetPowerUsage(device.handle, &power) //nolint:nlreturn
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "failed to get NVML device power usage")
	}

	return uint(power), nil
}

// GetEncoderStats retrieves statistics related to the encoder activity on the device.
// It returns the following statistics:
//   - sessionCount: the number of active encoder sessions.
//   - averageFps: the average frames per second across all active encoder sessions.
//   - averageLatency: the average latency (in milliseconds) across all active encoder sessions.
func (device *NVMLDevice) GetEncoderStats() (uint, uint, uint, error) {
	err := device.runner.symbolExists("nvmlDeviceGetEncoderStats")
	if err != nil {
		return 0, 0, 0, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var (
		sessionCount   C.uint
		averageFps     C.uint
		averageLatency C.uint
	)

	result := C.nvmlDeviceGetEncoderStats(device.handle, &sessionCount, &averageFps, &averageLatency) //nolint:nlreturn
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, 0, 0, errs.Wrap(err, "failed to get NVML device encoder stats")
	}

	return uint(sessionCount), uint(averageFps), uint(averageLatency), nil
}

// GetPowerManagementLimit retrieves the power management limit of the NVIDIA device in milliwatts.
func (device *NVMLDevice) GetPowerManagementLimit() (uint, error) {
	err := device.runner.symbolExists("nvmlDeviceGetPowerManagementLimit")
	if err != nil {
		return 0, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var powerLimit C.uint
	result := C.nvmlDeviceGetPowerManagementLimit(device.handle, &powerLimit) //nolint:nlreturn
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "failed to get NVML device power management limit")
	}

	return uint(powerLimit), nil
}

// GetTotalEnergyConsumption retrieves the total energy consumption of the NVIDIA device in millijoules.
func (device *NVMLDevice) GetTotalEnergyConsumption() (uint64, error) {
	err := device.runner.symbolExists("nvmlDeviceGetTotalEnergyConsumption")
	if err != nil {
		return 0, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var energy C.ulonglong
	result := C.nvmlDeviceGetTotalEnergyConsumption(device.handle, &energy) //nolint:nlreturn
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "failed to get NVML device total energy consumption")
	}

	return uint64(energy), nil
}

// GetPerformanceState retrieves the performance state (P-state) of the NVIDIA device.
func (device *NVMLDevice) GetPerformanceState() (uint, error) {
	err := device.runner.symbolExists("nvmlDeviceGetPerformanceState")
	if err != nil {
		return 0, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var perfState C.nvmlPstates_t
	result := C.nvmlDeviceGetPerformanceState(device.handle, &perfState) //nolint:nlreturn
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "failed to get NVML device performance state")
	}

	return uint(perfState), nil
}

// GetEncoderUtilization retrieves the encoder utilization statistics for the device.
// It returns the following values:
//   - utilization: the percentage of time over the past sampling period during which the encoder was active.
//   - samplingPeriodUs: the sampling period duration in microseconds,
//     indicating how long the utilization metric was measured.
func (device *NVMLDevice) GetEncoderUtilization() (uint, uint, error) {
	err := device.runner.symbolExists("nvmlDeviceGetEncoderUtilization")
	if err != nil {
		return 0, 0, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var (
		utilization      C.uint
		samplingPeriodUs C.uint
	)

	result := C.nvmlDeviceGetEncoderUtilization(
		device.handle, &utilization, &samplingPeriodUs, //nolint:nlreturn
	)

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, 0, errs.Wrap(err, "failed to get NVML device encoder utilization")
	}

	return uint(utilization), uint(samplingPeriodUs), nil
}

// GetDecoderUtilization retrieves the decoder utilization statistics for the device.
// It returns the following values:
//   - utilization: the percentage of time over the past sampling period during which the decoder was active.
//   - samplingPeriodUs: the sampling period duration in microseconds, indicating how long the utilization
//     metric was measured.
func (device *NVMLDevice) GetDecoderUtilization() (uint, uint, error) {
	err := device.runner.symbolExists("nvmlDeviceGetDecoderUtilization")
	if err != nil {
		return 0, 0, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var (
		utilization      C.uint
		samplingPeriodUs C.uint
	)

	result := C.nvmlDeviceGetDecoderUtilization(device.handle, &utilization, &samplingPeriodUs) //nolint:nlreturn
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, 0, errs.Wrap(err, "failed to get NVML device decoder utilization")
	}

	return uint(utilization), uint(samplingPeriodUs), nil
}

// GetMemoryErrorCounter retrieves the ECC memory error count for the specified error type,
// memory location, and counter type.
func (device *NVMLDevice) GetMemoryErrorCounter(
	errorType MemoryErrorType,
	memoryLocation MemoryLocation,
	counterType EccCounterType) (uint64, error) {
	err := device.runner.symbolExists("nvmlDeviceGetMemoryErrorCounter")
	if err != nil {
		return 0, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var errorCount C.ulonglong

	var (
		errorTypeC      = C.nvmlMemoryErrorType_t(errorType)
		memoryLocationC = C.nvmlMemoryLocation_t(memoryLocation)
		counterTypeC    = C.nvmlEccCounterType_t(counterType)
	)

	result := C.nvmlDeviceGetMemoryErrorCounter(
		device.handle,
		errorTypeC,
		counterTypeC,
		memoryLocationC,
		&errorCount, //nolint:nlreturn
	)

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "failed to get NVML memory error counter")
	}

	return uint64(errorCount), nil
}

// GetEccMode retrieves the current and pending ECC (Error Correction Code) modes for the device.
// ECC mode indicates whether error correction is enabled or disabled on the device.
//
// Returns:
//   - `currentEnabled` (bool): `true` if ECC is currently enabled, `false` if disabled.
//   - `pendingEnabled` (bool): `true` if ECC will be enabled on the next reboot, `false` if it will be disabled.
//   - `error` (error): An error if the function fails to retrieve the ECC mode, otherwise `nil`.
func (device *NVMLDevice) GetEccMode() (bool, bool, error) {
	err := device.runner.symbolExists("nvmlDeviceGetEccMode")
	if err != nil {
		return false, false, errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	var currentMode C.nvmlEnableState_t
	var pendingMode C.nvmlEnableState_t

	result := C.nvmlDeviceGetEccMode(device.handle, &currentMode, &pendingMode) //nolint:nlreturn
	err = mapNVMLResultToError(int(result))
	if err != nil {
		return false, false, errs.Wrap(err, "failed to get NVML ECC mode")
	}

	currentEnabled := currentMode == C.NVML_FEATURE_ENABLED
	pendingEnabled := pendingMode == C.NVML_FEATURE_ENABLED

	return currentEnabled, pendingEnabled, nil
}

// ShutdownNVML is a wrapper function to cleanly shut down NVML.
func (runner *NVMLRunner) ShutdownNVML() error {
	err := runner.symbolExists("nvmlShutdown")
	if err != nil {
		return errs.Wrap(err, "failed to verify existence of NVML symbol")
	}

	result := C.nvmlShutdown()

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return errs.Wrap(err, "failed to shutdown NVML")
	}

	return nil
}

// Close releases the resources associated with the dynamically loaded library.
func (runner *NVMLRunner) Close() error {
	C.dlclose(runner.dynamicLib) //nolint:nlreturn

	return nil
}

func (runner *NVMLRunner) symbolExists(funcName string) error {
	runner.procListMux.Lock()
	defer runner.procListMux.Unlock()

	_, ok := runner.procList[funcName]
	if ok {
		return nil
	}

	initSymbol := C.CString(funcName)
	defer C.free(unsafe.Pointer(initSymbol)) //nolint:nlreturn

	initPtr := C.dlsym(runner.dynamicLib, initSymbol) //nolint:nlreturn
	if initPtr == nil {
		return errs.Wrapf(ErrFunctionNotFound, "failed to get procedure %q", funcName)
	}

	runner.procList[funcName] = struct{}{}

	return nil
}

func loadLibrary() (unsafe.Pointer, error) {
	libName := C.CString("libnvidia-ml.so")
	defer C.free(unsafe.Pointer(libName)) //nolint:nlreturn

	handle := C.dlopen(libName, C.RTLD_LAZY|C.RTLD_GLOBAL)
	if handle == nil {
		return nil, ErrLibraryNotFound
	}

	return handle, nil
}
