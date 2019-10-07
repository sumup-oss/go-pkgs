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

import "github.com/sumup-oss/go-pkgs/logger"

var _ logger.Logger = (*NilLogger)(nil)

type NilLogger struct {
}

func (l *NilLogger) Debug(args ...interface{})                                   {}
func (l *NilLogger) Print(args ...interface{})                                   {}
func (l *NilLogger) Info(args ...interface{})                                    {}
func (l *NilLogger) Warn(args ...interface{})                                    {}
func (l *NilLogger) Warning(args ...interface{})                                 {}
func (l *NilLogger) Error(args ...interface{})                                   {}
func (l *NilLogger) Panic(args ...interface{})                                   {}
func (l *NilLogger) Fatal(args ...interface{})                                   {}
func (l *NilLogger) Debugf(format string, args ...interface{})                   {}
func (l *NilLogger) Printf(format string, args ...interface{})                   {}
func (l *NilLogger) Infof(format string, args ...interface{})                    {}
func (l *NilLogger) Warnf(format string, args ...interface{})                    {}
func (l *NilLogger) Warningf(format string, args ...interface{})                 {}
func (l *NilLogger) Errorf(format string, args ...interface{})                   {}
func (l *NilLogger) Panicf(format string, args ...interface{})                   {}
func (l *NilLogger) Fatalf(format string, args ...interface{})                   {}
func (l *NilLogger) Debugln(args ...interface{})                                 {}
func (l *NilLogger) Println(args ...interface{})                                 {}
func (l *NilLogger) Infoln(args ...interface{})                                  {}
func (l *NilLogger) Warnln(args ...interface{})                                  {}
func (l *NilLogger) Warningln(args ...interface{})                               {}
func (l *NilLogger) Errorln(args ...interface{})                                 {}
func (l *NilLogger) Panicln(args ...interface{})                                 {}
func (l *NilLogger) Fatalln(args ...interface{})                                 {}
func (l *NilLogger) Logf(level logger.Level, format string, args ...interface{}) {}
func (l *NilLogger) Log(level logger.Level, args ...interface{})                 {}
func (l *NilLogger) Logln(level logger.Level, args ...interface{})               {}
func (l *NilLogger) SetLevel(level logger.Level)                                 {}
func (l *NilLogger) GetLevel() logger.Level                                      { return logger.InfoLevel }
