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
	"context"
	stdlog "log"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.zabbix.com/plugin/nvidia/pkg/nvml"
	nvmlmock "golang.zabbix.com/plugin/nvidia/pkg/nvml-mock"
	"golang.zabbix.com/sdk/errs"
	"golang.zabbix.com/sdk/log"
	"golang.zabbix.com/sdk/metric"
	"golang.zabbix.com/sdk/plugin"
)

type MockCtxProvider struct {
	plugin.ContextProvider
	timeout int
}

type runnerSetMock struct {
	failed bool
}

func (m *MockCtxProvider) Timeout() int {
	return m.timeout
}

func (m *MockCtxProvider) LegacyTimeout() bool {
	return false
}

func (m runnerSetMock) init() error {
	if m.failed {
		return errs.New("fail")
	}

	return nil
}

func TestMain(m *testing.M) {
	log.DefaultLogger = stdlog.New(os.Stdout, "", stdlog.LstdFlags)
	exitVal := m.Run()
	os.Exit(exitVal)
}

func Test_nvmlPlugin_Export(t *testing.T) {
	t.Parallel()

	testParams := []*metric.Param{
		metric.NewConnParam("session", "Test session.").WithSession(),
		metric.NewConnParam("conn", "Test Conn."),
		metric.NewConnParam("default", "Test default.").WithDefault("default"),
	}

	type fields struct {
		returnErr bool
	}

	type args struct {
		key       string
		rawParams []string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			fields{
				returnErr: false,
			},
			args{
				key:       "test",
				rawParams: []string{},
			},
			"success",
			false,
		},
		{
			"-handlerErr",
			fields{
				returnErr: true,
			},
			args{
				key:       "test",
				rawParams: []string{},
			},
			nil,
			true,
		},
		{
			"-metricNotFound",
			fields{
				returnErr: false,
			},
			args{
				key:       "invalid",
				rawParams: []string{},
			},
			nil,
			true,
		},
		{
			"-invalidParams",
			fields{
				returnErr: false,
			},
			args{
				key:       "test",
				rawParams: []string{"one", "two", "three", "four"},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := &NvmlPlugin{
				metrics: map[string]*nvmlMetric{
					"test": {
						metric: metric.New("test metric", testParams, false),
						handler: func(
							ctx context.Context, metricParams map[string]string, extraParams ...string,
						) (any, error) {
							if tt.fields.returnErr {
								return "", errs.New("fail")
							}

							return "success", nil
						},
					},
				},
				config: &pluginConfig{},
			}

			ctxPrvider := MockCtxProvider{timeout: 2}

			got, err := p.Export(tt.args.key, tt.args.rawParams, &ctxPrvider)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NvmlPlugin.Export() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("NvmlPlugin.Export() = %s", diff)
			}
		})
	}
}

func Test_nvmlPlugin_registerMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			"+valid",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := (&NvmlPlugin{}).registerMetrics()
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"NvmlPlugin.registerMetrics() error = %v, wantErr %v",
					err, tt.wantErr,
				)
			}
		})
	}
}

func Test_nvmlPlugin_Stop(t *testing.T) {
	t.Parallel()

	type fields struct {
		runnerExpect []*nvmlmock.Expectation
	}

	tests := []struct {
		name   string
		fields fields
	}{
		{
			"+valid",
			fields{
				[]*nvmlmock.Expectation{
					nvmlmock.NewExpectation("ShutdownNVML").ProvideError(nil),
					nvmlmock.NewExpectation("Close").ProvideError(nil),
				},
			},
		},
		{
			"-nvmlRunnerShutdownNVMLError",
			fields{
				[]*nvmlmock.Expectation{
					nvmlmock.NewExpectation("ShutdownNVML").ProvideError(nvml.ErrNotFound),
					nvmlmock.NewExpectation("Close").ProvideError(nil),
				},
			},
		},
		{
			"-nvmlRunnerCloseError",
			fields{
				[]*nvmlmock.Expectation{
					nvmlmock.NewExpectation("ShutdownNVML").ProvideError(nil),
					nvmlmock.NewExpectation("Close").ProvideError(errs.New("fail")),
				},
			},
		},
		{
			"-allErrors",
			fields{
				[]*nvmlmock.Expectation{
					nvmlmock.NewExpectation("ShutdownNVML").ProvideError(nvml.ErrNotFound),
					nvmlmock.NewExpectation("Close").ProvideError(errs.New("fail")),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.fields.runnerExpect...)

			p := &NvmlPlugin{
				nvmlRunner: runner,
			}

			p.Logger = log.New("test")

			p.Stop()

			done := runner.ExpectedCallsDone()
			if !done {
				t.Fatal("NvmlPlugin.Stop() expected calls not done")
			}
		})
	}
}

