/*
**   Copyright 2001-2024 Zabbix SIA
**
**   Licensed under the Apache License, Version 2.0 (the "License");
**   you may not use this file except in compliance with the License.
**   You may obtain a copy of the License at
**
**       http://www.apache.org/licenses/LICENSE-2.0
**
**   Unless required by applicable law or agreed to in writing, software
**   distributed under the License is distributed on an "AS IS" BASIS,
**   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
**   See the License for the specific language governing permissions and
**   limitations under the License.
**/

package nvml

/*
#cgo CFLAGS: -I${SRCDIR}/nvml-sdk/include
#cgo CFLAGS: -DNVML_NO_UNVERSIONED_FUNC_DEFS=1

#include "nvml.h"
*/
import "C" //nolint:gocritic
import (
	"errors"
	"sync"
	"syscall"
	"unsafe" //nolint:gocritic

	"golang.org/x/sys/windows"
	"golang.zabbix.com/sdk/errs"
)

var (
	_ Device = (*NVMLDevice)(nil)
	_ Runner = (*NVMLRunner)(nil)
)

// NVMLRunner manages the loading and retrieval of functions from a Windows DLL.
// It is responsible for ensuring thread-safe access to the list of loaded
// procedures and facilitates dynamic function calls from the DLL.
type NVMLRunner struct {
	dll         *windows.DLL
	procListMux *sync.Mutex
	procList    map[string]*windows.Proc
}

// NVMLDevice represents an NVML GPU device, identified by a unique handle.
type NVMLDevice struct {
	handle uintptr
	runner *NVMLRunner // Reference to the Runner (formerly NVMLRunner)
}

// NewNVMLRunner creates a new NVML Runner instance, loading the NVML library.
func NewNVMLRunner() (*NVMLRunner, error) {
	dll, err := windows.LoadDLL("nvml.dll")
	if err != nil {
		return nil, errs.WrapConst(err, ErrLibraryNotFound) //nolint:wrapcheck
	}

	runner := &NVMLRunner{
		dll:         dll,
		procList:    make(map[string]*windows.Proc),
		procListMux: &sync.Mutex{},
	}

	return runner, nil
}

// Init initializes the NVML library using the older NVML interface.
func (runner *NVMLRunner) Init() error {
	err := runner.callProc("nvmlInit")
	if err != nil {
		return errs.Wrap(err, "failed while calling procedure")
	}

	return nil
}

// InitV2 initializes the NVML library using the NVML v2 interface.
func (runner *NVMLRunner) InitV2() error {
	err := runner.callProc("nvmlInit_v2")
	if err != nil {
		return errs.Wrap(err, "failed while calling procedure")
	}

	return nil
}

// GetNVMLVersion retrieves the version of the NVML library currently in use.
func (runner *NVMLRunner) GetNVMLVersion() (string, error) {
	var version [systemNVMLVersionBufferSize]byte

	err := runner.callProc("nvmlSystemGetNVMLVersion",
		uintptr(unsafe.Pointer(&version[0])),
		uintptr(systemNVMLVersionBufferSize),
	)
	if err != nil {
		return "", errs.Wrap(err, "failed while calling procedure")
	}

	return windows.ByteSliceToString(version[:]), nil
}

// GetDriverVersion retrieves the version of the NVIDIA driver currently in use.
func (runner *NVMLRunner) GetDriverVersion() (string, error) {
	var version [systemDriverVersionBufferSize]byte

	err := runner.callProc("nvmlSystemGetDriverVersion",
		uintptr(unsafe.Pointer(&version[0])),
		uintptr(systemDriverVersionBufferSize),
	)
	if err != nil {
		return "", errs.Wrap(err, "failed while calling procedure")
	}

	return windows.ByteSliceToString(version[:]), nil
}

// GetDeviceCountV2 retrieves the number of NVIDIA devices using the NVML v2 interface.
func (runner *NVMLRunner) GetDeviceCountV2() (uint, error) {
	var deviceCount C.uint

	err := runner.callProc("nvmlDeviceGetCount_v2", uintptr(unsafe.Pointer(&deviceCount)))
	if err != nil {
		return 0, errs.Wrap(err, "failed while calling procedure")
	}

	return uint(deviceCount), nil
}

