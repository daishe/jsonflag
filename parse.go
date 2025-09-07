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

import "strconv"

func parseInt(s string) (int, error) {
	v, err := strconv.ParseInt(s, 0, 0)
	return int(v), err
}

func parseInt8(s string) (int8, error) {
	v, err := strconv.ParseInt(s, 0, 8)
	return int8(v), err
}

func parseInt16(s string) (int16, error) {
	v, err := strconv.ParseInt(s, 0, 16)
	return int16(v), err
}

func parseInt32(s string) (int32, error) {
	v, err := strconv.ParseInt(s, 0, 32)
	return int32(v), err
}

func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 0, 64)
}

func parseUint(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 0, 0)
	return uint(v), err
}

func parseUint8(s string) (uint8, error) {
	v, err := strconv.ParseUint(s, 0, 8)
	return uint8(v), err
}

func parseUint16(s string) (uint16, error) {
	v, err := strconv.ParseUint(s, 0, 16)
	return uint16(v), err
}

func parseUint32(s string) (uint32, error) {
	v, err := strconv.ParseUint(s, 0, 32)
	return uint32(v), err
}

func parseUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 0, 64)
}

func parseFloat32(s string) (float32, error) {
	v, err := strconv.ParseFloat(s, 32)
	return float32(v), err
}

func parseFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func parseComplex64(s string) (complex64, error) {
	v, err := strconv.ParseComplex(s, 64)
	return complex64(v), err
}

func parseComplex128(s string) (complex128, error) {
	return strconv.ParseComplex(s, 128)
}
