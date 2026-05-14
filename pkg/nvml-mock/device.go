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

package nvmlmock

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.zabbix.com/plugin/nvidia/pkg/nvml"
)

var (
	_ nvml.Device = (*MockDevice)(nil)
	_ Mocker      = (*MockDevice)(nil)
)

// MockDevice is mock for NVML device.
type MockDevice struct {
	nvml.Device

	expectations []*Expectation
	callIdx      int
	t            *testing.T
}

// NewMockDevice returns new mock device.
func NewMockDevice(t *testing.T) *MockDevice {
	t.Helper()

	return &MockDevice{
		t:            t,
		expectations: []*Expectation{},
	}
}

// ExpectCalls sets calls that are expected by mock.
func (m *MockDevice) ExpectCalls(expectations ...*Expectation) *MockDevice {
	m.expectations = expectations

	return m
}

// ExpectedCallsDone checks if all expected calls of mock and it's submocks are done.
func (m *MockDevice) ExpectedCallsDone() bool {
	m.t.Helper()
	expected := len(m.expectations)
	received := m.callIdx

	if expected == received {
		return true
	}

	for _, e := range m.expectations[received:] {
		m.t.Errorf("Not called %q", e.funcName)
	}

	m.t.Errorf("received %d out of %d expected calls", received, expected)

	return false
}

// SubMocks returns submocks of the mock.
func (m *MockDevice) SubMocks() []Mocker {
	var subMocks []Mocker

	for _, expect := range m.expectations {
		for _, out := range expect.out {
			subMock, ok := out.(Mocker)
			if !ok {
				continue
			}

			subMocks = append(subMocks, subMock)
			subMocks = append(subMocks, subMock.SubMocks()...)
		}
	}

	return subMocks
}

// GetUUID is mock function.
func (m *MockDevice) GetUUID() (string, error) {
	m.t.Helper()

	res, err := m.handleFunctionCall("GetUUID")

	uuid, ok := res.out[0].(string)
	if !ok {
		m.t.Fatalf("expected string in GetUUID, got %T", res.out[0])
	}

	return uuid, err
}

// GetName is mock function.
func (m *MockDevice) GetName() (string, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetName")

	uuid, ok := res.out[0].(string)
	if !ok {
		m.t.Fatalf("expected string in GetName, got %T", res.out[0])
	}

	return uuid, err
}

// GetSerial is mock function.
func (m *MockDevice) GetSerial() (string, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetSerial")

	serial, ok := res.out[0].(string)
	if !ok {
		m.t.Fatalf("expected string in GetSerial, got %T", res.out[0])
	}

	return serial, err
}

// GetTemperature is mock function.
func (m *MockDevice) GetTemperature() (int, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetTemperature")

	temperature, ok := res.out[0].(int)
	if !ok {
		m.t.Fatalf("expected string in GetTemperature, got %T", res.out[0])
	}

	return temperature, err
}

// GetFanSpeed is mock function.
func (m *MockDevice) GetFanSpeed() (uint, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetFanSpeed")

	fanSpeed, ok := res.out[0].(uint)
	if !ok {
		m.t.Fatalf("expected string in GetFanSpeed, got %T", res.out[0])
	}

	return fanSpeed, err
}

// GetPerformanceState is mock function.
func (m *MockDevice) GetPerformanceState() (uint, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetPerformanceState")

	state, ok := res.out[0].(uint)
	if !ok {
		m.t.Fatalf("expected string in GetPerformanceState, got %T", res.out[0])
	}

	return state, err
}

// GetPowerManagementLimit is mock function.
func (m *MockDevice) GetPowerManagementLimit() (uint, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetPowerManagementLimit")

	limit, ok := res.out[0].(uint)
	if !ok {
		m.t.Fatalf("expected string in GetPowerManagementLimit, got %T", res.out[0])
	}

	return limit, err
}

// GetPowerUsage is mock function.
func (m *MockDevice) GetPowerUsage() (uint, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetPowerUsage")

	usage, ok := res.out[0].(uint)
	if !ok {
		m.t.Fatalf("expected string in GetPowerUsage, got %T", res.out[0])
	}

	return usage, err
}

// GetClockInfo is mock function.
func (m *MockDevice) GetClockInfo(clock nvml.ClockType) (uint, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetClockInfo", clock)

	c, ok := res.out[0].(uint)
	if !ok {
		m.t.Fatalf("expected string in GetClockInfo, got %T", res.out[0])
	}

	return c, err
}

// GetTotalEnergyConsumption is mock function.
func (m *MockDevice) GetTotalEnergyConsumption() (uint64, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetTotalEnergyConsumption")

	consumption, ok := res.out[0].(uint64)
	if !ok {
		m.t.Fatalf("expected uint in GetTotalEnergyConsumption, got %T", res.out[0])
	}

	return consumption, err
}

// GetUtilizationRates is mock function.
func (m *MockDevice) GetUtilizationRates() (uint, uint, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetUtilizationRates")

	util1, ok := res.out[0].(uint)
	if !ok {
		m.t.Fatalf("expected uint for util1 in GetUtilizationRates, got %T", res.out[0])
	}

	util2, ok := res.out[1].(uint)
	if !ok {
		m.t.Fatalf("expected uint for util2 in GetUtilizationRates, got %T", res.out[1])
	}

	return util1, util2, err
}

