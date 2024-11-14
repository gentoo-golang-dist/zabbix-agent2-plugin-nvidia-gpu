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

type MockRunner struct {
	nvml.Runner
	expectations []*Expectation
	subMocks     []Mocker
	callIdx      int
	t            *testing.T
}

func NewMockRunner(t *testing.T) *MockRunner {
	return &MockRunner{
		t:            t,
		expectations: []*Expectation{},
	}
}

func (m *MockRunner) ExpectCalls(expectations ...*Expectation) *MockRunner {
	m.expectations = expectations

	return m
}

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

func (m *MockRunner) handleFunctionCall(name funcName, receivedArgs ...any) (*Expectation, error) {
	if m.callIdx >= len(m.expectations) {
		m.t.Fatalf("no more calls expected but got call for %q", name)
	}

	expect := m.expectations[m.callIdx]
	m.callIdx++

	if expect.funcName != string(name) {
		m.t.Fatalf("got call for %q while expected for %q", name, expect.funcName)
	}

	if receivedArgs == nil {
		receivedArgs = []any{}
	}

	// Compare expectedArgs and receivedArgs using cmp
	if diff := cmp.Diff(expect.args, receivedArgs); diff != "" {
		m.t.Fatalf("arguments mismatch in %s call %d:\nexpected: %v\nreceived: %v\ndiff: %s", name, m.callIdx, expect.args, receivedArgs, diff)
		return &Expectation{}, nil
	}

	return expect, expect.err
}

func (m *MockRunner) ExpectedCallsDone() bool {
	done := true

	expected := len(m.expectations)
	received := m.callIdx

	if expected > received {
		m.t.Errorf("received %d out of %d expected calls", received, expected)

		// TODO: PRINT WHAT CALLS ARE NOT DONE
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

func (m *MockRunner) Init() error {
	_, err := m.handleFunctionCall("Init")
	if err != nil {
		return err
	}

	return nil
}

func (m *MockRunner) InitV2() error {
	_, err := m.handleFunctionCall("InitV2")
	if err != nil {
		return err
	}

	return nil
}

func (m *MockRunner) GetDriverVersion() (string, error) {
	res, err := m.handleFunctionCall("GetDriverVersion")

	// Type assertion to ensure res.resultArgs[0] is a string
	version, ok := res.out[0].(string)
	if !ok {
		m.t.Errorf("expected string in GetDriverVersion, got %T", res.out[0])
		return "", nil
	}

	return version, err
}

func (m *MockRunner) GetNVMLVersion() (string, error) {
	res, err := m.handleFunctionCall("GetNVMLVersion")

	// Type assertion to ensure res.resultArgs[0] is a string
	version, ok := res.out[0].(string)
	if !ok {
		m.t.Errorf("expected string in GetNVMLVersion, got %T", res.out[0])
		return "", nil
	}

	return version, err
}

func (m *MockRunner) GetDeviceCountV2() (uint, error) {
	res, err := m.handleFunctionCall("GetDeviceCountV2")

	// Type assertion to ensure res.resultArgs[0] is a string
	count, ok := res.out[0].(uint)
	if !ok {
		m.t.Errorf("expected uint in GetDeviceCountV2, got %T", res.out[0])
		return 0, nil
	}

	return count, err
}

func (m *MockRunner) GetDeviceByIndexV2(index uint) (nvml.Device, error) {
	res, err := m.handleFunctionCall("GetDeviceByIndexV2", index)

	if res.out[0] == nil {
		return nil, err
	}

	device, ok := res.out[0].(*MockDevice)
	if !ok {
		m.t.Errorf("expected *MockRunner in GetDeviceByIndexV2, got %T", res.out[0])
		return nil, nil
	}

	device.t = m.t

	return device, err
}

func (m *MockRunner) GetDeviceByUUID(uuid string) (nvml.Device, error) {
	res, err := m.handleFunctionCall("GetDeviceByUUID", uuid)

	if res.out[0] == nil {
		return nil, err
	}

	device, ok := res.out[0].(*MockDevice)
	if !ok {
		m.t.Errorf("expected *MockRunner in GetDeviceByUUID, got %T", res.out[0])
		return nil, nil
	}

	device.t = m.t

	return device, err
}
