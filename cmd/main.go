// find-latest-gopkg will query the Go Module Index for the latest version of one or more go packages and print
// selected fields to stdout.

package main

import (
	"fmt"
	"os"

	"github.com/rrrix/find-latest-gopkg/pkg/instance"
	"github.com/rrrix/find-latest-gopkg/pkg/moduleinfo"
	flag "github.com/spf13/pflag"
)

func Usage() {
	_, _ = fmt.Fprintf(os.Stderr, "Usage %s: [OPTIONS] MODULE [MODULE ...]\n\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {

	m := &instance.MainContext{
		Options: &instance.CLIOptions{},
	}
	// Define flags
	flag.BoolVarP(&m.Options.Version, "version", "V", false, "Print the Module Version field")
	flag.BoolVarP(&m.Options.Name, "name", "N", false, "Print the module name")
	flag.BoolVarP(&m.Options.Time, "time", "T", false, "Print the Time field")
	flag.BoolVarP(&m.Options.Repo, "repo", "r", false, "Print the Origin.URL field")
	flag.BoolVarP(&m.Options.Ref, "ref", "R", false, "Alias for --tag, Print the Origin.Ref field")
	flag.BoolVarP(&m.Options.Hash, "hash", "H", false, "Print the Origin.Hash field")
	flag.BoolVarP(&m.Options.Dump, "dump", "D", false, "Print the entire JSON response")

	flag.StringVarP(&m.Options.GoProxy, "goproxy", "g", "", "Go Module Mirror Endpoint (go env GOPROXY). Leave empty to use default.")
	flag.StringVarP(&m.Options.LogLevelName, "log-level", "l", "", "Log level")
	flag.BoolVarP(&m.Options.LogDebug, "debug", "d", false, "Increase logging verbosity (debug)")
	flag.BoolVarP(&m.Options.LogVerbose, "verbose", "v", false, "Increase logging verbosity (info)")

	flag.CommandLine.Usage = Usage

	// Parse flags and arguments
	flag.Parse()

	// Setup Logging
	m.BuildLogger()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}
	modules := args

	m.SetProxyEndpoints()

	// Iterate over modules, call findLatest on each, then printInfo on each
	for _, module := range modules {
		for _, proxy := range m.ProxyEndpoints {
			// TODO: Handle GOPRIVATE and GONOPROXY
			info, err := moduleinfo.FindLatest(proxy, module)
			if err != nil {
				continue
			}
			moduleinfo.PrintInfo(m, *info)
		}
	}
}
