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

	resp, err := http.Get(u.String()) //nolint:noctx // this is just an example
	if err != nil {
		fmt.Fprintf(os.Stderr, "Doing request error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if config.Verbose {
		respBytes, err := httputil.DumpResponse(resp, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Dumping response error: %v\n", err)
			os.Exit(1) //nolint:gocritic // this is just an example
		}
		fmt.Printf("%s\n", respBytes)
	}
}
