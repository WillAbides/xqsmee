package cmd

import (
	"bytes"

	"github.com/alecthomas/kong"
)

var (
	kongVars = kong.Vars{
		"defaultIfs": "\n",
	}
	usageParser = kong.Must(new(rootCmd), kong.Name("xqsmee"), kongVars)
)

//nolint: govet
type rootCmd struct {
	Version versionCmd `cmd help:"show the xqsmee version"`
	Server  serverCmd  `cmd help:"run a server"`
	Client  clientCmd  `cmd help:"run the client"`
}

//Execute executes rootCmd
func Execute() error {
	var cmd rootCmd
	ctx := kong.Parse(&cmd, kong.UsageOnError(), kongVars)
	return ctx.Run()
}

//CommandUsage returns the usage that would be output to stdout for the given command
// this is just for docgen
func CommandUsage(command ...string) ([]byte, error) {
	var stdOut bytes.Buffer
	usageParser.Stdout = &stdOut
	ctx, err := kong.Trace(usageParser, command)
	if err != nil {
		return nil, err
	}
	err = ctx.PrintUsage(false)
	return stdOut.Bytes(), err
}
