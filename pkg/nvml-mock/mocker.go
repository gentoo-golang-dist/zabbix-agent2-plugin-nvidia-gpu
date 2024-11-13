package nvmlmock

type Mocker interface {
	ExpectedCallsDone() bool
	SubMocks() []Mocker
}
