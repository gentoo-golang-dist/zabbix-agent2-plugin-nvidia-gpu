package nvml

import (
	"errors"
	"sync"
	"syscall"
	"unsafe"

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
		return nil, errs.Wrap(ErrFunctionNotFound, "error getting procedure: "+procName)
	}

	runner.procList[procName] = proc

	return proc, nil
}

// NewNVMLRunner creates a new NVML Runner instance, loading the NVML library.
func NewNVMLRunner() (*NVMLRunner, error) {
	dll, err := windows.LoadDLL("nvml.dll")
	if err != nil {
		return nil, ErrLibraryNotFound
	}

	runner := &NVMLRunner{
		dll:         dll,
		procList:    make(map[string]*windows.Proc),
		procListMux: &sync.Mutex{},
	}

	return runner, nil
}

// InitNVML initializes the NVML library using the older NVML interface.
func (runner *NVMLRunner) InitNVML() error {
	proc, err := runner.getProc("nvmlInit")
	if err != nil {
		return errs.Wrap(err, "error getting procedure")
	}

	result, _, callErr := proc.Call()

	err = checkCallError(callErr)
	if err != nil {
		return errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return errs.Wrap(err, "NVML returned error")
	}

	return nil
}

// InitNVMLv2 initializes the NVML library using the NVML v2 interface.
func (runner *NVMLRunner) InitNVMLv2() error {
	proc, err := runner.getProc("nvmlInit_v2")
	if err != nil {
		return errs.Wrap(err, "error getting procedure")
	}

	result, _, callErr := proc.Call()

	err = checkCallError(callErr)
	if err != nil {
		return errs.Wrap(err, "error making syscall")
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

// GetNVMLVersion retrieves the version of the NVML library currently in use.
func (runner *NVMLRunner) GetNVMLVersion() (string, error) {
	proc, err := runner.getProc("nvmlSystemGetNVMLVersion")
	if err != nil {
		return "", errs.Wrap(err, "error getting procedure")
	}

	var version [systemNVMLVersionBufferSize]byte

	result, _, callErr := proc.Call(
		uintptr(unsafe.Pointer(&version[0])), //nolint:gosec
		uintptr(systemNVMLVersionBufferSize),
	)

	err = checkCallError(callErr)
	if err != nil {
		return "", errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return "", errs.Wrap(err, "NVML returned error")
	}

	return windows.ByteSliceToString(version[:]), nil
}

// GetDriverVersion retrieves the version of the NVIDIA driver currently in use.
func (runner *NVMLRunner) GetDriverVersion() (string, error) {
	proc, err := runner.getProc("nvmlSystemGetDriverVersion")
	if err != nil {
		return "", errs.Wrap(err, "error getting procedure")
	}

	var version [systemDriverVersionBufferSize]byte

	result, _, callErr := proc.Call(
		uintptr(unsafe.Pointer(&version[0])), //nolint:gosec
		uintptr(systemDriverVersionBufferSize),
	)

	err = checkCallError(callErr)
	if err != nil {
		return "", errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return "", errs.Wrap(err, "NVML returned error")
	}

	versionString := windows.ByteSliceToString(version[:])

	return versionString, nil
}

// GetDeviceCountV2 retrieves the number of NVIDIA devices using the NVML v2 interface.
func (runner *NVMLRunner) GetDeviceCountV2() (uint, error) {
	proc, err := runner.getProc("nvmlDeviceGetCount_v2")
	if err != nil {
		return 0, errs.Wrap(err, "error getting procedure")
	}

	var deviceCount uint32

	result, _, callErr := proc.Call(uintptr(unsafe.Pointer(&deviceCount))) //nolint:gosec

	err = checkCallError(callErr)
	if err != nil {
		return 0, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "NVML returned error")
	}

	return uint(deviceCount), nil
}

// GetDeviceCount retrieves the number of NVIDIA devices using the standard NVML interface.
func (runner *NVMLRunner) GetDeviceCount() (uint, error) {
	proc, err := runner.getProc("nvmlDeviceGetCount")
	if err != nil {
		return 0, errs.Wrap(err, "error getting procedure")
	}

	var deviceCount uint32

	result, _, callErr := proc.Call(uintptr(unsafe.Pointer(&deviceCount))) //nolint:gosec

	err = checkCallError(callErr)
	if err != nil {
		return 0, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "NVML returned error")
	}

	return uint(deviceCount), nil
}

