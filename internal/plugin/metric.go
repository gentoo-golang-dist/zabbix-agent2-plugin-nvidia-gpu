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
	"golang.zabbix.com/plugin/nvidia/internal/plugin/handlers"
	"golang.zabbix.com/plugin/nvidia/internal/plugin/params"
	"golang.zabbix.com/sdk/errs"
	"golang.zabbix.com/sdk/metric"
	"golang.zabbix.com/sdk/plugin"
)

const (
	version           = "nvml.version"
	driverVersion     = "nvml.system.driver.version"
	devGet            = "nvml.device.get"
	devCount          = "nvml.device.count"
	devTemp           = "nvml.device.temperature"
	devSerial         = "nvml.device.serial"
	devfanSpeedAVG    = "nvml.device.fan.speed.avg"
	devPerfState      = "nvml.device.performance.state"
	devEneConsumption = "nvml.device.energy.consumption"
	devPowerLimit     = "nvml.device.power.limit"
	devPowerUsage     = "nvml.device.power.usage"
	devMemBar1        = "nvml.device.memory.bar1.get"
	devMemFBGet       = "nvml.device.memory.fb.get"
	devMemErrors      = "nvml.device.errors.memory"
	devRegErrors      = "nvml.device.errors.register"
	devPciUtil        = "nvml.device.pci.utilization"
	devEncStats       = "nvml.device.encoder.stats.get"
	devVidFreq        = "nvml.device.video.frequency"
	devGraphFreq      = "nvml.device.graphics.frequency"
	devSMFreq         = "nvml.device.sm.frequency"
	devMemFreq        = "nvml.device.memory.frequency"
	devEncoderUtil    = "nvml.device.encoder.utilization"
	devDecoderUtil    = "nvml.device.decoder.utilization"
	devUtil           = "nvml.device.utilization"
	devEccMode        = "nvml.device.ecc.mode"
)

func (p *nvmlPlugin) registerMetrics() error {
	p.metrics = map[string]*nvmlMetric{
		version: {
			metric: metric.New(
				"Returns local NVML version.",
				nil,
				false,
			),
		},
		driverVersion: {
			metric: metric.New(
				"Returns local NVIDIA driver version.",
				nil,
				false,
			),
		},
		devGet: {
			metric: metric.New(
				"Returns discovered devices.",
				nil,
				false,
			),
		},
		devCount: {
			metric: metric.New(
				"Returns device count.",
				nil,
				false,
			),
		},
		devTemp: {
			metric: metric.New(
				"Returns device temperature.",
				params.Params,
				false,
			),
		},
		devSerial: {
			metric: metric.New(
				"Returns device serial.",
				params.Params,
				false,
			),
		},
		devfanSpeedAVG: {
			metric: metric.New(
				"Returns device fan speed.",
				params.Params,
				false,
			),
		},
		devPerfState: {
			metric: metric.New(
				"Returns device performance state.",
				params.Params,
				false,
			),
		},
		devEneConsumption: {
			metric: metric.New(
				"Returns device energy consumption.",
				params.Params,
				false,
			),
		},
		devPowerLimit: {
			metric: metric.New(
				"Returns device power management limit.",
				params.Params,
				false,
			),
		},
		devPowerUsage: {
			metric: metric.New(
				"Returns device power usage.",
				params.Params,
				false,
			),
		},
		devMemBar1: {
			metric: metric.New(
				"Returns BAR1 memory info.",
				params.Params,
				false,
			),
		},
		devMemFBGet: {
			metric: metric.New(
				"Returns FB memory info.",
				params.Params,
				false,
			),
		},
		devMemErrors: {
			metric: metric.New(
				"Returns ECC error count in memory.",
				params.Params,
				false,
			),
		},
		devRegErrors: {
			metric: metric.New(
				"Returns ECC error count in register file.",
				params.Params,
				false,
			),
		},
		devPciUtil: {
			metric: metric.New(
				"Returns PCIe utilization.",
				params.Params,
				false,
			),
		},
		devEncStats: {
			metric: metric.New(
				"Returns Encoder utilization.",
				params.Params,
				false,
			),
		},
		devVidFreq: {
			metric: metric.New(
				"Returns Video frequency in MHz.",
				params.Params,
				false,
			),
		},
		devGraphFreq: {
			metric: metric.New(
				"Returns Graphics frequency in MHz.",
				params.Params,
				false,
			),
		},
		devSMFreq: {
			metric: metric.New(
				"Returns SM frequency in MHz.",
				params.Params,
				false,
			),
		},
		devMemFreq: {
			metric: metric.New(
				"Returns Memory frequency in MHz.",
				params.Params,
				false,
			),
		},
		devEncoderUtil: {
			metric: metric.New(
				"Returns Encoder utilization.",
				params.Params,
				false,
			),
		},
		devDecoderUtil: {
			metric: metric.New(
				"Returns Decoder utilization.",
				params.Params,
				false,
			),
		},
		devUtil: {
			metric: metric.New(
				"Returns Device utilization.",
				params.Params,
				false,
			),
		},
		devEccMode: {
			metric: metric.New(
				"Returns Device current and pending ECC mode.",
				params.Params,
				false,
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

func (p *nvmlPlugin) setMetricFunctions() {
	handler := handlers.New(p.nvmlRunner)
	p.metrics[version].handler = handler.GetNVMLVersion
	p.metrics[driverVersion].handler = handler.GetDriverVersion
	p.metrics[devGet].handler = handlers.WithJSONResponse(handler.DeviceDiscovery)
	p.metrics[devCount].handler = handler.GetDeviceCount
	p.metrics[devTemp].handler = handler.GetDeviceTemperature
	p.metrics[devSerial].handler = handler.GetDeviceSerial
	p.metrics[devfanSpeedAVG].handler = handler.GetDeviceFanSpeed
	p.metrics[devPerfState].handler = handler.GetDevicePerfState
	p.metrics[devEneConsumption].handler = handler.GetDeviceEnergyConsumption
	p.metrics[devPowerLimit].handler = handler.GetDevicePowerLimit
	p.metrics[devPowerUsage].handler = handler.GetDevicePowerUsage
	p.metrics[devMemBar1].handler = handlers.WithJSONResponse(handler.GetBAR1MemoryInfo)
	p.metrics[devMemFBGet].handler = handlers.WithJSONResponse(handler.GetFBMemoryInfo)
	p.metrics[devMemErrors].handler = handlers.WithJSONResponse(handler.GetMemoryErrors)
	p.metrics[devRegErrors].handler = handlers.WithJSONResponse(handler.GetRegisterErrors)
	p.metrics[devPciUtil].handler = handlers.WithJSONResponse(handler.GetPCIeThroughput)
	p.metrics[devEncStats].handler = handlers.WithJSONResponse(handler.GetEncoderStats)
	p.metrics[devVidFreq].handler = handler.GetVideoFrequency
	p.metrics[devGraphFreq].handler = handler.GetGraphicsFrequency
	p.metrics[devSMFreq].handler = handler.GetSMFrequency
	p.metrics[devMemFreq].handler = handler.GetMemoryFrequency
	p.metrics[devEncoderUtil].handler = handler.GetEncoderUtilization
	p.metrics[devDecoderUtil].handler = handler.GetDecoderUtilization
	p.metrics[devUtil].handler = handlers.WithJSONResponse(handler.GetDeviceUtilisation)
	p.metrics[devEccMode].handler = handlers.WithJSONResponse(handler.GetECCMode)
}
