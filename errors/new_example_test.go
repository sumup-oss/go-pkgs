package errors_test

import (
	"fmt"

	"github.com/sumup-oss/go-pkgs/errors"
)

var (
	NotFoundErr = errors.New("not found error") // Sentinel error created with New.
)

func ExampleNew() {
	fooErr := errors.Wrap(NotFoundErr, "foo failed")

	fmt.Printf("%v\n", fooErr)

	internalError := errors.New("internal failure code=%d", 500)
	barErr := errors.HideError(internalError, NotFoundErr)

	fmt.Printf("%v\n", barErr)

	// Output:
	// foo failed: not found error
	// not found error
}
