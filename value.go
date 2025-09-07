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

package jsonflag

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// New returns new flag value for the provided value. It returns nil if the value cannot be used as flag value.
func New(base any) *Value {
	if base == nil {
		return nil
	}
	v := reflect.ValueOf(base)
	if !v.CanSet() && (v.Kind() != reflect.Pointer || v.IsNil()) {
		return nil
	}
	return newValue(v, nil, nil)
}

// FilterFunc is a function that can be used to decide if a flag value should be included in recursive results as well as if flag values finding should descend into the sub-values.
type FilterFunc func(*Value) FilterResult

// FilterResult is a bit field holding filtering decision.
type FilterResult uint32

const (
	IncludeAndDescend FilterResult = 0b00 // indicates that the given flag value should be included and values finding should descend into sub-values
	IncludeNoDescend  FilterResult = 0b01 // indicates that the given flag value should be included, but values finding should NOT descend into sub-values
	SkipAndDescend    FilterResult = 0b10 // indicates that the given flag value should be skipped, but values finding should descend into sub-values
	SkipNoDescend     FilterResult = 0b11 // indicates that the given flag value should be skipped and values finding should NOT descend into sub-values
	noDescendMask     FilterResult = 0b01
	skipMask          FilterResult = 0b10
)

// Filter applies all filter function to the provided flag value and returns filtering decision.
func Filter(val *Value, filters ...FilterFunc) (r FilterResult) {
	if val == nil {
		return SkipNoDescend
	}
	for _, cond := range filters {
		r |= cond(val)
	}
	return r
}

// Recursive returns set of flag values for the provided value and all values within, recursively, according to the provided filters. Function silently skips all the values that cannot be used as flag values.
func Recursive(base any, filters ...FilterFunc) []*Value {
	if base == nil {
		return nil
	}
	v := reflect.ValueOf(base)
	if !v.CanSet() && (v.Kind() != reflect.Pointer || v.IsNil()) {
		return nil
	}
	return recursive(v, nil, nil, filters)
}

func recursive(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField, filters []FilterFunc) []*Value {
	values := []*Value(nil)
	val := newValue(base, slices.Clone(fieldsIndexes), slices.Clone(fields))
	filterResult := Filter(val, filters...)
	if filterResult&skipMask == 0 {
		values = append(values, val)
	}
	if filterResult&noDescendMask != 0 {
		return values
	}

	t := base.Type()
	if len(fields) > 0 {
		t = fields[len(fields)-1].Type
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return values
	}
	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		values = append(values, recursive(base, append(fieldsIndexes, i), append(fields, field), filters)...)
	}
	return values
}

func newValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	t := base.Type()
	if len(fields) > 0 {
		t = fields[len(fields)-1].Type
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() { //nolint:exhaustive // cases for only supported types
	case reflect.Bool:
		return newBoolValue(base, fieldsIndexes, fields)
	case reflect.Int:
		return newIntValue(base, fieldsIndexes, fields)
	case reflect.Int8:
		return newInt8Value(base, fieldsIndexes, fields)
	case reflect.Int16:
		return newInt16Value(base, fieldsIndexes, fields)
	case reflect.Int32:
		return newInt32Value(base, fieldsIndexes, fields)
	case reflect.Int64:
		return newInt64Value(base, fieldsIndexes, fields)
	case reflect.Uint:
		return newUintValue(base, fieldsIndexes, fields)
	case reflect.Uint8:
		return newUint8Value(base, fieldsIndexes, fields)
	case reflect.Uint16:
		return newUint16Value(base, fieldsIndexes, fields)
	case reflect.Uint32:
		return newUint32Value(base, fieldsIndexes, fields)
	case reflect.Uint64:
		return newUint64Value(base, fieldsIndexes, fields)
	case reflect.Float32:
		return newFloat32Value(base, fieldsIndexes, fields)
	case reflect.Float64:
		return newFloat64Value(base, fieldsIndexes, fields)
	case reflect.Complex64:
		return newComplex64Value(base, fieldsIndexes, fields)
	case reflect.Complex128:
		return newComplex128Value(base, fieldsIndexes, fields)
	case reflect.String:
		return newStringValue(base, fieldsIndexes, fields)
	case reflect.Slice:
		return newSliceValue(t, base, fieldsIndexes, fields)
	case reflect.Map:
		return newMapValue(base, fieldsIndexes, fields)
	case reflect.Struct:
		return newStructValue(base, fieldsIndexes, fields)
	}
	return nil
}