func Test_nvmlPlugin_Start(t *testing.T) {
	t.Parallel()

	type fields struct {
		runnerExpect []*nvmlmock.Expectation
		failed       bool
	}

	type expect struct {
		shouldPanic bool
	}

	tests := []struct {
		name   string
		fields fields
		expect expect
	}{
		{
			"+validWithInitV2",
			fields{
				[]*nvmlmock.Expectation{
					nvmlmock.NewExpectation("InitV2").ProvideError(nil),
				},
				false,
			},
			expect{
				shouldPanic: false,
			},
		},
		{
			"+validWithInit",
			fields{
				[]*nvmlmock.Expectation{
					nvmlmock.NewExpectation("InitV2").ProvideError(nvml.ErrFunctionNotFound),
					nvmlmock.NewExpectation("Init").ProvideError(nil),
				},
				false,
			},
			expect{
				shouldPanic: false,
			},
		},
		{
			"-nvmlInitError",
			fields{
				[]*nvmlmock.Expectation{
					nvmlmock.NewExpectation("InitV2").ProvideError(nvml.ErrFunctionNotFound),
					nvmlmock.NewExpectation("Init").ProvideError(nvml.ErrFunctionNotFound),
				},
				false,
			},
			expect{
				shouldPanic: true,
			},
		},
		{
			"-nvmlRunnerInitError",
			fields{
				[]*nvmlmock.Expectation{},
				true,
			},
			expect{
				shouldPanic: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.fields.runnerExpect...)

			rm := runnerSetMock{
				failed: tt.fields.failed,
			}

			p := &NvmlPlugin{
				nvmlRunner:    runner,
				setNvmlRunner: rm.init,
				metrics:       getMetrics(),
			}

			p.Logger = log.New("test")

			defer func() {
				r := recover()
				if tt.expect.shouldPanic && r == nil {
					t.Fatalf("NvmlPlugin.Start() expected panic did not occur")
				}

				if !tt.expect.shouldPanic && r != nil {
					t.Fatalf("NvmlPlugin.Start() unexpected panic occurred")
				}
			}()

			p.Start()

			done := runner.ExpectedCallsDone()
			if !done {
				t.Fatal("NvmlPlugin.Start() expected calls not done")
			}
		})
	}
}

func getMetrics() map[string]*nvmlMetric {
	return map[string]*nvmlMetric{
		version:           {},
		driverVersion:     {},
		devGet:            {},
		devCount:          {},
		devTemp:           {},
		devSerial:         {},
		devfanSpeedAVG:    {},
		devPerfState:      {},
		devEneConsumption: {},
		devPowerLimit:     {},
		devPowerUsage:     {},
		devMemBar1:        {},
		devMemFBGet:       {},
		devMemErrors:      {},
		devRegErrors:      {},
		devPciUtil:        {},
		devEncStats:       {},
		devVidFreq:        {},
		devGraphFreq:      {},
		devSMFreq:         {},
		devMemFreq:        {},
		devEncoderUtil:    {},
		devDecoderUtil:    {},
		devUtil:           {},
		devEccMode:        {},
	}
}
