package errors_test

import (
	"fmt"

	"github.com/sumup-oss/go-pkgs/errors"
)

type AsCustomError struct {
	msg  string
	code int
}

func NewAsCustomError(msg string, code int) error {
	return &AsCustomError{
		msg:  msg,
		code: code,
	}
}

func (e *AsCustomError) Error() string {
	return fmt.Sprintf("[%d] %s", e.code, e.msg)
}

func (e *AsCustomError) Message() string {
	return e.msg
}

func (e *AsCustomError) Code() int {
	return e.code
}

func ExampleAs() {
	fooError := NewAsCustomError("foo", 1)
	barError := errors.Wrap(fooError, "bar")
	bazError := errors.New("baz")
	quxError := errors.WrapError(barError, bazError)

	err := quxError

	var customErr *AsCustomError
	if errors.As(err, &customErr) {
		fmt.Printf("msg: %s, code: %d", customErr.Message(), customErr.Code())
	} else {
		fmt.Println("no customErr found")
	}

	// Output:
	// msg: foo, code: 1
}
