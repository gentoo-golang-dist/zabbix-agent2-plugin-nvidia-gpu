/*
**   Copyright 2001-2024 Zabbix SIA
**
**   Licensed under the Apache License, Version 2.0 (the "License");
**   you may not use this file except in compliance with the License.
**   You may obtain a copy of the License at
**
**       http://www.apache.org/licenses/LICENSE-2.0
**
**   Unless required by applicable law or agreed to in writing, software
**   distributed under the License is distributed on an "AS IS" BASIS,
**   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
**   See the License for the specific language governing permissions and
**   limitations under the License.
**/

package handlers

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.zabbix.com/plugin/nvidia/internal/plugin/params"
	"golang.zabbix.com/plugin/nvidia/pkg/nvml"
	nvmlmock "golang.zabbix.com/plugin/nvidia/pkg/nvml-mock"
	"golang.zabbix.com/sdk/errs"
)

func TestNew(t *testing.T) {
	t.Parallel()

	runner := nvmlmock.NewMockRunner(t).ExpectCalls()

	h := New(runner)

	if h.nvmlRunner != runner {
		t.Fatal("New() runner not as expected")
	}

	if h.deviceCache == nil {
		t.Fatal("New() device cache not set")
	}

	if h.deviceCacheMux == nil {
		t.Fatal("New() device cache mutex not set")
	}
}

func TestHandler_GetNVMLVersion(t *testing.T) {
	t.Parallel()

	type expect struct {
		expectations []*nvmlmock.Expectation
	}

	tests := []struct {
		name    string
		expect  expect
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				expectations: []*nvmlmock.Expectation{
					nvmlmock.NewExpectation("GetNVMLVersion").ProvideOutput("Mock NVML Version"),
				},
			},
			"Mock NVML Version",
			false,
		},
		{
			"-nvmlRunnerGetNVMLVersionError",
			expect{
				expectations: []*nvmlmock.Expectation{
					nvmlmock.NewExpectation("GetNVMLVersion").ProvideOutput("").ProvideError(errors.New("fail")),
				},
			},
			"",
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.expectations...)
			// Initialize Handler with the mocked nvmlRunner
			h := &Handler{
				nvmlRunner: runner,
			}

			// Call the method being tested
			got, err := h.GetNVMLVersion(context.Background(), nil, nil...)

			// Check for error match
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.DriverVersion() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check for result match
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.DriverVersion() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.DriverVersion() expected amount of calls not done")
			}
		})
	}
}

func TestHandler_GetDriverVersion(t *testing.T) {
	t.Parallel()

	type expect struct {
		expectations []*nvmlmock.Expectation
	}

	tests := []struct {
		name    string
		expect  expect
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				expectations: []*nvmlmock.Expectation{
					nvmlmock.NewExpectation("GetDriverVersion").ProvideOutput("Mock Driver Version"),
				},
			},
			"Mock Driver Version",
			false,
		},
		{
			"-nvmlRunnerGetDriverVersionError",
			expect{
				expectations: []*nvmlmock.Expectation{
					nvmlmock.NewExpectation("GetDriverVersion").ProvideOutput("").ProvideError(errors.New("fail")),
				},
			},
			"",
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.expectations...)
			// Initialize Handler with the mocked nvmlRunner
			h := &Handler{
				nvmlRunner: runner,
			}

			// Call the method being tested
			got, err := h.GetDriverVersion(context.Background(), nil, nil...)

			// Check for error match
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.DriverVersion() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check for result match
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.DriverVersion() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.DriverVersion() expected amount of calls not done")
			}
		})
	}
}

