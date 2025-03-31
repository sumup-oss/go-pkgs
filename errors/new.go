// Copyright 2021 SumUp Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import "fmt"

// New creates a new error.
//
// It can be used for creating sentinel errors, or as a replacement of the standard
// fmt.Errorf calls.
//
//	NOTE: New does not support the standard %w error wrapping.
//	      Use Wrap, WrapError, Hide and HideError instead.
func New(format string, a ...interface{}) error {
	return &wrapError{
		Frame: Caller(1),
		err: &stringError{
			msg: fmt.Sprintf(format, a...),
		},
		cause: nil,
		hide:  false,
	}
}

// Propagate creates a new error from existing error, by wrapping it with stack trace information.
//
// Most of the time users should use Wrap, WrapError, Hide and HideError.
//
// But sometimes the error does not need wrapping, since it contains enough context.
// And wrapping will create a chain of two errors with the same description
// (example: "foo failed: foo failed"). All it is needed is a stack trace.
// Propagate does exactly that, it just adds a stack trace.
func Propagate(err error) error {
	if err == nil {
		return nil
	}

	// The err is already wrapped.
	wrapped, ok := err.(*wrapError)
	if ok {
		return &wrapError{
			Frame: Caller(1),
			err:   wrapped.err,
			cause: err,
			hide:  false,
		}
	}

	return &wrapError{
		Frame: Caller(1),
		err:   err,
		cause: nil,
		hide:  false,
	}
}

// stringError is the default error type used for wrapping errors and creating sentinel errors.
type stringError struct {
	msg string
}

// Error returns the error string representation.
func (e *stringError) Error() string {
	return e.msg
}
