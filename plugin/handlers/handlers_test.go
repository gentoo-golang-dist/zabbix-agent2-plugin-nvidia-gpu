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
					nvmlmock.NewExpectation("GetDeviceCount").ProvideOutput(uint(1)),
				},
			},
			uint(1),
			false,
		},
		{
			"-invalid",
			expect{
				expectations: []*nvmlmock.Expectation{
					nvmlmock.NewExpectation("GetDeviceCount").ProvideOutput(uint(1)).ProvideError(nvml.ErrNotFound),
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
										ProvideOutput(int(55)), // Expectation for valid temperature retrieval
								),
							),
						// nvmlmock.NewExpectation("GetTemperature").
						// 	ProvideOutput(int(55)), // Expectation for valid temperature retrieval
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
						// nvmlmock.NewExpectation("GetTemperature").
						// 	ProvideOutput(int(55)), // Expectation for valid temperature retrieval
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
			name: "-errorGettingTemperature",
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
