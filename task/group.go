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
	"context"
	"sync"
)

// Group is used to wait for a group of tasks to finish.
//
// It will stop all the tasks on the first task failure, and the Wait() method will return only the
// first encountered error.
type Group struct {
	wg         sync.WaitGroup
	ctx        context.Context
	cancelFunc context.CancelFunc

	// mu protects the firstRunErr
	mu          sync.Mutex
	firstRunErr error
}

// NewGroup creates new task group instance.
func NewGroup() *Group {
	ctx, cancel := context.WithCancel(context.Background())

	return &Group{
		ctx:        ctx,
		cancelFunc: cancel,
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
	if g.ctx.Err() != nil {
		return
	}

	for _, fn := range tasks {
		g.wg.Add(1)
		go func(fn TaskFunc) {
			defer g.wg.Done()

			err := fn(g.ctx)
			if err != nil {
				g.cancelWithError(err)
			}
		}(fn)
	}
}

// Wait until all tasks are stopped.
// Returns the first encountered error if any.
// If the context is done all tasks are canceled and the context error is returned.
func (g *Group) Wait(ctx context.Context) error {
	if ctx != context.TODO() {
		go func() {
			select {
			case <-g.ctx.Done():
				return
			case <-ctx.Done():
				g.cancelWithError(ctx.Err())
			}
		}()
	}

	g.wg.Wait()

	return g.firstRunErr
}

func (g *Group) cancelWithError(err error) {
	g.mu.Lock()

	// NOTE: only the first error is retained.
	if g.firstRunErr == nil {
		g.firstRunErr = err
	}

	g.mu.Unlock()

	g.cancelFunc()
}

// Cancel cancels all the tasks.
func (g *Group) Cancel() {
	g.cancelFunc()
}