func newSliceValue(typ reflect.Type, base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	t := typ.Elem()
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() { //nolint:exhaustive // cases for only supported types
	case reflect.Bool:
		return newBoolSliceValue(base, fieldsIndexes, fields)
	case reflect.Int:
		return newIntSliceValue(base, fieldsIndexes, fields)
	case reflect.Int8:
		return newInt8SliceValue(base, fieldsIndexes, fields)
	case reflect.Int16:
		return newInt16SliceValue(base, fieldsIndexes, fields)
	case reflect.Int32:
		return newInt32SliceValue(base, fieldsIndexes, fields)
	case reflect.Int64:
		return newInt64SliceValue(base, fieldsIndexes, fields)
	case reflect.Uint:
		return newUintSliceValue(base, fieldsIndexes, fields)
	case reflect.Uint8:
		if typ.Elem().Kind() == reflect.Uint8 { // slice of bytes
			return newBytesValue(base, fieldsIndexes, fields)
		}
		return newUint8SliceValue(base, fieldsIndexes, fields)
	case reflect.Uint16:
		return newUint16SliceValue(base, fieldsIndexes, fields)
	case reflect.Uint32:
		return newUint32SliceValue(base, fieldsIndexes, fields)
	case reflect.Uint64:
		return newUint64SliceValue(base, fieldsIndexes, fields)
	case reflect.Float32:
		return newFloat32SliceValue(base, fieldsIndexes, fields)
	case reflect.Float64:
		return newFloat64SliceValue(base, fieldsIndexes, fields)
	case reflect.Complex64:
		return newComplex64SliceValue(base, fieldsIndexes, fields)
	case reflect.Complex128:
		return newComplex128SliceValue(base, fieldsIndexes, fields)
	case reflect.String:
		return newStringSliceValue(base, fieldsIndexes, fields)
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 { // slice of slices of bytes
			return newBytesSliceValue(base, fieldsIndexes, fields)
		}
		return newSliceSliceValue(base, fieldsIndexes, fields)
	case reflect.Map:
		return newMapSliceValue(base, fieldsIndexes, fields)
	case reflect.Struct:
		return newStructSliceValue(base, fieldsIndexes, fields)
	}
	return nil
}

type Value struct {
	base          reflect.Value
	fieldsIndexes []int
	fields        []reflect.StructField
	typeName      string
	stringFn      func(*Value) string
	encodeFn      func(any) ([]byte, error)
	setFn         func(*Value, string) error
	decodeFn      func([]byte, any) error
	isBool        bool
}

func (val *Value) Path() []reflect.StructField {
	if !val.isInitialized() {
		return nil
	}
	return val.fields
}

func (val *Value) Type() string {
	if !val.isInitialized() {
		return ""
	}
	return val.typeName
}

func (val *Value) Get() any {
	if !val.isInitialized() {
		return nil
	}
	return val.get().Interface()
}

func (val *Value) String() string {
	if !val.isInitialized() {
		return ""
	}
	if val.encodeFn != nil {
		b, err := val.encodeFn(val.get().Interface())
		if err != nil {
			return ""
		}
		return string(b)
	}
	return val.stringFn(val)
}

func (val *Value) Set(to string) error {
	if !val.isInitialized() {
		return nil
	}
	if val.decodeFn != nil {
		return val.decodeFn([]byte(to), elemIfPtr(val.get()).Addr().Interface())
	}
	return val.setFn(val, to)
}

func (val *Value) SetEncoder(fn func(any) ([]byte, error)) {
	if !val.isInitialized() {
		return
	}
	val.encodeFn = fn
}

func (val *Value) SetDecoder(fn func([]byte, any) error) {
	if !val.isInitialized() {
		return
	}
	val.decodeFn = fn
}

func (val *Value) IsBoolFlag() bool {
	if !val.isInitialized() {
		return false
	}
	return val.isBool
}

func (val *Value) isInitialized() bool {
	return val != nil && val.base.IsValid()
}

