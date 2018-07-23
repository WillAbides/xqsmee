package cmd

import (
	"context"
	"os"

	"github.com/WillAbides/xqsmee/client"
)

//nolint: govet
type clientCmd struct {
	Server   string `arg required help:"server ip or dns address" env:"XQSMEE_SERVER"`
	Queue    string `arg required help:"xqsmee queue to watch" env:"XQSMEE_QUEUE"`
	Port     int    `default:"9443" short:"p" help:"server grpc port"`
	Insecure bool   `help:"don't check for valid certificate"`
	NoTLS    bool   `help:"don't use tls (insecure)"`
	Ifs      string `default:"${defaultIfs}" help:"record separator"`
}

func (c *clientCmd) Run() error {
	return client.Run(context.Background(), &client.Config{
		Host:      c.Server,
		Port:      c.Port,
		Insecure:  c.Insecure,
		QueueName: c.Queue,
		Stdout:    os.Stdout,
		Separator: c.Ifs,
		UseTLS:    !c.NoTLS,
	})
}
