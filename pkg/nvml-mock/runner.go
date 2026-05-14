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
	_ nvml.Runner = (*MockRunner)(nil)
	_ Mocker      = (*MockRunner)(nil)
)

// Mocker any mock should implement to collect information if all its and submock calls are done.
type Mocker interface {
	ExpectedCallsDone() bool
	SubMocks() []Mocker
}

// MockRunner is mock for NVML Runner.
type MockRunner struct {
	nvml.Runner

	expectations []*Expectation
	callIdx      int
	t            *testing.T
}

// NewMockRunner creates new mock runner.
func NewMockRunner(t *testing.T) *MockRunner {
	t.Helper()

	return &MockRunner{
		t:            t,
		expectations: []*Expectation{},
	}
}

// ExpectCalls sets calls that are expected by mock.
func (m *MockRunner) ExpectCalls(expectations ...*Expectation) *MockRunner {
	m.expectations = expectations

	return m
}

// SubMocks returns submocks of the mock.
func (m *MockRunner) SubMocks() []Mocker {
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

// ExpectedCallsDone checks if all expected calls of mock and it's submocks are done.
func (m *MockRunner) ExpectedCallsDone() bool {
	done := true

	expected := len(m.expectations)
	received := m.callIdx

	if expected > received {
		m.t.Errorf("received %d out of %d expected calls", received, expected)

		for _, e := range m.expectations[received:] {
			m.t.Errorf("Not called %q", e.funcName)
		}

		done = false
	}

	for _, sub := range m.SubMocks() {
		if !sub.ExpectedCallsDone() {
			done = false
		}
	}

	return done
}

// Init is mock function.
func (m *MockRunner) Init() error {
	m.t.Helper()
	_, err := m.handleFunctionCall("Init")

	return err
}

// InitV2 is mock function.
func (m *MockRunner) InitV2() error {
	m.t.Helper()
	_, err := m.handleFunctionCall("InitV2")

	return err
}

// ShutdownNVML is mock function.
func (m *MockRunner) ShutdownNVML() error {
	m.t.Helper()
	_, err := m.handleFunctionCall("ShutdownNVML")

	return err
}

// GetDriverVersion is mock function.
func (m *MockRunner) GetDriverVersion() (string, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetDriverVersion")

	// Type assertion to ensure res.resultArgs[0] is a string
	version, ok := res.out[0].(string)
	if !ok {
		m.t.Fatalf("expected string in GetDriverVersion, got %T", res.out[0])
	}

	return version, err
}

// GetNVMLVersion is mock function.
func (m *MockRunner) GetNVMLVersion() (string, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetNVMLVersion")

	version, ok := res.out[0].(string)
	if !ok {
		m.t.Fatalf("expected string in GetNVMLVersion, got %T", res.out[0])
	}

	return version, err
}

// GetDeviceCountV2 is mock function.
func (m *MockRunner) GetDeviceCountV2() (uint, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetDeviceCountV2")

	count, ok := res.out[0].(uint)
	if !ok {
		m.t.Fatalf("expected uint in GetDeviceCountV2, got %T", res.out[0])
	}

	return count, err
}

// GetDeviceByIndexV2 is mock function.
//
//nolint:ireturn,nolintlint
func (m *MockRunner) GetDeviceByIndexV2(index uint) (nvml.Device, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetDeviceByIndexV2", index)

	if res.out[0] == nil {
		return nil, err
	}

	device, ok := res.out[0].(*MockDevice)
	if !ok {
		m.t.Fatalf("expected *MockRunner in GetDeviceByIndexV2, got %T", res.out[0])
	}

	device.t = m.t

	return device, err
}

// GetDeviceByUUID is mock function.
//
//nolint:ireturn,nolintlint
func (m *MockRunner) GetDeviceByUUID(uuid string) (nvml.Device, error) {
	m.t.Helper()
	res, err := m.handleFunctionCall("GetDeviceByUUID", uuid)

	if res.out[0] == nil {
		return nil, err
	}

	mockDevice, ok := res.out[0].(*MockDevice)
	if ok {
		return mockDevice, err
	}

	device, ok := res.out[0].(nvml.Device)
	if !ok {
		m.t.Fatalf("expected %T in GetDeviceByUUID, got %T", device, res.out[0])
	}

	return device, err
}

// Close is mock function.
func (m *MockRunner) Close() error {
	m.t.Helper()
	_, err := m.handleFunctionCall("Close")

	return err
}

// handleFunctionCall is handler for mock function calls.
// Takes in function name and any number of arguments function received.
func (m *MockRunner) handleFunctionCall(name string, receivedArgs ...any) (*Expectation, error) {
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
		m.t.Fatalf("arguments mismatch in %s call %d:\nexpected: %v\nreceived: %v\ndiff: %s",
			name,
			m.callIdx,
			expect.args,
			receivedArgs,
			diff,
		)
	}

	return expect, expect.err
}