func TestHandler_DeviceDiscovery(t *testing.T) {
	t.Parallel()

	type expect struct {
		expectations []*nvmlmock.Expectation
	}

	tests := []struct {
		name    string
		expect  expect
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				expectations: []*nvmlmock.Expectation{
					nvmlmock.NewExpectation("GetDeviceCountV2").
						ProvideOutput(uint(2)),
					nvmlmock.NewExpectation("GetDeviceByIndexV2").
						WithExpextedArgs(uint(0)).
						ProvideOutput(
							nvmlmock.NewMockDevice(t).ExpectCalls(
								nvmlmock.NewExpectation("GetUUID").ProvideOutput("UUID1"),
								nvmlmock.NewExpectation("GetName").ProvideOutput("Name1"),
							),
						),
					nvmlmock.NewExpectation("GetDeviceByIndexV2").
						WithExpextedArgs(uint(1)).
						ProvideOutput(
							nvmlmock.NewMockDevice(t).ExpectCalls(
								nvmlmock.NewExpectation("GetUUID").ProvideOutput("UUID2"),
								nvmlmock.NewExpectation("GetName").ProvideOutput("Name2"),
							),
						),
				},
			},
			[]DiscoveryDevice{{"UUID1", "Name1"}, {"UUID2", "Name2"}},
			false,
		},
		{
			"-getCountError",
			expect{
				expectations: []*nvmlmock.Expectation{
					nvmlmock.NewExpectation("GetDeviceCountV2").
						ProvideOutput(uint(0)).ProvideError(errors.New("fail")),
				},
			},
			nil,
			true,
		},
		{
			"-getDeviceByIndexError",
			expect{
				expectations: []*nvmlmock.Expectation{
					nvmlmock.NewExpectation("GetDeviceCountV2").
						ProvideOutput(uint(2)),
					nvmlmock.NewExpectation("GetDeviceByIndexV2").
						WithExpextedArgs(uint(0)).
						ProvideOutput(nil).
						ProvideError(errors.New("fail")),
				},
			},
			nil,
			true,
		},
		{
			"-getUUIDError",
			expect{
				expectations: []*nvmlmock.Expectation{
					nvmlmock.NewExpectation("GetDeviceCountV2").
						ProvideOutput(uint(2)),
					nvmlmock.NewExpectation("GetDeviceByIndexV2").
						WithExpextedArgs(uint(0)).
						ProvideOutput(
							nvmlmock.NewMockDevice(t).ExpectCalls(
								nvmlmock.NewExpectation("GetUUID").
									ProvideOutput("").
									ProvideError(errors.New("fail")),
							),
						),
				},
			},
			nil,
			true,
		},
		{
			"-getNameError",
			expect{
				expectations: []*nvmlmock.Expectation{
					nvmlmock.NewExpectation("GetDeviceCountV2").
						ProvideOutput(uint(2)),
					nvmlmock.NewExpectation("GetDeviceByIndexV2").
						WithExpextedArgs(uint(0)).
						ProvideOutput(
							nvmlmock.NewMockDevice(t).ExpectCalls(
								nvmlmock.NewExpectation("GetUUID").
									ProvideOutput("123"),
								nvmlmock.NewExpectation("GetName").
									ProvideOutput("").
									ProvideError(errors.New("fail")),
							),
						),
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.expectations...)

			h := &Handler{
				concurrentDeviceDiscoveries: 1,
				nvmlRunner:                  runner,
				deviceCacheMux:              &sync.Mutex{},
				deviceCache:                 make(map[string]nvml.Device),
			}

			got, err := h.DeviceDiscovery(context.Background(), nil, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.DeviceDiscovery() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.DeviceDiscovery() = %s", diff)
			}

			done := runner.ExpectedCallsDone()
			if !done {
				t.Fatal("Handler.DeviceDiscovery() expected calls not done")
			}
		})
	}
}

func TestHandler_GetDeviceCount(t *testing.T) {
	t.Parallel()

	type expect struct {
		expectations []*nvmlmock.Expectation
	}

	tests := []struct {
		name    string
		expect  expect
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				expectations: []*nvmlmock.Expectation{
					nvmlmock.NewExpectation("GetDeviceCountV2").ProvideOutput(uint(1)),
				},
			},
			uint(1),
			false,
		},
		{
			"-nvmlRunnerGetDeviceCountV2Error",
			expect{
				expectations: []*nvmlmock.Expectation{
					nvmlmock.NewExpectation("GetDeviceCountV2").ProvideOutput(uint(1)).ProvideError(errors.New("fail")),
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.expectations...)
			// Initialize Handler with the mocked nvmlRunner
			h := &Handler{
				nvmlRunner: runner,
			}

			// Call the method being tested
			got, err := h.GetDeviceCount(context.Background(), nil, nil...)

			// Check for error match
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.DriverVersion() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check for result match
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.DriverVersion() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.DriverVersion() expected amount of calls not done")
			}
		})
	}
}

