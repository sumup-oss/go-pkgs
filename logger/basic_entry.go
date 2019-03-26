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
	"bytes"
	"time"
)

type BasicEntry struct {
	Time    time.Time
	Level   Level
	Message string
	Buffer  *bytes.Buffer
	Fields  map[string]interface{}
}

func (entry *BasicEntry) GetTime() time.Time {
	return entry.Time
}

func (entry *BasicEntry) GetLevel() Level {
	return entry.Level
}

func (entry *BasicEntry) GetMessage() string {
	return entry.Message
}

func (entry *BasicEntry) GetBuffer() *bytes.Buffer {
	return entry.Buffer
}

func (entry *BasicEntry) GetFields() map[string]interface{} {
	return entry.Fields
}
