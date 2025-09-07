// Copyright 2025 Marek Dalewski
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package jsonflag is a tiny Go library that can turn any JSON-marshallable struct into a set of command‑line flags.
package jsonflag

import (
	"reflect"
	"strings"
	"unicode"
)

// Name creates a new flag name by joining all names of struct fields along the provided path.
func Name(path []reflect.StructField) string {
	b := strings.Builder{}
	for i, p := range path {
		if i != 0 {
			b.WriteByte('.')
		}
		b.WriteString(p.Name)
	}
	n := b.String()
	if n == "" {
		return "input"
	}
	return n
}

// JSONName creates a new flag name by joining all JSON names (as package "encoding/json" would generate them) of struct fields along the provided path.
func JSONName(path []reflect.StructField) string {
	b := strings.Builder{}
	dot := false
	for _, p := range path {
		n := jsonFieldName(p)
		if n == "" {
			dot = false
			continue
		}
		if dot {
			b.WriteByte('.')
		}
		dot = true
		b.WriteString(n)
	}
	n := b.String()
	if n == "" {
		return "input"
	}
	return n
}

//nolint:gocritic // values of reflect.StructField are passed by value
func jsonFieldName(sf reflect.StructField) string {
	tag, ok := sf.Tag.Lookup("json")
	if !ok {
		return sf.Name
	}
	tag, _, _ = strings.Cut(tag, ",")
	if tag == "" {
		return sf.Name
	}
	if tag == "-" {
		return ""
	}
	return tag
}

// Usage attempts to retrieve from the last element of path a tag value under key 'usage', 'description' or 'desc'.
func Usage(path []reflect.StructField) string {
	if len(path) == 0 {
		return ""
	}
	if v := path[len(path)-1].Tag.Get("usage"); v != "" {
		return v
	}
	if v := path[len(path)-1].Tag.Get("description"); v != "" {
		return v
	}
	if v := path[len(path)-1].Tag.Get("desc"); v != "" {
		return v
	}
	return ""
}

// JsonCamelCase converts Go camel case flag name into JSON (javascript) camel case flag name, eg. "Foo.FooBar.FooBarBaz" to "foo.fooBar.fooBarBaz".
func JsonCamelCase(s string) string {
	if s == "" {
		return ""
	}
	result := make([]rune, 0, len(s))
	needLower := true
	for _, r := range s {
		if r == '.' {
			result = append(result, r)
			needLower = true
			continue
		}
		if needLower {
			r = unicode.ToLower(r)
			needLower = false
		}
		result = append(result, r)
	}
	return string(result)
}

// SnakeCase converts Go camel case flag name into snake case flag name, eg. "Foo.FooBar.FooBarBaz" to "foo.foo_bar.foo_bar_baz".
func SnakeCase(s string) string {
	if s == "" {
		return ""
	}
	result := make([]rune, 0, len(s))
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 && unicode.IsLower(rune(s[i-1])) && i+1 < len(s) && unicode.IsLower(rune(s[i+1])) {
				result = append(result, '_')
			}
			r = unicode.ToLower(r)
		}
		result = append(result, r)
	}
	return string(result)
}

// DashCase converts Go camel case flag name into dash case flag name, eg. "Foo.FooBar.FooBarBaz" to "foo.foo-bar.foo-bar-baz".
func DashCase(s string) string {
	if s == "" {
		return ""
	}
	result := make([]rune, 0, len(s))
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 && unicode.IsLower(rune(s[i-1])) && i+1 < len(s) && unicode.IsLower(rune(s[i+1])) {
				result = append(result, '-')
			}
			r = unicode.ToLower(r)
		}
		result = append(result, r)
	}
	return string(result)
}