func (val *Value) typ() reflect.Type {
	if len(val.fields) == 0 {
		return val.base.Type()
	}
	return val.fields[len(val.fields)-1].Type
}

func (val *Value) get() reflect.Value {
	return fieldByIndex(val.base, val.fieldsIndexes)
}

func fieldByIndex(v reflect.Value, index []int) reflect.Value {
	if len(index) == 0 {
		if v.Kind() == reflect.Pointer && v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return v
	}
	if (v.Kind() != reflect.Pointer || v.Type().Elem().Kind() != reflect.Struct) && v.Kind() != reflect.Struct {
		panic("jsonflag: expected a struct or a pointer to struct, got " + v.Kind().String())
	}
	for _, x := range index {
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
		v = v.Field(x)
	}
	if v.Kind() == reflect.Pointer && v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}
	return v
}

func elemIfPtr(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Pointer {
		return v
	}
	if v.IsNil() {
		return reflect.Zero(v.Type().Elem())
	}
	return v.Elem()
}

func elemIfPtrType(t reflect.Type) reflect.Type {
	if t.Kind() != reflect.Pointer {
		return t
	}
	return t.Elem()
}

func reflectValueSet(dst, src reflect.Value) {
	if dst.Kind() == reflect.Pointer && !dst.CanSet() {
		dst = dst.Elem()
	}
	if dst.Kind() == reflect.Pointer && src.Kind() != reflect.Pointer {
		src = addressable(src).Addr()
	} else if dst.Kind() != reflect.Pointer && src.Kind() == reflect.Pointer {
		src = src.Elem()
	}
	dst.Set(src)
}

func reflectValueAppend(slc, v reflect.Value) reflect.Value {
	if elemIfPtrType(slc.Type()).Elem().Kind() == reflect.Pointer && v.Kind() != reflect.Pointer {
		v = addressable(v).Addr()
	} else if elemIfPtrType(slc.Type()).Elem().Kind() != reflect.Pointer && v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	return addressable(reflect.Append(elemIfPtr(slc), v))
}

func addressable(v reflect.Value) reflect.Value {
	if !v.CanAddr() {
		ptr := reflect.New(v.Type())
		ptr.Elem().Set(v)
		v = ptr.Elem()
	}
	return v
}

func newBoolValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "bool",
		stringFn:      boolValueString,
		setFn:         boolValueSet,
	}
}

func boolValueString(val *Value) string {
	v := elemIfPtr(val.get()).Bool()
	if !v {
		return ""
	}
	return strconv.FormatBool(v)
}

func boolValueSet(val *Value, to string) error {
	v, err := strconv.ParseBool(to)
	if err != nil {
		return err
	}
	reflectValueSet(val.get(), reflect.ValueOf(v))
	return nil
}

func newIntValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "int",
		stringFn:      intValueString,
		setFn:         intValueSet,
	}
}

func intValueString(val *Value) string {
	v := elemIfPtr(val.get()).Int()
	if v == 0 {
		return ""
	}
	return strconv.FormatInt(v, 10)
}

func intValueSet(val *Value, to string) error {
	v, err := parseInt(to)
	if err != nil {
		return err
	}
	reflectValueSet(val.get(), reflect.ValueOf(v))
	return nil
}

func newInt8Value(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "int8",
		stringFn:      intValueString,
		setFn:         int8ValueSet,
	}
}

func int8ValueSet(val *Value, to string) error {
	v, err := parseInt8(to)
	if err != nil {
		return err
	}
	reflectValueSet(val.get(), reflect.ValueOf(v))
	return nil
}

func newInt16Value(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "int16",
		stringFn:      intValueString,
		setFn:         int16ValueSet,
	}
}

func int16ValueSet(val *Value, to string) error {
	v, err := parseInt16(to)
	if err != nil {
		return err
	}
	reflectValueSet(val.get(), reflect.ValueOf(v))
	return nil
}

func newInt32Value(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "int32",
		stringFn:      intValueString,
		setFn:         int32ValueSet,
	}
}

func int32ValueSet(val *Value, to string) error {
	v, err := parseInt32(to)
	if err != nil {
		return err
	}
	reflectValueSet(val.get(), reflect.ValueOf(v))
	return nil
}

func newInt64Value(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "int64",
		stringFn:      intValueString,
		setFn:         int64ValueSet,
	}
}

