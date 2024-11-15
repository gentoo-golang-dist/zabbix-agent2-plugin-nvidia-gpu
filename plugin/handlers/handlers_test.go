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

package handlers

import (
	"context"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.zabbix.com/plugin/nvidia/pkg/nvml"
	nvmlmock "golang.zabbix.com/plugin/nvidia/pkg/nvml-mock"
	"golang.zabbix.com/plugin/nvidia/plugin/params"
	"golang.zabbix.com/sdk/errs"
)

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
			"-jsonMarshalErr",
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
			"-handlerErr",
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

func TestHandler_DriverVersion(t *testing.T) {
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
			"-invalid",
			expect{
				expectations: []*nvmlmock.Expectation{
					nvmlmock.NewExpectation("GetDriverVersion").ProvideOutput("").ProvideError(nvml.ErrNotFound),
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
			got, err := h.GetDriverVersion(context.TODO(), nil, nil...)

			// Check for error match
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.DriverVersion() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check for result match
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.DriverVersion() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected amount of calls not done")
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
						ProvideOutput(uint(0)).ProvideError(nvml.ErrUnknown),
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
						ProvideError(nvml.ErrUnknown),
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
									ProvideError(nvml.ErrUnknown),
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
									ProvideError(nvml.ErrUnknown),
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
				concurrentRuns: 1,
				nvmlRunner:     runner,
				deviceCacheMux: &sync.Mutex{},
				deviceCache:    make(map[string]nvml.Device),
			}

			got, err := h.DeviceDiscovery(context.TODO(), nil, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.DeviceDiscovery() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.DeviceDiscovery() = %s", diff)
			}

			done := runner.ExpectedCallsDone()
			if !done {
				t.Fatal("Expected calls not done")
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
			"-invalid",
			expect{
				expectations: []*nvmlmock.Expectation{
					nvmlmock.NewExpectation("GetDeviceCountV2").ProvideOutput(uint(1)).ProvideError(nvml.ErrNotFound),
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
			got, err := h.GetDeviceCount(context.TODO(), nil, nil...)

			// Check for error match
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.DriverVersion() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check for result match
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.DriverVersion() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected amount of calls not done")
			}
		})
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
			"-invalid",
			expect{
				expectations: []*nvmlmock.Expectation{
					nvmlmock.NewExpectation("GetNVMLVersion").ProvideOutput("").ProvideError(nvml.ErrNotFound),
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
			got, err := h.GetNVMLVersion(context.TODO(), nil, nil...)

			// Check for error match
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.DriverVersion() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check for result match
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.DriverVersion() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected amount of calls not done")
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
			name: "+valid",
			expect: expect{
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
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want:    int(55),
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInNVMLResponse",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetTemperature").
										ProvideOutput(int(0)).
										ProvideError(nvml.ErrCorruptedInforom),
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

			got, err := h.GetDeviceTemperature(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetDeviceTemperature() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetDeviceTemperature() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
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
			name: "+valid",
			expect: expect{
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
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want:    "12345",
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorGettingSerial",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetSerial").
										ProvideOutput("").
										ProvideError(nvml.ErrCorruptedInforom),
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

			got, err := h.GetDeviceSerial(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetDeviceSerial() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetDeviceSerial() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
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
			name: "+valid",
			expect: expect{
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
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want:    uint(55),
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInNVMLResponse",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetFanSpeed").
										ProvideOutput(uint(0)).
										ProvideError(nvml.ErrCorruptedInforom),
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

			got, err := h.GetDeviceFanSpeed(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetDeviceFanSpeed() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetDeviceFanSpeed() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
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
			name: "+valid",
			expect: expect{
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
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want:    uint(55),
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInNVMLResponse",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetPerformanceState").
										ProvideOutput(uint(0)).
										ProvideError(nvml.ErrCorruptedInforom),
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

			got, err := h.GetDevicePerfState(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetDevicePerfState() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetDevicePerfState() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
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
			name: "+valid",
			expect: expect{
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
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want:    uint(55),
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInNVMLResponse",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetPowerManagementLimit").
										ProvideOutput(uint(0)).
										ProvideError(nvml.ErrCorruptedInforom),
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

			got, err := h.GetDevicePowerLimit(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetDevicePowerLimit() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetDevicePowerLimit() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
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
			name: "+valid",
			expect: expect{
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
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want:    uint(55),
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInNVMLResponse",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetPowerUsage").
										ProvideOutput(uint(0)).
										ProvideError(nvml.ErrCorruptedInforom),
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

			got, err := h.GetDevicePowerUsage(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetDevicePowerUsage() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetDevicePowerUsage() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
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
			name: "+valid",
			expect: expect{
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
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want:    uint(55),
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInNVMLResponse",
			expect: expect{
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
										ProvideError(nvml.ErrCorruptedInforom),
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

			got, err := h.GetVideoFrequency(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetVideoFrequency() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetVideoFrequency() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
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
			name: "+valid",
			expect: expect{
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
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want:    uint(55),
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInNVMLResponse",
			expect: expect{
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
										ProvideError(nvml.ErrCorruptedInforom),
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

			got, err := h.GetGraphicsFrequency(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetGraphicsFrequency() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetGraphicsFrequency() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
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
			name: "+valid",
			expect: expect{
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
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want:    uint(55),
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInNVMLResponse",
			expect: expect{
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
										ProvideError(nvml.ErrCorruptedInforom),
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

			got, err := h.GetSMFrequency(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetSMFrequency() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetSMFrequency() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
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
			name: "+valid",
			expect: expect{
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
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want:    uint(55),
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInNVMLResponse",
			expect: expect{
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
										ProvideError(nvml.ErrCorruptedInforom),
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

			got, err := h.GetMemoryFrequency(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetMemoryFrequency() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetMemoryFrequency() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
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
			name: "+valid",
			expect: expect{
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
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want:    uint64(55),
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInNVMLResponse",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetTotalEnergyConsumption").
										ProvideOutput(uint64(0)).
										ProvideError(nvml.ErrCorruptedInforom),
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

			got, err := h.GetDeviceEnergyConsumption(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetDeviceEnergyConsumption() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetDeviceEnergyConsumption() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
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
			name: "+valid",
			expect: expect{
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
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want:    UtilisationRates{GPU: 55, Memory: 99},
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInNVMLResponse",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetUtilizationRates").
										ProvideOutput(uint(0), uint(0)).
										ProvideError(nvml.ErrCorruptedInforom),
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

			got, err := h.GetDeviceUtilisation(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetUtilizationRates() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetUtilizationRates() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
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
			name: "+valid",
			expect: expect{
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
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want:    ECCErrors{Corrected: 55, Uncorrected: 25},
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInFirstNVMLResponse",
			expect: expect{
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
										).ProvideOutput(uint64(0)).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInSecondNVMLResponse",
			expect: expect{
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
										).ProvideOutput(uint64(0)).ProvideError(nvml.ErrGpuIsLost),
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

			got, err := h.GetMemoryErrors(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetMemoryErrors() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetMemoryErrors() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
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
			name: "+valid",
			expect: expect{
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
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want:    ECCErrors{Corrected: 55, Uncorrected: 25},
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInFirstNVMLResponse",
			expect: expect{
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
										).ProvideOutput(uint64(0)).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInSecondNVMLResponse",
			expect: expect{
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
										).ProvideOutput(uint64(0)).ProvideError(nvml.ErrGpuIsLost),
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

			got, err := h.GetRegistryErrors(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetRegistryErrors() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetRegistryErrors() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
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
			name: "+valid",
			expect: expect{
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
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want:    uint(55),
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInNVMLResponse",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetEncoderUtilization").
										ProvideOutput(uint(0), uint(0)).
										ProvideError(nvml.ErrCorruptedInforom),
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

			got, err := h.GetEncoderUtilization(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetEncoderUtilization() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetEncoderUtilization() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
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
			name: "+valid",
			expect: expect{
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
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want:    uint(55),
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInNVMLResponse",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetDecoderUtilization").
										ProvideOutput(uint(0), uint(0)).
										ProvideError(nvml.ErrCorruptedInforom),
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

			got, err := h.GetDecoderUtilization(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetDecoderUtilization() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetDecoderUtilization() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
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
			name: "+valid",
			expect: expect{
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
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want:    ECCMode{Currect: true, Pending: false},
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInNVMLResponse",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetEccMode").
										ProvideOutput(false, false).
										ProvideError(nvml.ErrCorruptedInforom),
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

			got, err := h.GetECCMode(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetECCMode() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetECCMode() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
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
			name: "+valid",
			expect: expect{
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
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want:    PCIeUtil{Receive: 55, Transmit: 25},
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInFirstNVMLResponse",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetPCIeThroughput").
										WithExpextedArgs(nvml.RX).
										ProvideOutput(uint(0)).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInSecondNVMLResponse",
			expect: expect{
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
										ProvideOutput(uint(0)).ProvideError(nvml.ErrGpuIsLost),
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

			got, err := h.GetPCIeThroughput(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetPCIeThroughput() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetPCIeThroughput() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
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
			name: "+valid",
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
										}),
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
			want: &nvml.MemoryInfoV2{
				Total:    10,
				Used:     8,
				Free:     2,
				Reserved: 1,
			},
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInNVMLResponse",
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
										ProvideError(nvml.ErrCorruptedInforom),
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

			got, err := h.GetFBMemoryInfo(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetFBMemoryInfo() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetFBMemoryInfo() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
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
			name: "+valid",
			expect: expect{
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
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want: &nvml.MemoryInfo{
				Total: 10,
				Used:  8,
				Free:  2,
			},
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInNVMLResponse",
			expect: expect{
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
										ProvideError(nvml.ErrCorruptedInforom),
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

			got, err := h.GetBAR1MemoryInfo(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetBAR1MemoryInfo() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetBAR1MemoryInfo() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
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
			name: "+valid",
			expect: expect{
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
			args: args{
				metricParams: map[string]string{
					params.DeviceUUIDParamName: "test-uuid",
				},
			},
			want: EncoderStats{
				SessionCount: 1,
				FPS:          2,
				Latency:      3,
			},
			wantErr: false,
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
			name: "-deviceNotFound",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(nil).ProvideError(nvml.ErrGpuIsLost),
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
		{
			name: "-errorInNVMLResponse",
			expect: expect{
				device: device{
					deviceUUID: "test-uuid",
					expectations: []*nvmlmock.Expectation{
						nvmlmock.NewExpectation("GetDeviceByUUID").
							WithExpextedArgs("test-uuid").
							ProvideOutput(
								nvmlmock.NewMockDevice(t).ExpectCalls(
									nvmlmock.NewExpectation("GetEncoderStats").
										ProvideOutput(uint(0), uint(0), uint(0)).
										ProvideError(nvml.ErrCorruptedInforom),
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

			got, err := h.GetEncoderStats(context.TODO(), tt.args.metricParams, nil...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Handler.GetEncoderStats() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("Handler.GetEncoderStats() = %s", diff)
			}

			if !runner.ExpectedCallsDone() {
				t.Fatalf("Expected calls not done")
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	runner := nvmlmock.NewMockRunner(t).ExpectCalls()

	h := New(runner)

	if h.nvmlRunner != runner {
		t.Fatalf("Runner not as expected")
	}

	if h.deviceCache == nil {
		t.Fatalf("Device cache not set")
	}

	if h.deviceCacheMux == nil {
		t.Fatalf("Device cache mutex not set")
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
			name: "+deviceFromCache",
			fields: fields{
				runnerExpect: []*nvmlmock.Expectation{},
				deviceInCache: map[string]TestDevice{
					"test-1": {UUID: "test-1"},
					"test-2": {UUID: "test-2"},
					"test-3": {UUID: "test-3"},
				},
			},
			args: args{
				uuid: "test-2",
			},
			want:    TestDevice{UUID: "test-2"},
			wantErr: false,
		},
		{
			name: "+deviceFromNVML",
			fields: fields{
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
			args: args{
				uuid: "test-2",
			},
			want:    TestDevice{UUID: "test-2"},
			wantErr: false,
		},
		{
			name: "-noDeviceFound",
			fields: fields{
				runnerExpect: []*nvmlmock.Expectation{
					nvmlmock.NewExpectation("GetDeviceByUUID").
						WithExpextedArgs("test-2").
						ProvideOutput(nil).
						ProvideError(nvml.ErrNotFound),
				},
				deviceInCache: map[string]TestDevice{
					"test-1": {UUID: "test-1"},
					"test-3": {UUID: "test-3"},
				},
			},
			args: args{
				uuid: "test-2",
			},
			want:    nil,
			wantErr: true,
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
				t.Fatal("Expected calls not done")
			}
		})
	}
}
