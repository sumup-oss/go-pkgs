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

package task_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sumup/go-pkgs/task"
)

func TestIsRetryableError(t *testing.T) {
	t.Run("when the error is RetryableError, it returns true", func(t *testing.T) {
		t.Parallel()

		err := task.NewRetryableError(assert.AnError)
		assert.True(t, task.IsRetryableError(err))
	})

	t.Run("when the error is nil, it returns false", func(t *testing.T) {
		t.Parallel()

		assert.False(t, task.IsRetryableError(nil))
	})

	t.Run("when the error is not RetryableError, it returns false", func(t *testing.T) {
		t.Parallel()

		assert.False(t, task.IsRetryableError(assert.AnError))
	})
}

func TestRetry(t *testing.T) {
	t.Run("it retries until the task func returns no error", func(t *testing.T) {
		t.Parallel()

		cnt := 0
		pace := make(chan error, 10)
		cancel := make(chan struct{})
		repeat := task.Retry(1, func(cancel <-chan struct{}) error {
			cnt++
			return <-pace
		})

		pace <- task.NewRetryableError(errors.New("fooErr"))
		pace <- task.NewRetryableError(errors.New("barErr"))
		pace <- nil
		pace <- task.NewRetryableError(errors.New("bazErr"))

		err := repeat(cancel)

		assert.NoError(t, err)
		assert.Equal(t, 3, cnt)
	})

	t.Run("it retries until the task func returns non retriable error", func(t *testing.T) {
		t.Parallel()

		cnt := 0
		pace := make(chan error, 10)
		cancel := make(chan struct{})
		repeat := task.Retry(1, func(cancel <-chan struct{}) error {
			cnt++
			return <-pace
		})

		pace <- task.NewRetryableError(errors.New("fooErr"))
		pace <- task.NewRetryableError(errors.New("barErr"))
		pace <- errors.New("bazErr")

		err := repeat(cancel)

		assert.EqualError(t, err, "bazErr")
		assert.Equal(t, 3, cnt)
	})

	t.Run("it cancels the sleeping when cancel signal is fired", func(t *testing.T) {
		t.Parallel()

		cnt := 0
		pace := make(chan error, 10)
		cancel := make(chan struct{})
		repeat := task.Retry(time.Hour, func(cancel <-chan struct{}) error {
			cnt++
			return <-pace
		})

		pace <- task.NewRetryableError(errors.New("fooErr"))

		go func() {
			close(cancel)
		}()

		err := repeat(cancel)
		assert.NoError(t, err)
		assert.Equal(t, 1, cnt)
	})
}

func TestRetryUntil(t *testing.T) {
	t.Run(
		"when the task func returns no error, it stops retrying and returns no error",
		func(t *testing.T) {
			t.Parallel()

			cnt := 0
			pace := make(chan error, 10)
			cancel := make(chan struct{})
			repeat := task.RetryUntil(10, 1, func(cancel <-chan struct{}) error {
				cnt++
				return <-pace
			})

			pace <- task.NewRetryableError(errors.New("fooErr"))
			pace <- task.NewRetryableError(errors.New("barErr"))
			pace <- nil
			pace <- task.NewRetryableError(errors.New("bazErr"))

			err := repeat(cancel)

			assert.NoError(t, err)
			assert.Equal(t, 3, cnt)
		},
	)

	t.Run(
		"when the task func returns non retriable error, it stops retrying and returns the last error",
		func(t *testing.T) {
			t.Parallel()

			cnt := 0
			pace := make(chan error, 10)
			cancel := make(chan struct{})
			repeat := task.RetryUntil(10, 1, func(cancel <-chan struct{}) error {
				cnt++
				return <-pace
			})

			pace <- task.NewRetryableError(errors.New("fooErr"))
			pace <- task.NewRetryableError(errors.New("barErr"))
			pace <- errors.New("bazErr")

			err := repeat(cancel)

			assert.EqualError(t, err, "bazErr")
			assert.Equal(t, 3, cnt)
		},
	)

	t.Run(
		"when the task func returns retriable error maxAttempts in a row, "+
			"it stops retrying and returns MaxRetryExceedError",
		func(t *testing.T) {
			t.Parallel()

			cnt := 0
			pace := make(chan error, 10)
			cancel := make(chan struct{})
			repeat := task.RetryUntil(2, 1, func(cancel <-chan struct{}) error {
				cnt++
				return <-pace
			})

			pace <- task.NewRetryableError(errors.New("fooErr"))
			pace <- task.NewRetryableError(errors.New("barErr"))
			pace <- errors.New("bazErr")

			err := repeat(cancel)

			require.IsType(t, (*task.MaxRetryExceedError)(nil), err)
			assert.EqualError(t, err.(*task.MaxRetryExceedError).Cause(), "barErr")
			assert.EqualError(t, err, "max retry attempts 2 exceeded, last err: barErr")
			assert.Equal(t, 2, cnt)
		},
	)

	t.Run(
		"when cancel signal is fired, it cancels the sleeping between retries",
		func(t *testing.T) {
			t.Parallel()

			cnt := 0
			pace := make(chan error, 10)
			cancel := make(chan struct{})
			repeat := task.RetryUntil(10, time.Hour, func(cancel <-chan struct{}) error {
				cnt++
				return <-pace
			})

			pace <- task.NewRetryableError(errors.New("fooErr"))

			go func() {
				close(cancel)
			}()

			err := repeat(cancel)
			assert.NoError(t, err)
			assert.Equal(t, 1, cnt)
		},
	)
}

