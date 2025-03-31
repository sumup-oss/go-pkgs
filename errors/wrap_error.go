package errors

import "fmt"

// wrapError is the helper object for maintaining an error chain.
//
// It implements all the needed functionality, like Unwrap, UnwrapHidden, Is, As and Format method
// that is used for displaying the wrapped error info (including a stack trace).
type wrapError struct {
	Frame
	err   error
	cause error
	hide  bool
}

func (e *wrapError) Error() string {
	return e.err.Error()
}

func (e *wrapError) Unwrap() error {
	if e.hide {
		return nil
	}

	return e.cause
}

func (e *wrapError) UnwrapHidden() error {
	return e.cause
}

func (e *wrapError) Is(target error) bool {
	return Is(e.err, target)
}

func (e *wrapError) As(target interface{}) bool {
	return As(e.err, target)
}

// Format writes the wrapped error string representation.
//
// When the 'v' verb is used, it will print all the errors in the chain separated with ": ".
// When the '+v' verb is used, it will print all the errors in the chain along with call stack
// information.
func (e *wrapError) Format(s fmt.State, v rune) { //nolint:varnamelen
	switch err := e.err.(type) {
	case *wrapError:
		stringErr, ok := err.err.(*stringError)
		if ok {
			_, _ = s.Write([]byte(stringErr.Error()))
		} else {
			err.Format(s, v)
		}
	case fmt.Formatter:
		err.Format(s, v)
	default:
		_, _ = s.Write([]byte(e.err.Error()))
	}

	if v != 'v' {
		return
	}

	if !s.Flag('+') {
		next := Unwrap(e)
		if next != nil {
			_, _ = s.Write([]byte(": "))
			fmt.Fprintf(s, "%v", next)
		}

		return
	}

	_, _ = s.Write([]byte(":"))

	fn, file, line := e.Location()

	_, _ = s.Write([]byte("\n    " + fn))
	fmt.Fprintf(s, "\n        %s:%d", file, line)

	next := UnwrapHidden(e)
	if next != nil {
		fmt.Fprintf(s, "\n  - %+v", next)
	}
}
