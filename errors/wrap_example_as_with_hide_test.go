package errors_test

import (
	"fmt"

	"github.com/sumup-oss/go-pkgs/errors"
)

type AsWithHideCustomError struct {
	msg  string
	code int
}

func NewAsWithHideCustomError(msg string, code int) error {
	return &AsWithHideCustomError{
		msg:  msg,
		code: code,
	}
}

func (e *AsWithHideCustomError) Error() string {
	return fmt.Sprintf("[%d] %s", e.code, e.msg)
}

func (e *AsWithHideCustomError) Message() string {
	return e.msg
}

func (e *AsWithHideCustomError) Code() int {
	return e.code
}

func ExampleAs_withHide() {
	fooError := NewAsWithHideCustomError("foo", 1)
	barError := errors.Hide(fooError, "bar") // Since Unwrap will stop here, As cannot find foo.
	bazError := errors.New("baz")
	quxError := errors.WrapError(barError, bazError)

	err := quxError

	var customErr *AsWithHideCustomError
	if errors.As(err, &customErr) {
		fmt.Printf("msg: %s, code: %d", customErr.Message(), customErr.Code())
	} else {
		fmt.Println("no customErr found")
	}

	// Output:
	// no customErr found
}
