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

// TaskFunc specify the functions that are run by the task group.
type TaskFunc func(cancel <-chan struct{}) error

type TaskFuncDecorator func(fn TaskFunc) TaskFunc

// NewTaskFunc creates a TaskFunc decorated with the provided decorators.
func NewTaskFunc(fn TaskFunc, decorators ...TaskFuncDecorator) TaskFunc {
	for i := len(decorators) - 1; i >= 0; i-- {
		fn = decorators[i](fn)
	}

	return fn
}
