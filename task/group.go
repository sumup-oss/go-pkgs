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
	"errors"
	"sync"
	"time"
)

var (
	// ErrDeadline is returned when the group wait deadline is exceeded.
	ErrDeadline = errors.New("deadline exceeded")
)

// Group is used to wait for a group of tasks to finish.
//
// It will stop all the tasks on the first task failure, and the Wait() method will return only the
// first encountered error.
type Group struct {
	wg          sync.WaitGroup
	mu          sync.Mutex
	cancelCh    chan struct{}
	canceled    bool
	firstRunErr error
}

// NewGroup creates new task group instance.
func NewGroup() *Group {
	return &Group{
		cancelCh: make(chan struct{}),
	}
}

// Go runs tasks in the group.
//
// Every task is run in new goroutine.
// When a task returns an error, all the tasks in the group are canceled.
//
// Typically one should schedule tasks with the Group.Go() method and then wait for all of them to
// finish by using the Group.Wait() method.
func (g *Group) Go(tasks ...TaskFunc) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.canceled {
		return
	}

	for _, fn := range tasks {
		g.wg.Add(1)
		go func(fn TaskFunc) {
			defer g.wg.Done()

			err := fn(g.cancelCh)
			if err != nil {
				g.cancelWithError(err)
			}
		}(fn)
	}
}

// Wait until all tasks are stopped.
// Returns the first encountered error if any.
// If deadline is exceeded all tasks are canceled and the returned error is the deadline error.
// IsDeadlineError func can be used to check if the tasks were canceled due a deadline.
// Passing 0 for deadline means there will be no deadline, and Wait is blocked until all of the
// tasks are finished.
func (g *Group) Wait(deadline time.Duration) error {
	if deadline > 0 {
		timer := time.NewTimer(deadline)
		defer timer.Stop()

		go func() {
			select {
			case <-g.cancelCh:
				return
			case <-timer.C:
				g.cancelWithError(ErrDeadline)
			}
		}()
	}

	g.wg.Wait()
	return g.firstRunErr
}

func (g *Group) unsafeCancel() {
	if g.canceled {
		return
	}

	g.canceled = true
	close(g.cancelCh)
}

func (g *Group) cancelWithError(err error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// NOTE: only the first error is retained.
	if g.firstRunErr == nil {
		g.firstRunErr = err
	}

	g.unsafeCancel()
}

// Cancel cancels all the tasks.
func (g *Group) Cancel() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.unsafeCancel()
}

func IsDeadlineError(err error) bool {
	return err == ErrDeadline
}
