/*
** Copyright (C) 2001-2024 Zabbix SIA
**
** Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
** documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
** rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to
** permit persons to whom the Software is furnished to do so, subject to the following conditions:
**
** The above copyright notice and this permission notice shall be included in all copies or substantial portions
** of the Software.
**
** THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
** WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
** COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
** TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
** SOFTWARE.
**/

package plugin

import (
	"context"
	"time"

	"golang.zabbix.com/plugin/nvidia/pkg/nvml"
	"golang.zabbix.com/plugin/nvidia/plugin/handlers"
	"golang.zabbix.com/plugin/nvidia/plugin/params"
	"golang.zabbix.com/sdk/errs"
	"golang.zabbix.com/sdk/log"
	"golang.zabbix.com/sdk/metric"
	"golang.zabbix.com/sdk/plugin"
	"golang.zabbix.com/sdk/plugin/container"
	"golang.zabbix.com/sdk/zbxerr"
)

const (
	// Name of the plugin.
	Name = "NVIDIA"
)

var (
	_ plugin.Configurator = (*nvmlPlugin)(nil)
	_ plugin.Exporter     = (*nvmlPlugin)(nil)
	_ plugin.Runner       = (*nvmlPlugin)(nil)
)

type exampleMetric struct {
	metric  *metric.Metric
	handler handlers.HandlerFunc
}

type nvmlPlugin struct {
	plugin.Base
	config     *pluginConfig
	metrics    map[string]*exampleMetric
	nvmlRunner nvml.Runner
}

// Launch launches the Example plugin. Blocks until plugin execution has
// finished.
func Launch() error {
	runner, err := nvml.NewNVMLRunner()
	if err != nil {
		return err
	}

	p := &nvmlPlugin{
		nvmlRunner: runner,
	}

	err = p.registerMetrics()
	if err != nil {
		return err
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

	err = p.nvmlRunner.Close()
	if err != nil {
		return err
	}

	return nil
}

// Start starts the example plugin. Is required for plugin to match runner interface.
func (p *nvmlPlugin) Start() {
	p.Logger.Infof("Start called")

	err := initNVML(p.nvmlRunner, p.Logger)
	if err != nil {
		p.Logger.Errf("error initializing NVML library: %v", err)
		panic(err)
	}
}

// Stop stops the example plugin. Is required for plugin to match runner interface.
func (p *nvmlPlugin) Stop() {
	p.Logger.Infof("Stop called")

	err := p.nvmlRunner.ShutdownNVML()
	if err != nil {
		p.Logger.Errf("failed to shutdown nvml %v", err)
	}

}

func initNVML(runner nvml.Runner, loger log.Logger) error {
	err := runner.InitNVMLv2()
	if err == nil {
		return nil
	}

	loger.Debugf("failed to init runner with InitNVMLv2 %v", err)

	err = runner.InitNVML()
	if err != nil {
		return errs.Wrap(err, "failed to init NVML library")
	}

	return nil
}

// Export collects all the metrics.
func (p *nvmlPlugin) Export(key string, rawParams []string, _ plugin.ContextProvider) (any, error) {
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

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(p.config.Timeout)*time.Second,
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

	p.metrics = map[string]*exampleMetric{
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
				"Returns local Nvidia driver version.",
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
		"nvml.device.errors.registry": {
			metric: metric.New(
				"Returns ECC error count in registry file.",
				params.Params,
				false,
			),
			handler: handlers.WithJSONResponse(
				handler.GetRegistryErrors,
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
				"Returns Encoder utilisation.",
				params.Params,
				false,
			),
			handler: handler.GetEncoderUtilisation,
		},
		"nvml.device.decoder.utilization": {
			metric: metric.New(
				"Returns Decoder utilisation.",
				params.Params,
				false,
			),
			handler: handler.GetDecoderUtilisation,
		},
		"nvml.device.utilization": {
			metric: metric.New(
				"Returns Decoder utilisation.",
				params.Params,
				false,
			),
			handler: handlers.WithJSONResponse(
				handler.GetDeviceUtilisation,
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
