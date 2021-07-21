package errors_test

import (
	"fmt"

	"github.com/sumup-oss/go-pkgs/errors"
)

func ExampleUnwrap() {
	fooError := errors.New("foo")
	barError := errors.Hide(fooError, "bar") // Unwrap will stop here, foo won't be discoverable.
	bazError := errors.Wrap(barError, "baz")
	quxError := errors.Wrap(bazError, "qux")

	err := quxError
	for err != nil {
		fmt.Println(err.Error())

		err = errors.Unwrap(err)
	}

	// Output:
	// qux
	// baz
	// bar
}

func ExampleUnwrapHidden() {
	fooError := errors.New("foo")
	barError := errors.Hide(fooError, "bar") // UnwrapHidden will NOT stop here, foo will be discoverable.
	bazError := errors.Wrap(barError, "baz")
	quxError := errors.Wrap(bazError, "qux")

	err := quxError
	for err != nil {
		fmt.Println(err.Error())

		err = errors.UnwrapHidden(err)
	}

	// Output:
	// qux
	// baz
	// bar
	// foo
}

func ExampleIs() {
	fooError := errors.New("foo")
	barError := errors.Hide(fooError, "bar") // Since Unwrap will stop here, Is cannot find foo.
	bazError := errors.New("baz")
	quxError := errors.WrapError(barError, bazError)
	quuxError := errors.Wrap(quxError, "quux")

	err := quuxError

	if errors.Is(err, quuxError) {
		fmt.Println("quux found")
	}

	if errors.Is(err, quxError) {
		fmt.Println("qux found")
	}

	if errors.Is(err, bazError) {
		fmt.Println("baz found")
	}

	if errors.Is(err, barError) {
		fmt.Println("bar found")
	}

	if !errors.Is(err, fooError) {
		fmt.Println("foo NOT found")
	}

	// Output:
	// quux found
	// qux found
	// baz found
	// bar found
	// foo NOT found
}

func ExampleWrap() {
	fooError := errors.New("foo")
	barError := errors.Hide(fooError, "bar") // Unwrap will stop here, foo won't be discoverable
	bazError := errors.New("baz")
	quxError := errors.WrapError(barError, bazError)
	quuxError := errors.Wrap(quxError, "quux")

	err := quuxError

	for err != nil {
		fmt.Println(err.Error())

		err = errors.Unwrap(err)
	}

	// Output:
	// quux
	// baz
	// bar
}

func ExampleWrapError() {
	internalError := errors.New("internal")
	apiError := errors.New("api error")

	barError := errors.WrapError(internalError, apiError)
	bazError := errors.Wrap(barError, "baz")
	quxError := errors.Wrap(bazError, "qux")

	err := quxError
	for err != nil {
		fmt.Println(err.Error())

		err = errors.Unwrap(err)
	}

	// Output:
	// qux
	// baz
	// api error
	// internal
}

func ExampleHide() {
	fooError := errors.New("foo")
	barError := errors.Hide(fooError, "bar") // Unwrap will stop here, foo won't be discoverable
	bazError := errors.Wrap(barError, "baz")
	quxError := errors.Wrap(bazError, "qux")

	err := quxError
	for err != nil {
		fmt.Println(err.Error())

		err = errors.Unwrap(err)
	}

	// Output:
	// qux
	// baz
	// bar
}

func ExampleHideError() {
	internalError := errors.New("internal")
	apiError := errors.New("api error")

	// Unwrap will stop here, internalError won't be discoverable
	barError := errors.HideError(internalError, apiError)

	bazError := errors.Wrap(barError, "baz")
	quxError := errors.Wrap(bazError, "qux")

	err := quxError
	for err != nil {
		fmt.Println(err.Error())

		err = errors.Unwrap(err)
	}

	// Output:
	// qux
	// baz
	// api error
}
