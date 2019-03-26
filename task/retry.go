// Copyright 2019 SumUp Ltd.
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

package task

import (
	"fmt"
	"time"
)

// RetryableError error signify that the task can be retried.
type RetryableError interface {
	error
	IsRetryable() bool
}

// IsRetryableError checks if the error is retryable.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	retryableErr, ok := err.(RetryableError)
	return ok && retryableErr.IsRetryable()
}

// retryableError implements the RetryableError interface.
type retryableError struct {
	error
}

// IsRetryable verify that the error is in fact retryable.
func (err *retryableError) IsRetryable() bool {
	return true
}

// NewRetryableError wraps an error and makes it retryable.
func NewRetryableError(err error) error {
	return &retryableError{err}
}

// Retry retries a task until it returns no error or the returned error is non retriable.
// An error is retriable when it implements the RetryableError interface and its IsRetryable method
// returns true.
// The retryInterval specify how much time to wait between every retry.
func Retry(retryInterval time.Duration, retryFunc TaskFunc) TaskFunc {
	return func(cancel <-chan struct{}) error {
		for {
			err := retryFunc(cancel)
			if !IsRetryableError(err) {
				return err
			}

			retryTimer := time.NewTimer(retryInterval)
			select {
			case <-cancel:
				retryTimer.Stop()
				return nil
			case <-retryTimer.C:
			}
		}
	}
}

// MaxRetryExceedError is returned when RetryUntil could not complete successfully
// for a given maxAttempts retries.
type MaxRetryExceedError struct {
	maxAttempts int
	lastErr     error
}

// NewMaxRetryError creates MaxRetryError instance.
func NewMaxRetryError(maxAttempts int, lastErr error) error {
	return &MaxRetryExceedError{
		maxAttempts: maxAttempts,
		lastErr:     lastErr,
	}
}

// Error returns the error message.
func (err *MaxRetryExceedError) Error() string {
	return fmt.Sprintf("max retry attempts %d exceeded, last err: %v", err.maxAttempts, err.lastErr)
}

// Cause returns the last error before the max retry attempts is exceeded.
func (err *MaxRetryExceedError) Cause() error {
	return err.lastErr
}

// RetryUntil retries a task maxAttempts times until it returns no error or the returned error is non retriable.
// An error is retriable when it implements the RetryableError interface and its IsRetryable method
// returns true.
// The retryInterval specify how much time to wait between every retry.
// If the task do not complete for maxAttempts retries, RetryUntil will return MaxRetryExceedError.
//
// NOTE: when the cancel channel is closed, RetryUntil will not return an error, even if
// the retryFunc had failed couple of times so far.
func RetryUntil(maxAttempts int, retryInterval time.Duration, retryFunc TaskFunc) TaskFunc {
	return func(cancel <-chan struct{}) error {
		var err error
		attempts := maxAttempts

		for {
			if attempts < 1 {
				return NewMaxRetryError(maxAttempts, err)
			}

			err = retryFunc(cancel)
			if !IsRetryableError(err) {
				return err
			}

			attempts -= 1

			retryTimer := time.NewTimer(retryInterval)
			select {
			case <-cancel:
				retryTimer.Stop()
				return nil
			case <-retryTimer.C:
			}
		}
	}
}

// DeadlineRetryError is returned when RetryWithDeadline could not complete successfully
// within given deadline.
type DeadlineRetryError struct {
	deadline time.Duration
	lastErr  error
}

// NewDeadlineRetryError creates DeadlineRetryError instance.
func NewDeadlineError(deadline time.Duration, lastErr error) error {
	return &DeadlineRetryError{
		deadline: deadline,
		lastErr:  lastErr,
	}
}

// Error returns the error message.
func (err *DeadlineRetryError) Error() string {
	return fmt.Sprintf("deadline %v exceeded, last err: %v", err.deadline, err.lastErr)
}

// Cause returns the last error before the deadline is exceeded.
func (err *DeadlineRetryError) Cause() error {
	return err.lastErr
}

// RetryWithDeadline retries a task until it returns no error, or the returned error is non retriable,
// or timeoutDeadline is exceeded.
// An error is retriable when it implements the RetryableError interface and its IsRetryable method
// returns true.
// The retryInterval specify how much time to wait between every retry.
// If the task do not complete and timeoutDeadline is exceeded, RetryWithDeadline
// will return DeadlineRetryError.
//
// NOTE: when the cancel channel is closed, RetryWithDeadline will not return an error, even if
// the retryFunc had failed couple of times so far.
func RetryWithDeadline(
	timeoutDeadline time.Duration,
	retryInterval time.Duration,
	retryFunc TaskFunc,
) TaskFunc {
	return func(cancel <-chan struct{}) error {
		// when the task is complete or fails with non retryable error, doneChan is closed
		// for signaling the deadline monitoring to stop
		doneChan := make(chan struct{})
		// when the external cancel signal is triggered or the deadline is exceeded taskCancelChan
		// is closed for signaling the task and retryTimer to cancel
		taskCancelChan := make(chan struct{})

		var err error
		var lastErr error // last non-nil error
		go func() {
			defer close(doneChan)

			for {
				err = retryFunc(taskCancelChan)
				if err != nil {
					lastErr = err
				}
				if !IsRetryableError(err) {
					return
				}

				retryTimer := time.NewTimer(retryInterval)
				select {
				case <-taskCancelChan:
					retryTimer.Stop()
					err = nil
					return
				case <-retryTimer.C:
				}
			}
		}()

		deadlineTimer := time.NewTimer(timeoutDeadline)
		defer deadlineTimer.Stop()

		select {
		case <-doneChan:
			return err
		case <-cancel:
			close(taskCancelChan)
			<-doneChan
			return err
		case <-deadlineTimer.C:
			close(taskCancelChan)
			<-doneChan
			return NewDeadlineError(timeoutDeadline, lastErr)
		}
	}
}
