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