// GetDeviceByIndexV2 retrieves a handle to an NVIDIA device by its index using the NVML v2 interface.
func (runner *NVMLRunner) GetDeviceByIndexV2(index uint) (*NVMLDevice, error) {
	proc, err := runner.getProc("nvmlDeviceGetHandleByIndex_v2")
	if err != nil {
		return nil, errs.Wrap(err, "error getting procedure")
	}

	var deviceHandle uintptr

	result, _, callErr := proc.Call(
		uintptr(index),
		uintptr(unsafe.Pointer(&deviceHandle)), //nolint:gosec
	)

	err = checkCallError(callErr)
	if err != nil {
		return nil, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return nil, errs.Wrap(err, "NVML returned error")
	}

	device := &NVMLDevice{
		handle: deviceHandle,
		runner: runner,
	}

	return device, nil
}

// GetDeviceByUUID retrieves a handle to an NVIDIA device by its UUID.
func (runner *NVMLRunner) GetDeviceByUUID(uuid string) (*NVMLDevice, error) {
	proc, err := runner.getProc("nvmlDeviceGetHandleByUUID")
	if err != nil {
		return nil, errs.Wrap(err, "error getting procedure")
	}

	var deviceHandle uintptr

	cUUID, err := windows.ByteSliceFromString(uuid)
	if err != nil {
		return nil, errs.Wrap(err, "uuid contains terminator sign")
	}

	if len(cUUID) > deviceUUIDBufferSize {
		return nil, errs.New("uuid string too long")
	}

	result, _, callErr := proc.Call(
		uintptr(unsafe.Pointer(&cUUID[0])),     //nolint:gosec
		uintptr(unsafe.Pointer(&deviceHandle)), //nolint:gosec
	)

	err = checkCallError(callErr)
	if err != nil {
		return nil, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return nil, errs.Wrap(err, "NVML returned error")
	}

	device := &NVMLDevice{
		handle: deviceHandle,
		runner: runner,
	}

	return device, nil
}

// ShutdownNVML is a wrapper function to cleanly shut down NVML.
func (runner *NVMLRunner) ShutdownNVML() error {
	proc, err := runner.getProc("nvmlShutdown")
	if err != nil {
		return errs.Wrap(err, "error getting procedure")
	}

	result, _, callErr := proc.Call()

	err = checkCallError(callErr)
	if err != nil {
		return errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return errs.Wrap(err, "NVML returned error")
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
	proc, err := device.runner.getProc("nvmlDeviceGetTemperature")
	if err != nil {
		return 0, errs.Wrap(err, "error getting procedure")
	}

	var temperature uint32

	result, _, callErr := proc.Call(
		device.handle,
		0,
		uintptr(unsafe.Pointer(&temperature)), //nolint:gosec
	)

	err = checkCallError(callErr)
	if err != nil {
		return 0, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "NVML returned error")
	}

	return int(temperature), nil
}

// GetName retrieves the name of the NVIDIA device.
func (device *NVMLDevice) GetName() (string, error) {
	proc, err := device.runner.getProc("nvmlDeviceGetName")
	if err != nil {
		return "", errs.Wrap(err, "error getting procedure")
	}

	var name [deviceNameBufferSize]byte

	result, _, callErr := proc.Call(
		device.handle,
		uintptr(unsafe.Pointer(&name[0])), //nolint:gosec
		uintptr(deviceNameBufferSize),
	)

	err = checkCallError(callErr)
	if err != nil {
		return "", errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return "", errs.Wrap(err, "NVML returned error")
	}

	return windows.ByteSliceToString(name[:]), nil
}

// GetMemoryInfoV2 retrieves detailed memory information for the NVIDIA device using the NVML v2 interface.
func (device *NVMLDevice) GetMemoryInfoV2() (*MemoryInfoV2, error) {
	proc, err := device.runner.getProc("nvmlDeviceGetMemoryInfo_v2")
	if err != nil {
		return nil, errs.Wrap(err, "error getting procedure")
	}

	// Define the nvmlMemoryV2 structure
	type nvmlMemoryV2 struct {
		version  uint32 // Set this field explicitly
		_        [4]byte
		total    uint64 // Total physical device memory (in bytes)
		reserved uint64 // Memory reserved for system use (in bytes)
		free     uint64 // Unallocated device memory (in bytes)
		used     uint64 // Allocated device memory (in bytes)
	}

	// Initialize the structure
	var nvmlMemInfo nvmlMemoryV2

	// Manually set the version field
	// Structure size is 40 bytes, so we set it as: (40 | (2 << 24))
	nvmlMemInfo.version = 40 | (2 << 24)
	// Call the NVML function
	result, _, callErr := proc.Call(
		device.handle,
		uintptr(unsafe.Pointer(&nvmlMemInfo)), //nolint:gosec
	)

	// Check for errors
	err = checkCallError(callErr)
	if err != nil {
		return nil, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return nil, err
	}

	memInfo := &MemoryInfoV2{
		Total:    nvmlMemInfo.total,
		Reserved: nvmlMemInfo.reserved,
		Free:     nvmlMemInfo.free,
		Used:     nvmlMemInfo.used,
	}

	return memInfo, nil
}

