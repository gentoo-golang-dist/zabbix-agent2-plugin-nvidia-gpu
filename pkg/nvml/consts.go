package nvml

const (
	// DeviceUUIDBufferSize indicates buffer size
	// guaranteed to be large enough for nvmlDeviceGetUUID.
	deviceUUIDBufferSize = 80

	// SystemDriverVersionBufferSize indicates buffer size
	// guaranteed to be large enough for nvmlSystemGetDriverVersion.
	systemDriverVersionBufferSize = 80

	// SystemNVMLVersionBufferSize indicates buffer size
	// guaranteed to be large enough for nvmlSystemGetNVMLVersion.
	systemNVMLVersionBufferSize = 80

	// DeviceNameBufferSize indicates
	// buffer size guaranteed to be large enough for nvmlDeviceGetName.
	deviceNameBufferSize = 64

	// DeviceSerialBufferSize indicates buffer size
	// guaranteed to be large enough for nvmlDeviceGetSerial.
	deviceSerialBufferSize = 30
)

// Constants for PCIe Metric Types (TX and RX).
const (
	TX PcieMetricType = 0 // PCIe transmit throughput
	RX PcieMetricType = 1 // PCIe receive throughput
)

// Constants for ClockType representing various NVML clock types.
const (
	Graphics ClockType = 0 // Graphics clock
	SM       ClockType = 1 // SM (Streaming Multiprocessor) clock
	Memory   ClockType = 2 // Memory clock
	Video    ClockType = 3 // Video encoder/decoder clock
)

// Constants for MemoryErrorType.
const (
	MemoryErrorTypeCorrected   MemoryErrorType = 0 // Corrected memory errors
	MemoryErrorTypeUncorrected MemoryErrorType = 1 // Uncorrected memory errors
)

// Constants for MemoryLocation.
const (
	MemoryLocationDevice       MemoryLocation = 0 // Device memory
	MemoryLocationRegisterFile MemoryLocation = 1 // Register file memory
	MemoryLocationL1Cache      MemoryLocation = 2 // L1 cache memory
	MemoryLocationL2Cache      MemoryLocation = 3 // L2 cache memory
	// Add more memory locations as needed.
)

// Constants for EccCounterType.
const (
	EccCounterTypeVolatile  EccCounterType = 0 // Volatile ECC counter
	EccCounterTypeAggregate EccCounterType = 1 // Aggregate ECC counter
)

// PcieMetricType is a custom type that defines the PCIe throughput metric (TX or RX).
type PcieMetricType uint32

// ClockType is a custom type representing the different clock types available for a device.
type ClockType uint32

// MemoryErrorType represents the type of memory errors (corrected or uncorrected).
type MemoryErrorType uint32

// MemoryLocation represents the location of the memory error (device memory, register file, etc.).
type MemoryLocation uint32

// EccCounterType represents the type of ECC counter (volatile or aggregate).
type EccCounterType uint32