func TestHandler_GetDeviceTemperature(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetTemperature").
										ProvideOutput(int(55)),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			int(55),
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-nvmlDeviceGetTemperatureError",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetTemperature").
										ProvideOutput(int(0)).
										ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetDeviceTemperature(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetDeviceTemperature() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetDeviceTemperature() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetDeviceTemperature() expected calls not done")
			}
		})
	}
}

func TestHandler_GetDeviceSerial(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetSerial").
										ProvideOutput("12345"),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			"12345",
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-errorGettingSerial",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetSerial").
										ProvideOutput("").
										ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetDeviceSerial(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetDeviceSerial() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetDeviceSerial() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetDeviceSerial() expected calls not done")
			}
		})
	}
}

func TestHandler_GetDeviceFanSpeed(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetFanSpeed").
										ProvideOutput(uint(55)),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			uint(55),
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-nvmlDeviceGetFanSpeedError",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetFanSpeed").
										ProvideOutput(uint(0)).
										ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetDeviceFanSpeed(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetDeviceFanSpeed() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetDeviceFanSpeed() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetDeviceFanSpeed() expected calls not done")
			}
		})
	}
}

func TestHandler_GetDevicePerfState(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetPerformanceState").
										ProvideOutput(uint(55)),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			uint(55),
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-nvmlDeviceGetPerformanceStateError",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetPerformanceState").
										ProvideOutput(uint(0)).
										ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetDevicePerfState(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetDevicePerfState() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetDevicePerfState() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetDevicePerfState() expected calls not done")
			}
		})
	}
}

func TestHandler_GetDeviceEnergyConsumption(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetTotalEnergyConsumption").
										ProvideOutput(uint64(55)),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			uint64(55),
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-nvmlDeviceGetTotalEnergyConsumptionError",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetTotalEnergyConsumption").
										ProvideOutput(uint64(0)).
										ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetDeviceEnergyConsumption(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetDeviceEnergyConsumption() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetDeviceEnergyConsumption() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetDeviceEnergyConsumption() expected calls not done")
			}
		})
	}
}

func TestHandler_GetDevicePowerLimit(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetPowerManagementLimit").
										ProvideOutput(uint(55)),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			uint(55),
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-nvmlDeviceGetPowerManagementLimitError",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetPowerManagementLimit").
										ProvideOutput(uint(0)).
										ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetDevicePowerLimit(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetDevicePowerLimit() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetDevicePowerLimit() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetDevicePowerLimit() expected calls not done")
			}
		})
	}
}

func TestHandler_GetDevicePowerUsage(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetPowerUsage").
										ProvideOutput(uint(55)),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			uint(55),
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-nvmlDeviceGetPowerUsageError",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetPowerUsage").
										ProvideOutput(uint(0)).
										ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetDevicePowerUsage(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetDevicePowerUsage() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetDevicePowerUsage() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetDevicePowerUsage() expected calls not done")
			}
		})
	}
}

func TestHandler_GetBAR1MemoryInfo(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetBAR1MemoryInfo").
										ProvideOutput(&nvml.MemoryInfo{
											Total: 10,
											Used:  8,
											Free:  2,
										}),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			&nvml.MemoryInfo{
				Total: 10,
				Used:  8,
				Free:  2,
			},
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-nvmlDeviceGetBAR1MemoryInfoError",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetBAR1MemoryInfo").
										ProvideOutput(&nvml.MemoryInfo{
											Total: 10,
											Used:  8,
											Free:  2,
										}).
										ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetBAR1MemoryInfo(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetBAR1MemoryInfo() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetBAR1MemoryInfo() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetBAR1MemoryInfo() expected calls not done")
			}
		})
	}
}

func TestHandler_GetFBMemoryInfo(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetMemoryInfoV2").
										ProvideOutput(&nvml.MemoryInfoV2{
											Total:    10,
											Used:     8,
											Free:     2,
											Reserved: 1,
										}),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			&nvml.MemoryInfoV2{
				Total:    10,
				Used:     8,
				Free:     2,
				Reserved: 1,
			},
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			name: "-nvmlDeviceGetMemoryInfoV2Error",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetMemoryInfoV2").
										ProvideOutput(&nvml.MemoryInfoV2{
											Total:    10,
											Used:     8,
											Free:     2,
											Reserved: 1,
										}).
										ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetFBMemoryInfo(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetFBMemoryInfo() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetFBMemoryInfo() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetFBMemoryInfo() expected calls not done")
			}
		})
	}
}

