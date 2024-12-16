/*
** Zabbix
** Copyright (C) 2001-2024 Zabbix SIA
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

package plugin

import (
	"golang.zabbix.com/sdk/conf"
	"golang.zabbix.com/sdk/errs"
	"golang.zabbix.com/sdk/plugin"
)

type pluginConfig struct {
	plugin.SystemOptions `conf:"optional,name=System"`
	// Timeout is plugin connection timeout.
	Timeout int `conf:"optional,range=1:30"`
}

// Configure implements the Configurator interface.
// Initializes configuration structures.
func (p *nvmlPlugin) Configure(global *plugin.GlobalOptions, options any) {
	pConfig := &pluginConfig{}

	err := conf.Unmarshal(options, pConfig)
	if err != nil {
		p.Errf("cannot unmarshal configuration options: %s", err.Error())

		return
	}

	p.config = pConfig

	if p.config.Timeout == 0 {
		p.config.Timeout = global.Timeout
	}
}

// Validate implements the Configurator interface.
// Returns an error if validation of a plugin's configuration is failed.
func (*nvmlPlugin) Validate(options any) error {
	err := conf.Unmarshal(options, &pluginConfig{})
	if err != nil {
		return errs.Wrap(err, "failed to unmarshal configuration options")
	}

	return nil
}