// GetMemoryInfo retrieves memory information for the NVIDIA device.
func (device *NVMLDevice) GetMemoryInfo() (*MemoryInfo, error) {
	proc, err := device.runner.getProc("nvmlDeviceGetMemoryInfo")
	if err != nil {
		return nil, errs.Wrap(err, "error getting procedure")
	}

	var memInfo MemoryInfo

	result, _, callErr := proc.Call(
		device.handle,
		uintptr(unsafe.Pointer(&memInfo)), //nolint:gosec
	)

	err = checkCallError(callErr)
	if err != nil {
		return nil, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return nil, err
	}

	return &memInfo, nil
}

// GetBAR1MemoryInfo retrieves BAR1 memory information for the NVIDIA device.
func (device *NVMLDevice) GetBAR1MemoryInfo() (*MemoryInfo, error) {
	proc, err := device.runner.getProc("nvmlDeviceGetBAR1MemoryInfo")
	if err != nil {
		return nil, errs.Wrap(err, "error getting procedure")
	}

	var memInfo MemoryInfo

	result, _, callErr := proc.Call(
		device.handle,
		uintptr(unsafe.Pointer(&memInfo)), //nolint:gosec
	)

	err = checkCallError(callErr)
	if err != nil {
		return nil, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return nil, errs.Wrap(err, "NVML returned error")
	}

	return &memInfo, nil
}

// GetFanSpeed retrieves the current fan speed of the NVIDIA device as a percentage of its maximum speed.
func (device *NVMLDevice) GetFanSpeed() (uint, error) {
	proc, err := device.runner.getProc("nvmlDeviceGetFanSpeed")
	if err != nil {
		return 0, errs.Wrap(err, "error getting procedure")
	}

	var speed uint32

	result, _, callErr := proc.Call(
		device.handle,
		uintptr(unsafe.Pointer(&speed)), //nolint:gosec
	)

	err = checkCallError(callErr)
	if err != nil {
		return 0, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "NVML returned error")
	}

	return uint(speed), nil
}

// GetPcieThroughput retrieves the PCIe throughput for the NVIDIA device, based on the specified metric type.
func (device *NVMLDevice) GetPcieThroughput(metricType PcieMetricType) (uint, error) {
	proc, err := device.runner.getProc("nvmlDeviceGetPcieThroughput")
	if err != nil {
		return 0, errs.Wrap(err, "error getting procedure")
	}

	var throughput uint32

	result, _, callErr := proc.Call(
		device.handle,
		uintptr(metricType),
		uintptr(unsafe.Pointer(&throughput)), //nolint:gosec
	)

	err = checkCallError(callErr)
	if err != nil {
		return 0, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "NVML returned error")
	}

	return uint(throughput), nil
}

// GetUUID retrieves the UUID of the NVIDIA device.
func (device *NVMLDevice) GetUUID() (string, error) {
	proc, err := device.runner.getProc("nvmlDeviceGetUUID")
	if err != nil {
		return "", errs.Wrap(err, "error getting procedure")
	}

	var uuid [deviceUUIDBufferSize]byte

	result, _, callErr := proc.Call(
		device.handle,
		uintptr(unsafe.Pointer(&uuid[0])), //nolint:gosec
		uintptr(deviceUUIDBufferSize),
	)

	err = checkCallError(callErr)
	if err != nil {
		return "", errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return "", errs.Wrap(err, "NVML returned error")
	}

	return windows.ByteSliceToString(uuid[:]), nil
}

// GetSerial retrieves the serial number of the NVIDIA device.
func (device *NVMLDevice) GetSerial() (string, error) {
	proc, err := device.runner.getProc("nvmlDeviceGetSerial")
	if err != nil {
		return "", errs.Wrap(err, "error getting procedure")
	}

	var serial [deviceSerialBufferSize]byte

	result, _, callErr := proc.Call(
		device.handle,
		uintptr(unsafe.Pointer(&serial[0])), //nolint:gosec
		uintptr(deviceSerialBufferSize),
	)

	err = checkCallError(callErr)
	if err != nil {
		return "", errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return "", errs.Wrap(err, "NVML returned error")
	}

	return windows.ByteSliceToString(serial[:]), nil
}

