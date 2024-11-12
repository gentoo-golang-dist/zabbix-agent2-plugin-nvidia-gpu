package nvmlmock

type funcCall struct {
	expectations []*Expectation
	callIdx      int
}

type funcName string

type Expectation struct {
	funcName string
	args     []any
	out      []any
	err      error
}

type option func(*Expectation)

func NewExpectation(funcName string) *Expectation {
	exp := &Expectation{
		funcName: funcName,
		args:     []any{},
		out:      []any{},
		err:      nil,
	}

	return exp
}

func (e *Expectation) WithExpextedArgs(args ...any) *Expectation {
	e.args = args
	return e
}

func (e *Expectation) ProvideOutput(out ...any) *Expectation {
	e.out = out
	return e
}

func (e *Expectation) ProvideError(err error) *Expectation {
	e.err = err
	return e
}
