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

package jsonflag_test

import (
	"reflect"
	"testing"

	"github.com/daishe/jsonflag"
	"github.com/stretchr/testify/require"
)

func TestName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		given []reflect.StructField
		want  string
	}{
		{
			name:  "empty-path",
			given: []reflect.StructField{},
			want:  "input",
		},
		{
			name:  "single-field",
			given: []reflect.StructField{{Name: "Foo"}},
			want:  "Foo",
		},
		{
			name:  "multiple-fields",
			given: []reflect.StructField{{Name: "Foo"}, {Name: "Bar"}},
			want:  "Foo.Bar",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := jsonflag.Name(test.given)
			require.Equal(t, test.want, got)
		})
	}
}

func TestJSONName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		given []reflect.StructField
		want  string
	}{
		{
			name:  "empty-path",
			given: []reflect.StructField{},
			want:  "input",
		},
		{
			name:  "single-field/tag",
			given: []reflect.StructField{{Name: "Foo", Tag: `json:"foo"`}},
			want:  "foo",
		},
		{
			name:  "single-field/tag-with-options",
			given: []reflect.StructField{{Name: "Foo", Tag: `json:"foo,omitempty"`}},
			want:  "foo",
		},
		{
			name:  "single-field/tag-skips-field",
			given: []reflect.StructField{{Name: "Foo", Tag: `json:"-"`}},
			want:  "input",
		},
		{
			name:  "single-field/tag-empty-name",
			given: []reflect.StructField{{Name: "Foo", Tag: `json:""`}},
			want:  "Foo",
		},
		{
			name:  "single-field/missing-tag",
			given: []reflect.StructField{{Name: "Foo"}},
			want:  "Foo",
		},
		{
			name:  "multiple-fields/tag",
			given: []reflect.StructField{{Name: "Foo", Tag: `json:"foo"`}, {Name: "Bar", Tag: `json:"bar"`}},
			want:  "foo.bar",
		},
		{
			name:  "multiple-fields/tag-with-options",
			given: []reflect.StructField{{Name: "Foo", Tag: `json:"foo"`}, {Name: "Bar", Tag: `json:"bar,omitempty"`}},
			want:  "foo.bar",
		},
		{
			name:  "multiple-fields/tag-skips-field",
			given: []reflect.StructField{{Name: "Foo", Tag: `json:"foo"`}, {Name: "Bar", Tag: `json:"-"`}},
			want:  "foo",
		},
		{
			name:  "multiple-fields/tag-empty-name",
			given: []reflect.StructField{{Name: "Foo", Tag: `json:"foo"`}, {Name: "Bar", Tag: `json:""`}},
			want:  "foo.Bar",
		},
		{
			name:  "multiple-fields/missing-tag",
			given: []reflect.StructField{{Name: "Foo", Tag: `json:"foo"`}, {Name: "Bar"}},
			want:  "foo.Bar",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := jsonflag.JSONName(test.given)
			require.Equal(t, test.want, got)
		})
	}
}

func TestUsage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		given []reflect.StructField
		want  string
	}{
		{
			name:  "empty-path",
			given: []reflect.StructField{},
			want:  "",
		},
		{
			name:  "single-field/no-tag",
			given: []reflect.StructField{{Name: "Foo", Tag: ``}},
			want:  "",
		},
		{
			name:  "single-field/tag-usage",
			given: []reflect.StructField{{Name: "Foo", Tag: `usage:"usageString"`}},
			want:  "usageString",
		},
		{
			name:  "single-field/tag-description",
			given: []reflect.StructField{{Name: "Foo", Tag: `description:"descriptionString"`}},
			want:  "descriptionString",
		},
		{
			name:  "single-field/tag-desc",
			given: []reflect.StructField{{Name: "Foo", Tag: `desc:"descString"`}},
			want:  "descString",
		},
		{
			name:  "multiple-fields/no-tag",
			given: []reflect.StructField{{Name: "Foo", Tag: `usage:"fooUsageString"`}, {Name: "Bar", Tag: ``}},
			want:  "",
		},
		{
			name:  "multiple-fields/tag-usage",
			given: []reflect.StructField{{Name: "Foo", Tag: `usage:"fooUsageString"`}, {Name: "Bar", Tag: `usage:"usageString"`}},
			want:  "usageString",
		},
		{
			name:  "multiple-fields/tag-description",
			given: []reflect.StructField{{Name: "Foo", Tag: `description:"fooDescriptionString"`}, {Name: "Bar", Tag: `description:"descriptionString"`}},
			want:  "descriptionString",
		},
		{
			name:  "multiple-fields/tag-desc",
			given: []reflect.StructField{{Name: "Foo", Tag: `desc:"fooDescString"`}, {Name: "Bar", Tag: `desc:"descString"`}},
			want:  "descString",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := jsonflag.Usage(test.given)
			require.Equal(t, test.want, got)
		})
	}
}

func TestJsonCamelCase(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		given string
		want  string
	}{
		{name: "empty", given: "", want: ""},
		{name: "singlePart", given: "Foo", want: "foo"},
		{name: "doublePart", given: "FooBar", want: "fooBar"},
		{name: "singlePart/singlePart", given: "Foo.Bar", want: "foo.bar"},
		{name: "doublePart/doublePart", given: "FooBar.FooBaz", want: "fooBar.fooBaz"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := jsonflag.JsonCamelCase(test.given)
			require.Equal(t, test.want, got)
		})
	}
}

func TestSnakeCase(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		given string
		want  string
	}{
		{name: "empty", given: "", want: ""},
		{name: "singlePart", given: "Foo", want: "foo"},
		{name: "doublePart", given: "FooBar", want: "foo_bar"},
		{name: "singlePart/singlePart", given: "Foo.Bar", want: "foo.bar"},
		{name: "doublePart/doublePart", given: "FooBar.FooBaz", want: "foo_bar.foo_baz"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := jsonflag.SnakeCase(test.given)
			require.Equal(t, test.want, got)
		})
	}
}

func TestDashCase(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		given string
		want  string
	}{
		{name: "empty", given: "", want: ""},
		{name: "singlePart", given: "Foo", want: "foo"},
		{name: "doublePart", given: "FooBar", want: "foo-bar"},
		{name: "singlePart/singlePart", given: "Foo.Bar", want: "foo.bar"},
		{name: "doublePart/doublePart", given: "FooBar.FooBaz", want: "foo-bar.foo-baz"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := jsonflag.DashCase(test.given)
			require.Equal(t, test.want, got)
		})
	}
}
