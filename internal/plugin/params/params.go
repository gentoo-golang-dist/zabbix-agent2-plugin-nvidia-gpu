/*
TODO: Add Apache-2.0 licence
*/

package params

import "golang.zabbix.com/sdk/metric"

const (
	// DeviceUUIDParamName device UUID parameter name.
	DeviceUUIDParamName = "DeviceUUID"
)

//nolint:gochecknoglobals // global constants.
var (
	// Params groups all base parameters common for all connections.
	Params = []*metric.Param{DeviceUUID}

	// DeviceUUID is device UUID parameter.
	DeviceUUID = metric.NewParam(DeviceUUIDParamName, "Device UUID to get information from.").SetRequired()
)
