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

import "context"

type TaskInterface interface {
	Run(ctx context.Context) error
}

// Task adapts a TaskInterface to TaskFunc required by the task.Group to run tasks.
func Task(t TaskInterface) TaskFunc {
	//nolint:gocritic
	return func(ctx context.Context) error {
		return t.Run(ctx)
	}
}
