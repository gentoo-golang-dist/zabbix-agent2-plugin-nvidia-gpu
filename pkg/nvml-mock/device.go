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

type MockDevice struct {
	nvml.Device
	expectations []*Expectation
	callIdx      int
	t            *testing.T
}

func NewMockDevice(t *testing.T) *MockDevice {
	return &MockDevice{
		t:            t,
		expectations: []*Expectation{},
	}
}

func (m *MockDevice) ExpectCalls(expectations ...*Expectation) *MockDevice {
	m.expectations = expectations

	return m
}

func (m *MockDevice) handleFunctionCall(name funcName, receivedArgs ...any) (*Expectation, error) {
	if m.callIdx >= len(m.expectations) {
		m.t.Errorf("no more calls expected but got call for %q", name)
	}

	expect := m.expectations[m.callIdx]
	m.callIdx++

	if receivedArgs == nil {
		receivedArgs = []any{}
	}

	// Compare expectedArgs and receivedArgs using cmp
	if diff := cmp.Diff(expect.args, receivedArgs); diff != "" {
		m.t.Errorf("arguments mismatch in %s call %d:\nexpected: %v\nreceived: %v\ndiff: %s", name, m.callIdx, expect.args, receivedArgs, diff)
		return &Expectation{}, nil
	}

	return expect, expect.err
}

func (m *MockDevice) ExpectedCallsDone() bool {
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

func (m *MockDevice) GetUUID() (string, error) {
	res, err := m.handleFunctionCall("GetUUID")

	uuid, ok := res.out[0].(string)
	if !ok {
		m.t.Errorf("expected string in GetUUID, got %T", res.out[0])
		return "", nil
	}

	return uuid, err
}

func (m *MockDevice) GetName() (string, error) {
	res, err := m.handleFunctionCall("GetName")

	uuid, ok := res.out[0].(string)
	if !ok {
		m.t.Errorf("expected string in GetName, got %T", res.out[0])
		return "", nil
	}

	return uuid, err
}

func (m *MockDevice) GetTemperature() (int, error) {
	res, err := m.handleFunctionCall("GetTemperature")

	temperature, ok := res.out[0].(int)
	if !ok {
		m.t.Errorf("expected string in GetTemperature, got %T", res.out[0])
		return 0, nil
	}

	return temperature, err
}
