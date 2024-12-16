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

package nvmlmock

// Expectation contains expectations of mocks.
type Expectation struct {
	funcName string
	args     []any
	out      []any
	err      error
}

// NewExpectation creates new expectations for mock.
func NewExpectation(funcName string) *Expectation {
	exp := &Expectation{
		funcName: funcName,
		args:     []any{},
		out:      []any{},
		err:      nil,
	}

	return exp
}

// WithExpextedArgs should be called to provide arguments mock should receive.
func (e *Expectation) WithExpextedArgs(args ...any) *Expectation {
	e.args = args

	return e
}

// ProvideOutput should be called to provide outputs mock should return.
func (e *Expectation) ProvideOutput(out ...any) *Expectation {
	e.out = out

	return e
}

// ProvideError should be called to provide error mock should return.
func (e *Expectation) ProvideError(err error) *Expectation {
	e.err = err

	return e
}
