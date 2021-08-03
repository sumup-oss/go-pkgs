package errors_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/sumup-oss/go-pkgs/errors"
)

func Example_printf() {
	fooError := errors.New("foo")
	barError := errors.Hide(fooError, "bar")
	bazError := errors.Hide(barError, "baz")
	quxError := errors.Wrap(bazError, "qux")

	fmt.Println("String:")
	fmt.Printf("%s\n", quxError)
	fmt.Println()

	fmt.Println("Verbose:")
	fmt.Printf("%v\n", quxError)
	fmt.Println()

	// Only +v will go trough the Hide barriers and print the whole error chain.
	fmt.Println("Verbose with stacktrace:")

	curDir, _ := os.Getwd()

	output := fmt.Sprintf("%+v\n", quxError)

	// Clean base path for Output test.
	fmt.Print(strings.ReplaceAll(output, curDir, "/path"))

	// Output:
	// String:
	// qux
	//
	// Verbose:
	// qux: baz
	//
	// Verbose with stacktrace:
	// qux:
	//     github.com/sumup-oss/go-pkgs/errors_test.Example_printf
	//         /path/example_printf_test.go:15
	//   - baz:
	//     github.com/sumup-oss/go-pkgs/errors_test.Example_printf
	//         /path/example_printf_test.go:14
	//   - bar:
	//     github.com/sumup-oss/go-pkgs/errors_test.Example_printf
	//         /path/example_printf_test.go:13
	//   - foo:
	//     github.com/sumup-oss/go-pkgs/errors_test.Example_printf
	//         /path/example_printf_test.go:12
}
