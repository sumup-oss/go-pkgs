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
func (e *wrapError) Format(s fmt.State, v rune) {
	switch err := e.err.(type) {
	case *wrapError:
		stringErr, ok := err.err.(*stringError)
		if ok {
			s.Write([]byte(stringErr.Error()))
		} else {
			err.Format(s, v)
		}
	case fmt.Formatter:
		err.Format(s, v)
	default:
		s.Write([]byte(e.err.Error()))
	}

	if v != 'v' {
		return
	}

	if !s.Flag('+') {
		next := Unwrap(e)
		if next != nil {
			s.Write([]byte(": "))
			s.Write([]byte(fmt.Sprintf("%v", next)))
		}

		return
	}

	s.Write([]byte(":"))

	fn, file, line := e.Location()

	s.Write([]byte("\n    " + fn))
	s.Write([]byte(fmt.Sprintf("\n        %s:%d", file, line)))

	next := UnwrapHidden(e)
	if next != nil {
		s.Write([]byte(fmt.Sprintf("\n  - %+v", next)))
	}
}
