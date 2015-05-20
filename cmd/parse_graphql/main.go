package main

import (
	"os"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	Verbose []bool `short:"v" long:"verbose" description:"Be verbose"`
}

var globalOptions Options

var optionsParser = flags.NewNamedParser("parse_graphql", flags.Default)

func init() {
	optionsParser.AddGroup("Global options", "", &globalOptions)
}

func main() {
	if _, err := optionsParser.Parse(); err != nil {
		os.Exit(1)
	}
}