func int64ValueSet(val *Value, to string) error {
	v, err := parseInt64(to)
	if err != nil {
		return err
	}
	reflectValueSet(val.get(), reflect.ValueOf(v))
	return nil
}

func newUintValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "uint",
		stringFn:      uintValueString,
		setFn:         uintValueSet,
	}
}

func uintValueString(val *Value) string {
	v := elemIfPtr(val.get()).Uint()
	if v == 0 {
		return ""
	}
	return strconv.FormatUint(v, 10)
}

func uintValueSet(val *Value, to string) error {
	v, err := parseUint(to)
	if err != nil {
		return err
	}
	reflectValueSet(val.get(), reflect.ValueOf(v))
	return nil
}

func newUint8Value(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "uint8",
		stringFn:      uintValueString,
		setFn:         uint8ValueSet,
	}
}

func uint8ValueSet(val *Value, to string) error {
	v, err := parseUint8(to)
	if err != nil {
		return err
	}
	reflectValueSet(val.get(), reflect.ValueOf(v))
	return nil
}

func newUint16Value(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "uint16",
		stringFn:      uintValueString,
		setFn:         uint16ValueSet,
	}
}

func uint16ValueSet(val *Value, to string) error {
	v, err := parseUint16(to)
	if err != nil {
		return err
	}
	reflectValueSet(val.get(), reflect.ValueOf(v))
	return nil
}

func newUint32Value(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "uint32",
		stringFn:      uintValueString,
		setFn:         uint32ValueSet,
	}
}

func uint32ValueSet(val *Value, to string) error {
	v, err := parseUint32(to)
	if err != nil {
		return err
	}
	reflectValueSet(val.get(), reflect.ValueOf(v))
	return nil
}

func newUint64Value(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "uint64",
		stringFn:      uintValueString,
		setFn:         uint64ValueSet,
	}
}

func uint64ValueSet(val *Value, to string) error {
	v, err := parseUint64(to)
	if err != nil {
		return err
	}
	reflectValueSet(val.get(), reflect.ValueOf(v))
	return nil
}

func newFloat32Value(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "float32",
		stringFn:      float32ValueString,
		setFn:         float32ValueSet,
	}
}

func float32ValueString(val *Value) string {
	v := elemIfPtr(val.get()).Float()
	if v == 0 {
		return ""
	}
	return strconv.FormatFloat(v, 'g', -1, 32)
}

func float32ValueSet(val *Value, to string) error {
	v, err := parseFloat32(to)
	if err != nil {
		return err
	}
	reflectValueSet(val.get(), reflect.ValueOf(v))
	return nil
}

func newFloat64Value(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "float64",
		stringFn:      float64ValueString,
		setFn:         float64ValueSet,
	}
}

func float64ValueString(val *Value) string {
	v := elemIfPtr(val.get()).Float()
	if v == 0 {
		return ""
	}
	return strconv.FormatFloat(v, 'g', -1, 64)
}

func float64ValueSet(val *Value, to string) error {
	v, err := parseFloat64(to)
	if err != nil {
		return err
	}
	reflectValueSet(val.get(), reflect.ValueOf(v))
	return nil
}

func newComplex64Value(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "complex64",
		stringFn:      complex64ValueString,
		setFn:         complex64ValueSet,
	}
}

func complex64ValueString(val *Value) string {
	v := elemIfPtr(val.get()).Complex()
	if v == 0 {
		return ""
	}
	return strconv.FormatComplex(v, 'g', -1, 64)
}

func complex64ValueSet(val *Value, to string) error {
	v, err := parseComplex64(to)
	if err != nil {
		return err
	}
	reflectValueSet(val.get(), reflect.ValueOf(v))
	return nil
}

func newComplex128Value(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "complex128",
		stringFn:      complex128ValueString,
		setFn:         complex128ValueSet,
	}
}

func complex128ValueString(val *Value) string {
	v := elemIfPtr(val.get()).Complex()
	if v == 0 {
		return ""
	}
	return strconv.FormatComplex(v, 'g', -1, 128)
}

func complex128ValueSet(val *Value, to string) error {
	v, err := parseComplex128(to)
	if err != nil {
		return err
	}
	reflectValueSet(val.get(), reflect.ValueOf(v))
	return nil
}

func newStringValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "string",
		stringFn:      stringValueString,
		setFn:         stringValueSet,
	}
}

func stringValueString(val *Value) string {
	return elemIfPtr(val.get()).String()
}

func stringValueSet(val *Value, to string) error {
	reflectValueSet(val.get(), reflect.ValueOf(to))
	return nil
}

func newBytesValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "base64",
		stringFn:      bytesValueString,
		setFn:         bytesValueSet,
	}
}

func bytesValueString(val *Value) string {
	v := elemIfPtr(val.get()).Interface().([]byte) //nolint:forcetypeassert // no nee to check type conversion result - correct value should be selected by the newValue function
	if len(v) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(v)
}

func bytesValueSet(val *Value, to string) error {
	v, err := base64.StdEncoding.DecodeString(to)
	if err != nil {
		return err
	}
	x := val.get()
	reflectValueSet(x, reflect.ValueOf(v))
	return nil
}

func newMapValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "JSON object",
		stringFn:      mapValueString,
		setFn:         mapValueSet,
	}
}

func mapValueString(val *Value) string {
	v := elemIfPtr(val.get())
	if v.Len() == 0 {
		return ""
	}
	b, err := json.Marshal(v.Interface())
	if err != nil {
		return ""
	}
	return string(b)
}

func mapValueSet(val *Value, to string) error {
	t := elemIfPtrType(val.typ())
	v := addressable(reflect.MakeMap(t)).Addr()
	if err := json.Unmarshal([]byte(to), v.Interface()); err != nil {
		return err
	}
	reflectValueSet(val.get(), v)
	return nil
}

func newStructValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "JSON object",
		stringFn:      structValueString,
		setFn:         structValueSet,
	}
}

func structValueString(val *Value) string {
	b, err := json.Marshal(val.get().Interface())
	if err != nil {
		return ""
	}
	str := string(b)
	if str == "{}" || str == "[]" || str == `""` || str == "0" || str == "false" {
		return ""
	}
	return str
}

func structValueSet(val *Value, to string) error {
	t := elemIfPtrType(val.typ())
	v := reflect.New(t)
	if err := json.Unmarshal([]byte(to), v.Interface()); err != nil {
		return err
	}
	reflectValueSet(val.get(), v)
	return nil
}

func sliceValueString(val *Value) string {
	v := elemIfPtr(val.get())
	if v.Len() == 0 {
		return ""
	}
	b, err := json.Marshal(v.Interface())
	if err != nil {
		return ""
	}
	return string(b)
}

func newBoolSliceValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "bool (JSON list)",
		stringFn:      sliceValueString,
		setFn:         boolSliceValueSet,
		isBool:        true,
	}
}

func boolSliceValueSet(val *Value, to string) error {
	v, err := strconv.ParseBool(to)
	if err != nil {
		return err
	}
	x := val.get()
	reflectValueSet(x, reflectValueAppend(elemIfPtr(x), reflect.ValueOf(v)))
	return nil
}

func newIntSliceValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "int (JSON list)",
		stringFn:      sliceValueString,
		setFn:         intSliceValueSet,
	}
}

func intSliceValueSet(val *Value, to string) error {
	v, err := parseInt(to)
	if err != nil {
		return err
	}
	x := val.get()
	reflectValueSet(x, reflectValueAppend(elemIfPtr(x), reflect.ValueOf(v)))
	return nil
}

func newInt8SliceValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "int8 (JSON list)",
		stringFn:      sliceValueString,
		setFn:         int8SliceValueSet,
	}
}

func int8SliceValueSet(val *Value, to string) error {
	v, err := parseInt8(to)
	if err != nil {
		return err
	}
	x := val.get()
	reflectValueSet(x, reflectValueAppend(elemIfPtr(x), reflect.ValueOf(v)))
	return nil
}

func newInt16SliceValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "int16 (JSON list)",
		stringFn:      sliceValueString,
		setFn:         int16SliceValueSet,
	}
}

func int16SliceValueSet(val *Value, to string) error {
	v, err := parseInt16(to)
	if err != nil {
		return err
	}
	x := val.get()
	reflectValueSet(x, reflectValueAppend(elemIfPtr(x), reflect.ValueOf(v)))
	return nil
}