// GetEncoderUtilization retrieves the encoder utilization statistics for the device.
// It returns the following values:
//   - utilization: the percentage of time over the past sampling period during which the encoder was active.
//   - samplingPeriodUs: the sampling period duration in microseconds,
//     indicating how long the utilization metric was measured.
func (device *NVMLDevice) GetEncoderUtilization() (uint, uint, error) {
	proc, err := device.runner.getProc("nvmlDeviceGetEncoderUtilization")
	if err != nil {
		return 0, 0, errs.Wrap(err, "error getting procedure")
	}

	var (
		utilization      uint32
		samplingPeriodUs uint32
	)

	result, _, callErr := proc.Call(
		device.handle,
		uintptr(unsafe.Pointer(&utilization)),      //nolint:gosec
		uintptr(unsafe.Pointer(&samplingPeriodUs)), //nolint:gosec
	)

	err = checkCallError(callErr)
	if err != nil {
		return 0, 0, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, 0, errs.Wrap(err, "NVML returned error")
	}

	return uint(utilization), uint(samplingPeriodUs), nil
}

// GetDecoderUtilization retrieves the decoder utilization statistics for the device.
// It returns the following values:
//   - utilization: the percentage of time over the past sampling period during which the decoder was active.
//   - samplingPeriodUs: the sampling period duration in microseconds, indicating how long the utilization
//     metric was measured.
func (device *NVMLDevice) GetDecoderUtilization() (uint, uint, error) {
	proc, err := device.runner.getProc("nvmlDeviceGetDecoderUtilization")
	if err != nil {
		return 0, 0, errs.Wrap(err, "error getting procedure")
	}

	var (
		utilization      uint32
		samplingPeriodUs uint32
	)

	result, _, callErr := proc.Call(
		device.handle,
		uintptr(unsafe.Pointer(&utilization)),      //nolint:gosec
		uintptr(unsafe.Pointer(&samplingPeriodUs)), //nolint:gosec
	)

	err = checkCallError(callErr)
	if err != nil {
		return 0, 0, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, 0, errs.Wrap(err, "NVML returned error")
	}

	return uint(utilization), uint(samplingPeriodUs), nil
}

// GetMemoryErrorCounter retrieves the ECC memory error count for the specified error type,
// memory location, and counter type.
func (device *NVMLDevice) GetMemoryErrorCounter(
	errorType MemoryErrorType,
	memoryLocation MemoryLocation,
	counterType EccCounterType) (uint64, error) {
	proc, err := device.runner.getProc("nvmlDeviceGetMemoryErrorCounter")
	if err != nil {
		return 0, errs.Wrap(err, "error getting procedure")
	}

	var errorCount uint64

	result, _, callErr := proc.Call(
		device.handle,
		uintptr(errorType),
		uintptr(counterType),
		uintptr(memoryLocation),
		uintptr(unsafe.Pointer(&errorCount)), //nolint:gosec
	)

	err = checkCallError(callErr)
	if err != nil {
		return 0, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "NVML returned error")
	}

	return errorCount, nil
}

// GetTotalEnergyConsumption retrieves the total energy consumption of the NVIDIA device in millijoules.
func (device *NVMLDevice) GetTotalEnergyConsumption() (uint64, error) {
	proc, err := device.runner.getProc("nvmlDeviceGetTotalEnergyConsumption")
	if err != nil {
		return 0, errs.Wrap(err, "error getting procedure")
	}

	var energy uint64

	result, _, callErr := proc.Call(
		device.handle,
		uintptr(unsafe.Pointer(&energy)), //nolint:gosec
	)

	err = checkCallError(callErr)
	if err != nil {
		return 0, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "NVML returned error")
	}

	return energy, nil
}

// GetPerformanceState retrieves the performance state (P-state) of the NVIDIA device.
func (device *NVMLDevice) GetPerformanceState() (uint, error) {
	proc, err := device.runner.getProc("nvmlDeviceGetPerformanceState")
	if err != nil {
		return 0, errs.Wrap(err, "error getting procedure")
	}

	var perfState uint32

	result, _, callErr := proc.Call(
		device.handle,
		uintptr(unsafe.Pointer(&perfState)), //nolint:gosec
	)

	err = checkCallError(callErr)
	if err != nil {
		return 0, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "NVML returned error")
	}

	return uint(perfState), nil
}

