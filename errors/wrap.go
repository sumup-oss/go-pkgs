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

import (
	stdErrors "errors"
	"fmt"
)

// Unwrap is just a wrapper of the standard Unwrap.
func Unwrap(err error) error {
	return stdErrors.Unwrap(err)
}

// UnwrapHidden works the same way as Unwrap, but instead of using method Unwrap, it
// expects method called UnwrapHidden.
//
// This "duplication" is required in order to support the error hiding feature.
// When an error is created with Hide and HideError functions, the Unwrap loop will stop
// to the current error and all older errors in the chain will not be discoverable with Unwrap.
// On the other hand they will continue to be discoverable with UnwrapHidden, which makes it
// possible to extract the full stacktrace (frame info) even if the error is hiding older errors.
func UnwrapHidden(err error) error {
	hidden, ok := err.(interface {
		UnwrapHidden() error
	})

	if ok {
		return hidden.UnwrapHidden()
	}

	u, ok := err.(interface {
		Unwrap() error
	})

	if ok {
		return u.Unwrap()
	}

	return nil
}

// Is is just a wrapper of the standard Is.
func Is(err, target error) bool {
	return stdErrors.Is(err, target)
}

// As is just a wrapper of the standard As.
func As(err error, target interface{}) bool {
	return stdErrors.As(err, target)
}

// Wrap creates a new error wrapping the err.
// If the err is nil it will return nil.
func Wrap(err error, format string, a ...interface{}) error {
	if err == nil {
		return nil
	}

	return &wrapError{
		err: &stringError{
			msg: fmt.Sprintf(format, a...),
		},
		cause: err,
		Frame: Caller(1),
	}
}

// WrapError creates a new error wrapping the err with the provided wrapper.
//
// This can be useful if you want to wrap an error with a sentinel error.
func WrapError(err error, wrapper error) error {
	if err == nil {
		return nil
	}

	return &wrapError{
		Frame: Caller(1),
		err:   wrapper,
		cause: err,
	}
}

// Hide creates a new error wrapping the err, but the err will not be discoverable with Unwrap calls.
//
// This means that Is and As will not be able to detect any older error in the chain.
//
// Nevertheless the wrapped err and all older errors in the chain are discoverable by using
// UnwrapHidden calls, meaning that the stacktrace (frame info) is accessible.
func Hide(err error, format string, a ...interface{}) error {
	if err == nil {
		return nil
	}

	return &wrapError{
		Frame: Caller(1),
		err: &stringError{
			msg: fmt.Sprintf(format, a...),
		},
		cause: err,
		hide:  true,
	}
}

// HideError creates a new error wrapping the err with provided wrapper, but the err will not
// be discoverable with Unwrap calls.
//
// This can be useful if you want to hide an error with sentinel error.
//
// This means that Is and As will not be able to detect any older error in the chain.
//
// Nevertheless the wrapped err and all older errors in the chain are discoverable by using
// UnwrapHidden calls, meaning that the stacktrace (frame info) is accessible.
func HideError(err error, wrapper error) error {
	if err == nil {
		return nil
	}

	return &wrapError{
		Frame: Caller(1),
		err:   wrapper,
		cause: err,
		hide:  true,
	}
}
