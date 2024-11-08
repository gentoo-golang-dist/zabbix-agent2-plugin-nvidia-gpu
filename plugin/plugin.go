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
	Name = "Nvidia"
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
	p := &nvmlPlugin{}

	err := p.registerMetrics()
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

	return nil
}

// Start starts the example plugin. Is required for plugin to match runner interface.
func (p *nvmlPlugin) Start() {
	p.Logger.Infof("Start called")

	runner, err := nvml.NewNVMLRunner()
	if err != nil {
		p.Logger.Critf("error creating NVML runner: %v", err)
		panic(err)
	}

	err = initNVML(runner, p.Logger)
	if err != nil {
		p.Logger.Critf("error initializing NVML library: %v", err)
		panic(err)
	}

	p.nvmlRunner = runner
}

// Stop stops the example plugin. Is required for plugin to match runner interface.
func (p *nvmlPlugin) Stop() {
	p.Logger.Infof("Stop called")

	err := p.nvmlRunner.ShutdownNVML()
	if err != nil {
		p.Logger.Errf("failed to shutdown nvml %v", err)
	}

	err = p.nvmlRunner.Close()
	if err != nil {
		p.Logger.Errf("failed to close dynamic library %v", err)
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

	metricParams, extraParams, hardcodedParams, err := m.metric.EvalParams(rawParams, p.config.Sessions)
	if err != nil {
		return nil, errs.Wrap(err, "failed to evaluate metric parameters")
	}

	err = metric.SetDefaults(metricParams, hardcodedParams, p.config.Default)
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
				params.Params,
				false,
			),
			handler: handlers.WithJSONResponse(
				handler.NVMLVersion,
			),
		},
		"nvml.system.driver.version": {
			metric: metric.New(
				"Returns local Nvidia driver version.",
				params.Params,
				false,
			),
			handler: handlers.WithJSONResponse(
				handler.DriverVersion,
			),
		},
		// "example.go.env": {
		// 	metric: metric.New(
		// 		"Returns the result rows of a custom query.",
		// 		params.Params,
		// 		true,
		// 	),
		// 	handler: handlers.WithJSONResponse(
		// 		handlers.WithCredentialValidation(handler.GoEnvironment),
		// 	),
		// },
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