// GetDeviceCount retrieves the number of NVIDIA devices using the standard NVML interface.
func (runner *NVMLRunner) GetDeviceCount() (uint, error) {
	var deviceCount C.uint

	err := runner.callProc("nvmlDeviceGetCount", uintptr(unsafe.Pointer(&deviceCount)))
	if err != nil {
		return 0, errs.Wrap(err, "failed while calling procedure")
	}

	return uint(deviceCount), nil
}

// GetDeviceByIndexV2 retrieves a handle to an NVIDIA device by its index using the NVML v2 interface.
//
//nolint:ireturn
func (runner *NVMLRunner) GetDeviceByIndexV2(index uint) (Device, error) {
	var deviceHandle uintptr

	err := runner.callProc("nvmlDeviceGetHandleByIndex_v2",
		uintptr(index),
		uintptr(unsafe.Pointer(&deviceHandle)),
	)
	if err != nil {
		return nil, errs.Wrap(err, "failed while calling procedure")
	}

	device := &NVMLDevice{
		handle: deviceHandle,
		runner: runner,
	}

	return device, nil
}

// GetDeviceByUUID retrieves a handle to an NVIDIA device by its UUID.
//
//nolint:ireturn
func (runner *NVMLRunner) GetDeviceByUUID(uuid string) (Device, error) {
	var deviceHandle uintptr

	cUUID, err := windows.ByteSliceFromString(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "uuid contains terminator sign")
	}

	if len(cUUID) > deviceUUIDBufferSize {
		return nil, errs.New("uuid string too long")
	}

	err = runner.callProc("nvmlDeviceGetHandleByUUID",
		uintptr(unsafe.Pointer(&cUUID[0])),
		uintptr(unsafe.Pointer(&deviceHandle)),
	)
	if err != nil {
		return nil, errs.Wrap(err, "failed while calling procedure")
	}

	device := &NVMLDevice{
		handle: deviceHandle,
		runner: runner,
	}

	return device, nil
}

// ShutdownNVML is a wrapper function to cleanly shut down NVML.
func (runner *NVMLRunner) ShutdownNVML() error {
	err := runner.callProc("nvmlShutdown")
	if err != nil {
		return errs.Wrap(err, "failed while calling procedure")
	}

	return nil
}

// Close releases the resources associated with the loaded DLL in the Runner.
func (runner *NVMLRunner) Close() error {
	callErr := runner.dll.Release()

	err := checkCallError(callErr)
	if err != nil {
		return err
	}

	return nil
}

// GetTemperature retrieves the temperature of the NVIDIA device using the default sensor.
func (device *NVMLDevice) GetTemperature() (int, error) {
	var temperature C.uint

	err := device.runner.callProc("nvmlDeviceGetTemperature",
		device.handle,
		0,
		uintptr(unsafe.Pointer(&temperature)),
	)
	if err != nil {
		return 0, errs.Wrap(err, "failed while calling procedure")
	}

	return int(temperature), nil
}

// GetName retrieves the name of the NVIDIA device.
func (device *NVMLDevice) GetName() (string, error) {
	var name [deviceNameBufferSize]byte

	err := device.runner.callProc("nvmlDeviceGetName",
		device.handle,
		uintptr(unsafe.Pointer(&name[0])),
		uintptr(deviceNameBufferSize),
	)
	if err != nil {
		return "", errs.Wrap(err, "failed while calling procedure")
	}

	return windows.ByteSliceToString(name[:]), nil
}

// GetMemoryInfoV2 retrieves detailed memory information for the NVIDIA device using the NVML v2 interface.
func (device *NVMLDevice) GetMemoryInfoV2() (*MemoryInfoV2, error) {
	var nvmlMemInfo C.nvmlMemory_v2_t

	// there is a macto in nvml.h called NVML_STRUCT_VERSION,
	// it does not work for CGO, so doing manually.
	nvmlMemInfo.version = C.uint(unsafe.Sizeof(nvmlMemInfo)) | (2 << 24)

	err := device.runner.callProc("nvmlDeviceGetMemoryInfo_v2",
		device.handle,
		uintptr(unsafe.Pointer(&nvmlMemInfo)),
	)
	if err != nil {
		return nil, errs.Wrap(err, "failed while calling procedure")
	}

	memInfo := &MemoryInfoV2{
		Total:    uint64(nvmlMemInfo.total),
		Reserved: uint64(nvmlMemInfo.reserved),
		Free:     uint64(nvmlMemInfo.free),
		Used:     uint64(nvmlMemInfo.used),
	}

	return memInfo, nil
}