func TestHandler_GetMemoryErrors(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetMemoryErrorCounter").
										WithExpextedArgs(
											nvml.MemoryErrorTypeCorrected,
											nvml.MemoryLocationDevice,
											nvml.EccCounterTypeAggregate,
										).ProvideOutput(uint64(55)),
									nvmlmock.NewExpectation("GetMemoryErrorCounter").
										WithExpextedArgs(
											nvml.MemoryErrorTypeUncorrected,
											nvml.MemoryLocationDevice,
											nvml.EccCounterTypeAggregate,
										).ProvideOutput(uint64(25)),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			ECCErrors{Corrected: 55, Uncorrected: 25},
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-errorInFirstNVMLResponse",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetMemoryErrorCounter").
										WithExpextedArgs(
											nvml.MemoryErrorTypeCorrected,
											nvml.MemoryLocationDevice,
											nvml.EccCounterTypeAggregate,
										).ProvideOutput(uint64(0)).ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-errorInSecondNVMLResponse",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetMemoryErrorCounter").
										WithExpextedArgs(
											nvml.MemoryErrorTypeCorrected,
											nvml.MemoryLocationDevice,
											nvml.EccCounterTypeAggregate,
										).ProvideOutput(uint64(55)),
									nvmlmock.NewExpectation("GetMemoryErrorCounter").
										WithExpextedArgs(
											nvml.MemoryErrorTypeUncorrected,
											nvml.MemoryLocationDevice,
											nvml.EccCounterTypeAggregate,
										).ProvideOutput(uint64(0)).ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetMemoryErrors(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetMemoryErrors() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetMemoryErrors() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetMemoryErrors() expected calls not done")
			}
		})
	}
}

func TestHandler_GetRegistryErrors(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetMemoryErrorCounter").
										WithExpextedArgs(
											nvml.MemoryErrorTypeCorrected,
											nvml.MemoryLocationRegisterFile,
											nvml.EccCounterTypeAggregate,
										).ProvideOutput(uint64(55)),
									nvmlmock.NewExpectation("GetMemoryErrorCounter").
										WithExpextedArgs(
											nvml.MemoryErrorTypeUncorrected,
											nvml.MemoryLocationRegisterFile,
											nvml.EccCounterTypeAggregate,
										).ProvideOutput(uint64(25)),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			ECCErrors{Corrected: 55, Uncorrected: 25},
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-errorInFirstNVMLGetMemoryErrorCounter",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetMemoryErrorCounter").
										WithExpextedArgs(
											nvml.MemoryErrorTypeCorrected,
											nvml.MemoryLocationRegisterFile,
											nvml.EccCounterTypeAggregate,
										).ProvideOutput(uint64(0)).ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-errorInSecondNVMLGetMemoryErrorCounter",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetMemoryErrorCounter").
										WithExpextedArgs(
											nvml.MemoryErrorTypeCorrected,
											nvml.MemoryLocationRegisterFile,
											nvml.EccCounterTypeAggregate,
										).ProvideOutput(uint64(55)),
									nvmlmock.NewExpectation("GetMemoryErrorCounter").
										WithExpextedArgs(
											nvml.MemoryErrorTypeUncorrected,
											nvml.MemoryLocationRegisterFile,
											nvml.EccCounterTypeAggregate,
										).ProvideOutput(uint64(0)).ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetRegisterErrors(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetRegisterErrors() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetRegisterErrors() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetRegisterErrors() expected calls not done")
			}
		})
	}
}

func TestHandler_GetPCIeThroughput(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetPCIeThroughput").
										WithExpextedArgs(nvml.RX).
										ProvideOutput(uint(55)),
									nvmlmock.NewExpectation("GetPCIeThroughput").
										WithExpextedArgs(nvml.TX).
										ProvideOutput(uint(25)),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			PCIeUtil{Receive: 55, Transmit: 25},
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-errorInFirstNVMLGetPCIeThroughput",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetPCIeThroughput").
										WithExpextedArgs(nvml.RX).
										ProvideOutput(uint(0)).ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-errorInSecondNVMLGetPCIeThroughput",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetPCIeThroughput").
										WithExpextedArgs(nvml.RX).
										ProvideOutput(uint(55)),
									nvmlmock.NewExpectation("GetPCIeThroughput").
										WithExpextedArgs(nvml.TX).
										ProvideOutput(uint(0)).ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetPCIeThroughput(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetPCIeThroughput() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetPCIeThroughput() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetPCIeThroughput() expected calls not done")
			}
		})
	}
}

