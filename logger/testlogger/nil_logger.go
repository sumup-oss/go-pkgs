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

package testlogger

import "github.com/sumup/go-pkgs/logger"

var _ logger.Logger = (*NilLogger)(nil)

type NilLogger struct {
}

func (logger *NilLogger) Debug(args ...interface{})                                   {}
func (logger *NilLogger) Print(args ...interface{})                                   {}
func (logger *NilLogger) Info(args ...interface{})                                    {}
func (logger *NilLogger) Warn(args ...interface{})                                    {}
func (logger *NilLogger) Warning(args ...interface{})                                 {}
func (logger *NilLogger) Error(args ...interface{})                                   {}
func (logger *NilLogger) Panic(args ...interface{})                                   {}
func (logger *NilLogger) Fatal(args ...interface{})                                   {}
func (logger *NilLogger) Debugf(format string, args ...interface{})                   {}
func (logger *NilLogger) Printf(format string, args ...interface{})                   {}
func (logger *NilLogger) Infof(format string, args ...interface{})                    {}
func (logger *NilLogger) Warnf(format string, args ...interface{})                    {}
func (logger *NilLogger) Warningf(format string, args ...interface{})                 {}
func (logger *NilLogger) Errorf(format string, args ...interface{})                   {}
func (logger *NilLogger) Panicf(format string, args ...interface{})                   {}
func (logger *NilLogger) Fatalf(format string, args ...interface{})                   {}
func (logger *NilLogger) Debugln(args ...interface{})                                 {}
func (logger *NilLogger) Println(args ...interface{})                                 {}
func (logger *NilLogger) Infoln(args ...interface{})                                  {}
func (logger *NilLogger) Warnln(args ...interface{})                                  {}
func (logger *NilLogger) Warningln(args ...interface{})                               {}
func (logger *NilLogger) Errorln(args ...interface{})                                 {}
func (logger *NilLogger) Panicln(args ...interface{})                                 {}
func (logger *NilLogger) Fatalln(args ...interface{})                                 {}
func (logger *NilLogger) Logf(level logger.Level, format string, args ...interface{}) {}
func (logger *NilLogger) Log(level logger.Level, args ...interface{})                 {}
func (logger *NilLogger) Logln(level logger.Level, args ...interface{})               {}
func (logger *NilLogger) SetLevel(level logger.Level)                                 {}