// GetMemoryInfo retrieves memory information for the NVIDIA device.
func (device *NVMLDevice) GetMemoryInfo() (*MemoryInfo, error) {
	var memInfo C.nvmlMemory_t

	err := device.runner.callProc("nvmlDeviceGetMemoryInfo",
		device.handle,
		uintptr(unsafe.Pointer(&memInfo)),
	)
	if err != nil {
		return nil, errs.Wrap(err, "failed while calling procedure")
	}

	info := &MemoryInfo{
		Total: uint64(memInfo.total),
		Free:  uint64(memInfo.free),
		Used:  uint64(memInfo.used),
	}

	return info, nil
}

// GetBAR1MemoryInfo retrieves BAR1 memory information for the NVIDIA device.
func (device *NVMLDevice) GetBAR1MemoryInfo() (*MemoryInfo, error) {
	var memInfo C.nvmlBAR1Memory_t

	err := device.runner.callProc("nvmlDeviceGetBAR1MemoryInfo",
		device.handle,
		uintptr(unsafe.Pointer(&memInfo)),
	)
	if err != nil {
		return nil, errs.Wrap(err, "failed while calling procedure")
	}

	info := &MemoryInfo{
		Total: uint64(memInfo.bar1Total),
		Free:  uint64(memInfo.bar1Free),
		Used:  uint64(memInfo.bar1Used),
	}

	return info, nil
}

// GetFanSpeed retrieves the current fan speed of the NVIDIA device as a percentage of its maximum speed.
func (device *NVMLDevice) GetFanSpeed() (uint, error) {
	var speed C.uint

	err := device.runner.callProc("nvmlDeviceGetFanSpeed",
		device.handle,
		uintptr(unsafe.Pointer(&speed)),
	)
	if err != nil {
		return 0, errs.Wrap(err, "failed while calling procedure")
	}

	return uint(speed), nil
}

// GetPCIeThroughput retrieves the PCIe throughput for the NVIDIA device, based on the specified metric type.
func (device *NVMLDevice) GetPCIeThroughput(metricType PcieMetricType) (uint, error) {
	var throughput C.uint

	err := device.runner.callProc("nvmlDeviceGetPcieThroughput",
		device.handle,
		uintptr(metricType),
		uintptr(unsafe.Pointer(&throughput)),
	)
	if err != nil {
		return 0, errs.Wrap(err, "failed while calling procedure")
	}

	return uint(throughput), nil
}

// GetUUID retrieves the UUID of the NVIDIA device.
func (device *NVMLDevice) GetUUID() (string, error) {
	var uuid [deviceUUIDBufferSize]byte

	err := device.runner.callProc("nvmlDeviceGetUUID",
		device.handle,
		uintptr(unsafe.Pointer(&uuid[0])),
		uintptr(deviceUUIDBufferSize),
	)
	if err != nil {
		return "", errs.Wrap(err, "failed while calling procedure")
	}

	return windows.ByteSliceToString(uuid[:]), nil
}

// GetSerial retrieves the serial number of the NVIDIA device.
func (device *NVMLDevice) GetSerial() (string, error) {
	var serial [deviceSerialBufferSize]byte

	err := device.runner.callProc("nvmlDeviceGetSerial",
		device.handle,
		uintptr(unsafe.Pointer(&serial[0])),
		uintptr(deviceSerialBufferSize),
	)
	if err != nil {
		return "", errs.Wrap(err, "failed while calling procedure")
	}

	return windows.ByteSliceToString(serial[:]), nil
}

// GetEncoderUtilization retrieves the encoder utilization statistics for the device.
// It returns the following values:
//   - utilization: the percentage of time over the past sampling period during which the encoder was active.
//   - samplingPeriodUs: the sampling period duration in microseconds,
//     indicating how long the utilization metric was measured.
func (device *NVMLDevice) GetEncoderUtilization() (uint, uint, error) {
	var (
		utilization      C.uint
		samplingPeriodUs C.uint
	)

	err := device.runner.callProc("nvmlDeviceGetEncoderUtilization",
		device.handle,
		uintptr(unsafe.Pointer(&utilization)),
		uintptr(unsafe.Pointer(&samplingPeriodUs)),
	)
	if err != nil {
		return 0, 0, errs.Wrap(err, "failed while calling procedure")
	}

	return uint(utilization), uint(samplingPeriodUs), nil
}