func TestHandler_GetEncoderStats(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetEncoderStats").
										ProvideOutput(uint(1), uint(2), uint(3)),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			EncoderStats{
				SessionCount: 1,
				FPS:          2,
				Latency:      3,
			},
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-nvmlDeviceGetEncoderStatsError",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetEncoderStats").
										ProvideOutput(uint(0), uint(0), uint(0)).
										ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetEncoderStats(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetEncoderStats() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetEncoderStats() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetEncoderStats() expected calls not done")
			}
		})
	}
}

func TestHandler_GetVideoFrequency(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetClockInfo").
										WithExpextedArgs(nvml.Video).
										ProvideOutput(uint(55)),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			uint(55),
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-nvmlDeviceGetClockInfoError",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetClockInfo").
										WithExpextedArgs(nvml.Video).
										ProvideOutput(uint(0)).
										ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetVideoFrequency(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetVideoFrequency() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetVideoFrequency() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetVideoFrequency() expected calls not done")
			}
		})
	}
}

func TestHandler_GetGraphicsFrequency(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetClockInfo").
										WithExpextedArgs(nvml.Graphics).
										ProvideOutput(uint(55)),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			uint(55),
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-nvmlDeviceGetClockInfoError",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetClockInfo").
										WithExpextedArgs(nvml.Graphics).
										ProvideOutput(uint(0)).
										ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetGraphicsFrequency(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetGraphicsFrequency() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetGraphicsFrequency() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetGraphicsFrequency() expected calls not done")
			}
		})
	}
}

func TestHandler_GetSMFrequency(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetClockInfo").
										WithExpextedArgs(nvml.SM).
										ProvideOutput(uint(55)),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			uint(55),
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-nvmlDeviceGetClockInfoError",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetClockInfo").
										WithExpextedArgs(nvml.SM).
										ProvideOutput(uint(0)).
										ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetSMFrequency(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetSMFrequency() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetSMFrequency() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetSMFrequency() expected calls not done")
			}
		})
	}
}

func TestHandler_GetMemoryFrequency(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetClockInfo").
										WithExpextedArgs(nvml.Memory).
										ProvideOutput(uint(55)),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			uint(55),
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-nvmlDeviceGetClockInfoError",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetClockInfo").
										WithExpextedArgs(nvml.Memory).
										ProvideOutput(uint(0)).
										ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetMemoryFrequency(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetMemoryFrequency() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetMemoryFrequency() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetMemoryFrequency() expected calls not done")
			}
		})
	}
}

func TestHandler_GetEncoderUtilization(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetEncoderUtilization").
										ProvideOutput(uint(55), uint(99)),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			uint(55),
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-nvmlDeviceGetEncoderUtilizationError",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetEncoderUtilization").
										ProvideOutput(uint(0), uint(0)).
										ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetEncoderUtilization(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetEncoderUtilization() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetEncoderUtilization() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetEncoderUtilization() expected calls not done")
			}
		})
	}
}

func TestHandler_GetDecoderUtilization(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetDecoderUtilization").
										ProvideOutput(uint(55), uint(99)),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			uint(55),
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-nvmlDeviceGetDecoderUtilizationError",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetDecoderUtilization").
										ProvideOutput(uint(0), uint(0)).
										ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetDecoderUtilization(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetDecoderUtilization() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetDecoderUtilization() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetDecoderUtilization() expected calls not done")
			}
		})
	}
}

func TestHandler_GetDeviceUtilisation(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetUtilizationRates").
										ProvideOutput(uint(55), uint(99)),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			UtilisationRates{Device: 55, Memory: 99},
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-nvmlDeviceGetUtilizationRatesError",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetUtilizationRates").
										ProvideOutput(uint(0), uint(0)).
										ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetDeviceUtilisation(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetUtilizationRates() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetUtilizationRates() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetUtilizationRates() expected calls not done")
			}
		})
	}
}

