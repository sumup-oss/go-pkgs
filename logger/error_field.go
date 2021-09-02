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

package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/sumup-oss/go-pkgs/errors"
)

// ErrorField creates a zap.Field for the corresponding error.
//
// By default it creates zapcore.ErrorType field the same way zap.Error() does it.
//
// If the error implements the Location interface from github.com/sumup-oss/go-pkgs/errors,
// it will add a `trace` field in the log with the error stack trace.
//
// The Location interface looks like this:
//   interface {
//		Location() (function, file string, line int)
//	}
func ErrorField(err error) zap.Field {
	if err == nil {
		return zap.Skip()
	}

	_, ok := err.(interface {
		Location() (function, file string, line int)
	})

	if ok {
		return zap.Field{
			Type: zapcore.InlineMarshalerType,
			Interface: &errorObjectField{
				err: err,
			},
		}
	}

	return zap.Field{
		Key:       "error",
		Type:      zapcore.ErrorType,
		Interface: err,
	}
}

type errorStacktrace struct {
	err error
}

func (e *errorStacktrace) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("err", fmt.Sprintf("%s", e.err))

	locator, ok := e.err.(interface {
		Location() (function, file string, line int)
	})

	if ok {
		_, file, line := locator.Location()
		enc.AddString("loc", fmt.Sprintf("%s:%d", file, line))
	}

	return nil
}

type errorObjectField struct {
	err error
}

func (e *errorObjectField) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("error", fmt.Sprintf("%s", e.err))
	enc.AddArray("trace", e)

	return nil
}

func (e *errorObjectField) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	err := e.err
	for err != nil {
		enc.AppendObject(&errorStacktrace{err: err})
		err = errors.UnwrapHidden(err)
	}

	return nil
}
