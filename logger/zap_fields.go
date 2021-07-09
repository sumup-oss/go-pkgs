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

package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ErrorField creates a zap.Field for the corresponding error.
//
// The sole purpose of this func is to provide extensibility, on the way errors logging is handled.
// Currently it does not do anything different than zap.Error().
//
// Future version should wrap the error with an object that supports Errors interface:
//   type Errors interface {
//     Errors() []error
//   }
func ErrorField(err error) zap.Field {
	if err == nil {
		return zap.Skip()
	}

	return zap.Field{
		Key:       "error",
		Type:      zapcore.ErrorType,
		Interface: err,
	}
}