// GetMemoryErrorCounter is mock function.
func (m *MockDevice) GetMemoryErrorCounter(
	memoryType nvml.MemoryErrorType,
	memoryLocation nvml.MemoryLocation,
	counterType nvml.EccCounterType) (uint64, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetMemoryErrorCounter",
		memoryType, memoryLocation, counterType,
	)

	rate, ok := res.out[0].(uint64)
	if !ok {
		m.t.Fatalf("expected uint64 for rate in GetMemoryErrorCounter, got %T", res.out[0])
	}

	return rate, err
}

// GetEncoderUtilization is mock function.
func (m *MockDevice) GetEncoderUtilization() (uint, uint, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetEncoderUtilization")

	util1, ok := res.out[0].(uint)
	if !ok {
		m.t.Fatalf("expected uint for util1 in GetEncoderUtilisation, got %T", res.out[0])
	}

	util2, ok := res.out[1].(uint)
	if !ok {
		m.t.Fatalf("expected uint for util2 in GetEncoderUtilisation, got %T", res.out[1])
	}

	return util1, util2, err
}

// GetDecoderUtilization is mock function.
func (m *MockDevice) GetDecoderUtilization() (uint, uint, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetDecoderUtilization")

	util1, ok := res.out[0].(uint)
	if !ok {
		m.t.Fatalf("expected uint for util1 in GetDecoderUtilisation, got %T", res.out[0])
	}

	util2, ok := res.out[1].(uint)
	if !ok {
		m.t.Fatalf("expected uint for util2 in GetDecoderUtilisation, got %T", res.out[1])
	}

	return util1, util2, err
}

// GetEccMode is mock function.
func (m *MockDevice) GetEccMode() (bool, bool, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetEccMode")

	current, ok := res.out[0].(bool)
	if !ok {
		m.t.Fatalf("expected bool in GetEccMode, got %T", res.out[0])
	}

	pending, ok := res.out[1].(bool)
	if !ok {
		m.t.Fatalf("expected bool in GetEccMode, got %T", res.out[1])
	}

	return current, pending, err
}

// GetPCIeThroughput is mock function.
func (m *MockDevice) GetPCIeThroughput(pcie nvml.PcieMetricType) (uint, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetPCIeThroughput", pcie)

	throughput, ok := res.out[0].(uint)
	if !ok {
		m.t.Fatalf("expected uint in GetPCIeThroughput, got %T", res.out[0])
	}

	return throughput, err
}

// GetMemoryInfoV2 is mock function.
func (m *MockDevice) GetMemoryInfoV2() (*nvml.MemoryInfoV2, error) {
	m.t.Helper()

	res, err := m.handleFunctionCall("GetMemoryInfoV2")
	if err != nil {
		return nil, err
	}

	info, ok := res.out[0].(*nvml.MemoryInfoV2)
	if !ok {
		m.t.Fatalf("expected %T in GetMemoryInfoV2, got %T", info, res.out[0])
	}

	return info, err
}

// GetMemoryInfo is mock function.
func (m *MockDevice) GetMemoryInfo() (*nvml.MemoryInfo, error) {
	m.t.Helper()

	res, err := m.handleFunctionCall("GetMemoryInfo")
	if err != nil {
		return nil, err
	}

	info, ok := res.out[0].(*nvml.MemoryInfo)
	if !ok {
		m.t.Fatalf("expected %T in GetMemoryInfo, got %T", info, res.out[0])
	}

	return info, err
}

// GetBAR1MemoryInfo is mock function.
func (m *MockDevice) GetBAR1MemoryInfo() (*nvml.MemoryInfo, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetBAR1MemoryInfo")

	info, ok := res.out[0].(*nvml.MemoryInfo)
	if !ok {
		m.t.Fatalf("expected %T in GetBAR1MemoryInfo, got %T", info, res.out[0])
	}

	return info, err
}

// GetEncoderStats is mock function.
func (m *MockDevice) GetEncoderStats() (uint, uint, uint, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetEncoderStats")

	sessions, ok := res.out[0].(uint)
	if !ok {
		m.t.Fatalf("expected %T in GetEncoderStats, got %T", sessions, res.out[0])
	}

	fps, ok := res.out[1].(uint)
	if !ok {
		m.t.Fatalf("expected %T in GetEncoderStats, got %T", fps, res.out[1])
	}

	latency, ok := res.out[2].(uint)
	if !ok {
		m.t.Fatalf("expected %T in GetEncoderStats, got %T", latency, res.out[2])
	}

	return sessions, fps, latency, err
}

// handleFunctionCall is handler for mock function calls.
// Takes in function name and any number of arguments function received.
func (m *MockDevice) handleFunctionCall(name string, receivedArgs ...any) (*Expectation, error) {
	m.t.Helper()

	if m.callIdx >= len(m.expectations) {
		m.t.Fatalf("no more calls expected but got call for %q", name)
	}

	expect := m.expectations[m.callIdx]
	m.callIdx++

	if expect.funcName != name {
		m.t.Fatalf("got call for %q while expected for %q", name, expect.funcName)
	}

	if receivedArgs == nil {
		receivedArgs = []any{}
	}

	// Compare expectedArgs and receivedArgs using cmp
	if diff := cmp.Diff(expect.args, receivedArgs); diff != "" {
		m.t.Fatalf(`arguments mismatch in %s call %d:\nexpected: %v\nreceived: %v\ndiff: %s`,
			name,
			m.callIdx,
			expect.args,
			receivedArgs,
			diff,
		)
	}

	return expect, expect.err
}
