/*
TODO: Add Apache-2.0 licence
*/

package main

import (
	"errors"
	"fmt"
	"os"

	"golang.zabbix.com/plugin/nvidia/internal/plugin"
	"golang.zabbix.com/sdk/plugin/flag"
	"golang.zabbix.com/sdk/zbxerr"
)

/*
TODO: Add Apache-2.0 licence
*/
const copyrightMessage = //
`
Copyright 2001-%d Zabbix SIA
`

//nolint:gochecknoglobals,revive // required ALL_CAPS by build scripts
var (
	PLUGIN_VERSION_MAJOR = 7
	PLUGIN_VERSION_MINOR = 2
	PLUGIN_VERSION_PATCH = 0
	PLUGIN_VERSION_RC    = "alpha1"
	PLUGIN_LICENSE_YEAR  = 2024
)

func main() {
	err := flag.HandleFlags(
		plugin.Name,
		os.Args[0],
		fmt.Sprintf(copyrightMessage, PLUGIN_LICENSE_YEAR),
		PLUGIN_VERSION_RC,
		PLUGIN_VERSION_MAJOR,
		PLUGIN_VERSION_MINOR,
		PLUGIN_VERSION_PATCH,
	)
	if err != nil {
		if errors.Is(err, zbxerr.ErrorOSExitZero) {
			return
		}

		panic(err)
	}

	err = plugin.Launch()
	if err != nil {
		panic(err)
	}
}
