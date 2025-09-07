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
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/daishe/jsonflag"
)

type FlagValueData struct {
	PathNames []string
	Type      string
	Get       any
	String    string
}

func (d *FlagValueData) Path() string {
	return strings.Join(d.PathNames, ".")
}

func DataOfFlagValue(val *jsonflag.Value) *FlagValueData {
	pathNames := []string{}
	for _, x := range val.Path() {
		pathNames = append(pathNames, x.Name)
	}
	return &FlagValueData{
		PathNames: pathNames,
		Type:      val.Type(),
		Get:       clone(val.Get()),
		String:    val.String(),
	}
}

func DataOfFlagValues(vals []*jsonflag.Value) (d []*FlagValueData) {
	for _, f := range vals {
		d = append(d, DataOfFlagValue(f))
	}
	return d
}

func RequireDataOfFlagValueEqual(t *testing.T, want, got *FlagValueData) {
	t.Helper()
	require.Equal(t, want, got, "mismatched flag value data for path %q", want.Path())
}

func RequireDataOfFlagValuesEqual(t *testing.T, want, got []*FlagValueData) {
	t.Helper()
	wantPaths, gotPaths := make([]string, len(want)), make([]string, len(got))
	for i := range want {
		wantPaths[i] = want[i].Path()
	}
	for i := range got {
		gotPaths[i] = got[i].Path()
	}
	require.Equal(t, wantPaths, gotPaths)
	for i := range want {
		RequireDataOfFlagValueEqual(t, want[i], got[i])
	}
}

func Type[T any]() reflect.Type {
	var v *T
	return reflect.TypeOf(v).Elem()
}

func Ptr[T any](v T) *T {
	return &v
}

func clone(x any) any {
	if x == nil {
		return x //nolint:gocritic // keep underlying type (if present)
	}
	return cloneValue(reflect.ValueOf(x)).Interface()
}

func cloneValue(x reflect.Value) reflect.Value {
	switch x.Kind() {
	case reflect.Pointer:
		if x.IsNil() {
			return reflect.New(x.Type()).Elem()
		}
		c := reflect.New(x.Type())
		c.Elem().Set(cloneValue(x.Elem()).Addr())
		return c.Elem()

	case reflect.Slice:
		if x.IsNil() {
			return reflect.New(x.Type()).Elem()
		}
		c := reflect.New(x.Type())
		c.Elem().Set(reflect.MakeSlice(x.Type(), x.Len(), x.Cap()))
		c = c.Elem()
		for i, el := range x.Seq2() {
			c.Index(int(i.Int())).Set(cloneValue(el))
		}
		return c

	case reflect.Array:
		c := reflect.New(x.Type()).Elem()
		for i := range x.Len() {
			c.Index(i).Set(cloneValue(x.Index(i)))
		}
		return c

	case reflect.Map:
		if x.IsNil() {
			return reflect.New(x.Type()).Elem()
		}
		c := reflect.New(x.Type())
		c.Elem().Set(reflect.MakeMap(x.Type()))
		c = c.Elem()
		for k, el := range x.Seq2() {
			c.SetMapIndex(cloneValue(k), cloneValue(el))
		}
		return c

	case reflect.Struct:
		c := reflect.New(x.Type()).Elem()
		for i, t := 0, x.Type(); i < t.NumField(); i++ {
			if !t.Field(i).IsExported() {
				continue
			}
			c.Field(i).Set(cloneValue(x.Field(i)))
		}
		return c

	default:
		if !x.CanAddr() {
			c := reflect.New(x.Type())
			c.Elem().Set(x)
			return c.Elem()
		}
		return x
	}
}

type TestBase struct {
	SkipValue  int                                 `json:"SkipValue,omitempty"`
	SkipPtr    *int                                `json:"SkipPtr,omitempty"`
	Bool       *TestGenericType[bool]              `json:"Bool,omitempty"`
	Int        *TestGenericType[int]               `json:"Int,omitempty"`
	Int8       *TestGenericType[int8]              `json:"Int8,omitempty"`
	Int16      *TestGenericType[int16]             `json:"Int16,omitempty"`
	Int32      *TestGenericType[int32]             `json:"Int32,omitempty"`
	Int64      *TestGenericType[int64]             `json:"Int64,omitempty"`
	Uint       *TestGenericType[uint]              `json:"Uint,omitempty"`
	Uint8      *TestGenericType[uint8]             `json:"Uint8,omitempty"`
	Uint16     *TestGenericType[uint16]            `json:"Uint16,omitempty"`
	Uint32     *TestGenericType[uint32]            `json:"Uint32,omitempty"`
	Uint64     *TestGenericType[uint64]            `json:"Uint64,omitempty"`
	Float32    *TestGenericType[float32]           `json:"Float32,omitempty"`
	Float64    *TestGenericType[float64]           `json:"Float64,omitempty"`
	Complex64  *TestGenericType[complex64]         `json:"Complex64,omitempty"`
	Complex128 *TestGenericType[complex128]        `json:"Complex128,omitempty"`
	String     *TestGenericType[string]            `json:"String,omitempty"`
	Bytes      *TestGenericType[[]byte]            `json:"Byte,omitempty"`
	Slice      *TestGenericType[[]string]          `json:"Slice,omitempty"`
	Map        *TestGenericType[map[string]string] `json:"Map,omitempty"`
	Struct     *TestGenericType[TestStruct]        `json:"Struct,omitempty"`
}

type TestGenericType[T any] struct {
	Value              T     `json:"Value,omitempty"`
	Ptr                *T    `json:"Ptr,omitempty"`
	SliceOfValues      []T   `json:"SliceOfValues,omitempty"`
	PtrToSliceOfValues *[]T  `json:"PtrToSliceOfValues,omitempty"`
	SliceOfPtrs        []*T  `json:"SliceOfPtrs,omitempty"`
	PtrToSliceOfPtrs   *[]*T `json:"PtrToSliceOfPtrs,omitempty"`
}

type TestStruct struct {
	Value string `json:"Value,omitempty"`
}

type TestUnexported struct {
	unexported int `json:"unexported,omitempty"` //nolint:govet // tagged for testing purpose
}

func TestRecursiveFlagValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		given     any
		filters   []jsonflag.FilterFunc
		postGiven any
		want      []*FlagValueData
	}{
		{
			name:      "nil",
			given:     nil,
			postGiven: nil,
			want:      []*FlagValueData(nil),
		},
		{
			name:      "string",
			given:     "",
			postGiven: "",
			want:      []*FlagValueData(nil),
		},
		{
			name:      "nil-ptr-to-string",
			given:     (*string)(nil),
			postGiven: (*string)(nil),
			want:      []*FlagValueData(nil),
		},
		{
			name:      "ptr-to-string",
			given:     Ptr(""),
			postGiven: Ptr(""),
			want: []*FlagValueData{
				{PathNames: []string{}, Type: "string", Get: Ptr(""), String: ``},
			},
		},
		{
			name:      "slice",
			given:     []string{},
			postGiven: []string{},
			want:      []*FlagValueData(nil),
		},
		{
			name:      "nil-ptr-to-slice",
			given:     (*[]string)(nil),
			postGiven: (*[]string)(nil),
			want:      []*FlagValueData(nil),
		},
		{
			name:      "ptr-to-slice",
			given:     &[]string{},
			postGiven: &[]string{},
			want: []*FlagValueData{
				{PathNames: []string{}, Type: "string (JSON list)", Get: &[]string{}, String: ``},
			},
		},
		{
			name:      "struct",
			given:     TestBase{},
			postGiven: TestBase{},
			want:      []*FlagValueData(nil),
		},
		{
			name:      "nil-ptr-to-struct",
			given:     (*TestBase)(nil),
			postGiven: (*TestBase)(nil),
			want:      []*FlagValueData(nil),
		},
		{
			name: "ptr-to-struct",
			filters: []jsonflag.FilterFunc{
				func(val *jsonflag.Value) jsonflag.FilterResult { // skip and do not descend into fields starting with "Skip"
					if p := val.Path(); len(p) > 0 && strings.HasPrefix(p[len(p)-1].Name, "Skip") {
						return jsonflag.SkipNoDescend
					}
					return jsonflag.IncludeAndDescend
				},
			},
			given: &TestBase{},
			postGiven: &TestBase{
				SkipValue:  0,
				SkipPtr:    nil,
				Bool:       &TestGenericType[bool]{Ptr: Ptr(false), PtrToSliceOfValues: Ptr([]bool(nil)), PtrToSliceOfPtrs: Ptr([]*bool(nil))},
				Int:        &TestGenericType[int]{Ptr: Ptr(int(0)), PtrToSliceOfValues: Ptr([]int(nil)), PtrToSliceOfPtrs: Ptr([]*int(nil))},
				Int8:       &TestGenericType[int8]{Ptr: Ptr(int8(0)), PtrToSliceOfValues: Ptr([]int8(nil)), PtrToSliceOfPtrs: Ptr([]*int8(nil))},
				Int16:      &TestGenericType[int16]{Ptr: Ptr(int16(0)), PtrToSliceOfValues: Ptr([]int16(nil)), PtrToSliceOfPtrs: Ptr([]*int16(nil))},
				Int32:      &TestGenericType[int32]{Ptr: Ptr(int32(0)), PtrToSliceOfValues: Ptr([]int32(nil)), PtrToSliceOfPtrs: Ptr([]*int32(nil))},
				Int64:      &TestGenericType[int64]{Ptr: Ptr(int64(0)), PtrToSliceOfValues: Ptr([]int64(nil)), PtrToSliceOfPtrs: Ptr([]*int64(nil))},
				Uint:       &TestGenericType[uint]{Ptr: Ptr(uint(0)), PtrToSliceOfValues: Ptr([]uint(nil)), PtrToSliceOfPtrs: Ptr([]*uint(nil))},
				Uint8:      &TestGenericType[uint8]{Ptr: Ptr(uint8(0)), PtrToSliceOfValues: Ptr([]uint8(nil)), PtrToSliceOfPtrs: Ptr([]*uint8(nil))},
				Uint16:     &TestGenericType[uint16]{Ptr: Ptr(uint16(0)), PtrToSliceOfValues: Ptr([]uint16(nil)), PtrToSliceOfPtrs: Ptr([]*uint16(nil))},
				Uint32:     &TestGenericType[uint32]{Ptr: Ptr(uint32(0)), PtrToSliceOfValues: Ptr([]uint32(nil)), PtrToSliceOfPtrs: Ptr([]*uint32(nil))},
				Uint64:     &TestGenericType[uint64]{Ptr: Ptr(uint64(0)), PtrToSliceOfValues: Ptr([]uint64(nil)), PtrToSliceOfPtrs: Ptr([]*uint64(nil))},
				Float32:    &TestGenericType[float32]{Ptr: Ptr(float32(0)), PtrToSliceOfValues: Ptr([]float32(nil)), PtrToSliceOfPtrs: Ptr([]*float32(nil))},
				Float64:    &TestGenericType[float64]{Ptr: Ptr(float64(0)), PtrToSliceOfValues: Ptr([]float64(nil)), PtrToSliceOfPtrs: Ptr([]*float64(nil))},
				Complex64:  &TestGenericType[complex64]{Ptr: Ptr(complex64(0)), PtrToSliceOfValues: Ptr([]complex64(nil)), PtrToSliceOfPtrs: Ptr([]*complex64(nil))},
				Complex128: &TestGenericType[complex128]{Ptr: Ptr(complex128(0)), PtrToSliceOfValues: Ptr([]complex128(nil)), PtrToSliceOfPtrs: Ptr([]*complex128(nil))},
				String:     &TestGenericType[string]{Ptr: Ptr(""), PtrToSliceOfValues: Ptr([]string(nil)), PtrToSliceOfPtrs: Ptr([]*string(nil))},
				Bytes:      &TestGenericType[[]byte]{Ptr: Ptr([]byte(nil)), PtrToSliceOfValues: Ptr([][]byte(nil)), PtrToSliceOfPtrs: Ptr([]*[]byte(nil))},
				Slice:      &TestGenericType[[]string]{Ptr: Ptr([]string(nil)), PtrToSliceOfValues: Ptr([][]string(nil)), PtrToSliceOfPtrs: Ptr([]*[]string(nil))},
				Map:        &TestGenericType[map[string]string]{Ptr: Ptr(map[string]string(nil)), PtrToSliceOfValues: Ptr([]map[string]string(nil)), PtrToSliceOfPtrs: Ptr([]*map[string]string(nil))},
				Struct:     &TestGenericType[TestStruct]{Ptr: &TestStruct{}, PtrToSliceOfValues: Ptr([]TestStruct(nil)), PtrToSliceOfPtrs: Ptr([]*TestStruct(nil))},
			},
			want: []*FlagValueData{
				{PathNames: []string{}, Type: "JSON object", Get: &TestBase{}, String: ``},
				{PathNames: []string{"Bool"}, Type: "JSON object", Get: &TestGenericType[bool]{}, String: ``},
				{PathNames: []string{"Bool", "Value"}, Type: "bool", Get: false, String: ``},
				{PathNames: []string{"Bool", "Ptr"}, Type: "bool", Get: Ptr(false), String: ``},
				{PathNames: []string{"Bool", "SliceOfValues"}, Type: "bool (JSON list)", Get: []bool(nil), String: ``},
				{PathNames: []string{"Bool", "PtrToSliceOfValues"}, Type: "bool (JSON list)", Get: Ptr([]bool(nil)), String: ``},
				{PathNames: []string{"Bool", "SliceOfPtrs"}, Type: "bool (JSON list)", Get: []*bool(nil), String: ``},
				{PathNames: []string{"Bool", "PtrToSliceOfPtrs"}, Type: "bool (JSON list)", Get: Ptr([]*bool(nil)), String: ``},
				{PathNames: []string{"Int"}, Type: "JSON object", Get: &TestGenericType[int]{}, String: ``},
				{PathNames: []string{"Int", "Value"}, Type: "int", Get: int(0), String: ``},
				{PathNames: []string{"Int", "Ptr"}, Type: "int", Get: Ptr(int(0)), String: ``},
				{PathNames: []string{"Int", "SliceOfValues"}, Type: "int (JSON list)", Get: []int(nil), String: ``},
				{PathNames: []string{"Int", "PtrToSliceOfValues"}, Type: "int (JSON list)", Get: Ptr([]int(nil)), String: ``},
				{PathNames: []string{"Int", "SliceOfPtrs"}, Type: "int (JSON list)", Get: []*int(nil), String: ``},
				{PathNames: []string{"Int", "PtrToSliceOfPtrs"}, Type: "int (JSON list)", Get: Ptr([]*int(nil)), String: ``},
				{PathNames: []string{"Int8"}, Type: "JSON object", Get: &TestGenericType[int8]{}, String: ``},
				{PathNames: []string{"Int8", "Value"}, Type: "int8", Get: int8(0), String: ``},
				{PathNames: []string{"Int8", "Ptr"}, Type: "int8", Get: Ptr(int8(0)), String: ``},
				{PathNames: []string{"Int8", "SliceOfValues"}, Type: "int8 (JSON list)", Get: []int8(nil), String: ``},
				{PathNames: []string{"Int8", "PtrToSliceOfValues"}, Type: "int8 (JSON list)", Get: Ptr([]int8(nil)), String: ``},
				{PathNames: []string{"Int8", "SliceOfPtrs"}, Type: "int8 (JSON list)", Get: []*int8(nil), String: ``},
				{PathNames: []string{"Int8", "PtrToSliceOfPtrs"}, Type: "int8 (JSON list)", Get: Ptr([]*int8(nil)), String: ``},
				{PathNames: []string{"Int16"}, Type: "JSON object", Get: &TestGenericType[int16]{}, String: ``},
				{PathNames: []string{"Int16", "Value"}, Type: "int16", Get: int16(0), String: ``},
				{PathNames: []string{"Int16", "Ptr"}, Type: "int16", Get: Ptr(int16(0)), String: ``},
				{PathNames: []string{"Int16", "SliceOfValues"}, Type: "int16 (JSON list)", Get: []int16(nil), String: ``},
				{PathNames: []string{"Int16", "PtrToSliceOfValues"}, Type: "int16 (JSON list)", Get: Ptr([]int16(nil)), String: ``},
				{PathNames: []string{"Int16", "SliceOfPtrs"}, Type: "int16 (JSON list)", Get: []*int16(nil), String: ``},
				{PathNames: []string{"Int16", "PtrToSliceOfPtrs"}, Type: "int16 (JSON list)", Get: Ptr([]*int16(nil)), String: ``},
				{PathNames: []string{"Int32"}, Type: "JSON object", Get: &TestGenericType[int32]{}, String: ``},
				{PathNames: []string{"Int32", "Value"}, Type: "int32", Get: int32(0), String: ``},
				{PathNames: []string{"Int32", "Ptr"}, Type: "int32", Get: Ptr(int32(0)), String: ``},
				{PathNames: []string{"Int32", "SliceOfValues"}, Type: "int32 (JSON list)", Get: []int32(nil), String: ``},
				{PathNames: []string{"Int32", "PtrToSliceOfValues"}, Type: "int32 (JSON list)", Get: Ptr([]int32(nil)), String: ``},
				{PathNames: []string{"Int32", "SliceOfPtrs"}, Type: "int32 (JSON list)", Get: []*int32(nil), String: ``},
				{PathNames: []string{"Int32", "PtrToSliceOfPtrs"}, Type: "int32 (JSON list)", Get: Ptr([]*int32(nil)), String: ``},
				{PathNames: []string{"Int64"}, Type: "JSON object", Get: &TestGenericType[int64]{}, String: ``},
				{PathNames: []string{"Int64", "Value"}, Type: "int64", Get: int64(0), String: ``},
				{PathNames: []string{"Int64", "Ptr"}, Type: "int64", Get: Ptr(int64(0)), String: ``},
				{PathNames: []string{"Int64", "SliceOfValues"}, Type: "int64 (JSON list)", Get: []int64(nil), String: ``},
				{PathNames: []string{"Int64", "PtrToSliceOfValues"}, Type: "int64 (JSON list)", Get: Ptr([]int64(nil)), String: ``},
				{PathNames: []string{"Int64", "SliceOfPtrs"}, Type: "int64 (JSON list)", Get: []*int64(nil), String: ``},
				{PathNames: []string{"Int64", "PtrToSliceOfPtrs"}, Type: "int64 (JSON list)", Get: Ptr([]*int64(nil)), String: ``},
				{PathNames: []string{"Uint"}, Type: "JSON object", Get: &TestGenericType[uint]{}, String: ``},
				{PathNames: []string{"Uint", "Value"}, Type: "uint", Get: uint(0), String: ``},
				{PathNames: []string{"Uint", "Ptr"}, Type: "uint", Get: Ptr(uint(0)), String: ``},
				{PathNames: []string{"Uint", "SliceOfValues"}, Type: "uint (JSON list)", Get: []uint(nil), String: ``},
				{PathNames: []string{"Uint", "PtrToSliceOfValues"}, Type: "uint (JSON list)", Get: Ptr([]uint(nil)), String: ``},
				{PathNames: []string{"Uint", "SliceOfPtrs"}, Type: "uint (JSON list)", Get: []*uint(nil), String: ``},
				{PathNames: []string{"Uint", "PtrToSliceOfPtrs"}, Type: "uint (JSON list)", Get: Ptr([]*uint(nil)), String: ``},
				{PathNames: []string{"Uint8"}, Type: "JSON object", Get: &TestGenericType[uint8]{}, String: ``},
				{PathNames: []string{"Uint8", "Value"}, Type: "uint8", Get: uint8(0), String: ``},
				{PathNames: []string{"Uint8", "Ptr"}, Type: "uint8", Get: Ptr(uint8(0)), String: ``},
				{PathNames: []string{"Uint8", "SliceOfValues"}, Type: "base64", Get: []uint8(nil), String: ``},
				{PathNames: []string{"Uint8", "PtrToSliceOfValues"}, Type: "base64", Get: Ptr([]uint8(nil)), String: ``},
				{PathNames: []string{"Uint8", "SliceOfPtrs"}, Type: "uint8 (JSON list)", Get: []*uint8(nil), String: ``},
				{PathNames: []string{"Uint8", "PtrToSliceOfPtrs"}, Type: "uint8 (JSON list)", Get: Ptr([]*uint8(nil)), String: ``},
				{PathNames: []string{"Uint16"}, Type: "JSON object", Get: &TestGenericType[uint16]{}, String: ``},
				{PathNames: []string{"Uint16", "Value"}, Type: "uint16", Get: uint16(0), String: ``},
				{PathNames: []string{"Uint16", "Ptr"}, Type: "uint16", Get: Ptr(uint16(0)), String: ``},
				{PathNames: []string{"Uint16", "SliceOfValues"}, Type: "uint16 (JSON list)", Get: []uint16(nil), String: ``},
				{PathNames: []string{"Uint16", "PtrToSliceOfValues"}, Type: "uint16 (JSON list)", Get: Ptr([]uint16(nil)), String: ``},
				{PathNames: []string{"Uint16", "SliceOfPtrs"}, Type: "uint16 (JSON list)", Get: []*uint16(nil), String: ``},
				{PathNames: []string{"Uint16", "PtrToSliceOfPtrs"}, Type: "uint16 (JSON list)", Get: Ptr([]*uint16(nil)), String: ``},
				{PathNames: []string{"Uint32"}, Type: "JSON object", Get: &TestGenericType[uint32]{}, String: ``},
				{PathNames: []string{"Uint32", "Value"}, Type: "uint32", Get: uint32(0), String: ``},
				{PathNames: []string{"Uint32", "Ptr"}, Type: "uint32", Get: Ptr(uint32(0)), String: ``},
				{PathNames: []string{"Uint32", "SliceOfValues"}, Type: "uint32 (JSON list)", Get: []uint32(nil), String: ``},
				{PathNames: []string{"Uint32", "PtrToSliceOfValues"}, Type: "uint32 (JSON list)", Get: Ptr([]uint32(nil)), String: ``},
				{PathNames: []string{"Uint32", "SliceOfPtrs"}, Type: "uint32 (JSON list)", Get: []*uint32(nil), String: ``},
				{PathNames: []string{"Uint32", "PtrToSliceOfPtrs"}, Type: "uint32 (JSON list)", Get: Ptr([]*uint32(nil)), String: ``},
				{PathNames: []string{"Uint64"}, Type: "JSON object", Get: &TestGenericType[uint64]{}, String: ``},
				{PathNames: []string{"Uint64", "Value"}, Type: "uint64", Get: uint64(0), String: ``},
				{PathNames: []string{"Uint64", "Ptr"}, Type: "uint64", Get: Ptr(uint64(0)), String: ``},
				{PathNames: []string{"Uint64", "SliceOfValues"}, Type: "uint64 (JSON list)", Get: []uint64(nil), String: ``},
				{PathNames: []string{"Uint64", "PtrToSliceOfValues"}, Type: "uint64 (JSON list)", Get: Ptr([]uint64(nil)), String: ``},
				{PathNames: []string{"Uint64", "SliceOfPtrs"}, Type: "uint64 (JSON list)", Get: []*uint64(nil), String: ``},
				{PathNames: []string{"Uint64", "PtrToSliceOfPtrs"}, Type: "uint64 (JSON list)", Get: Ptr([]*uint64(nil)), String: ``},
				{PathNames: []string{"Float32"}, Type: "JSON object", Get: &TestGenericType[float32]{}, String: ``},
				{PathNames: []string{"Float32", "Value"}, Type: "float32", Get: float32(0), String: ``},
				{PathNames: []string{"Float32", "Ptr"}, Type: "float32", Get: Ptr(float32(0)), String: ``},
				{PathNames: []string{"Float32", "SliceOfValues"}, Type: "float32 (JSON list)", Get: []float32(nil), String: ``},
				{PathNames: []string{"Float32", "PtrToSliceOfValues"}, Type: "float32 (JSON list)", Get: Ptr([]float32(nil)), String: ``},
				{PathNames: []string{"Float32", "SliceOfPtrs"}, Type: "float32 (JSON list)", Get: []*float32(nil), String: ``},
				{PathNames: []string{"Float32", "PtrToSliceOfPtrs"}, Type: "float32 (JSON list)", Get: Ptr([]*float32(nil)), String: ``},
				{PathNames: []string{"Float64"}, Type: "JSON object", Get: &TestGenericType[float64]{}, String: ``},
				{PathNames: []string{"Float64", "Value"}, Type: "float64", Get: float64(0), String: ``},
				{PathNames: []string{"Float64", "Ptr"}, Type: "float64", Get: Ptr(float64(0)), String: ``},
				{PathNames: []string{"Float64", "SliceOfValues"}, Type: "float64 (JSON list)", Get: []float64(nil), String: ``},
				{PathNames: []string{"Float64", "PtrToSliceOfValues"}, Type: "float64 (JSON list)", Get: Ptr([]float64(nil)), String: ``},
				{PathNames: []string{"Float64", "SliceOfPtrs"}, Type: "float64 (JSON list)", Get: []*float64(nil), String: ``},
				{PathNames: []string{"Float64", "PtrToSliceOfPtrs"}, Type: "float64 (JSON list)", Get: Ptr([]*float64(nil)), String: ``},
				{PathNames: []string{"Complex64"}, Type: "JSON object", Get: &TestGenericType[complex64]{}, String: ``},
				{PathNames: []string{"Complex64", "Value"}, Type: "complex64", Get: complex64(0), String: ``},
				{PathNames: []string{"Complex64", "Ptr"}, Type: "complex64", Get: Ptr(complex64(0)), String: ``},
				{PathNames: []string{"Complex64", "SliceOfValues"}, Type: "complex64 (JSON list)", Get: []complex64(nil), String: ``},
				{PathNames: []string{"Complex64", "PtrToSliceOfValues"}, Type: "complex64 (JSON list)", Get: Ptr([]complex64(nil)), String: ``},
				{PathNames: []string{"Complex64", "SliceOfPtrs"}, Type: "complex64 (JSON list)", Get: []*complex64(nil), String: ``},
				{PathNames: []string{"Complex64", "PtrToSliceOfPtrs"}, Type: "complex64 (JSON list)", Get: Ptr([]*complex64(nil)), String: ``},
				{PathNames: []string{"Complex128"}, Type: "JSON object", Get: &TestGenericType[complex128]{}, String: ``},
				{PathNames: []string{"Complex128", "Value"}, Type: "complex128", Get: complex128(0), String: ``},
				{PathNames: []string{"Complex128", "Ptr"}, Type: "complex128", Get: Ptr(complex128(0)), String: ``},
				{PathNames: []string{"Complex128", "SliceOfValues"}, Type: "complex128 (JSON list)", Get: []complex128(nil), String: ``},
				{PathNames: []string{"Complex128", "PtrToSliceOfValues"}, Type: "complex128 (JSON list)", Get: Ptr([]complex128(nil)), String: ``},
				{PathNames: []string{"Complex128", "SliceOfPtrs"}, Type: "complex128 (JSON list)", Get: []*complex128(nil), String: ``},
				{PathNames: []string{"Complex128", "PtrToSliceOfPtrs"}, Type: "complex128 (JSON list)", Get: Ptr([]*complex128(nil)), String: ``},
				{PathNames: []string{"String"}, Type: "JSON object", Get: &TestGenericType[string]{}, String: ``},
				{PathNames: []string{"String", "Value"}, Type: "string", Get: "", String: ``},
				{PathNames: []string{"String", "Ptr"}, Type: "string", Get: Ptr(""), String: ``},
				{PathNames: []string{"String", "SliceOfValues"}, Type: "string (JSON list)", Get: []string(nil), String: ``},
				{PathNames: []string{"String", "PtrToSliceOfValues"}, Type: "string (JSON list)", Get: Ptr([]string(nil)), String: ``},
				{PathNames: []string{"String", "SliceOfPtrs"}, Type: "string (JSON list)", Get: []*string(nil), String: ``},
				{PathNames: []string{"String", "PtrToSliceOfPtrs"}, Type: "string (JSON list)", Get: Ptr([]*string(nil)), String: ``},
				{PathNames: []string{"Bytes"}, Type: "JSON object", Get: &TestGenericType[[]byte]{}, String: ``},
				{PathNames: []string{"Bytes", "Value"}, Type: "base64", Get: []byte(nil), String: ``},
				{PathNames: []string{"Bytes", "Ptr"}, Type: "base64", Get: Ptr([]byte(nil)), String: ``},
				{PathNames: []string{"Bytes", "SliceOfValues"}, Type: "base64 (JSON list)", Get: [][]byte(nil), String: ``},
				{PathNames: []string{"Bytes", "PtrToSliceOfValues"}, Type: "base64 (JSON list)", Get: Ptr([][]byte(nil)), String: ``},
				{PathNames: []string{"Bytes", "SliceOfPtrs"}, Type: "base64 (JSON list)", Get: []*[]byte(nil), String: ``},
				{PathNames: []string{"Bytes", "PtrToSliceOfPtrs"}, Type: "base64 (JSON list)", Get: Ptr([]*[]byte(nil)), String: ``},
				{PathNames: []string{"Slice"}, Type: "JSON object", Get: &TestGenericType[[]string]{}, String: ``},
				{PathNames: []string{"Slice", "Value"}, Type: "string (JSON list)", Get: []string(nil), String: ``},
				{PathNames: []string{"Slice", "Ptr"}, Type: "string (JSON list)", Get: Ptr([]string(nil)), String: ``},
				{PathNames: []string{"Slice", "SliceOfValues"}, Type: "JSON list", Get: [][]string(nil), String: ``},
				{PathNames: []string{"Slice", "PtrToSliceOfValues"}, Type: "JSON list", Get: Ptr([][]string(nil)), String: ``},
				{PathNames: []string{"Slice", "SliceOfPtrs"}, Type: "JSON list", Get: []*[]string(nil), String: ``},
				{PathNames: []string{"Slice", "PtrToSliceOfPtrs"}, Type: "JSON list", Get: Ptr([]*[]string(nil)), String: ``},
				{PathNames: []string{"Map"}, Type: "JSON object", Get: &TestGenericType[map[string]string]{}, String: ``},
				{PathNames: []string{"Map", "Value"}, Type: "JSON object", Get: map[string]string(nil), String: ``},
				{PathNames: []string{"Map", "Ptr"}, Type: "JSON object", Get: Ptr(map[string]string(nil)), String: ``},
				{PathNames: []string{"Map", "SliceOfValues"}, Type: "JSON object (JSON list)", Get: []map[string]string(nil), String: ``},
				{PathNames: []string{"Map", "PtrToSliceOfValues"}, Type: "JSON object (JSON list)", Get: Ptr([]map[string]string(nil)), String: ``},
				{PathNames: []string{"Map", "SliceOfPtrs"}, Type: "JSON object (JSON list)", Get: []*map[string]string(nil), String: ``},
				{PathNames: []string{"Map", "PtrToSliceOfPtrs"}, Type: "JSON object (JSON list)", Get: Ptr([]*map[string]string(nil)), String: ``},
				{PathNames: []string{"Struct"}, Type: "JSON object", Get: &TestGenericType[TestStruct]{}, String: `{"Value":{}}`},
				{PathNames: []string{"Struct", "Value"}, Type: "JSON object", Get: TestStruct{}, String: ``},
				{PathNames: []string{"Struct", "Value", "Value"}, Type: "string", Get: "", String: ``},
				{PathNames: []string{"Struct", "Ptr"}, Type: "JSON object", Get: &TestStruct{}, String: ``},
				{PathNames: []string{"Struct", "Ptr", "Value"}, Type: "string", Get: "", String: ``},
				{PathNames: []string{"Struct", "SliceOfValues"}, Type: "JSON object (JSON list)", Get: []TestStruct(nil), String: ``},
				{PathNames: []string{"Struct", "PtrToSliceOfValues"}, Type: "JSON object (JSON list)", Get: Ptr([]TestStruct(nil)), String: ``},
				{PathNames: []string{"Struct", "SliceOfPtrs"}, Type: "JSON object (JSON list)", Get: []*TestStruct(nil), String: ``},
				{PathNames: []string{"Struct", "PtrToSliceOfPtrs"}, Type: "JSON object (JSON list)", Get: Ptr([]*TestStruct(nil)), String: ``},
			},
		},
		{
			name: "ptr-to-struct-non-default",
			filters: []jsonflag.FilterFunc{
				func(val *jsonflag.Value) jsonflag.FilterResult { // skip and do not descend into fields starting with "Skip"
					if p := val.Path(); len(p) > 0 && strings.HasPrefix(p[len(p)-1].Name, "Skip") {
						return jsonflag.SkipNoDescend
					}
					return jsonflag.IncludeAndDescend
				},
			},
			given: &TestBase{
				SkipValue:  1,
				SkipPtr:    Ptr(1),
				Bool:       &TestGenericType[bool]{Value: true, Ptr: Ptr(true), SliceOfValues: []bool{true}, PtrToSliceOfValues: &[]bool{true}, SliceOfPtrs: []*bool{Ptr(true)}, PtrToSliceOfPtrs: &[]*bool{Ptr(true)}},
				Int:        &TestGenericType[int]{Value: 1, Ptr: Ptr(int(1)), SliceOfValues: []int{1}, PtrToSliceOfValues: &[]int{1}, SliceOfPtrs: []*int{Ptr(int(1))}, PtrToSliceOfPtrs: &[]*int{Ptr(int(1))}},
				Int8:       &TestGenericType[int8]{Value: 1, Ptr: Ptr(int8(1)), SliceOfValues: []int8{1}, PtrToSliceOfValues: &[]int8{1}, SliceOfPtrs: []*int8{Ptr(int8(1))}, PtrToSliceOfPtrs: &[]*int8{Ptr(int8(1))}},
				Int16:      &TestGenericType[int16]{Value: 1, Ptr: Ptr(int16(1)), SliceOfValues: []int16{1}, PtrToSliceOfValues: &[]int16{1}, SliceOfPtrs: []*int16{Ptr(int16(1))}, PtrToSliceOfPtrs: &[]*int16{Ptr(int16(1))}},
				Int32:      &TestGenericType[int32]{Value: 1, Ptr: Ptr(int32(1)), SliceOfValues: []int32{1}, PtrToSliceOfValues: &[]int32{1}, SliceOfPtrs: []*int32{Ptr(int32(1))}, PtrToSliceOfPtrs: &[]*int32{Ptr(int32(1))}},
				Int64:      &TestGenericType[int64]{Value: 1, Ptr: Ptr(int64(1)), SliceOfValues: []int64{1}, PtrToSliceOfValues: &[]int64{1}, SliceOfPtrs: []*int64{Ptr(int64(1))}, PtrToSliceOfPtrs: &[]*int64{Ptr(int64(1))}},
				Uint:       &TestGenericType[uint]{Value: 1, Ptr: Ptr(uint(1)), SliceOfValues: []uint{1}, PtrToSliceOfValues: &[]uint{1}, SliceOfPtrs: []*uint{Ptr(uint(1))}, PtrToSliceOfPtrs: &[]*uint{Ptr(uint(1))}},
				Uint8:      &TestGenericType[uint8]{Value: 1, Ptr: Ptr(uint8(1)), SliceOfValues: []uint8{1}, PtrToSliceOfValues: &[]uint8{1}, SliceOfPtrs: []*uint8{Ptr(uint8(1))}, PtrToSliceOfPtrs: &[]*uint8{Ptr(uint8(1))}},
				Uint16:     &TestGenericType[uint16]{Value: 1, Ptr: Ptr(uint16(1)), SliceOfValues: []uint16{1}, PtrToSliceOfValues: &[]uint16{1}, SliceOfPtrs: []*uint16{Ptr(uint16(1))}, PtrToSliceOfPtrs: &[]*uint16{Ptr(uint16(1))}},
				Uint32:     &TestGenericType[uint32]{Value: 1, Ptr: Ptr(uint32(1)), SliceOfValues: []uint32{1}, PtrToSliceOfValues: &[]uint32{1}, SliceOfPtrs: []*uint32{Ptr(uint32(1))}, PtrToSliceOfPtrs: &[]*uint32{Ptr(uint32(1))}},
				Uint64:     &TestGenericType[uint64]{Value: 1, Ptr: Ptr(uint64(1)), SliceOfValues: []uint64{1}, PtrToSliceOfValues: &[]uint64{1}, SliceOfPtrs: []*uint64{Ptr(uint64(1))}, PtrToSliceOfPtrs: &[]*uint64{Ptr(uint64(1))}},
				Float32:    &TestGenericType[float32]{Value: 1, Ptr: Ptr(float32(1)), SliceOfValues: []float32{1}, PtrToSliceOfValues: &[]float32{1}, SliceOfPtrs: []*float32{Ptr(float32(1))}, PtrToSliceOfPtrs: &[]*float32{Ptr(float32(1))}},
				Float64:    &TestGenericType[float64]{Value: 1, Ptr: Ptr(float64(1)), SliceOfValues: []float64{1}, PtrToSliceOfValues: &[]float64{1}, SliceOfPtrs: []*float64{Ptr(float64(1))}, PtrToSliceOfPtrs: &[]*float64{Ptr(float64(1))}},
				Complex64:  &TestGenericType[complex64]{Value: 1, Ptr: Ptr(complex64(1)), SliceOfValues: []complex64{1}, PtrToSliceOfValues: &[]complex64{1}, SliceOfPtrs: []*complex64{Ptr(complex64(1))}, PtrToSliceOfPtrs: &[]*complex64{Ptr(complex64(1))}},
				Complex128: &TestGenericType[complex128]{Value: 1, Ptr: Ptr(complex128(1)), SliceOfValues: []complex128{1}, PtrToSliceOfValues: &[]complex128{1}, SliceOfPtrs: []*complex128{Ptr(complex128(1))}, PtrToSliceOfPtrs: &[]*complex128{Ptr(complex128(1))}},
				String:     &TestGenericType[string]{Value: "a", Ptr: Ptr("a"), SliceOfValues: []string{"a"}, PtrToSliceOfValues: &[]string{"a"}, SliceOfPtrs: []*string{Ptr("a")}, PtrToSliceOfPtrs: &[]*string{Ptr("a")}},
				Bytes:      &TestGenericType[[]byte]{Value: []byte{1}, Ptr: &[]byte{1}, SliceOfValues: [][]byte{{1}}, PtrToSliceOfValues: &[][]byte{{1}}, SliceOfPtrs: []*[]byte{{1}}, PtrToSliceOfPtrs: &[]*[]byte{{1}}},
				Slice:      &TestGenericType[[]string]{Value: []string{"a"}, Ptr: &[]string{"a"}, SliceOfValues: [][]string{{"a"}}, PtrToSliceOfValues: &[][]string{{"a"}}, SliceOfPtrs: []*[]string{{"a"}}, PtrToSliceOfPtrs: &[]*[]string{{"a"}}},
				Map:        &TestGenericType[map[string]string]{Value: map[string]string{"k": "a"}, Ptr: &map[string]string{"k": "a"}, SliceOfValues: []map[string]string{{"k": "a"}}, PtrToSliceOfValues: &[]map[string]string{{"k": "a"}}, SliceOfPtrs: []*map[string]string{{"k": "a"}}, PtrToSliceOfPtrs: &[]*map[string]string{{"k": "a"}}},
				Struct:     &TestGenericType[TestStruct]{Value: TestStruct{Value: "a"}, Ptr: &TestStruct{Value: "a"}, SliceOfValues: []TestStruct{{Value: "a"}}, PtrToSliceOfValues: &[]TestStruct{{Value: "a"}}, SliceOfPtrs: []*TestStruct{{Value: "a"}}, PtrToSliceOfPtrs: &[]*TestStruct{{Value: "a"}}},
			},
			postGiven: &TestBase{
				SkipValue:  1,
				SkipPtr:    Ptr(1),
				Bool:       &TestGenericType[bool]{Value: true, Ptr: Ptr(true), SliceOfValues: []bool{true}, PtrToSliceOfValues: &[]bool{true}, SliceOfPtrs: []*bool{Ptr(true)}, PtrToSliceOfPtrs: &[]*bool{Ptr(true)}},
				Int:        &TestGenericType[int]{Value: 1, Ptr: Ptr(int(1)), SliceOfValues: []int{1}, PtrToSliceOfValues: &[]int{1}, SliceOfPtrs: []*int{Ptr(int(1))}, PtrToSliceOfPtrs: &[]*int{Ptr(int(1))}},
				Int8:       &TestGenericType[int8]{Value: 1, Ptr: Ptr(int8(1)), SliceOfValues: []int8{1}, PtrToSliceOfValues: &[]int8{1}, SliceOfPtrs: []*int8{Ptr(int8(1))}, PtrToSliceOfPtrs: &[]*int8{Ptr(int8(1))}},
				Int16:      &TestGenericType[int16]{Value: 1, Ptr: Ptr(int16(1)), SliceOfValues: []int16{1}, PtrToSliceOfValues: &[]int16{1}, SliceOfPtrs: []*int16{Ptr(int16(1))}, PtrToSliceOfPtrs: &[]*int16{Ptr(int16(1))}},
				Int32:      &TestGenericType[int32]{Value: 1, Ptr: Ptr(int32(1)), SliceOfValues: []int32{1}, PtrToSliceOfValues: &[]int32{1}, SliceOfPtrs: []*int32{Ptr(int32(1))}, PtrToSliceOfPtrs: &[]*int32{Ptr(int32(1))}},
				Int64:      &TestGenericType[int64]{Value: 1, Ptr: Ptr(int64(1)), SliceOfValues: []int64{1}, PtrToSliceOfValues: &[]int64{1}, SliceOfPtrs: []*int64{Ptr(int64(1))}, PtrToSliceOfPtrs: &[]*int64{Ptr(int64(1))}},
				Uint:       &TestGenericType[uint]{Value: 1, Ptr: Ptr(uint(1)), SliceOfValues: []uint{1}, PtrToSliceOfValues: &[]uint{1}, SliceOfPtrs: []*uint{Ptr(uint(1))}, PtrToSliceOfPtrs: &[]*uint{Ptr(uint(1))}},
				Uint8:      &TestGenericType[uint8]{Value: 1, Ptr: Ptr(uint8(1)), SliceOfValues: []uint8{1}, PtrToSliceOfValues: &[]uint8{1}, SliceOfPtrs: []*uint8{Ptr(uint8(1))}, PtrToSliceOfPtrs: &[]*uint8{Ptr(uint8(1))}},
				Uint16:     &TestGenericType[uint16]{Value: 1, Ptr: Ptr(uint16(1)), SliceOfValues: []uint16{1}, PtrToSliceOfValues: &[]uint16{1}, SliceOfPtrs: []*uint16{Ptr(uint16(1))}, PtrToSliceOfPtrs: &[]*uint16{Ptr(uint16(1))}},
				Uint32:     &TestGenericType[uint32]{Value: 1, Ptr: Ptr(uint32(1)), SliceOfValues: []uint32{1}, PtrToSliceOfValues: &[]uint32{1}, SliceOfPtrs: []*uint32{Ptr(uint32(1))}, PtrToSliceOfPtrs: &[]*uint32{Ptr(uint32(1))}},
				Uint64:     &TestGenericType[uint64]{Value: 1, Ptr: Ptr(uint64(1)), SliceOfValues: []uint64{1}, PtrToSliceOfValues: &[]uint64{1}, SliceOfPtrs: []*uint64{Ptr(uint64(1))}, PtrToSliceOfPtrs: &[]*uint64{Ptr(uint64(1))}},
				Float32:    &TestGenericType[float32]{Value: 1, Ptr: Ptr(float32(1)), SliceOfValues: []float32{1}, PtrToSliceOfValues: &[]float32{1}, SliceOfPtrs: []*float32{Ptr(float32(1))}, PtrToSliceOfPtrs: &[]*float32{Ptr(float32(1))}},
				Float64:    &TestGenericType[float64]{Value: 1, Ptr: Ptr(float64(1)), SliceOfValues: []float64{1}, PtrToSliceOfValues: &[]float64{1}, SliceOfPtrs: []*float64{Ptr(float64(1))}, PtrToSliceOfPtrs: &[]*float64{Ptr(float64(1))}},
				Complex64:  &TestGenericType[complex64]{Value: 1, Ptr: Ptr(complex64(1)), SliceOfValues: []complex64{1}, PtrToSliceOfValues: &[]complex64{1}, SliceOfPtrs: []*complex64{Ptr(complex64(1))}, PtrToSliceOfPtrs: &[]*complex64{Ptr(complex64(1))}},
				Complex128: &TestGenericType[complex128]{Value: 1, Ptr: Ptr(complex128(1)), SliceOfValues: []complex128{1}, PtrToSliceOfValues: &[]complex128{1}, SliceOfPtrs: []*complex128{Ptr(complex128(1))}, PtrToSliceOfPtrs: &[]*complex128{Ptr(complex128(1))}},
				String:     &TestGenericType[string]{Value: "a", Ptr: Ptr("a"), SliceOfValues: []string{"a"}, PtrToSliceOfValues: &[]string{"a"}, SliceOfPtrs: []*string{Ptr("a")}, PtrToSliceOfPtrs: &[]*string{Ptr("a")}},
				Bytes:      &TestGenericType[[]byte]{Value: []byte{1}, Ptr: &[]byte{1}, SliceOfValues: [][]byte{{1}}, PtrToSliceOfValues: &[][]byte{{1}}, SliceOfPtrs: []*[]byte{{1}}, PtrToSliceOfPtrs: &[]*[]byte{{1}}},
				Slice:      &TestGenericType[[]string]{Value: []string{"a"}, Ptr: &[]string{"a"}, SliceOfValues: [][]string{{"a"}}, PtrToSliceOfValues: &[][]string{{"a"}}, SliceOfPtrs: []*[]string{{"a"}}, PtrToSliceOfPtrs: &[]*[]string{{"a"}}},
				Map:        &TestGenericType[map[string]string]{Value: map[string]string{"k": "a"}, Ptr: &map[string]string{"k": "a"}, SliceOfValues: []map[string]string{{"k": "a"}}, PtrToSliceOfValues: &[]map[string]string{{"k": "a"}}, SliceOfPtrs: []*map[string]string{{"k": "a"}}, PtrToSliceOfPtrs: &[]*map[string]string{{"k": "a"}}},
				Struct:     &TestGenericType[TestStruct]{Value: TestStruct{Value: "a"}, Ptr: &TestStruct{Value: "a"}, SliceOfValues: []TestStruct{{Value: "a"}}, PtrToSliceOfValues: &[]TestStruct{{Value: "a"}}, SliceOfPtrs: []*TestStruct{{Value: "a"}}, PtrToSliceOfPtrs: &[]*TestStruct{{Value: "a"}}},
			},
			want: []*FlagValueData{
				{
					PathNames: []string{},
					Type:      "JSON object",
					Get: &TestBase{
						SkipValue:  1,
						SkipPtr:    Ptr(1),
						Bool:       &TestGenericType[bool]{Value: true, Ptr: Ptr(true), SliceOfValues: []bool{true}, PtrToSliceOfValues: &[]bool{true}, SliceOfPtrs: []*bool{Ptr(true)}, PtrToSliceOfPtrs: &[]*bool{Ptr(true)}},
						Int:        &TestGenericType[int]{Value: 1, Ptr: Ptr(int(1)), SliceOfValues: []int{1}, PtrToSliceOfValues: &[]int{1}, SliceOfPtrs: []*int{Ptr(int(1))}, PtrToSliceOfPtrs: &[]*int{Ptr(int(1))}},
						Int8:       &TestGenericType[int8]{Value: 1, Ptr: Ptr(int8(1)), SliceOfValues: []int8{1}, PtrToSliceOfValues: &[]int8{1}, SliceOfPtrs: []*int8{Ptr(int8(1))}, PtrToSliceOfPtrs: &[]*int8{Ptr(int8(1))}},
						Int16:      &TestGenericType[int16]{Value: 1, Ptr: Ptr(int16(1)), SliceOfValues: []int16{1}, PtrToSliceOfValues: &[]int16{1}, SliceOfPtrs: []*int16{Ptr(int16(1))}, PtrToSliceOfPtrs: &[]*int16{Ptr(int16(1))}},
						Int32:      &TestGenericType[int32]{Value: 1, Ptr: Ptr(int32(1)), SliceOfValues: []int32{1}, PtrToSliceOfValues: &[]int32{1}, SliceOfPtrs: []*int32{Ptr(int32(1))}, PtrToSliceOfPtrs: &[]*int32{Ptr(int32(1))}},
						Int64:      &TestGenericType[int64]{Value: 1, Ptr: Ptr(int64(1)), SliceOfValues: []int64{1}, PtrToSliceOfValues: &[]int64{1}, SliceOfPtrs: []*int64{Ptr(int64(1))}, PtrToSliceOfPtrs: &[]*int64{Ptr(int64(1))}},
						Uint:       &TestGenericType[uint]{Value: 1, Ptr: Ptr(uint(1)), SliceOfValues: []uint{1}, PtrToSliceOfValues: &[]uint{1}, SliceOfPtrs: []*uint{Ptr(uint(1))}, PtrToSliceOfPtrs: &[]*uint{Ptr(uint(1))}},
						Uint8:      &TestGenericType[uint8]{Value: 1, Ptr: Ptr(uint8(1)), SliceOfValues: []uint8{1}, PtrToSliceOfValues: &[]uint8{1}, SliceOfPtrs: []*uint8{Ptr(uint8(1))}, PtrToSliceOfPtrs: &[]*uint8{Ptr(uint8(1))}},
						Uint16:     &TestGenericType[uint16]{Value: 1, Ptr: Ptr(uint16(1)), SliceOfValues: []uint16{1}, PtrToSliceOfValues: &[]uint16{1}, SliceOfPtrs: []*uint16{Ptr(uint16(1))}, PtrToSliceOfPtrs: &[]*uint16{Ptr(uint16(1))}},
						Uint32:     &TestGenericType[uint32]{Value: 1, Ptr: Ptr(uint32(1)), SliceOfValues: []uint32{1}, PtrToSliceOfValues: &[]uint32{1}, SliceOfPtrs: []*uint32{Ptr(uint32(1))}, PtrToSliceOfPtrs: &[]*uint32{Ptr(uint32(1))}},
						Uint64:     &TestGenericType[uint64]{Value: 1, Ptr: Ptr(uint64(1)), SliceOfValues: []uint64{1}, PtrToSliceOfValues: &[]uint64{1}, SliceOfPtrs: []*uint64{Ptr(uint64(1))}, PtrToSliceOfPtrs: &[]*uint64{Ptr(uint64(1))}},
						Float32:    &TestGenericType[float32]{Value: 1, Ptr: Ptr(float32(1)), SliceOfValues: []float32{1}, PtrToSliceOfValues: &[]float32{1}, SliceOfPtrs: []*float32{Ptr(float32(1))}, PtrToSliceOfPtrs: &[]*float32{Ptr(float32(1))}},
						Float64:    &TestGenericType[float64]{Value: 1, Ptr: Ptr(float64(1)), SliceOfValues: []float64{1}, PtrToSliceOfValues: &[]float64{1}, SliceOfPtrs: []*float64{Ptr(float64(1))}, PtrToSliceOfPtrs: &[]*float64{Ptr(float64(1))}},
						Complex64:  &TestGenericType[complex64]{Value: 1, Ptr: Ptr(complex64(1)), SliceOfValues: []complex64{1}, PtrToSliceOfValues: &[]complex64{1}, SliceOfPtrs: []*complex64{Ptr(complex64(1))}, PtrToSliceOfPtrs: &[]*complex64{Ptr(complex64(1))}},
						Complex128: &TestGenericType[complex128]{Value: 1, Ptr: Ptr(complex128(1)), SliceOfValues: []complex128{1}, PtrToSliceOfValues: &[]complex128{1}, SliceOfPtrs: []*complex128{Ptr(complex128(1))}, PtrToSliceOfPtrs: &[]*complex128{Ptr(complex128(1))}},
						String:     &TestGenericType[string]{Value: "a", Ptr: Ptr("a"), SliceOfValues: []string{"a"}, PtrToSliceOfValues: &[]string{"a"}, SliceOfPtrs: []*string{Ptr("a")}, PtrToSliceOfPtrs: &[]*string{Ptr("a")}},
						Bytes:      &TestGenericType[[]byte]{Value: []byte{1}, Ptr: &[]byte{1}, SliceOfValues: [][]byte{{1}}, PtrToSliceOfValues: &[][]byte{{1}}, SliceOfPtrs: []*[]byte{{1}}, PtrToSliceOfPtrs: &[]*[]byte{{1}}},
						Slice:      &TestGenericType[[]string]{Value: []string{"a"}, Ptr: &[]string{"a"}, SliceOfValues: [][]string{{"a"}}, PtrToSliceOfValues: &[][]string{{"a"}}, SliceOfPtrs: []*[]string{{"a"}}, PtrToSliceOfPtrs: &[]*[]string{{"a"}}},
						Map:        &TestGenericType[map[string]string]{Value: map[string]string{"k": "a"}, Ptr: &map[string]string{"k": "a"}, SliceOfValues: []map[string]string{{"k": "a"}}, PtrToSliceOfValues: &[]map[string]string{{"k": "a"}}, SliceOfPtrs: []*map[string]string{{"k": "a"}}, PtrToSliceOfPtrs: &[]*map[string]string{{"k": "a"}}},
						Struct:     &TestGenericType[TestStruct]{Value: TestStruct{Value: "a"}, Ptr: &TestStruct{Value: "a"}, SliceOfValues: []TestStruct{{Value: "a"}}, PtrToSliceOfValues: &[]TestStruct{{Value: "a"}}, SliceOfPtrs: []*TestStruct{{Value: "a"}}, PtrToSliceOfPtrs: &[]*TestStruct{{Value: "a"}}},
					},
					String: ``,
				},
				{PathNames: []string{"Bool"}, Type: "JSON object", Get: &TestGenericType[bool]{Value: true, Ptr: Ptr(true), SliceOfValues: []bool{true}, PtrToSliceOfValues: &[]bool{true}, SliceOfPtrs: []*bool{Ptr(true)}, PtrToSliceOfPtrs: &[]*bool{Ptr(true)}}, String: `{"Value":true,"Ptr":true,"SliceOfValues":[true],"PtrToSliceOfValues":[true],"SliceOfPtrs":[true],"PtrToSliceOfPtrs":[true]}`},
				{PathNames: []string{"Bool", "Value"}, Type: "bool", Get: true, String: `true`},
				{PathNames: []string{"Bool", "Ptr"}, Type: "bool", Get: Ptr(true), String: `true`},
				{PathNames: []string{"Bool", "SliceOfValues"}, Type: "bool (JSON list)", Get: []bool{true}, String: `[true]`},
				{PathNames: []string{"Bool", "PtrToSliceOfValues"}, Type: "bool (JSON list)", Get: &[]bool{true}, String: `[true]`},
				{PathNames: []string{"Bool", "SliceOfPtrs"}, Type: "bool (JSON list)", Get: []*bool{Ptr(true)}, String: `[true]`},
				{PathNames: []string{"Bool", "PtrToSliceOfPtrs"}, Type: "bool (JSON list)", Get: &[]*bool{Ptr(true)}, String: `[true]`},
				{PathNames: []string{"Int"}, Type: "JSON object", Get: &TestGenericType[int]{Value: 1, Ptr: Ptr(int(1)), SliceOfValues: []int{1}, PtrToSliceOfValues: &[]int{1}, SliceOfPtrs: []*int{Ptr(int(1))}, PtrToSliceOfPtrs: &[]*int{Ptr(int(1))}}, String: `{"Value":1,"Ptr":1,"SliceOfValues":[1],"PtrToSliceOfValues":[1],"SliceOfPtrs":[1],"PtrToSliceOfPtrs":[1]}`},
				{PathNames: []string{"Int", "Value"}, Type: "int", Get: int(1), String: `1`},
				{PathNames: []string{"Int", "Ptr"}, Type: "int", Get: Ptr(int(1)), String: `1`},
				{PathNames: []string{"Int", "SliceOfValues"}, Type: "int (JSON list)", Get: []int{1}, String: `[1]`},
				{PathNames: []string{"Int", "PtrToSliceOfValues"}, Type: "int (JSON list)", Get: &[]int{1}, String: `[1]`},
				{PathNames: []string{"Int", "SliceOfPtrs"}, Type: "int (JSON list)", Get: []*int{Ptr(int(1))}, String: `[1]`},
				{PathNames: []string{"Int", "PtrToSliceOfPtrs"}, Type: "int (JSON list)", Get: &[]*int{Ptr(int(1))}, String: `[1]`},
				{PathNames: []string{"Int8"}, Type: "JSON object", Get: &TestGenericType[int8]{Value: 1, Ptr: Ptr(int8(1)), SliceOfValues: []int8{1}, PtrToSliceOfValues: &[]int8{1}, SliceOfPtrs: []*int8{Ptr(int8(1))}, PtrToSliceOfPtrs: &[]*int8{Ptr(int8(1))}}, String: `{"Value":1,"Ptr":1,"SliceOfValues":[1],"PtrToSliceOfValues":[1],"SliceOfPtrs":[1],"PtrToSliceOfPtrs":[1]}`},
				{PathNames: []string{"Int8", "Value"}, Type: "int8", Get: int8(1), String: `1`},
				{PathNames: []string{"Int8", "Ptr"}, Type: "int8", Get: Ptr(int8(1)), String: `1`},
				{PathNames: []string{"Int8", "SliceOfValues"}, Type: "int8 (JSON list)", Get: []int8{1}, String: `[1]`},
				{PathNames: []string{"Int8", "PtrToSliceOfValues"}, Type: "int8 (JSON list)", Get: &[]int8{1}, String: `[1]`},
				{PathNames: []string{"Int8", "SliceOfPtrs"}, Type: "int8 (JSON list)", Get: []*int8{Ptr(int8(1))}, String: `[1]`},
				{PathNames: []string{"Int8", "PtrToSliceOfPtrs"}, Type: "int8 (JSON list)", Get: &[]*int8{Ptr(int8(1))}, String: `[1]`},
				{PathNames: []string{"Int16"}, Type: "JSON object", Get: &TestGenericType[int16]{Value: 1, Ptr: Ptr(int16(1)), SliceOfValues: []int16{1}, PtrToSliceOfValues: &[]int16{1}, SliceOfPtrs: []*int16{Ptr(int16(1))}, PtrToSliceOfPtrs: &[]*int16{Ptr(int16(1))}}, String: `{"Value":1,"Ptr":1,"SliceOfValues":[1],"PtrToSliceOfValues":[1],"SliceOfPtrs":[1],"PtrToSliceOfPtrs":[1]}`},
				{PathNames: []string{"Int16", "Value"}, Type: "int16", Get: int16(1), String: `1`},
				{PathNames: []string{"Int16", "Ptr"}, Type: "int16", Get: Ptr(int16(1)), String: `1`},
				{PathNames: []string{"Int16", "SliceOfValues"}, Type: "int16 (JSON list)", Get: []int16{1}, String: `[1]`},
				{PathNames: []string{"Int16", "PtrToSliceOfValues"}, Type: "int16 (JSON list)", Get: &[]int16{1}, String: `[1]`},
				{PathNames: []string{"Int16", "SliceOfPtrs"}, Type: "int16 (JSON list)", Get: []*int16{Ptr(int16(1))}, String: `[1]`},
				{PathNames: []string{"Int16", "PtrToSliceOfPtrs"}, Type: "int16 (JSON list)", Get: &[]*int16{Ptr(int16(1))}, String: `[1]`},
				{PathNames: []string{"Int32"}, Type: "JSON object", Get: &TestGenericType[int32]{Value: 1, Ptr: Ptr(int32(1)), SliceOfValues: []int32{1}, PtrToSliceOfValues: &[]int32{1}, SliceOfPtrs: []*int32{Ptr(int32(1))}, PtrToSliceOfPtrs: &[]*int32{Ptr(int32(1))}}, String: `{"Value":1,"Ptr":1,"SliceOfValues":[1],"PtrToSliceOfValues":[1],"SliceOfPtrs":[1],"PtrToSliceOfPtrs":[1]}`},
				{PathNames: []string{"Int32", "Value"}, Type: "int32", Get: int32(1), String: `1`},
				{PathNames: []string{"Int32", "Ptr"}, Type: "int32", Get: Ptr(int32(1)), String: `1`},
				{PathNames: []string{"Int32", "SliceOfValues"}, Type: "int32 (JSON list)", Get: []int32{1}, String: `[1]`},
				{PathNames: []string{"Int32", "PtrToSliceOfValues"}, Type: "int32 (JSON list)", Get: &[]int32{1}, String: `[1]`},
				{PathNames: []string{"Int32", "SliceOfPtrs"}, Type: "int32 (JSON list)", Get: []*int32{Ptr(int32(1))}, String: `[1]`},
				{PathNames: []string{"Int32", "PtrToSliceOfPtrs"}, Type: "int32 (JSON list)", Get: &[]*int32{Ptr(int32(1))}, String: `[1]`},
				{PathNames: []string{"Int64"}, Type: "JSON object", Get: &TestGenericType[int64]{Value: 1, Ptr: Ptr(int64(1)), SliceOfValues: []int64{1}, PtrToSliceOfValues: &[]int64{1}, SliceOfPtrs: []*int64{Ptr(int64(1))}, PtrToSliceOfPtrs: &[]*int64{Ptr(int64(1))}}, String: `{"Value":1,"Ptr":1,"SliceOfValues":[1],"PtrToSliceOfValues":[1],"SliceOfPtrs":[1],"PtrToSliceOfPtrs":[1]}`},
				{PathNames: []string{"Int64", "Value"}, Type: "int64", Get: int64(1), String: `1`},
				{PathNames: []string{"Int64", "Ptr"}, Type: "int64", Get: Ptr(int64(1)), String: `1`},
				{PathNames: []string{"Int64", "SliceOfValues"}, Type: "int64 (JSON list)", Get: []int64{1}, String: `[1]`},
				{PathNames: []string{"Int64", "PtrToSliceOfValues"}, Type: "int64 (JSON list)", Get: &[]int64{1}, String: `[1]`},
				{PathNames: []string{"Int64", "SliceOfPtrs"}, Type: "int64 (JSON list)", Get: []*int64{Ptr(int64(1))}, String: `[1]`},
				{PathNames: []string{"Int64", "PtrToSliceOfPtrs"}, Type: "int64 (JSON list)", Get: &[]*int64{Ptr(int64(1))}, String: `[1]`},
				{PathNames: []string{"Uint"}, Type: "JSON object", Get: &TestGenericType[uint]{Value: 1, Ptr: Ptr(uint(1)), SliceOfValues: []uint{1}, PtrToSliceOfValues: &[]uint{1}, SliceOfPtrs: []*uint{Ptr(uint(1))}, PtrToSliceOfPtrs: &[]*uint{Ptr(uint(1))}}, String: `{"Value":1,"Ptr":1,"SliceOfValues":[1],"PtrToSliceOfValues":[1],"SliceOfPtrs":[1],"PtrToSliceOfPtrs":[1]}`},
				{PathNames: []string{"Uint", "Value"}, Type: "uint", Get: uint(1), String: `1`},
				{PathNames: []string{"Uint", "Ptr"}, Type: "uint", Get: Ptr(uint(1)), String: `1`},
				{PathNames: []string{"Uint", "SliceOfValues"}, Type: "uint (JSON list)", Get: []uint{1}, String: `[1]`},
				{PathNames: []string{"Uint", "PtrToSliceOfValues"}, Type: "uint (JSON list)", Get: &[]uint{1}, String: `[1]`},
				{PathNames: []string{"Uint", "SliceOfPtrs"}, Type: "uint (JSON list)", Get: []*uint{Ptr(uint(1))}, String: `[1]`},
				{PathNames: []string{"Uint", "PtrToSliceOfPtrs"}, Type: "uint (JSON list)", Get: &[]*uint{Ptr(uint(1))}, String: `[1]`},
				{PathNames: []string{"Uint8"}, Type: "JSON object", Get: &TestGenericType[uint8]{Value: 1, Ptr: Ptr(uint8(1)), SliceOfValues: []uint8{1}, PtrToSliceOfValues: &[]uint8{1}, SliceOfPtrs: []*uint8{Ptr(uint8(1))}, PtrToSliceOfPtrs: &[]*uint8{Ptr(uint8(1))}}, String: `{"Value":1,"Ptr":1,"SliceOfValues":"AQ==","PtrToSliceOfValues":"AQ==","SliceOfPtrs":[1],"PtrToSliceOfPtrs":[1]}`},
				{PathNames: []string{"Uint8", "Value"}, Type: "uint8", Get: uint8(1), String: `1`},
				{PathNames: []string{"Uint8", "Ptr"}, Type: "uint8", Get: Ptr(uint8(1)), String: `1`},
				{PathNames: []string{"Uint8", "SliceOfValues"}, Type: "base64", Get: []uint8{1}, String: `AQ==`},
				{PathNames: []string{"Uint8", "PtrToSliceOfValues"}, Type: "base64", Get: &[]uint8{1}, String: `AQ==`},
				{PathNames: []string{"Uint8", "SliceOfPtrs"}, Type: "uint8 (JSON list)", Get: []*uint8{Ptr(uint8(1))}, String: `[1]`},
				{PathNames: []string{"Uint8", "PtrToSliceOfPtrs"}, Type: "uint8 (JSON list)", Get: &[]*uint8{Ptr(uint8(1))}, String: `[1]`},
				{PathNames: []string{"Uint16"}, Type: "JSON object", Get: &TestGenericType[uint16]{Value: 1, Ptr: Ptr(uint16(1)), SliceOfValues: []uint16{1}, PtrToSliceOfValues: &[]uint16{1}, SliceOfPtrs: []*uint16{Ptr(uint16(1))}, PtrToSliceOfPtrs: &[]*uint16{Ptr(uint16(1))}}, String: `{"Value":1,"Ptr":1,"SliceOfValues":[1],"PtrToSliceOfValues":[1],"SliceOfPtrs":[1],"PtrToSliceOfPtrs":[1]}`},
				{PathNames: []string{"Uint16", "Value"}, Type: "uint16", Get: uint16(1), String: `1`},
				{PathNames: []string{"Uint16", "Ptr"}, Type: "uint16", Get: Ptr(uint16(1)), String: `1`},
				{PathNames: []string{"Uint16", "SliceOfValues"}, Type: "uint16 (JSON list)", Get: []uint16{1}, String: `[1]`},
				{PathNames: []string{"Uint16", "PtrToSliceOfValues"}, Type: "uint16 (JSON list)", Get: &[]uint16{1}, String: `[1]`},
				{PathNames: []string{"Uint16", "SliceOfPtrs"}, Type: "uint16 (JSON list)", Get: []*uint16{Ptr(uint16(1))}, String: `[1]`},
				{PathNames: []string{"Uint16", "PtrToSliceOfPtrs"}, Type: "uint16 (JSON list)", Get: &[]*uint16{Ptr(uint16(1))}, String: `[1]`},
				{PathNames: []string{"Uint32"}, Type: "JSON object", Get: &TestGenericType[uint32]{Value: 1, Ptr: Ptr(uint32(1)), SliceOfValues: []uint32{1}, PtrToSliceOfValues: &[]uint32{1}, SliceOfPtrs: []*uint32{Ptr(uint32(1))}, PtrToSliceOfPtrs: &[]*uint32{Ptr(uint32(1))}}, String: `{"Value":1,"Ptr":1,"SliceOfValues":[1],"PtrToSliceOfValues":[1],"SliceOfPtrs":[1],"PtrToSliceOfPtrs":[1]}`},
				{PathNames: []string{"Uint32", "Value"}, Type: "uint32", Get: uint32(1), String: `1`},
				{PathNames: []string{"Uint32", "Ptr"}, Type: "uint32", Get: Ptr(uint32(1)), String: `1`},
				{PathNames: []string{"Uint32", "SliceOfValues"}, Type: "uint32 (JSON list)", Get: []uint32{1}, String: `[1]`},
				{PathNames: []string{"Uint32", "PtrToSliceOfValues"}, Type: "uint32 (JSON list)", Get: &[]uint32{1}, String: `[1]`},
				{PathNames: []string{"Uint32", "SliceOfPtrs"}, Type: "uint32 (JSON list)", Get: []*uint32{Ptr(uint32(1))}, String: `[1]`},
				{PathNames: []string{"Uint32", "PtrToSliceOfPtrs"}, Type: "uint32 (JSON list)", Get: &[]*uint32{Ptr(uint32(1))}, String: `[1]`},
				{PathNames: []string{"Uint64"}, Type: "JSON object", Get: &TestGenericType[uint64]{Value: 1, Ptr: Ptr(uint64(1)), SliceOfValues: []uint64{1}, PtrToSliceOfValues: &[]uint64{1}, SliceOfPtrs: []*uint64{Ptr(uint64(1))}, PtrToSliceOfPtrs: &[]*uint64{Ptr(uint64(1))}}, String: `{"Value":1,"Ptr":1,"SliceOfValues":[1],"PtrToSliceOfValues":[1],"SliceOfPtrs":[1],"PtrToSliceOfPtrs":[1]}`},
				{PathNames: []string{"Uint64", "Value"}, Type: "uint64", Get: uint64(1), String: `1`},
				{PathNames: []string{"Uint64", "Ptr"}, Type: "uint64", Get: Ptr(uint64(1)), String: `1`},
				{PathNames: []string{"Uint64", "SliceOfValues"}, Type: "uint64 (JSON list)", Get: []uint64{1}, String: `[1]`},
				{PathNames: []string{"Uint64", "PtrToSliceOfValues"}, Type: "uint64 (JSON list)", Get: &[]uint64{1}, String: `[1]`},
				{PathNames: []string{"Uint64", "SliceOfPtrs"}, Type: "uint64 (JSON list)", Get: []*uint64{Ptr(uint64(1))}, String: `[1]`},
				{PathNames: []string{"Uint64", "PtrToSliceOfPtrs"}, Type: "uint64 (JSON list)", Get: &[]*uint64{Ptr(uint64(1))}, String: `[1]`},
				{PathNames: []string{"Float32"}, Type: "JSON object", Get: &TestGenericType[float32]{Value: 1, Ptr: Ptr(float32(1)), SliceOfValues: []float32{1}, PtrToSliceOfValues: &[]float32{1}, SliceOfPtrs: []*float32{Ptr(float32(1))}, PtrToSliceOfPtrs: &[]*float32{Ptr(float32(1))}}, String: `{"Value":1,"Ptr":1,"SliceOfValues":[1],"PtrToSliceOfValues":[1],"SliceOfPtrs":[1],"PtrToSliceOfPtrs":[1]}`},
				{PathNames: []string{"Float32", "Value"}, Type: "float32", Get: float32(1), String: `1`},
				{PathNames: []string{"Float32", "Ptr"}, Type: "float32", Get: Ptr(float32(1)), String: `1`},
				{PathNames: []string{"Float32", "SliceOfValues"}, Type: "float32 (JSON list)", Get: []float32{1}, String: `[1]`},
				{PathNames: []string{"Float32", "PtrToSliceOfValues"}, Type: "float32 (JSON list)", Get: &[]float32{1}, String: `[1]`},
				{PathNames: []string{"Float32", "SliceOfPtrs"}, Type: "float32 (JSON list)", Get: []*float32{Ptr(float32(1))}, String: `[1]`},
				{PathNames: []string{"Float32", "PtrToSliceOfPtrs"}, Type: "float32 (JSON list)", Get: &[]*float32{Ptr(float32(1))}, String: `[1]`},
				{PathNames: []string{"Float64"}, Type: "JSON object", Get: &TestGenericType[float64]{Value: 1, Ptr: Ptr(float64(1)), SliceOfValues: []float64{1}, PtrToSliceOfValues: &[]float64{1}, SliceOfPtrs: []*float64{Ptr(float64(1))}, PtrToSliceOfPtrs: &[]*float64{Ptr(float64(1))}}, String: `{"Value":1,"Ptr":1,"SliceOfValues":[1],"PtrToSliceOfValues":[1],"SliceOfPtrs":[1],"PtrToSliceOfPtrs":[1]}`},
				{PathNames: []string{"Float64", "Value"}, Type: "float64", Get: float64(1), String: `1`},
				{PathNames: []string{"Float64", "Ptr"}, Type: "float64", Get: Ptr(float64(1)), String: `1`},
				{PathNames: []string{"Float64", "SliceOfValues"}, Type: "float64 (JSON list)", Get: []float64{1}, String: `[1]`},
				{PathNames: []string{"Float64", "PtrToSliceOfValues"}, Type: "float64 (JSON list)", Get: &[]float64{1}, String: `[1]`},
				{PathNames: []string{"Float64", "SliceOfPtrs"}, Type: "float64 (JSON list)", Get: []*float64{Ptr(float64(1))}, String: `[1]`},
				{PathNames: []string{"Float64", "PtrToSliceOfPtrs"}, Type: "float64 (JSON list)", Get: &[]*float64{Ptr(float64(1))}, String: `[1]`},
				{PathNames: []string{"Complex64"}, Type: "JSON object", Get: &TestGenericType[complex64]{Value: 1, Ptr: Ptr(complex64(1)), SliceOfValues: []complex64{1}, PtrToSliceOfValues: &[]complex64{1}, SliceOfPtrs: []*complex64{Ptr(complex64(1))}, PtrToSliceOfPtrs: &[]*complex64{Ptr(complex64(1))}}, String: ``},
				{PathNames: []string{"Complex64", "Value"}, Type: "complex64", Get: complex64(1), String: `(1+0i)`},
				{PathNames: []string{"Complex64", "Ptr"}, Type: "complex64", Get: Ptr(complex64(1)), String: `(1+0i)`},
				{PathNames: []string{"Complex64", "SliceOfValues"}, Type: "complex64 (JSON list)", Get: []complex64{1}, String: `["(1+0i)"]`},
				{PathNames: []string{"Complex64", "PtrToSliceOfValues"}, Type: "complex64 (JSON list)", Get: &[]complex64{1}, String: `["(1+0i)"]`},
				{PathNames: []string{"Complex64", "SliceOfPtrs"}, Type: "complex64 (JSON list)", Get: []*complex64{Ptr(complex64(1))}, String: `["(1+0i)"]`},
				{PathNames: []string{"Complex64", "PtrToSliceOfPtrs"}, Type: "complex64 (JSON list)", Get: &[]*complex64{Ptr(complex64(1))}, String: `["(1+0i)"]`},
				{PathNames: []string{"Complex128"}, Type: "JSON object", Get: &TestGenericType[complex128]{Value: 1, Ptr: Ptr(complex128(1)), SliceOfValues: []complex128{1}, PtrToSliceOfValues: &[]complex128{1}, SliceOfPtrs: []*complex128{Ptr(complex128(1))}, PtrToSliceOfPtrs: &[]*complex128{Ptr(complex128(1))}}, String: ``},
				{PathNames: []string{"Complex128", "Value"}, Type: "complex128", Get: complex128(1), String: `(1+0i)`},
				{PathNames: []string{"Complex128", "Ptr"}, Type: "complex128", Get: Ptr(complex128(1)), String: `(1+0i)`},
				{PathNames: []string{"Complex128", "SliceOfValues"}, Type: "complex128 (JSON list)", Get: []complex128{1}, String: `["(1+0i)"]`},
				{PathNames: []string{"Complex128", "PtrToSliceOfValues"}, Type: "complex128 (JSON list)", Get: &[]complex128{1}, String: `["(1+0i)"]`},
				{PathNames: []string{"Complex128", "SliceOfPtrs"}, Type: "complex128 (JSON list)", Get: []*complex128{Ptr(complex128(1))}, String: `["(1+0i)"]`},
				{PathNames: []string{"Complex128", "PtrToSliceOfPtrs"}, Type: "complex128 (JSON list)", Get: &[]*complex128{Ptr(complex128(1))}, String: `["(1+0i)"]`},
				{PathNames: []string{"String"}, Type: "JSON object", Get: &TestGenericType[string]{Value: "a", Ptr: Ptr("a"), SliceOfValues: []string{"a"}, PtrToSliceOfValues: &[]string{"a"}, SliceOfPtrs: []*string{Ptr("a")}, PtrToSliceOfPtrs: &[]*string{Ptr("a")}}, String: `{"Value":"a","Ptr":"a","SliceOfValues":["a"],"PtrToSliceOfValues":["a"],"SliceOfPtrs":["a"],"PtrToSliceOfPtrs":["a"]}`},
				{PathNames: []string{"String", "Value"}, Type: "string", Get: "a", String: `a`},
				{PathNames: []string{"String", "Ptr"}, Type: "string", Get: Ptr("a"), String: `a`},
				{PathNames: []string{"String", "SliceOfValues"}, Type: "string (JSON list)", Get: []string{"a"}, String: `["a"]`},
				{PathNames: []string{"String", "PtrToSliceOfValues"}, Type: "string (JSON list)", Get: &[]string{"a"}, String: `["a"]`},
				{PathNames: []string{"String", "SliceOfPtrs"}, Type: "string (JSON list)", Get: []*string{Ptr("a")}, String: `["a"]`},
				{PathNames: []string{"String", "PtrToSliceOfPtrs"}, Type: "string (JSON list)", Get: &[]*string{Ptr("a")}, String: `["a"]`},
				{PathNames: []string{"Bytes"}, Type: "JSON object", Get: &TestGenericType[[]byte]{Value: []byte{1}, Ptr: &[]byte{1}, SliceOfValues: [][]byte{{1}}, PtrToSliceOfValues: &[][]byte{{1}}, SliceOfPtrs: []*[]byte{{1}}, PtrToSliceOfPtrs: &[]*[]byte{{1}}}, String: `{"Value":"AQ==","Ptr":"AQ==","SliceOfValues":["AQ=="],"PtrToSliceOfValues":["AQ=="],"SliceOfPtrs":["AQ=="],"PtrToSliceOfPtrs":["AQ=="]}`},
				{PathNames: []string{"Bytes", "Value"}, Type: "base64", Get: []byte{1}, String: `AQ==`},
				{PathNames: []string{"Bytes", "Ptr"}, Type: "base64", Get: &[]byte{1}, String: `AQ==`},
				{PathNames: []string{"Bytes", "SliceOfValues"}, Type: "base64 (JSON list)", Get: [][]byte{{1}}, String: `["AQ=="]`},
				{PathNames: []string{"Bytes", "PtrToSliceOfValues"}, Type: "base64 (JSON list)", Get: &[][]byte{{1}}, String: `["AQ=="]`},
				{PathNames: []string{"Bytes", "SliceOfPtrs"}, Type: "base64 (JSON list)", Get: []*[]byte{{1}}, String: `["AQ=="]`},
				{PathNames: []string{"Bytes", "PtrToSliceOfPtrs"}, Type: "base64 (JSON list)", Get: &[]*[]byte{{1}}, String: `["AQ=="]`},
				{PathNames: []string{"Slice"}, Type: "JSON object", Get: &TestGenericType[[]string]{Value: []string{"a"}, Ptr: &[]string{"a"}, SliceOfValues: [][]string{{"a"}}, PtrToSliceOfValues: &[][]string{{"a"}}, SliceOfPtrs: []*[]string{{"a"}}, PtrToSliceOfPtrs: &[]*[]string{{"a"}}}, String: `{"Value":["a"],"Ptr":["a"],"SliceOfValues":[["a"]],"PtrToSliceOfValues":[["a"]],"SliceOfPtrs":[["a"]],"PtrToSliceOfPtrs":[["a"]]}`},
				{PathNames: []string{"Slice", "Value"}, Type: "string (JSON list)", Get: []string{"a"}, String: `["a"]`},
				{PathNames: []string{"Slice", "Ptr"}, Type: "string (JSON list)", Get: &[]string{"a"}, String: `["a"]`},
				{PathNames: []string{"Slice", "SliceOfValues"}, Type: "JSON list", Get: [][]string{{"a"}}, String: `[["a"]]`},
				{PathNames: []string{"Slice", "PtrToSliceOfValues"}, Type: "JSON list", Get: &[][]string{{"a"}}, String: `[["a"]]`},
				{PathNames: []string{"Slice", "SliceOfPtrs"}, Type: "JSON list", Get: []*[]string{{"a"}}, String: `[["a"]]`},
				{PathNames: []string{"Slice", "PtrToSliceOfPtrs"}, Type: "JSON list", Get: &[]*[]string{{"a"}}, String: `[["a"]]`},
				{PathNames: []string{"Map"}, Type: "JSON object", Get: &TestGenericType[map[string]string]{Value: map[string]string{"k": "a"}, Ptr: &map[string]string{"k": "a"}, SliceOfValues: []map[string]string{{"k": "a"}}, PtrToSliceOfValues: &[]map[string]string{{"k": "a"}}, SliceOfPtrs: []*map[string]string{{"k": "a"}}, PtrToSliceOfPtrs: &[]*map[string]string{{"k": "a"}}}, String: `{"Value":{"k":"a"},"Ptr":{"k":"a"},"SliceOfValues":[{"k":"a"}],"PtrToSliceOfValues":[{"k":"a"}],"SliceOfPtrs":[{"k":"a"}],"PtrToSliceOfPtrs":[{"k":"a"}]}`},
				{PathNames: []string{"Map", "Value"}, Type: "JSON object", Get: map[string]string{"k": "a"}, String: `{"k":"a"}`},
				{PathNames: []string{"Map", "Ptr"}, Type: "JSON object", Get: &map[string]string{"k": "a"}, String: `{"k":"a"}`},
				{PathNames: []string{"Map", "SliceOfValues"}, Type: "JSON object (JSON list)", Get: []map[string]string{{"k": "a"}}, String: `[{"k":"a"}]`},
				{PathNames: []string{"Map", "PtrToSliceOfValues"}, Type: "JSON object (JSON list)", Get: &[]map[string]string{{"k": "a"}}, String: `[{"k":"a"}]`},
				{PathNames: []string{"Map", "SliceOfPtrs"}, Type: "JSON object (JSON list)", Get: []*map[string]string{{"k": "a"}}, String: `[{"k":"a"}]`},
				{PathNames: []string{"Map", "PtrToSliceOfPtrs"}, Type: "JSON object (JSON list)", Get: &[]*map[string]string{{"k": "a"}}, String: `[{"k":"a"}]`},
				{PathNames: []string{"Struct"}, Type: "JSON object", Get: &TestGenericType[TestStruct]{Value: TestStruct{Value: "a"}, Ptr: &TestStruct{Value: "a"}, SliceOfValues: []TestStruct{{Value: "a"}}, PtrToSliceOfValues: &[]TestStruct{{Value: "a"}}, SliceOfPtrs: []*TestStruct{{Value: "a"}}, PtrToSliceOfPtrs: &[]*TestStruct{{Value: "a"}}}, String: `{"Value":{"Value":"a"},"Ptr":{"Value":"a"},"SliceOfValues":[{"Value":"a"}],"PtrToSliceOfValues":[{"Value":"a"}],"SliceOfPtrs":[{"Value":"a"}],"PtrToSliceOfPtrs":[{"Value":"a"}]}`},
				{PathNames: []string{"Struct", "Value"}, Type: "JSON object", Get: TestStruct{Value: "a"}, String: `{"Value":"a"}`},
				{PathNames: []string{"Struct", "Value", "Value"}, Type: "string", Get: "a", String: `a`},
				{PathNames: []string{"Struct", "Ptr"}, Type: "JSON object", Get: &TestStruct{Value: "a"}, String: `{"Value":"a"}`},
				{PathNames: []string{"Struct", "Ptr", "Value"}, Type: "string", Get: "a", String: `a`},
				{PathNames: []string{"Struct", "SliceOfValues"}, Type: "JSON object (JSON list)", Get: []TestStruct{{Value: "a"}}, String: `[{"Value":"a"}]`},
				{PathNames: []string{"Struct", "PtrToSliceOfValues"}, Type: "JSON object (JSON list)", Get: &[]TestStruct{{Value: "a"}}, String: `[{"Value":"a"}]`},
				{PathNames: []string{"Struct", "SliceOfPtrs"}, Type: "JSON object (JSON list)", Get: []*TestStruct{{Value: "a"}}, String: `[{"Value":"a"}]`},
				{PathNames: []string{"Struct", "PtrToSliceOfPtrs"}, Type: "JSON object (JSON list)", Get: &[]*TestStruct{{Value: "a"}}, String: `[{"Value":"a"}]`},
			},
		},
		{
			name: "unexported",
			filters: []jsonflag.FilterFunc{
				func(val *jsonflag.Value) jsonflag.FilterResult { // skip base
					if len(val.Path()) == 0 {
						return jsonflag.SkipAndDescend
					}
					return jsonflag.IncludeAndDescend
				},
			},
			given:     &TestUnexported{unexported: 0},
			postGiven: &TestUnexported{unexported: 0},
			want:      []*FlagValueData(nil),
		},
		{
			name: "unsupported",
			filters: []jsonflag.FilterFunc{
				func(val *jsonflag.Value) jsonflag.FilterResult { // skip base
					if len(val.Path()) == 0 {
						return jsonflag.SkipAndDescend
					}
					return jsonflag.IncludeAndDescend
				},
			},
			given:     &TestGenericType[chan struct{}]{},
			postGiven: &TestGenericType[chan struct{}]{},
			want:      []*FlagValueData(nil),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := jsonflag.Recursive(test.given, test.filters...)
			RequireDataOfFlagValuesEqual(t, test.want, DataOfFlagValues(got))
			require.Equal(t, test.postGiven, test.given)
		})
	}
}

func TestNewFlagValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		given       any
		noFlagValue bool
		encoder     func(any) ([]byte, error)
		decoder     func([]byte, any) error
		pre         *FlagValueData
		setTo       string
		setError    string
		want        *FlagValueData
	}{
		{
			name:        "nil",
			given:       nil,
			noFlagValue: true,
		},
		{
			name:        "non-pointer",
			given:       false,
			noFlagValue: true,
		},
		{
			name:  "bool/empty",
			given: Ptr(false),
			pre:   &FlagValueData{PathNames: []string{}, Type: "bool", Get: Ptr(false), String: ``},
			setTo: "true",
			want:  &FlagValueData{PathNames: []string{}, Type: "bool", Get: Ptr(true), String: `true`},
		},
		{
			name:  "bool/non-empty",
			given: Ptr(true),
			pre:   &FlagValueData{PathNames: []string{}, Type: "bool", Get: Ptr(true), String: `true`},
			setTo: "true",
			want:  &FlagValueData{PathNames: []string{}, Type: "bool", Get: Ptr(true), String: `true`},
		},
		{
			name:     "bool/invalid-value",
			given:    Ptr(false),
			pre:      &FlagValueData{PathNames: []string{}, Type: "bool", Get: Ptr(false), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "bool", Get: Ptr(false), String: ``},
		},
		{
			name:  "int/empty",
			given: Ptr(int(0)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "int", Get: Ptr(int(0)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int", Get: Ptr(int(2)), String: `2`},
		},
		{
			name:  "int/non-empty",
			given: Ptr(int(1)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "int", Get: Ptr(int(1)), String: `1`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int", Get: Ptr(int(2)), String: `2`},
		},
		{
			name:     "int/invalid-value",
			given:    Ptr(int(0)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "int", Get: Ptr(int(0)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "int", Get: Ptr(int(0)), String: ``},
		},
		{
			name:  "int8/empty",
			given: Ptr(int8(0)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "int8", Get: Ptr(int8(0)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int8", Get: Ptr(int8(2)), String: `2`},
		},
		{
			name:  "int8/non-empty",
			given: Ptr(int8(1)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "int8", Get: Ptr(int8(1)), String: `1`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int8", Get: Ptr(int8(2)), String: `2`},
		},
		{
			name:     "int8/invalid-value",
			given:    Ptr(int8(0)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "int8", Get: Ptr(int8(0)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "int8", Get: Ptr(int8(0)), String: ``},
		},
		{
			name:  "int16/empty",
			given: Ptr(int16(0)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "int16", Get: Ptr(int16(0)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int16", Get: Ptr(int16(2)), String: `2`},
		},
		{
			name:  "int16/non-empty",
			given: Ptr(int16(1)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "int16", Get: Ptr(int16(1)), String: `1`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int16", Get: Ptr(int16(2)), String: `2`},
		},
		{
			name:     "int16/invalid-value",
			given:    Ptr(int16(0)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "int16", Get: Ptr(int16(0)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "int16", Get: Ptr(int16(0)), String: ``},
		},
		{
			name:  "int32/empty",
			given: Ptr(int32(0)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "int32", Get: Ptr(int32(0)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int32", Get: Ptr(int32(2)), String: `2`},
		},
		{
			name:  "int32/non-empty",
			given: Ptr(int32(1)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "int32", Get: Ptr(int32(1)), String: `1`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int32", Get: Ptr(int32(2)), String: `2`},
		},
		{
			name:     "int32/invalid-value",
			given:    Ptr(int32(0)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "int32", Get: Ptr(int32(0)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "int32", Get: Ptr(int32(0)), String: ``},
		},
		{
			name:  "int64/empty",
			given: Ptr(int64(0)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "int64", Get: Ptr(int64(0)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int64", Get: Ptr(int64(2)), String: `2`},
		},
		{
			name:  "int64/non-empty",
			given: Ptr(int64(1)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "int64", Get: Ptr(int64(1)), String: `1`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int64", Get: Ptr(int64(2)), String: `2`},
		},
		{
			name:     "int64/invalid-value",
			given:    Ptr(int64(0)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "int64", Get: Ptr(int64(0)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "int64", Get: Ptr(int64(0)), String: ``},
		},
		{
			name:  "uint/empty",
			given: Ptr(uint(0)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "uint", Get: Ptr(uint(0)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "uint", Get: Ptr(uint(2)), String: `2`},
		},
		{
			name:  "uint/non-empty",
			given: Ptr(uint(1)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "uint", Get: Ptr(uint(1)), String: `1`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "uint", Get: Ptr(uint(2)), String: `2`},
		},
		{
			name:     "uint/invalid-value",
			given:    Ptr(uint(0)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "uint", Get: Ptr(uint(0)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "uint", Get: Ptr(uint(0)), String: ``},
		},
		{
			name:  "uint8/empty",
			given: Ptr(uint8(0)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "uint8", Get: Ptr(uint8(0)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "uint8", Get: Ptr(uint8(2)), String: `2`},
		},
		{
			name:  "uint8/non-empty",
			given: Ptr(uint8(1)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "uint8", Get: Ptr(uint8(1)), String: `1`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "uint8", Get: Ptr(uint8(2)), String: `2`},
		},
		{
			name:     "uint8/invalid-value",
			given:    Ptr(uint8(0)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "uint8", Get: Ptr(uint8(0)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "uint8", Get: Ptr(uint8(0)), String: ``},
		},
		{
			name:  "uint16/empty",
			given: Ptr(uint16(0)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "uint16", Get: Ptr(uint16(0)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "uint16", Get: Ptr(uint16(2)), String: `2`},
		},
		{
			name:  "uint16/non-empty",
			given: Ptr(uint16(1)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "uint16", Get: Ptr(uint16(1)), String: `1`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "uint16", Get: Ptr(uint16(2)), String: `2`},
		},
		{
			name:     "uint16/invalid-value",
			given:    Ptr(uint16(0)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "uint16", Get: Ptr(uint16(0)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "uint16", Get: Ptr(uint16(0)), String: ``},
		},
		{
			name:  "uint32/empty",
			given: Ptr(uint32(0)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "uint32", Get: Ptr(uint32(0)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "uint32", Get: Ptr(uint32(2)), String: `2`},
		},
		{
			name:  "uint32/non-empty",
			given: Ptr(uint32(1)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "uint32", Get: Ptr(uint32(1)), String: `1`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "uint32", Get: Ptr(uint32(2)), String: `2`},
		},
		{
			name:     "uint32/invalid-value",
			given:    Ptr(uint32(0)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "uint32", Get: Ptr(uint32(0)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "uint32", Get: Ptr(uint32(0)), String: ``},
		},
		{
			name:  "uint64/empty",
			given: Ptr(uint64(0)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "uint64", Get: Ptr(uint64(0)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "uint64", Get: Ptr(uint64(2)), String: `2`},
		},
		{
			name:  "uint64/non-empty",
			given: Ptr(uint64(1)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "uint64", Get: Ptr(uint64(1)), String: `1`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "uint64", Get: Ptr(uint64(2)), String: `2`},
		},
		{
			name:     "uint64/invalid-value",
			given:    Ptr(uint64(0)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "uint64", Get: Ptr(uint64(0)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "uint64", Get: Ptr(uint64(0)), String: ``},
		},
		{
			name:  "float32/empty",
			given: Ptr(float32(0)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "float32", Get: Ptr(float32(0)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "float32", Get: Ptr(float32(2)), String: `2`},
		},
		{
			name:  "float32/non-empty",
			given: Ptr(float32(1)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "float32", Get: Ptr(float32(1)), String: `1`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "float32", Get: Ptr(float32(2)), String: `2`},
		},
		{
			name:     "float32/invalid-value",
			given:    Ptr(float32(0)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "float32", Get: Ptr(float32(0)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "float32", Get: Ptr(float32(0)), String: ``},
		},
		{
			name:  "float64/empty",
			given: Ptr(float64(0)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "float64", Get: Ptr(float64(0)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "float64", Get: Ptr(float64(2)), String: `2`},
		},
		{
			name:  "float64/non-empty",
			given: Ptr(float64(1)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "float64", Get: Ptr(float64(1)), String: `1`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "float64", Get: Ptr(float64(2)), String: `2`},
		},
		{
			name:     "float64/invalid-value",
			given:    Ptr(float64(0)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "float64", Get: Ptr(float64(0)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "float64", Get: Ptr(float64(0)), String: ``},
		},
		{
			name:  "complex64/empty",
			given: Ptr(complex64(0)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "complex64", Get: Ptr(complex64(0)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "complex64", Get: Ptr(complex64(2)), String: `(2+0i)`},
		},
		{
			name:  "complex64/non-empty",
			given: Ptr(complex64(1)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "complex64", Get: Ptr(complex64(1)), String: `(1+0i)`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "complex64", Get: Ptr(complex64(2)), String: `(2+0i)`},
		},
		{
			name:     "complex64/invalid-value",
			given:    Ptr(complex64(0)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "complex64", Get: Ptr(complex64(0)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "complex64", Get: Ptr(complex64(0)), String: ``},
		},
		{
			name:  "complex128/empty",
			given: Ptr(complex128(0)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "complex128", Get: Ptr(complex128(0)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "complex128", Get: Ptr(complex128(2)), String: `(2+0i)`},
		},
		{
			name:  "complex128/non-empty",
			given: Ptr(complex128(1)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "complex128", Get: Ptr(complex128(1)), String: `(1+0i)`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "complex128", Get: Ptr(complex128(2)), String: `(2+0i)`},
		},
		{
			name:     "complex128/invalid-value",
			given:    Ptr(complex128(0)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "complex128", Get: Ptr(complex128(0)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "complex128", Get: Ptr(complex128(0)), String: ``},
		},
		{
			name:  "string/empty",
			given: Ptr(""),
			pre:   &FlagValueData{PathNames: []string{}, Type: "string", Get: Ptr(""), String: ``},
			setTo: "a",
			want:  &FlagValueData{PathNames: []string{}, Type: "string", Get: Ptr("a"), String: `a`},
		},
		{
			name:  "string/non-empty",
			given: Ptr("z"),
			pre:   &FlagValueData{PathNames: []string{}, Type: "string", Get: Ptr("z"), String: `z`},
			setTo: "a",
			want:  &FlagValueData{PathNames: []string{}, Type: "string", Get: Ptr("a"), String: `a`},
		},
		{
			name:  "base64/empty",
			given: Ptr([]byte(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "base64", Get: Ptr([]byte(nil)), String: ``},
			setTo: "Ag==",
			want:  &FlagValueData{PathNames: []string{}, Type: "base64", Get: Ptr([]byte{2}), String: `Ag==`},
		},
		{
			name:  "base64/non-empty",
			given: Ptr([]byte{1}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "base64", Get: Ptr([]byte{1}), String: `AQ==`},
			setTo: "Ag==",
			want:  &FlagValueData{PathNames: []string{}, Type: "base64", Get: Ptr([]byte{2}), String: `Ag==`},
		},
		{
			name:     "base64/invalid-value",
			given:    Ptr([]byte(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "base64", Get: Ptr([]byte(nil)), String: ``},
			setTo:    "invalid",
			setError: "illegal base64 data",
			want:     &FlagValueData{PathNames: []string{}, Type: "base64", Get: Ptr([]byte(nil)), String: ``},
		},
		{
			name:  "map/empty",
			given: Ptr(map[string]string(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "JSON object", Get: Ptr(map[string]string(nil)), String: ``},
			setTo: `{"k":"a"}`,
			want:  &FlagValueData{PathNames: []string{}, Type: "JSON object", Get: Ptr(map[string]string{"k": "a"}), String: `{"k":"a"}`},
		},
		{
			name:  "map/non-empty",
			given: Ptr(map[string]string{"w": "z"}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "JSON object", Get: Ptr(map[string]string{"w": "z"}), String: `{"w":"z"}`},
			setTo: `{"k":"a"}`,
			want:  &FlagValueData{PathNames: []string{}, Type: "JSON object", Get: Ptr(map[string]string{"k": "a"}), String: `{"k":"a"}`},
		},
		{
			name:     "map/invalid-value",
			given:    Ptr(map[string]string(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "JSON object", Get: Ptr(map[string]string(nil)), String: ``},
			setTo:    `invalid`,
			setError: "invalid character",
			want:     &FlagValueData{PathNames: []string{}, Type: "JSON object", Get: Ptr(map[string]string(nil)), String: ``},
		},
		{
			name:  "struct/empty",
			given: &TestBase{},
			pre:   &FlagValueData{PathNames: []string{}, Type: "JSON object", Get: &TestBase{}, String: ``},
			setTo: `{"String":{"Value":"a"}}`,
			want:  &FlagValueData{PathNames: []string{}, Type: "JSON object", Get: &TestBase{String: &TestGenericType[string]{Value: "a"}}, String: `{"String":{"Value":"a"}}`},
		},
		{
			name:  "struct/non-empty",
			given: &TestBase{Bool: &TestGenericType[bool]{Value: true}},
			pre:   &FlagValueData{PathNames: []string{}, Type: "JSON object", Get: &TestBase{Bool: &TestGenericType[bool]{Value: true}}, String: `{"Bool":{"Value":true}}`},
			setTo: `{"String":{"Value":"a"}}`,
			want:  &FlagValueData{PathNames: []string{}, Type: "JSON object", Get: &TestBase{String: &TestGenericType[string]{Value: "a"}}, String: `{"String":{"Value":"a"}}`},
		},
		{
			name:     "struct/invalid-value",
			given:    &TestBase{},
			pre:      &FlagValueData{PathNames: []string{}, Type: "JSON object", Get: &TestBase{}, String: ``},
			setTo:    `invalid`,
			setError: "invalid character",
			want:     &FlagValueData{PathNames: []string{}, Type: "JSON object", Get: &TestBase{}, String: ``},
		},
		{
			name:  "sliceOfBools/empty",
			given: Ptr([]bool(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "bool (JSON list)", Get: Ptr([]bool(nil)), String: ``},
			setTo: "true",
			want:  &FlagValueData{PathNames: []string{}, Type: "bool (JSON list)", Get: Ptr([]bool{true}), String: `[true]`},
		},
		{
			name:  "sliceOfBools/non-empty",
			given: Ptr([]bool{true}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "bool (JSON list)", Get: Ptr([]bool{true}), String: `[true]`},
			setTo: "true",
			want:  &FlagValueData{PathNames: []string{}, Type: "bool (JSON list)", Get: Ptr([]bool{true, true}), String: `[true,true]`},
		},
		{
			name:     "sliceOfBools/invalid-value",
			given:    Ptr([]bool(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "bool (JSON list)", Get: Ptr([]bool(nil)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "bool (JSON list)", Get: Ptr([]bool(nil)), String: ``},
		},
		{
			name:  "sliceOfInts/empty",
			given: Ptr([]int(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "int (JSON list)", Get: Ptr([]int(nil)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int (JSON list)", Get: Ptr([]int{2}), String: `[2]`},
		},
		{
			name:  "sliceOfInts/non-empty",
			given: Ptr([]int{1}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "int (JSON list)", Get: Ptr([]int{1}), String: `[1]`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int (JSON list)", Get: Ptr([]int{1, 2}), String: `[1,2]`},
		},
		{
			name:     "sliceOfInts/invalid-value",
			given:    Ptr([]int(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "int (JSON list)", Get: Ptr([]int(nil)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "int (JSON list)", Get: Ptr([]int(nil)), String: ``},
		},
		{
			name:  "sliceOfInt8s/empty",
			given: Ptr([]int8(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "int8 (JSON list)", Get: Ptr([]int8(nil)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int8 (JSON list)", Get: Ptr([]int8{2}), String: `[2]`},
		},
		{
			name:  "sliceOfInt8s/non-empty",
			given: Ptr([]int8{1}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "int8 (JSON list)", Get: Ptr([]int8{1}), String: `[1]`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int8 (JSON list)", Get: Ptr([]int8{1, 2}), String: `[1,2]`},
		},
		{
			name:     "sliceOfInt8s/invalid-value",
			given:    Ptr([]int8(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "int8 (JSON list)", Get: Ptr([]int8(nil)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "int8 (JSON list)", Get: Ptr([]int8(nil)), String: ``},
		},
		{
			name:  "sliceOfInt16s/empty",
			given: Ptr([]int16(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "int16 (JSON list)", Get: Ptr([]int16(nil)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int16 (JSON list)", Get: Ptr([]int16{2}), String: `[2]`},
		},
		{
			name:  "sliceOfInt16s/non-empty",
			given: Ptr([]int16{1}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "int16 (JSON list)", Get: Ptr([]int16{1}), String: `[1]`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int16 (JSON list)", Get: Ptr([]int16{1, 2}), String: `[1,2]`},
		},
		{
			name:     "sliceOfInt16s/invalid-value",
			given:    Ptr([]int16(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "int16 (JSON list)", Get: Ptr([]int16(nil)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "int16 (JSON list)", Get: Ptr([]int16(nil)), String: ``},
		},
		{
			name:  "sliceOfInt32s/empty",
			given: Ptr([]int32(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "int32 (JSON list)", Get: Ptr([]int32(nil)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int32 (JSON list)", Get: Ptr([]int32{2}), String: `[2]`},
		},
		{
			name:  "sliceOfInt32s/non-empty",
			given: Ptr([]int32{1}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "int32 (JSON list)", Get: Ptr([]int32{1}), String: `[1]`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int32 (JSON list)", Get: Ptr([]int32{1, 2}), String: `[1,2]`},
		},
		{
			name:     "sliceOfInt32s/invalid-value",
			given:    Ptr([]int32(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "int32 (JSON list)", Get: Ptr([]int32(nil)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "int32 (JSON list)", Get: Ptr([]int32(nil)), String: ``},
		},
		{
			name:  "sliceOfInt64s/empty",
			given: Ptr([]int64(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "int64 (JSON list)", Get: Ptr([]int64(nil)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int64 (JSON list)", Get: Ptr([]int64{2}), String: `[2]`},
		},
		{
			name:  "sliceOfInt64s/non-empty",
			given: Ptr([]int64{1}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "int64 (JSON list)", Get: Ptr([]int64{1}), String: `[1]`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int64 (JSON list)", Get: Ptr([]int64{1, 2}), String: `[1,2]`},
		},
		{
			name:     "sliceOfInt64s/invalid-value",
			given:    Ptr([]int64(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "int64 (JSON list)", Get: Ptr([]int64(nil)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "int64 (JSON list)", Get: Ptr([]int64(nil)), String: ``},
		},
		{
			name:  "sliceOfUints/empty",
			given: Ptr([]uint(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "uint (JSON list)", Get: Ptr([]uint(nil)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "uint (JSON list)", Get: Ptr([]uint{2}), String: `[2]`},
		},
		{
			name:  "sliceOfUints/non-empty",
			given: Ptr([]uint{1}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "uint (JSON list)", Get: Ptr([]uint{1}), String: `[1]`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "uint (JSON list)", Get: Ptr([]uint{1, 2}), String: `[1,2]`},
		},
		{
			name:     "sliceOfUints/invalid-value",
			given:    Ptr([]uint(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "uint (JSON list)", Get: Ptr([]uint(nil)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "uint (JSON list)", Get: Ptr([]uint(nil)), String: ``},
		},
		{
			name:  "sliceOfUint8s/empty",
			given: Ptr([]*uint8(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "uint8 (JSON list)", Get: Ptr([]*uint8(nil)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "uint8 (JSON list)", Get: Ptr([]*uint8{Ptr(uint8(2))}), String: `[2]`},
		},
		{
			name:  "sliceOfUint8s/non-empty",
			given: Ptr([]*uint8{Ptr(uint8(1))}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "uint8 (JSON list)", Get: Ptr([]*uint8{Ptr(uint8(1))}), String: `[1]`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "uint8 (JSON list)", Get: Ptr([]*uint8{Ptr(uint8(1)), Ptr(uint8(2))}), String: `[1,2]`},
		},
		{
			name:     "sliceOfUint8s/invalid-value",
			given:    Ptr([]*uint8(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "uint8 (JSON list)", Get: Ptr([]*uint8(nil)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "uint8 (JSON list)", Get: Ptr([]*uint8(nil)), String: ``},
		},
		{
			name:  "sliceOfUint16s/empty",
			given: Ptr([]uint16(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "uint16 (JSON list)", Get: Ptr([]uint16(nil)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "uint16 (JSON list)", Get: Ptr([]uint16{2}), String: `[2]`},
		},
		{
			name:  "sliceOfUint16s/non-empty",
			given: Ptr([]uint16{1}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "uint16 (JSON list)", Get: Ptr([]uint16{1}), String: `[1]`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "uint16 (JSON list)", Get: Ptr([]uint16{1, 2}), String: `[1,2]`},
		},
		{
			name:     "sliceOfUint16s/invalid-value",
			given:    Ptr([]uint16(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "uint16 (JSON list)", Get: Ptr([]uint16(nil)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "uint16 (JSON list)", Get: Ptr([]uint16(nil)), String: ``},
		},
		{
			name:  "sliceOfUint32s/empty",
			given: Ptr([]uint32(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "uint32 (JSON list)", Get: Ptr([]uint32(nil)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "uint32 (JSON list)", Get: Ptr([]uint32{2}), String: `[2]`},
		},
		{
			name:  "sliceOfUint32s/non-empty",
			given: Ptr([]uint32{1}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "uint32 (JSON list)", Get: Ptr([]uint32{1}), String: `[1]`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "uint32 (JSON list)", Get: Ptr([]uint32{1, 2}), String: `[1,2]`},
		},
		{
			name:     "sliceOfUint32s/invalid-value",
			given:    Ptr([]uint32(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "uint32 (JSON list)", Get: Ptr([]uint32(nil)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "uint32 (JSON list)", Get: Ptr([]uint32(nil)), String: ``},
		},
		{
			name:  "sliceOfUint64s/empty",
			given: Ptr([]uint64(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "uint64 (JSON list)", Get: Ptr([]uint64(nil)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "uint64 (JSON list)", Get: Ptr([]uint64{2}), String: `[2]`},
		},
		{
			name:  "sliceOfUint64s/non-empty",
			given: Ptr([]uint64{1}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "uint64 (JSON list)", Get: Ptr([]uint64{1}), String: `[1]`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "uint64 (JSON list)", Get: Ptr([]uint64{1, 2}), String: `[1,2]`},
		},
		{
			name:     "sliceOfUint64s/invalid-value",
			given:    Ptr([]uint64(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "uint64 (JSON list)", Get: Ptr([]uint64(nil)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "uint64 (JSON list)", Get: Ptr([]uint64(nil)), String: ``},
		},
		{
			name:  "sliceOfFloat32s/empty",
			given: Ptr([]float32(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "float32 (JSON list)", Get: Ptr([]float32(nil)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "float32 (JSON list)", Get: Ptr([]float32{2}), String: `[2]`},
		},
		{
			name:  "sliceOfFloat32s/non-empty",
			given: Ptr([]float32{1}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "float32 (JSON list)", Get: Ptr([]float32{1}), String: `[1]`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "float32 (JSON list)", Get: Ptr([]float32{1, 2}), String: `[1,2]`},
		},
		{
			name:     "sliceOfFloat32s/invalid-value",
			given:    Ptr([]float32(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "float32 (JSON list)", Get: Ptr([]float32(nil)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "float32 (JSON list)", Get: Ptr([]float32(nil)), String: ``},
		},
		{
			name:  "sliceOfFloat64s/empty",
			given: Ptr([]float64(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "float64 (JSON list)", Get: Ptr([]float64(nil)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "float64 (JSON list)", Get: Ptr([]float64{2}), String: `[2]`},
		},
		{
			name:  "sliceOfFloat64s/non-empty",
			given: Ptr([]float64{1}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "float64 (JSON list)", Get: Ptr([]float64{1}), String: `[1]`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "float64 (JSON list)", Get: Ptr([]float64{1, 2}), String: `[1,2]`},
		},
		{
			name:     "sliceOfFloat64s/invalid-value",
			given:    Ptr([]float64(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "float64 (JSON list)", Get: Ptr([]float64(nil)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "float64 (JSON list)", Get: Ptr([]float64(nil)), String: ``},
		},
		{
			name:  "sliceOfComplex64s/empty",
			given: Ptr([]complex64(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "complex64 (JSON list)", Get: Ptr([]complex64(nil)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "complex64 (JSON list)", Get: Ptr([]complex64{2}), String: `["(2+0i)"]`},
		},
		{
			name:  "sliceOfComplex64s/non-empty",
			given: Ptr([]complex64{1}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "complex64 (JSON list)", Get: Ptr([]complex64{1}), String: `["(1+0i)"]`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "complex64 (JSON list)", Get: Ptr([]complex64{1, 2}), String: `["(1+0i)","(2+0i)"]`},
		},
		{
			name:     "sliceOfComplex64s/invalid-value",
			given:    Ptr([]complex64(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "complex64 (JSON list)", Get: Ptr([]complex64(nil)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "complex64 (JSON list)", Get: Ptr([]complex64(nil)), String: ``},
		},
		{
			name:  "sliceOfComplex128s/empty",
			given: Ptr([]complex128(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "complex128 (JSON list)", Get: Ptr([]complex128(nil)), String: ``},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "complex128 (JSON list)", Get: Ptr([]complex128{2}), String: `["(2+0i)"]`},
		},
		{
			name:  "sliceOfComplex128s/non-empty",
			given: Ptr([]complex128{1}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "complex128 (JSON list)", Get: Ptr([]complex128{1}), String: `["(1+0i)"]`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "complex128 (JSON list)", Get: Ptr([]complex128{1, 2}), String: `["(1+0i)","(2+0i)"]`},
		},
		{
			name:     "sliceOfComplex128s/invalid-value",
			given:    Ptr([]complex128(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "complex128 (JSON list)", Get: Ptr([]complex128(nil)), String: ``},
			setTo:    "invalid",
			setError: "invalid syntax",
			want:     &FlagValueData{PathNames: []string{}, Type: "complex128 (JSON list)", Get: Ptr([]complex128(nil)), String: ``},
		},
		{
			name:  "sliceOfStrings/empty",
			given: Ptr([]string(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "string (JSON list)", Get: Ptr([]string(nil)), String: ``},
			setTo: "a",
			want:  &FlagValueData{PathNames: []string{}, Type: "string (JSON list)", Get: Ptr([]string{"a"}), String: `["a"]`},
		},
		{
			name:  "sliceOfStrings/non-empty",
			given: Ptr([]string{"z"}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "string (JSON list)", Get: Ptr([]string{"z"}), String: `["z"]`},
			setTo: "a",
			want:  &FlagValueData{PathNames: []string{}, Type: "string (JSON list)", Get: Ptr([]string{"z", "a"}), String: `["z","a"]`},
		},
		{
			name:  "sliceOfBase64s/empty",
			given: Ptr([][]byte(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "base64 (JSON list)", Get: Ptr([][]byte(nil)), String: ``},
			setTo: "Ag==",
			want:  &FlagValueData{PathNames: []string{}, Type: "base64 (JSON list)", Get: Ptr([][]byte{{2}}), String: `["Ag=="]`},
		},
		{
			name:  "sliceOfBase64s/non-empty",
			given: Ptr([][]byte{{1}}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "base64 (JSON list)", Get: Ptr([][]byte{{1}}), String: `["AQ=="]`},
			setTo: "Ag==",
			want:  &FlagValueData{PathNames: []string{}, Type: "base64 (JSON list)", Get: Ptr([][]byte{{1}, {2}}), String: `["AQ==","Ag=="]`},
		},
		{
			name:     "sliceOfBase64s/invalid-value",
			given:    Ptr([][]byte(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "base64 (JSON list)", Get: Ptr([][]byte(nil)), String: ``},
			setTo:    "invalid",
			setError: "illegal base64 data",
			want:     &FlagValueData{PathNames: []string{}, Type: "base64 (JSON list)", Get: Ptr([][]byte(nil)), String: ``},
		},
		{
			name:  "sliceOfMaps/empty",
			given: Ptr([]map[string]string(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "JSON object (JSON list)", Get: Ptr([]map[string]string(nil)), String: ``},
			setTo: `{"k":"a"}`,
			want:  &FlagValueData{PathNames: []string{}, Type: "JSON object (JSON list)", Get: Ptr([]map[string]string{{"k": "a"}}), String: `[{"k":"a"}]`},
		},
		{
			name:  "sliceOfMaps/non-empty",
			given: Ptr([]map[string]string{{"w": "z"}}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "JSON object (JSON list)", Get: Ptr([]map[string]string{{"w": "z"}}), String: `[{"w":"z"}]`},
			setTo: `{"k":"a"}`,
			want:  &FlagValueData{PathNames: []string{}, Type: "JSON object (JSON list)", Get: Ptr([]map[string]string{{"w": "z"}, {"k": "a"}}), String: `[{"w":"z"},{"k":"a"}]`},
		},
		{
			name:     "sliceOfMaps/invalid-value",
			given:    Ptr([]map[string]string(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "JSON object (JSON list)", Get: Ptr([]map[string]string(nil)), String: ``},
			setTo:    "invalid",
			setError: "invalid character",
			want:     &FlagValueData{PathNames: []string{}, Type: "JSON object (JSON list)", Get: Ptr([]map[string]string(nil)), String: ``},
		},
		{
			name:  "sliceOfStructs/empty",
			given: Ptr([]*TestBase(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "JSON object (JSON list)", Get: Ptr([]*TestBase(nil)), String: ``},
			setTo: `{"String":{"Value":"a"}}`,
			want:  &FlagValueData{PathNames: []string{}, Type: "JSON object (JSON list)", Get: Ptr([]*TestBase{{String: &TestGenericType[string]{Value: "a"}}}), String: `[{"String":{"Value":"a"}}]`},
		},
		{
			name:  "sliceOfStructs/non-empty",
			given: Ptr([]*TestBase{{Bool: &TestGenericType[bool]{Value: true}}}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "JSON object (JSON list)", Get: Ptr([]*TestBase{{Bool: &TestGenericType[bool]{Value: true}}}), String: `[{"Bool":{"Value":true}}]`},
			setTo: `{"String":{"Value":"a"}}`,
			want:  &FlagValueData{PathNames: []string{}, Type: "JSON object (JSON list)", Get: Ptr([]*TestBase{{Bool: &TestGenericType[bool]{Value: true}}, {String: &TestGenericType[string]{Value: "a"}}}), String: `[{"Bool":{"Value":true}},{"String":{"Value":"a"}}]`},
		},
		{
			name:     "sliceOfStructs/invalid-value",
			given:    Ptr([]*TestBase(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "JSON object (JSON list)", Get: Ptr([]*TestBase(nil)), String: ``},
			setTo:    "invalid",
			setError: "invalid character",
			want:     &FlagValueData{PathNames: []string{}, Type: "JSON object (JSON list)", Get: Ptr([]*TestBase(nil)), String: ``},
		},
		{
			name:  "sliceOfSlicesOfStrings/empty",
			given: Ptr([][]string(nil)),
			pre:   &FlagValueData{PathNames: []string{}, Type: "JSON list", Get: Ptr([][]string(nil)), String: ``},
			setTo: `["a"]`,
			want:  &FlagValueData{PathNames: []string{}, Type: "JSON list", Get: Ptr([][]string{{"a"}}), String: `[["a"]]`},
		},
		{
			name:  "sliceOfSlicesOfStrings/non-empty",
			given: Ptr([][]string{{"z"}}),
			pre:   &FlagValueData{PathNames: []string{}, Type: "JSON list", Get: Ptr([][]string{{"z"}}), String: `[["z"]]`},
			setTo: `["a"]`,
			want:  &FlagValueData{PathNames: []string{}, Type: "JSON list", Get: Ptr([][]string{{"z"}, {"a"}}), String: `[["z"],["a"]]`},
		},
		{
			name:     "sliceOfSlicesOfStrings/invalid-value",
			given:    Ptr([][]string(nil)),
			pre:      &FlagValueData{PathNames: []string{}, Type: "JSON list", Get: Ptr([][]string(nil)), String: ``},
			setTo:    "invalid",
			setError: "invalid character",
			want:     &FlagValueData{PathNames: []string{}, Type: "JSON list", Get: Ptr([][]string(nil)), String: ``},
		},
		{
			name:    "customEncoder/value",
			given:   Ptr(int(1)),
			encoder: func(x any) ([]byte, error) { return strconv.AppendBool(nil, *(x.(*int)) > 1), nil },
			pre:     &FlagValueData{PathNames: []string{}, Type: "int", Get: Ptr(int(1)), String: `false`},
			setTo:   "2",
			want:    &FlagValueData{PathNames: []string{}, Type: "int", Get: Ptr(int(2)), String: `true`},
		},
		{
			name:    "customEncoder/error",
			given:   Ptr(int(1)),
			encoder: func(_ any) ([]byte, error) { return nil, errors.New("encoding error") },
			pre:     &FlagValueData{PathNames: []string{}, Type: "int", Get: Ptr(int(1)), String: ``},
			setTo:   "2",
			want:    &FlagValueData{PathNames: []string{}, Type: "int", Get: Ptr(int(2)), String: ``},
		},
		{
			name:  "customDecoder/value",
			given: Ptr(int(1)),
			decoder: func(_ []byte, x any) error {
				*(x.(*int)) = 3
				return nil
			},
			pre:   &FlagValueData{PathNames: []string{}, Type: "int", Get: Ptr(int(1)), String: `1`},
			setTo: "2",
			want:  &FlagValueData{PathNames: []string{}, Type: "int", Get: Ptr(int(3)), String: `3`},
		},
		{
			name:     "customDecoder/error",
			given:    Ptr(int(1)),
			decoder:  func(_ []byte, _ any) error { return errors.New("decoding error") },
			pre:      &FlagValueData{PathNames: []string{}, Type: "int", Get: Ptr(int(1)), String: `1`},
			setTo:    "2",
			setError: "decoding error",
			want:     &FlagValueData{PathNames: []string{}, Type: "int", Get: Ptr(int(1)), String: `1`},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			val := jsonflag.New(test.given)
			if test.noFlagValue {
				require.Nil(t, val)
				return
			}

			require.NotNil(t, val)
			if test.encoder != nil {
				val.SetEncoder(test.encoder)
			}
			if test.decoder != nil {
				val.SetDecoder(test.decoder)
			}

			RequireDataOfFlagValueEqual(t, test.pre, DataOfFlagValue(val))

			err := val.Set(test.setTo)
			if test.setError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.setError)
			} else {
				require.NoError(t, err)
			}

			RequireDataOfFlagValueEqual(t, test.want, DataOfFlagValue(val))
			require.Equal(t, test.want.Get, test.given)
		})
	}
}

func TestZeroFlagValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value *jsonflag.Value
	}{
		{name: "nil-pointer", value: nil},
		{name: "zero-value", value: &jsonflag.Value{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.value.SetEncoder(nil)
			test.value.SetDecoder(nil)
			require.Zero(t, test.value.Path())
			require.Zero(t, test.value.Type()) //nolint:testifylint // use require.Zero for code self-similarity
			require.Zero(t, test.value.Get())
			require.Zero(t, test.value.String()) //nolint:testifylint // use require.Zero for code self-similarity
			require.Zero(t, test.value.Set(""))  //nolint:testifylint // use require.Zero for code self-similarity
			require.Zero(t, test.value.IsBoolFlag())
		})
	}
}