func newInt32SliceValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "int32 (JSON list)",
		stringFn:      sliceValueString,
		setFn:         int32SliceValueSet,
	}
}

func int32SliceValueSet(val *Value, to string) error {
	v, err := parseInt32(to)
	if err != nil {
		return err
	}
	x := val.get()
	reflectValueSet(x, reflectValueAppend(elemIfPtr(x), reflect.ValueOf(v)))
	return nil
}

func newInt64SliceValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "int64 (JSON list)",
		stringFn:      sliceValueString,
		setFn:         int64SliceValueSet,
	}
}

func int64SliceValueSet(val *Value, to string) error {
	v, err := parseInt64(to)
	if err != nil {
		return err
	}
	x := val.get()
	reflectValueSet(x, reflectValueAppend(elemIfPtr(x), reflect.ValueOf(v)))
	return nil
}

func newUintSliceValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "uint (JSON list)",
		stringFn:      sliceValueString,
		setFn:         uintSliceValueSet,
	}
}

func uintSliceValueSet(val *Value, to string) error {
	v, err := parseUint(to)
	if err != nil {
		return err
	}
	x := val.get()
	reflectValueSet(x, reflectValueAppend(elemIfPtr(x), reflect.ValueOf(v)))
	return nil
}

func newUint8SliceValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "uint8 (JSON list)",
		stringFn:      sliceValueString,
		setFn:         uint8SliceValueSet,
	}
}

func uint8SliceValueSet(val *Value, to string) error {
	v, err := parseUint8(to)
	if err != nil {
		return err
	}
	x := val.get()
	reflectValueSet(x, reflectValueAppend(elemIfPtr(x), reflect.ValueOf(v)))
	return nil
}

func newUint16SliceValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "uint16 (JSON list)",
		stringFn:      sliceValueString,
		setFn:         uint16SliceValueSet,
	}
}

func uint16SliceValueSet(val *Value, to string) error {
	v, err := parseUint16(to)
	if err != nil {
		return err
	}
	x := val.get()
	reflectValueSet(x, reflectValueAppend(elemIfPtr(x), reflect.ValueOf(v)))
	return nil
}

func newUint32SliceValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "uint32 (JSON list)",
		stringFn:      sliceValueString,
		setFn:         uint32SliceValueSet,
	}
}

func uint32SliceValueSet(val *Value, to string) error {
	v, err := parseUint32(to)
	if err != nil {
		return err
	}
	x := val.get()
	reflectValueSet(x, reflectValueAppend(elemIfPtr(x), reflect.ValueOf(v)))
	return nil
}

func newUint64SliceValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "uint64 (JSON list)",
		stringFn:      sliceValueString,
		setFn:         uint64SliceValueSet,
	}
}

func uint64SliceValueSet(val *Value, to string) error {
	v, err := parseUint64(to)
	if err != nil {
		return err
	}
	x := val.get()
	reflectValueSet(x, reflectValueAppend(elemIfPtr(x), reflect.ValueOf(v)))
	return nil
}

func newFloat32SliceValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "float32 (JSON list)",
		stringFn:      sliceValueString,
		setFn:         float32SliceValueSet,
	}
}

func float32SliceValueSet(val *Value, to string) error {
	v, err := parseFloat32(to)
	if err != nil {
		return err
	}
	x := val.get()
	reflectValueSet(x, reflectValueAppend(elemIfPtr(x), reflect.ValueOf(v)))
	return nil
}

func newFloat64SliceValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "float64 (JSON list)",
		stringFn:      sliceValueString,
		setFn:         float64SliceValueSet,
	}
}

func float64SliceValueSet(val *Value, to string) error {
	v, err := parseFloat64(to)
	if err != nil {
		return err
	}
	x := val.get()
	reflectValueSet(x, reflectValueAppend(elemIfPtr(x), reflect.ValueOf(v)))
	return nil
}

func newComplex64SliceValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "complex64 (JSON list)",
		stringFn:      complex64SliceValueString,
		setFn:         complex64SliceValueSet,
	}
}

