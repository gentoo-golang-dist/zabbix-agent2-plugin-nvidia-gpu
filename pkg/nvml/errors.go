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

import (
	"golang.zabbix.com/sdk/errs"
)

// Go idiomatic error constants.
var (
	ErrSuccess       error // Represents successful operation (no error)
	ErrUninitialized = errs.New(
		"NVML error: NVML was not first initialized with nvmlInit()",
	)
	ErrInvalidArgument = errs.New(
		"NVML error: A supplied argument is invalid",
	)
	ErrNotSupported = errs.New(
		"NVML error: The requested operation is not available on target device",
	)
	ErrNoPermission = errs.New(
		"NVML error: The current user does not have permission for operation",
	)
	ErrAlreadyInitialized = errs.New(
		"NVML error: Multiple initializations are now allowed through ref counting",
	)
	ErrNotFound = errs.New(
		"NVML error: A query to find an object was unsuccessful",
	)
	ErrInsufficientSize = errs.New(
		"NVML error: An input argument is not large enough",
	)
	ErrInsufficientPower = errs.New(
		"NVML error: A device's external power cables are not properly attached",
	)
	ErrDriverNotLoaded = errs.New(
		"NVML error: NVIDIA driver is not loaded",
	)
	ErrTimeout = errs.New(
		"NVML error: User provided timeout passed",
	)
	ErrIrqIssue = errs.New(
		"NVML error: NVIDIA Kernel detected an interrupt issue with a GPU",
	)
	ErrLibraryNotFound = errs.New(
		"NVML error: NVML Shared Library couldn't be found or loaded",
	)
	ErrFunctionNotFound = errs.New(
		"NVML error: Local version of NVML doesn't implement this function",
	)
	ErrCorruptedInforom = errs.New(
		"NVML error: infoROM is corrupted",
	)
	ErrGpuIsLost = errs.New(
		"NVML error: The GPU has fallen off the bus or has otherwise become inaccessible",
	)
	ErrResetRequired = errs.New(
		"NVML error: The GPU requires a reset before it can be used again",
	)
	ErrOperatingSystem = errs.New(
		"NVML error: The GPU control device has been blocked by the operating system/cgroups",
	)
	ErrLibRmVersionMismatch = errs.New(
		"NVML error: RM detects a driver/library version mismatch",
	)
	ErrInUse = errs.New(
		"NVML error: An operation cannot be performed because the GPU is currently in use",
	)
	ErrMemory = errs.New(
		"NVML error: Insufficient memory",
	)
	ErrNoData = errs.New(
		"NVML error: No data",
	)
	ErrVgpuEccNotSupported = errs.New(
		"NVML error: The requested vgpu operation is not available on target device because ECC is enabled",
	)
	ErrInsufficientResources = errs.New(
		"NVML error: Ran out of critical resources, other than memory",
	)
	ErrFreqNotSupported = errs.New(
		"NVML error: Frequency not supported",
	)
	ErrArgumentVersionMismatch = errs.New(
		"NVML error: The provided version is invalid/unsupported",
	)
	ErrDeprecated = errs.New(
		"NVML error: The requested functionality has been deprecated",
	)
	ErrNotReady = errs.New(
		"NVML error: The system is not ready for the request",
	)
	ErrGpuNotFound = errs.New(
		"NVML error: No GPUs were found",
	)
	ErrInvalidState = errs.New(
		"NVML error: Resource not in correct state to perform requested operation",
	)
	ErrUnknown = errs.New(
		"NVML error: An internal driver error occurred",
	)
)

// mapNVMLResultToError maps NVML error codes to Go error types.
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
	case 16:
		return ErrResetRequired
	case 17:
		return ErrOperatingSystem
	case 18:
		return ErrLibRmVersionMismatch
	case 19:
		return ErrInUse
	case 20:
		return ErrMemory
	case 21:
		return ErrNoData
	case 22:
		return ErrVgpuEccNotSupported
	case 23:
		return ErrInsufficientResources
	case 24:
		return ErrFreqNotSupported
	case 25:
		return ErrArgumentVersionMismatch
	case 26:
		return ErrDeprecated
	case 27:
		return ErrNotReady
	case 28:
		return ErrGpuNotFound
	case 29:
		return ErrInvalidState
	case 999:
		return ErrUnknown
	default:
		return errs.Errorf("NVML error: Unknown error code %d", code)
	}
}
