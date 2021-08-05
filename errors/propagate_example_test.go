package errors_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/sumup-oss/go-pkgs/errors"
)

func ExamplePropagate() {
	fooError := errors.New("foo failed")

	// Creating error with stack trace without a wrapping string.
	barErr := errors.Propagate(fooError)

	curDir, _ := os.Getwd()

	output := fmt.Sprintf("%+v\n", barErr)

	// Clean base path for the Output test.
	fmt.Print(strings.ReplaceAll(output, curDir, "/path"))

	// Output:
	// foo failed:
	//     github.com/sumup-oss/go-pkgs/errors_test.ExamplePropagate
	//         /path/propagate_example_test.go:15
	//   - foo failed:
	//     github.com/sumup-oss/go-pkgs/errors_test.ExamplePropagate
	//         /path/propagate_example_test.go:12
}
