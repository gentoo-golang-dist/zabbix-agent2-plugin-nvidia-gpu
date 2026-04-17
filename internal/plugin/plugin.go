/*
** Zabbix
** Copyright (C) 2001-2026 Zabbix SIA
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
	"time"

	"golang.zabbix.com/plugin/nvidia/internal/plugin/handlers"
	"golang.zabbix.com/plugin/nvidia/pkg/nvml"
	"golang.zabbix.com/sdk/errs"
	"golang.zabbix.com/sdk/log"
	"golang.zabbix.com/sdk/metric"
	"golang.zabbix.com/sdk/plugin"
	"golang.zabbix.com/sdk/plugin/container"
	"golang.zabbix.com/sdk/zbxerr"
)

// Name of the plugin.
const Name = "NVIDIA"

var (
	_ plugin.Configurator = (*NvmlPlugin)(nil)
	_ plugin.Exporter     = (*NvmlPlugin)(nil)
	_ plugin.Runner       = (*NvmlPlugin)(nil)
)

// NvmlPlugin holds plugin parameters.
type NvmlPlugin struct {
	plugin.Base
	config        *pluginConfig
	metrics       map[string]*nvmlMetric
	nvmlRunner    nvml.Runner
	setNvmlRunner func() error
}

type nvmlMetric struct {
	metric  *metric.Metric
	handler handlers.HandlerFunc
}

// New creates a new plugin implementation.
func New() (*NvmlPlugin, error) {
	p := &NvmlPlugin{}
	p.setNvmlRunner = p.setRunner

	err := log.Open(log.Console, log.Info, "", 0)
	if err != nil {
		return nil, errs.Wrap(err, "failed to open log")
	}

	p.Logger = log.New(Name)

	err = p.registerMetrics()
	if err != nil {
		return nil, errs.Wrap(err, "failed to register metrics")
	}

	return p, nil
}

// Run starts the plugin.
func (p *NvmlPlugin) Run() error {
	h, err := container.NewHandler(Name)
	if err != nil {
		return errs.Wrap(err, "failed to create new handler")
	}

	p.Logger = h

	err = h.Execute()
	if err != nil {
		return errs.Wrap(err, "failed to execute plugin handler")
	}

	return nil
}

// Start starts the NVIDIA plugin. Is required for plugin to match runner interface.
func (p *NvmlPlugin) Start() {
	p.Infof("Start called")

	// this is needed for testing purposes, no way to mock it unless it's a callback, and can not pass it as a parameter
	// since Start is needed for plugin runner interface.
	err := p.setNvmlRunner()
	if err != nil {
		wrappedErr := errs.Wrap(err, "failed to init NVML runner")
		p.Errf("%s", wrappedErr.Error())
		panic(wrappedErr)
	}

	p.setMetricFunctions()
	// Try to initialize NVML using InitV2, fallback to Init if it fails
	err = p.nvmlRunner.InitV2()
	if err != nil {
		p.Debugf("failed to init runner with InitNVMLv2: %s", err.Error())

		// Fallback to Init if InitV2 fails
		err = p.nvmlRunner.Init()
		if err != nil {
			wrappedErr := errs.Wrap(err, "failed to init NVML library")
			p.Errf("%s", wrappedErr.Error())
			panic(wrappedErr)
		}
	}
}

// Stop stops the NVIDIA plugin. Is required for plugin to match runner interface.
func (p *NvmlPlugin) Stop() {
	p.Infof("Stop called")

	err := p.nvmlRunner.ShutdownNVML()
	if err != nil {
		p.Errf("failed to shutdown nvml %s", err.Error())
	}

	err = p.nvmlRunner.Close()
	if err != nil {
		p.Errf("failed to shutdown nvml runner %s", err.Error())
	}
}

// Export collects all the metrics.
func (p *NvmlPlugin) Export(key string, rawParams []string, ctx plugin.ContextProvider) (any, error) {
	m, ok := p.metrics[key]
	if !ok {
		return nil, errs.Wrapf(zbxerr.ErrorUnsupportedMetric, "unknown metric %q", key)
	}

	metricParams, extraParams, hardcodedParams, err := m.metric.EvalParams(rawParams, nil)
	if err != nil {
		return nil, errs.Wrap(err, "failed to evaluate metric parameters")
	}

	err = metric.SetDefaults(metricParams, hardcodedParams, nil)
	if err != nil {
		return nil, errs.Wrap(err, "failed to set default params")
	}

	if ctx.LegacyTimeout() {
		p.Debugf("using legacy timeout")

		ctx = plugin.OverrideTimeout(ctx, time.Now(), p.config.LegacyTimeout)
	}

	p.Tracef("request timeout set to: %d", ctx.Timeout())

	res, err := m.handler(ctx, metricParams, extraParams...)
	if err != nil {
		return nil, errs.Wrap(err, "failed to execute handler")
	}

	return res, nil
}

func (p *NvmlPlugin) setRunner() error {
	runner, err := nvml.NewNVMLRunner()
	if err != nil {
		return errs.Wrap(err, "failed to create new nvml runner")
	}

	p.nvmlRunner = runner

	return nil
}