// GetDecoderUtilization retrieves the decoder utilization statistics for the device.
// It returns the following values:
//   - utilization: the percentage of time over the past sampling period during which the decoder was active.
//   - samplingPeriodUs: the sampling period duration in microseconds, indicating how long the utilization
//     metric was measured.
func (device *NVMLDevice) GetDecoderUtilization() (uint, uint, error) {
	var (
		utilization      C.uint
		samplingPeriodUs C.uint
	)

	err := device.runner.callProc("nvmlDeviceGetDecoderUtilization",
		device.handle,
		uintptr(unsafe.Pointer(&utilization)),
		uintptr(unsafe.Pointer(&samplingPeriodUs)),
	)
	if err != nil {
		return 0, 0, errs.Wrap(err, "failed while calling procedure")
	}

	return uint(utilization), uint(samplingPeriodUs), nil
}

// GetMemoryErrorCounter retrieves the ECC memory error count for the specified error type,
// memory location, and counter type.
func (device *NVMLDevice) GetMemoryErrorCounter(
	errorType MemoryErrorType,
	memoryLocation MemoryLocation,
	counterType EccCounterType) (uint64, error) {
	var errorCount C.ulonglong

	err := device.runner.callProc("nvmlDeviceGetMemoryErrorCounter",
		device.handle,
		uintptr(errorType),
		uintptr(counterType),
		uintptr(memoryLocation),
		uintptr(unsafe.Pointer(&errorCount)),
	)
	if err != nil {
		return 0, errs.Wrap(err, "failed while calling procedure")
	}

	return uint64(errorCount), nil
}

// GetTotalEnergyConsumption retrieves the total energy consumption of the NVIDIA device in millijoules.
func (device *NVMLDevice) GetTotalEnergyConsumption() (uint64, error) {
	var energy C.ulonglong

	err := device.runner.callProc("nvmlDeviceGetTotalEnergyConsumption",
		device.handle,
		uintptr(unsafe.Pointer(&energy)),
	)
	if err != nil {
		return 0, errs.Wrap(err, "failed while calling procedure")
	}

	return uint64(energy), nil
}

// GetPerformanceState retrieves the performance state (P-state) of the NVIDIA device.
func (device *NVMLDevice) GetPerformanceState() (uint, error) {
	var perfState C.uint

	err := device.runner.callProc("nvmlDeviceGetPerformanceState",
		device.handle,
		uintptr(unsafe.Pointer(&perfState)),
	)
	if err != nil {
		return 0, errs.Wrap(err, "failed while calling procedure")
	}

	return uint(perfState), nil
}

// GetClockInfo retrieves the clock rate for the specified clock type of the NVIDIA device.
func (device *NVMLDevice) GetClockInfo(clockType ClockType) (uint, error) {
	var clockRate C.uint

	err := device.runner.callProc("nvmlDeviceGetClockInfo",
		device.handle,
		uintptr(clockType),
		uintptr(unsafe.Pointer(&clockRate)),
	)
	if err != nil {
		return 0, errs.Wrap(err, "failed while calling procedure")
	}

	return uint(clockRate), nil
}

// GetPowerUsage retrieves the power usage of the NVIDIA device in milliwatts.
func (device *NVMLDevice) GetPowerUsage() (uint, error) {
	var power C.uint

	err := device.runner.callProc("nvmlDeviceGetPowerUsage",
		device.handle,
		uintptr(unsafe.Pointer(&power)),
	)
	if err != nil {
		return 0, errs.Wrap(err, "failed while calling procedure")
	}

	return uint(power), nil
}