func TestHandler_GetECCMode(t *testing.T) {
	t.Parallel()

	type device struct {
		deviceUUID   string
		expectations []*nvmlmock.Expectation
	}

	type expect struct {
		device device
	}

	type args struct {
		metricParams map[string]string
	}

	tests := []struct {
		name    string
		expect  expect
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetEccMode").
										ProvideOutput(true, false),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			ECCMode{Current: true, Pending: false},
			false,
		},
		{
			"-noInMetricParams",
			expect{},
			args{
				map[string]string{},
			},
			nil,
			true,
		},
		{
			"-deviceNotFound",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(errors.New("fail")),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
		{
			"-nvmlDeviceGetEccModeError",
			expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetEccMode").
										ProvideOutput(false, false).
										ProvideError(errors.New("fail")),
								),
							),
					},
				},
			},
			args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.expect.device.expectations...)

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.GetECCMode(context.Background(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetECCMode() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetECCMode() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.GetECCMode() expected calls not done")
			}
		})
	}
}

func TestWithJSONResponse(t *testing.T) {
	t.Parallel()

	type args struct {
		value  any
		gotErr bool
	}

	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			"+valid",
			args{
				value: "foobar",
			},
			`"foobar"`,
			false,
		},
		{
			"+jsonObject",
			args{
				value: map[string]string{
					"foo":  "bar",
					"test": "true",
				},
			},
			`{"foo":"bar","test":"true"}`,
			false,
		},
		{
			"-jsonMarshalError",
			args{
				value: map[struct{ test string }]string{
					{test: "1"}: "bar",
					{test: "2"}: "foo",
				},
			},
			nil,
			true,
		},
		{
			"-handlerError",
			args{gotErr: true},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := WithJSONResponse(
				func(ctx context.Context, metricParams map[string]string, extraParams ...string) (any, error) {
					if tt.args.gotErr {
						return nil, errs.New("fail")
					}

					return tt.args.value, nil
				},
			)(context.Background(), nil)

			if (err != nil) != tt.wantErr {
				t.Fatalf("WithJSONResponse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("WithJSONResponse() = %s", diff)
			}
		})
	}
}

func TestHandler_getDeviceByUUID(t *testing.T) {
	t.Parallel()

	type TestDevice struct {
		nvml.Device
		UUID string
	}

	type fields struct {
		runnerExpect  []*nvmlmock.Expectation
		deviceInCache map[string]TestDevice
	}

	type args struct {
		uuid string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    any
		wantErr bool
	}{
		{
			"+deviceFromCache",
			fields{
				runnerExpect: []*nvmlmock.Expectation{},
				deviceInCache: map[string]TestDevice{
					"test-1": {UUID: "test-1"},
					"test-2": {UUID: "test-2"},
					"test-3": {UUID: "test-3"},
				},
			},
			args{
				uuid: "test-2",
			},
			TestDevice{UUID: "test-2"},
			false,
		},
		{
			"+deviceFromNVML",
			fields{
				runnerExpect: []*nvmlmock.Expectation{
					nvmlmock.NewExpectation("GetDeviceByUUID").
						WithExpextedArgs("test-2").
						ProvideOutput(TestDevice{UUID: "test-2"}).
						ProvideError(nil),
				},
				deviceInCache: map[string]TestDevice{
					"test-1": {UUID: "test-1"},
					"test-3": {UUID: "test-3"},
				},
			},
			args{
				uuid: "test-2",
			},
			TestDevice{UUID: "test-2"},
			false,
		},
		{
			"-noDeviceFound",
			fields{
				runnerExpect: []*nvmlmock.Expectation{
					nvmlmock.NewExpectation("GetDeviceByUUID").
						WithExpextedArgs("test-2").
						ProvideOutput(nil).
						ProvideError(errors.New("fail")),
				},
				deviceInCache: map[string]TestDevice{
					"test-1": {UUID: "test-1"},
					"test-3": {UUID: "test-3"},
				},
			},
			args{
				uuid: "test-2",
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := nvmlmock.NewMockRunner(t).ExpectCalls(tt.fields.runnerExpect...)

			deviceCache := make(map[string]nvml.Device)

			for key, device := range tt.fields.deviceInCache {
				device := device
				deviceCache[key] = device
			}

			h := &Handler{
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    deviceCache,
			}

			got, err := h.getDeviceByUUID(tt.args.uuid)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.getDeviceByUUID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.getDeviceByUUID() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatal("Handler.getDeviceByUUID() expected calls not done")
			}
		})
	}
}
