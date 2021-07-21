package errors_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/sumup-oss/go-pkgs/errors"
)

func ExampleCaller() {
	curDir, _ := os.Getwd()

	frame := errors.Caller(0)

	function, file, line := frame.Location()

	fmt.Println(function)
	fmt.Println(strings.ReplaceAll(file, curDir, "/path")) // Clean base path for Output test.
	fmt.Println(line)

	// Output:
	// github.com/sumup-oss/go-pkgs/errors_test.ExampleCaller
	// /path/frame_example_test.go
	// 14
}

func ExampleFrame_Location() {
	curDir, _ := os.Getwd()

	frame := errors.Caller(0)

	function, file, line := frame.Location()

	fmt.Println(function)
	fmt.Println(strings.ReplaceAll(file, curDir, "/path")) // Clean base path for output test.
	fmt.Println(line)

	// Output:
	// github.com/sumup-oss/go-pkgs/errors_test.ExampleFrame_Location
	// /path/frame_example_test.go
	// 31
}
