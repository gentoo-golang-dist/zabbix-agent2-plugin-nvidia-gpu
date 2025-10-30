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

package main

import (
	"errors"
	"fmt"
	"os"

	"golang.zabbix.com/plugin/nvidia/internal/plugin"
	"golang.zabbix.com/sdk/errs"
	sdkplugin "golang.zabbix.com/sdk/plugin"
	"golang.zabbix.com/sdk/plugin/flag"
)

const copyrightMessage = //
`
   Copyright 2001-%d Zabbix SIA

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
`

//nolint:gochecknoglobals,revive // required ALL_CAPS by build scripts
var (
	PLUGIN_VERSION_MAJOR = 8
	PLUGIN_VERSION_MINOR = 0
	PLUGIN_VERSION_PATCH = 0
	PLUGIN_VERSION_RC    = "alpha2"
	PLUGIN_LICENSE_YEAR  = 2025
)

func main() {
	args, err := flag.HandleFlags()
	if err != nil {
		exitWithError(errs.Wrap(err, "failed to handle flags: "))
	}

	pluginInfo := &sdkplugin.Info{
		Name:             plugin.Name,
		BinName:          os.Args[0],
		CopyrightMessage: fmt.Sprintf(copyrightMessage, PLUGIN_LICENSE_YEAR),
		MajorVersion:     PLUGIN_VERSION_MAJOR,
		MinorVersion:     PLUGIN_VERSION_MINOR,
		PatchVersion:     PLUGIN_VERSION_PATCH,
		Alphatag:         PLUGIN_VERSION_RC,
	}

	p, err := plugin.New()
	if err != nil {
		exitWithError(errs.Wrap(err, "failed to initialize plugin: "))
	}

	err = flag.DecideActionFromFlags(args, p, pluginInfo, nil)
	if err != nil {
		if errors.Is(err, errs.ErrExitGracefully) {
			// exit gracefully if parameter supposed to exit after execution
			exitGracefully()
		}

		exitWithError(errs.Wrap(err, "failed to execute plugin functions: "))
	}

	err = p.Run()
	if err != nil {
		exitWithError(errs.Wrap(err, "failed to run plugin: "))
	}
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	os.Exit(1)
}

func exitGracefully() {
	os.Exit(0)
}
