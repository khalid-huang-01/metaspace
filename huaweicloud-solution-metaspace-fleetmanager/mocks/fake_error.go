package mocks

type FakeError struct {
	ErrorMsg string
}

func (e *FakeError) Error() string {
	return e.ErrorMsg
}
