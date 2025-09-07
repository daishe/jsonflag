# jsonflag

**jsonflag** is a tiny Go library that can turn any JSON-marshallable struct into a set of command‑line flags. It works with the standard flag package and with the popular [pflag package](https://github.com/spf13/pflag), handling nested structs, slices, maps and pointer fields automatically.

[![Go Reference](https://pkg.go.dev/badge/github.com/daishe/jsonflag.svg)](https://pkg.go.dev/github.com/daishe/jsonflag)
[![Go Report Card](https://goreportcard.com/badge/github.com/daishe/jsonflag)](https://goreportcard.com/report/github.com/daishe/jsonflag)

## Adding to project

First, use `go get` to download and add the latest version of the library to the project.

```sh
go get -u github.com/daishe/jsonflag
```

Then include in your source code.

```go
import "github.com/daishe/jsonflag"
```

## Usage example (with standard go `flag` package)

```go
package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/daishe/jsonflag"
)

func main() {
	config := struct {
		URL struct {
			Scheme string `json:"scheme"`
			Host   string `json:"host"`
			Port   int    `json:"port"`
			Path   string `json:"path"`
		} `json:"url"`
		Verbose bool `json:"verbose"`
	}{}

	// default values
	config.URL.Scheme = "https"
	config.URL.Host = "localhost"
	config.URL.Port = 8080

	fs := flag.NewFlagSet("", flag.ExitOnError) // you can also add it directly to the root flag set
	for _, val := range jsonflag.Recursive(&config) {
		fs.Var(val, jsonflag.JSONName(val.Path()), jsonflag.Usage(val.Path()))
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Arguments parsing error: %v\n", err)
		os.Exit(1)
	}

	u := url.URL{
		Scheme: config.URL.Scheme,
		Host:   fmt.Sprintf("%s:%d", config.URL.Host, config.URL.Port),
		Path:   config.URL.Path,
	}

	resp, err := http.Get(u.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Doing request error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if config.Verbose {
		respBytes, err := httputil.DumpResponse(resp, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Dumping response error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%s\n", respBytes)
	}
}
```

## License

The project is released under the **Apache License, Version 2.0**. See the full LICENSE file for the complete terms and conditions.
