package nvml

import (
	"golang.zabbix.com/sdk/errs"
)

// Go idiomatic error constants.
var (
	ErrSuccess            error
	ErrUninitialized      = errs.New("NVML error: NVML was not first initialized with nvmlInit()")
	ErrInvalidArgument    = errs.New("NVML error: A supplied argument is invalid")
	ErrNotSupported       = errs.New("NVML error: The requested operation is not available on target device")
	ErrNoPermission       = errs.New("NVML error: The current user does not have permission for operation")
	ErrAlreadyInitialized = errs.New("NVML error: Multiple initializations are now allowed through ref counting")
	ErrNotFound           = errs.New("NVML error: A query to find an object was unsuccessful")
	ErrInsufficientSize   = errs.New("NVML error: An input argument is not large enough")
	ErrInsufficientPower  = errs.New("NVML error: A device's external power cables are not properly attached")
	ErrDriverNotLoaded    = errs.New("NVML error: NVIDIA driver is not loaded")
	ErrTimeout            = errs.New("NVML error: User provided timeout passed")
	ErrIrqIssue           = errs.New("NVML error: NVIDIA Kernel detected an interrupt issue with a GPU")
	ErrLibraryNotFound    = errs.New("NVML error: NVML Shared Library couldn't be found or loaded")
	ErrFunctionNotFound   = errs.New("NVML error: Local version of NVML doesn't implement this function")
	ErrCorruptedInforom   = errs.New("NVML error: infoROM is corrupted")
	ErrGpuIsLost          = errs.New("NVML error: The GPU has fallen off the bus or has otherwise become inaccessible")
	ErrUnknown            = errs.New("NVML error: An internal driver error occurred")
)

// MapNVMLResultToError maps NVML error codes to Go error types.
//
//nolint:gocyclo,cyclop
func mapNVMLResultToError(code int) error {
	switch code {
	case 0:
		return ErrSuccess
	case 1:
		return ErrUninitialized
	case 2:
		return ErrInvalidArgument
	case 3:
		return ErrNotSupported
	case 4:
		return ErrNoPermission
	case 5:
		return ErrAlreadyInitialized
	case 6:
		return ErrNotFound
	case 7:
		return ErrInsufficientSize
	case 8:
		return ErrInsufficientPower
	case 9:
		return ErrDriverNotLoaded
	case 10:
		return ErrTimeout
	case 11:
		return ErrIrqIssue
	case 12:
		return ErrLibraryNotFound
	case 13:
		return ErrFunctionNotFound
	case 14:
		return ErrCorruptedInforom
	case 15:
		return ErrGpuIsLost
	case 999:
		return ErrUnknown
	default:
		return errs.Errorf("NVML error: Unknown error code %d", code)
	}
}