func complex64SliceValueString(val *Value) string {
	v := elemIfPtr(val.get())
	if v.Len() == 0 {
		return ""
	}
	li := make([]string, 0, v.Len())
	for _, x := range v.Seq2() {
		li = append(li, `"`+strconv.FormatComplex(elemIfPtr(x).Complex(), 'g', -1, 64)+`"`)
	}
	return "[" + strings.Join(li, ",") + "]"
}

func complex64SliceValueSet(val *Value, to string) error {
	v, err := parseComplex64(to)
	if err != nil {
		return err
	}
	x := val.get()
	reflectValueSet(x, reflectValueAppend(elemIfPtr(x), reflect.ValueOf(v)))
	return nil
}

func newComplex128SliceValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "complex128 (JSON list)",
		stringFn:      complex128SliceValueString,
		setFn:         complex128SliceValueSet,
	}
}

func complex128SliceValueString(val *Value) string {
	v := elemIfPtr(val.get())
	if v.Len() == 0 {
		return ""
	}
	li := make([]string, 0, v.Len())
	for _, x := range v.Seq2() {
		li = append(li, `"`+strconv.FormatComplex(elemIfPtr(x).Complex(), 'g', -1, 128)+`"`)
	}
	return "[" + strings.Join(li, ",") + "]"
}

func complex128SliceValueSet(val *Value, to string) error {
	v, err := parseComplex128(to)
	if err != nil {
		return err
	}
	x := val.get()
	reflectValueSet(x, reflectValueAppend(elemIfPtr(x), reflect.ValueOf(v)))
	return nil
}

func newStringSliceValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "string (JSON list)",
		stringFn:      sliceValueString,
		setFn:         stringSliceValueSet,
	}
}

func stringSliceValueSet(val *Value, to string) error {
	x := val.get()
	reflectValueSet(x, reflectValueAppend(elemIfPtr(x), reflect.ValueOf(to)))
	return nil
}
func newBytesSliceValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "base64 (JSON list)",
		stringFn:      sliceValueString,
		setFn:         bytesSliceValueSet,
	}
}

func bytesSliceValueSet(val *Value, to string) error {
	v, err := base64.StdEncoding.DecodeString(to)
	if err != nil {
		return err
	}
	x := val.get()
	reflectValueSet(x, reflectValueAppend(elemIfPtr(x), reflect.ValueOf(v)))
	return nil
}

func newSliceSliceValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "JSON list",
		stringFn:      sliceValueString,
		setFn:         sliceSliceValueSet,
	}
}

func sliceSliceValueSet(val *Value, to string) error {
	sliceType := elemIfPtrType(val.typ())
	t := elemIfPtrType(sliceType.Elem()) // slice element type
	v := addressable(reflect.MakeSlice(t, 0, 0)).Addr()
	if err := json.Unmarshal([]byte(to), v.Interface()); err != nil {
		return err
	}
	x := val.get()
	reflectValueSet(x, reflectValueAppend(elemIfPtr(x), v))
	return nil
}

func newMapSliceValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "JSON object (JSON list)",
		stringFn:      sliceValueString,
		setFn:         mapSliceValueSet,
	}
}

func mapSliceValueSet(val *Value, to string) error {
	sliceType := elemIfPtrType(val.typ())
	t := elemIfPtrType(sliceType.Elem()) // slice element type
	v := addressable(reflect.MakeMap(t)).Addr()
	if err := json.Unmarshal([]byte(to), v.Interface()); err != nil {
		return err
	}
	x := val.get()
	reflectValueSet(x, reflectValueAppend(elemIfPtr(x), v))
	return nil
}

func newStructSliceValue(base reflect.Value, fieldsIndexes []int, fields []reflect.StructField) *Value {
	return &Value{
		base:          base,
		fieldsIndexes: fieldsIndexes,
		fields:        fields,
		typeName:      "JSON object (JSON list)",
		stringFn:      sliceValueString,
		setFn:         structSliceValueSet,
	}
}

func structSliceValueSet(val *Value, to string) error {
	sliceType := elemIfPtrType(val.typ())
	t := elemIfPtrType(sliceType.Elem()) // slice element type
	v := reflect.New(t)
	if err := json.Unmarshal([]byte(to), v.Interface()); err != nil {
		return err
	}
	x := val.get()
	reflectValueSet(x, reflectValueAppend(elemIfPtr(x), v))
	return nil
}
