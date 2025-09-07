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
	"encoding/json"
	"flag"
	"fmt"

	"github.com/daishe/jsonflag"
	"github.com/spf13/pflag"
)

func ExampleRecursive_withGoFlag() {
	type Input struct {
		Foo string `json:"Foo"`
		Bar string `json:"Bar"`
		Baz string `json:"Baz"`
	}

	fs := flag.NewFlagSet("", flag.ExitOnError)
	i := &Input{}
	for _, val := range jsonflag.Recursive(i) {
		fs.Var(val, jsonflag.Name(val.Path()), jsonflag.Usage(val.Path()))
	}
	if err := fs.Parse([]string{"--Foo=foo value", "--Baz=baz value"}); err != nil {
		panic(err)
	}

	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))

	// Output:
	// {
	//   "Foo": "foo value",
	//   "Bar": "",
	//   "Baz": "baz value"
	// }
}

func ExampleRecursive_withPFlag() {
	type Input struct {
		Foo string `json:"Foo"`
		Bar string `json:"Bar"`
		Baz string `json:"Baz"`
	}

	fs := pflag.NewFlagSet("", pflag.ExitOnError)
	i := &Input{}
	for _, val := range jsonflag.Recursive(i) {
		pf := &pflag.Flag{
			Name:     jsonflag.Name(val.Path()),
			Usage:    jsonflag.Usage(val.Path()),
			Value:    val,
			DefValue: val.String(),
		}
		if val.IsBoolFlag() {
			pf.NoOptDefVal = "true"
		}
		fs.AddFlag(pf)
	}
	if err := fs.Parse([]string{"--Foo=foo value", "--Baz=baz value", "--Baz=another baz value"}); err != nil {
		panic(err)
	}

	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))

	// Output:
	// {
	//   "Foo": "foo value",
	//   "Bar": "",
	//   "Baz": "another baz value"
	// }
}