// GetEncoderStats retrieves statistics related to the encoder activity on the device.
// It returns the following statistics:
//   - sessionCount: the number of active encoder sessions.
//   - averageFps: the average frames per second across all active encoder sessions.
//   - averageLatency: the average latency (in milliseconds) across all active encoder sessions.
func (device *NVMLDevice) GetEncoderStats() (uint, uint, uint, error) {
	var (
		sessionCount   C.uint
		averageFps     C.uint
		averageLatency C.uint
	)

	err := device.runner.callProc("nvmlDeviceGetPowerUsage",
		device.handle,
		uintptr(unsafe.Pointer(&sessionCount)),
		uintptr(unsafe.Pointer(&averageFps)),
		uintptr(unsafe.Pointer(&averageLatency)),
	)
	if err != nil {
		return 0, 0, 0, errs.Wrap(err, "failed while calling procedure")
	}

	return uint(sessionCount), uint(averageFps), uint(averageLatency), nil
}

// GetPowerManagementLimit retrieves the power management limit of the NVIDIA device in milliwatts.
func (device *NVMLDevice) GetPowerManagementLimit() (uint, error) {
	var powerLimit C.uint

	err := device.runner.callProc("nvmlDeviceGetPowerManagementLimit",
		device.handle,
		uintptr(unsafe.Pointer(&powerLimit)),
	)
	if err != nil {
		return 0, errs.Wrap(err, "failed while calling procedure")
	}

	return uint(powerLimit), nil
}

// GetEccMode retrieves the current and pending ECC (Error Correction Code) modes for the device.
// ECC mode indicates whether error correction is enabled or disabled on the device.
//
// Returns:
//   - `currentEnabled` (bool): `true` if ECC is currently enabled, `false` if disabled.
//   - `pendingEnabled` (bool): `true` if ECC will be enabled on the next reboot, `false` if it will be disabled.
//   - `error` (error): An error if the function fails to retrieve the ECC mode, otherwise `nil`.
func (device *NVMLDevice) GetEccMode() (bool, bool, error) {
	var (
		currentMode C.uint
		pendingMode C.uint
	)

	err := device.runner.callProc("nvmlDeviceGetEccMode",
		device.handle,
		uintptr(unsafe.Pointer(&currentMode)),
		uintptr(unsafe.Pointer(&pendingMode)),
	)
	if err != nil {
		return false, false, errs.Wrap(err, "failed while calling procedure")
	}

	currentEnabled := currentMode == C.NVML_FEATURE_ENABLED
	pendingEnabled := pendingMode == C.NVML_FEATURE_ENABLED

	return currentEnabled, pendingEnabled, nil
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
	var util C.nvmlUtilization_t

	err := device.runner.callProc("nvmlDeviceGetUtilizationRates",
		device.handle,
		uintptr(unsafe.Pointer(&util)),
	)
	if err != nil {
		return 0, 0, errs.Wrap(err, "failed while calling procedure")
	}

	return uint(util.gpu), uint(util.memory), nil
}

// getProc retrieves the specified procedure from the Windows DLL.
// If the procedure is already cached in procList, it returns the cached
// *windows.Proc. Otherwise, it finds the procedure using dll.FindProc, caches it,
// and returns it.
func (runner *NVMLRunner) getProc(procName string) (*windows.Proc, error) {
	runner.procListMux.Lock()
	defer runner.procListMux.Unlock()

	proc, ok := runner.procList[procName]
	if ok {
		return proc, nil
	}

	proc, err := runner.dll.FindProc(procName)
	if err != nil {
		return nil, errs.Wrapf(ErrFunctionNotFound, "failed to get procedure %q", procName)
	}

	runner.procList[procName] = proc

	return proc, nil
}

func (runner *NVMLRunner) callProc(procName string, args ...uintptr) error {
	proc, err := runner.getProc(procName)
	if err != nil {
		return errs.Wrap(err, "failed getting procedure")
	}

	result, _, callErr := proc.Call(args...)
	err = checkCallError(callErr)
	if err != nil {
		return errs.Wrap(err, "failed making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return errs.Wrap(err, "NVML returned error")
	}

	return nil
}

// checkCallError checks for and interprets errors returned from system calls.
//
// Parameters:
// - callErr: The error returned from a system call, which may be of type syscall.Errno.
func checkCallError(callErr error) error {
	if callErr == nil {
		return nil
	}

	var errno syscall.Errno
	if errors.As(callErr, &errno) {
		if errno != windows.ERROR_SUCCESS {
			return errs.Errorf("failed with error code %d: %v", errno, callErr)
		}

		return nil
	}

	return callErr
}
