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

package plugin

import (
	"context"
	"time"

	"golang.zabbix.com/plugin/nvidia/internal/plugin/handlers"
	"golang.zabbix.com/plugin/nvidia/internal/plugin/params"
	"golang.zabbix.com/plugin/nvidia/pkg/nvml"
	"golang.zabbix.com/sdk/errs"
	"golang.zabbix.com/sdk/metric"
	"golang.zabbix.com/sdk/plugin"
	"golang.zabbix.com/sdk/plugin/container"
	"golang.zabbix.com/sdk/zbxerr"
)

// Name of the plugin.
const Name = "NVIDIA"

var (
	_ plugin.Configurator = (*nvmlPlugin)(nil)
	_ plugin.Exporter     = (*nvmlPlugin)(nil)
	_ plugin.Runner       = (*nvmlPlugin)(nil)
)

type nvmlMetric struct {
	metric  *metric.Metric
	handler handlers.HandlerFunc
}

type nvmlPlugin struct {
	plugin.Base
	config     *pluginConfig
	metrics    map[string]*nvmlMetric
	nvmlRunner nvml.Runner
}

// Launch launches the NVIDIA plugin. Blocks until plugin execution has
// finished.
func Launch() error {
	runner, err := nvml.NewNVMLRunner()
	if err != nil {
		runner = nil
	}

	p := &nvmlPlugin{
		nvmlRunner: runner,
	}

	err = p.registerMetrics()
	if err != nil {
		return errs.Wrap(err, "failed to register metrics")
	}

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
func (p *nvmlPlugin) Start() {
	p.Logger.Infof("Start called")

	// Runner is nil in case library was not loaded
	if p.nvmlRunner == nil {
		panic("nvml runner is nil")
	}

	// Try to initialize NVML using InitV2, fallback to Init if it fails
	err := p.nvmlRunner.InitV2()
	if err != nil {
		p.Logger.Debugf("failed to init runner with InitNVMLv2: %s", err.Error())

		// Fallback to Init if InitV2 fails
		err = p.nvmlRunner.Init()
		if err != nil {
			wrappedErr := errs.Wrap(err, "failed to init NVML library")
			p.Logger.Errf("%s", wrappedErr.Error())
			panic(wrappedErr)
		}
	}
}

// Stop stops the NVIDIA plugin. Is required for plugin to match runner interface.
func (p *nvmlPlugin) Stop() {
	p.Logger.Infof("Stop called")

	if p.nvmlRunner == nil {
		return
	}

	err := p.nvmlRunner.ShutdownNVML()
	if err != nil {
		p.Logger.Errf("failed to shutdown nvml: %s", err.Error())
	}

	err = p.nvmlRunner.Close()
	if err != nil {
		p.Logger.Errf("failed to close nvml runner: %s", err.Error())
	}
}

// Export collects all the metrics.
func (p *nvmlPlugin) Export(key string, rawParams []string, pluginCtx plugin.ContextProvider) (any, error) {
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

	timeout := time.Second * time.Duration(p.config.Timeout)
	if timeout < time.Second*time.Duration(pluginCtx.Timeout()) {
		timeout = time.Second * time.Duration(pluginCtx.Timeout())
	}

	ctx, cancel := context.WithTimeout(
		context.Background(), timeout,
	)
	defer cancel()

	res, err := m.handler(ctx, metricParams, extraParams...)
	if err != nil {
		return nil, errs.Wrap(err, "failed to execute handler")
	}

	return res, nil
}

func (p *nvmlPlugin) registerMetrics() error {
	handler := handlers.New(p.nvmlRunner)

	p.metrics = map[string]*nvmlMetric{
		"nvml.version": {
			metric: metric.New(
				"Returns local NVML version.",
				nil,
				false,
			),
			handler: handler.GetNVMLVersion,
		},
		"nvml.system.driver.version": {
			metric: metric.New(
				"Returns local NVIDIA driver version.",
				nil,
				false,
			),
			handler: handler.GetDriverVersion,
		},
		"nvml.device.get": {
			metric: metric.New(
				"Returns discovered devices.",
				nil,
				false,
			),
			handler: handlers.WithJSONResponse(
				handler.DeviceDiscovery,
			),
		},
		"nvml.device.count": {
			metric: metric.New(
				"Returns device count.",
				nil,
				false,
			),
			handler: handler.GetDeviceCount,
		},
		"nvml.device.temperature": {
			metric: metric.New(
				"Returns device temperature.",
				params.Params,
				false,
			),
			handler: handler.GetDeviceTemperature,
		},
		"nvml.device.serial": {
			metric: metric.New(
				"Returns device serial.",
				params.Params,
				false,
			),
			handler: handler.GetDeviceSerial,
		},
		"nvml.device.fan.speed.avg": {
			metric: metric.New(
				"Returns device fan speed.",
				params.Params,
				false,
			),
			handler: handler.GetDeviceFanSpeed,
		},
		"nvml.device.performance.state": {
			metric: metric.New(
				"Returns device performance state.",
				params.Params,
				false,
			),
			handler: handler.GetDevicePerfState,
		},
		"nvml.device.energy.consumption": {
			metric: metric.New(
				"Returns device energy consumption.",
				params.Params,
				false,
			),
			handler: handler.GetDeviceEnergyConsumption,
		},
		"nvml.device.power.limit": {
			metric: metric.New(
				"Returns device power management limit.",
				params.Params,
				false,
			),
			handler: handler.GetDevicePowerLimit,
		},
		"nvml.device.power.usage": {
			metric: metric.New(
				"Returns device power usage.",
				params.Params,
				false,
			),
			handler: handler.GetDevicePowerUsage,
		},
		"nvml.device.memory.bar1.get": {
			metric: metric.New(
				"Returns BAR1 memory info.",
				params.Params,
				false,
			),
			handler: handlers.WithJSONResponse(
				handler.GetBAR1MemoryInfo,
			),
		},
		"nvml.device.memory.fb.get": {
			metric: metric.New(
				"Returns FB memory info.",
				params.Params,
				false,
			),
			handler: handlers.WithJSONResponse(
				handler.GetFBMemoryInfo,
			),
		},
		"nvml.device.errors.memory": {
			metric: metric.New(
				"Returns ECC error count in memory.",
				params.Params,
				false,
			),
			handler: handlers.WithJSONResponse(
				handler.GetMemoryErrors,
			),
		},
		"nvml.device.errors.register": {
			metric: metric.New(
				"Returns ECC error count in register file.",
				params.Params,
				false,
			),
			handler: handlers.WithJSONResponse(
				handler.GetRegisterErrors,
			),
		},
		"nvml.device.pci.utilization": {
			metric: metric.New(
				"Returns PCIe utilization.",
				params.Params,
				false,
			),
			handler: handlers.WithJSONResponse(
				handler.GetPCIeThroughput,
			),
		},
		"nvml.device.encoder.stats.get": {
			metric: metric.New(
				"Returns Encoder utilization.",
				params.Params,
				false,
			),
			handler: handlers.WithJSONResponse(
				handler.GetEncoderStats,
			),
		},
		"nvml.device.video.frequency": {
			metric: metric.New(
				"Returns Video frequency in MHz.",
				params.Params,
				false,
			),
			handler: handler.GetVideoFrequency,
		},
		"nvml.device.graphics.frequency": {
			metric: metric.New(
				"Returns Graphics frequency in MHz.",
				params.Params,
				false,
			),
			handler: handler.GetGraphicsFrequency,
		},
		"nvml.device.sm.frequency": {
			metric: metric.New(
				"Returns SM frequency in MHz.",
				params.Params,
				false,
			),
			handler: handler.GetSMFrequency,
		},
		"nvml.device.memory.frequency": {
			metric: metric.New(
				"Returns Memory frequency in MHz.",
				params.Params,
				false,
			),
			handler: handler.GetMemoryFrequency,
		},
		"nvml.device.encoder.utilization": {
			metric: metric.New(
				"Returns Encoder utilization.",
				params.Params,
				false,
			),
			handler: handler.GetEncoderUtilization,
		},
		"nvml.device.decoder.utilization": {
			metric: metric.New(
				"Returns Decoder utilization.",
				params.Params,
				false,
			),
			handler: handler.GetDecoderUtilization,
		},
		"nvml.device.utilization": {
			metric: metric.New(
				"Returns Device utilization.",
				params.Params,
				false,
			),
			handler: handlers.WithJSONResponse(
				handler.GetDeviceUtilisation,
			),
		},
		"nvml.device.ecc.mode": {
			metric: metric.New(
				"Returns Device current and pending ECC mode.",
				params.Params,
				false,
			),
			handler: handlers.WithJSONResponse(
				handler.GetECCMode,
			),
		},
	}

	metricSet := metric.MetricSet{}

	for k, m := range p.metrics {
		metricSet[k] = m.metric
	}

	err := plugin.RegisterMetrics(p, Name, metricSet.List()...)
	if err != nil {
		return errs.Wrap(err, "failed to register metrics")
	}

	return nil
}
