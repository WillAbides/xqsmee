package cmd

import "github.com/alecthomas/kong"

//nolint: govet
type rootCmd struct {
	Server serverCmd `cmd help:"run a server"`
	Client clientCmd `cmd help:"run the client"`
}

//Execute executes rootCmd
func Execute() error {
	var cmd rootCmd
	ctx := kong.Parse(&cmd, kong.UsageOnError())
	return ctx.Run()
}