// GetClockInfo retrieves the clock rate for the specified clock type of the NVIDIA device.
func (device *NVMLDevice) GetClockInfo(clockType ClockType) (uint, error) {
	proc, err := device.runner.getProc("nvmlDeviceGetClockInfo")
	if err != nil {
		return 0, errs.Wrap(err, "error getting procedure")
	}

	var clockRate uint32

	result, _, callErr := proc.Call(
		device.handle,
		uintptr(clockType),
		uintptr(unsafe.Pointer(&clockRate)), //nolint:gosec
	)

	err = checkCallError(callErr)
	if err != nil {
		return 0, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "NVML returned error")
	}

	return uint(clockRate), nil
}

// GetPowerUsage retrieves the power usage of the NVIDIA device in milliwatts.
func (device *NVMLDevice) GetPowerUsage() (uint, error) {
	proc, err := device.runner.getProc("nvmlDeviceGetPowerUsage")
	if err != nil {
		return 0, errs.Wrap(err, "error getting procedure")
	}

	var power uint32

	result, _, callErr := proc.Call(
		device.handle,
		uintptr(unsafe.Pointer(&power)), //nolint:gosec
	)

	err = checkCallError(callErr)
	if err != nil {
		return 0, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "NVML returned error")
	}

	return uint(power), nil
}

// GetEncoderStats retrieves statistics related to the encoder activity on the device.
// It returns the following statistics:
//   - sessionCount: the number of active encoder sessions.
//   - averageFps: the average frames per second across all active encoder sessions.
//   - averageLatency: the average latency (in milliseconds) across all active encoder sessions.
func (device *NVMLDevice) GetEncoderStats() (uint, uint, uint, error) {
	proc, err := device.runner.getProc("nvmlDeviceGetPowerUsage")
	if err != nil {
		return 0, 0, 0, errs.Wrap(err, "error getting procedure")
	}

	var (
		sessionCount   uint32
		averageFps     uint32
		averageLatency uint32
	)

	result, _, callErr := proc.Call(
		device.handle,
		uintptr(unsafe.Pointer(&sessionCount)),   //nolint:gosec
		uintptr(unsafe.Pointer(&averageFps)),     //nolint:gosec
		uintptr(unsafe.Pointer(&averageLatency)), //nolint:gosec
	)

	err = checkCallError(callErr)
	if err != nil {
		return 0, 0, 0, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, 0, 0, errs.Wrap(err, "NVML returned error")
	}

	return uint(sessionCount), uint(averageFps), uint(averageLatency), nil
}

// GetPowerManagementLimit retrieves the power management limit of the NVIDIA device in milliwatts.
func (device *NVMLDevice) GetPowerManagementLimit() (uint, error) {
	proc, err := device.runner.getProc("nvmlDeviceGetPowerManagementLimit")
	if err != nil {
		return 0, errs.Wrap(err, "error getting procedure")
	}

	var powerLimit uint32

	result, _, callErr := proc.Call(
		device.handle,
		uintptr(unsafe.Pointer(&powerLimit)), //nolint:gosec
	)

	err = checkCallError(callErr)
	if err != nil {
		return 0, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, errs.Wrap(err, "NVML returned error")
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
	proc, err := device.runner.getProc("nvmlDeviceGetEccMode")
	if err != nil {
		return false, false, errs.Wrap(err, "error getting procedure")
	}

	var (
		currentMode uint32
		pendingMode uint32
	)

	result, _, callErr := proc.Call(
		device.handle,
		uintptr(unsafe.Pointer(&currentMode)), //nolint:gosec
		uintptr(unsafe.Pointer(&pendingMode)), //nolint:gosec
	)

	err = checkCallError(callErr)
	if err != nil {
		return false, false, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return false, false, errs.Wrap(err, "NVML returned error")
	}

	currentEnabled := currentMode == 1
	pendingEnabled := pendingMode == 1

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
	proc, err := device.runner.getProc("nvmlDeviceGetUtilizationRates")
	if err != nil {
		return 0, 0, errs.Wrap(err, "error getting procedure")
	}

	type utilization struct {
		gpu    uint32
		memory uint32
	}

	var util utilization

	result, _, callErr := proc.Call(
		device.handle,
		uintptr(unsafe.Pointer(&util)), //nolint:gosec
	)

	err = checkCallError(callErr)
	if err != nil {
		return 0, 0, errs.Wrap(err, "error making syscall")
	}

	err = mapNVMLResultToError(int(result))
	if err != nil {
		return 0, 0, errs.Wrap(err, "NVML returned error")
	}

	return uint(util.gpu), uint(util.memory), nil
}
