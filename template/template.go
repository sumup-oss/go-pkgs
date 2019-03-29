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

package template

import (
	"html/template"
)

func Noescape(str string) template.HTML {
	//nolint:gosec
	return template.HTML(str)
}

var FuncMap = template.FuncMap{
	"noescape": Noescape,
}

func MustParseTemplate(name, content string) *template.Template {
	t, err := template.
		New(name).
		Funcs(FuncMap).
		Option("missingkey=error").
		Parse(content)
	return template.Must(t, err)
}

func MustParseTemplateWithFuncs(name string, funcs template.FuncMap, content string) *template.Template {
	funcMap := template.FuncMap{}
	for k, v := range FuncMap {
		funcMap[k] = v
	}

	for k, v := range funcs {
		funcMap[k] = v
	}

	t, err := template.
		New(name).
		Funcs(funcMap).
		Option("missingkey=error").
		Parse(content)
	return template.Must(t, err)
}
