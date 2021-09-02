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
	"net/http"
	"net/http/httputil"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type httpRequestDumpField struct {
	req      *http.Request
	dumpBody bool
}

func (r *httpRequestDumpField) String() string {
	dump, err := httputil.DumpRequest(r.req, r.dumpBody)
	if err != nil {
		return fmt.Sprintf("request dump failed, err: %v", err)
	}

	return string(dump)
}

// HTTPRequestDump creates a zap.Field that can dump http request lazily.
//
// It is typically used with logger Debug calls.
// NOTE: The http request body is not dumped. For that use HTTRequestDumpBody.
func HTTPRequestDump(key string, req *http.Request) zap.Field {
	return zap.Field{
		Key:  key,
		Type: zapcore.StringerType,
		Interface: &httpRequestDumpField{
			req:      req,
			dumpBody: false,
		},
	}
}

// HTTPRequestDumpBody creates a zap.Field that can dump http request along with its body lazily.
//
// It is typically used with logger Debug calls.
// NOTE: If you want to dump the request path and headers only, use HTTPRequestDump.
func HTTPRequestDumpBody(key string, req *http.Request) zap.Field {
	return zap.Field{
		Key:  key,
		Type: zapcore.StringerType,
		Interface: &httpRequestDumpField{
			req:      req,
			dumpBody: true,
		},
	}
}

type httpResponseDumpField struct {
	req      *http.Response
	dumpBody bool
}

func (r *httpResponseDumpField) String() string {
	dump, err := httputil.DumpResponse(r.req, r.dumpBody)
	if err != nil {
		return fmt.Sprintf("response dump failed, err: %v", err)
	}

	return string(dump)
}

// HTTPResponseDump creates a zap.Field that can dump http response lazily.
//
// It is typically used with logger Debug calls.
// NOTE: The http response body is not dumped. For that use HTTResponseDumpBody.
func HTTPResponseDump(key string, req *http.Response) zap.Field {
	return zap.Field{
		Key:  key,
		Type: zapcore.StringerType,
		Interface: &httpResponseDumpField{
			req:      req,
			dumpBody: false,
		},
	}
}

// HTTPResponseDumpBody creates a zap.Field that can dump http request along with its body lazily.
//
// It is typically used with logger Debug calls.
// NOTE: If you want to dump the response path and headers only, use HTTPResponseDump.
func HTTPResponseDumpBody(key string, req *http.Response) zap.Field {
	return zap.Field{
		Key:  key,
		Type: zapcore.StringerType,
		Interface: &httpResponseDumpField{
			req:      req,
			dumpBody: true,
		},
	}
}