func TestRetryWithDeadline(t *testing.T) {
	t.Run(
		"when the task func returns no error, it stops retrying and returns no error",
		func(t *testing.T) {
			t.Parallel()

			cnt := 0
			pace := make(chan error, 10)
			cancel := make(chan struct{})
			repeat := task.RetryWithDeadline(time.Hour, 1, func(cancel <-chan struct{}) error {
				cnt++
				return <-pace
			})

			pace <- task.NewRetryableError(errors.New("fooErr"))
			pace <- task.NewRetryableError(errors.New("barErr"))
			pace <- nil
			pace <- task.NewRetryableError(errors.New("bazErr"))

			err := repeat(cancel)

			assert.NoError(t, err)
			assert.Equal(t, 3, cnt)
		},
	)

	t.Run(
		"when the task func returns non retriable error, it stops retrying and returns the last error",
		func(t *testing.T) {
			t.Parallel()

			cnt := 0
			pace := make(chan error, 10)
			cancel := make(chan struct{})
			repeat := task.RetryWithDeadline(time.Hour, 1, func(cancel <-chan struct{}) error {
				cnt++
				return <-pace
			})

			pace <- task.NewRetryableError(errors.New("fooErr"))
			pace <- task.NewRetryableError(errors.New("barErr"))
			pace <- errors.New("bazErr")

			err := repeat(cancel)

			assert.EqualError(t, err, "bazErr")
			assert.Equal(t, 3, cnt)
		},
	)

	t.Run(
		"when the task func returns only retriable errors and the deadline is reached, "+
			"it stops retrying and returns DeadlineRetryError",
		func(t *testing.T) {
			t.Parallel()

			cnt := 0
			pace := make(chan error, 10)
			cancel := make(chan struct{})
			repeat := task.RetryWithDeadline(1, time.Hour, func(cancel <-chan struct{}) error {
				cnt++
				return <-pace
			})

			pace <- task.NewRetryableError(errors.New("fooErr"))

			err := repeat(cancel)

			require.IsType(t, (*task.DeadlineRetryError)(nil), err)
			assert.EqualError(t, err.(*task.DeadlineRetryError).Cause(), "fooErr")
			assert.EqualError(t, err, "deadline 1ns exceeded, last err: fooErr")
			assert.Equal(t, 1, cnt)
		},
	)

	t.Run(
		"when the task func does not complete within the deadline, "+
			"it cancels the task and returns DeadlineRetryError",
		func(t *testing.T) {
			t.Parallel()

			cancel := make(chan struct{})
			repeat := task.RetryWithDeadline(1, time.Hour, func(cancel <-chan struct{}) error {
				<-cancel
				return nil
			})

			err := repeat(cancel)

			require.IsType(t, (*task.DeadlineRetryError)(nil), err)
			assert.NoError(t, err.(*task.DeadlineRetryError).Cause())
			assert.EqualError(t, err, "deadline 1ns exceeded, last err: <nil>")
		},
	)

	t.Run(
		"when cancel signal is fired, it cancels the sleeping between retries",
		func(t *testing.T) {
			t.Parallel()

			cnt := 0
			pace := make(chan error, 10)
			cancel := make(chan struct{})
			repeat := task.RetryWithDeadline(time.Hour, time.Hour, func(cancel <-chan struct{}) error {
				cnt++
				return <-pace
			})

			pace <- task.NewRetryableError(errors.New("fooErr"))

			go func() {
				close(cancel)
			}()

			err := repeat(cancel)
			assert.NoError(t, err)
			assert.Equal(t, 1, cnt)
		},
	)
}
