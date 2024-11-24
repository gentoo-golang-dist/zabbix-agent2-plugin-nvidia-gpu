/*
TODO: Add Apache-2.0 licence
*/

package params

import "golang.zabbix.com/sdk/metric"

const (
	// DeviceUUIDParamName device uuid parameter name.
	DeviceUUIDParamName = "DeviceUUID"
)

//nolint:gochecknoglobals // global constants.
var (
	// Params groups all base parameters common for all connections.
	Params = []*metric.Param{DeviceUUID}

	// DeviceUUID is device uuid parameter.
	DeviceUUID = metric.NewConnParam(DeviceUUIDParamName, "Example parameter mimics a password.").WithDefault("")
)
