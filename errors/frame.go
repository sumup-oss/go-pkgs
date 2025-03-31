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
	"runtime"
)

// Frame is a frame from the call stack, that can report the call location.
type Frame struct {
	pc [1]uintptr
}

// Caller returns a Frame from the call stack.
//
// The argument skip is the number of stack frames
// to ascend, with 0 identifying the caller of Caller.
func Caller(skip int) Frame {
	var frame Frame

	runtime.Callers(skip+2, frame.pc[:]) //nolint:mnd

	return frame
}

// Location reports the file, line, and function of a frame.
func (f *Frame) Location() (string, string, int) {
	frame, _ := runtime.CallersFrames(f.pc[:]).Next()

	return frame.Function, frame.File, frame.Line
}
