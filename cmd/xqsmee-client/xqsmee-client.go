package main

import (
	"context"
	"fmt"
	"github.com/WillAbides/xqsmee/client"
	"github.com/spf13/cobra"
	"os"
)

var (
	queueName  string
	serverHost string
	serverPort int
	insecure   bool
	separator  string
)

var cmd = &cobra.Command{
	Use: "xqsmee-client",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := &client.Config{
			Host:         serverHost,
			Port:         serverPort,
			WithInsecure: insecure,
			QueueName:    queueName,
			Stdout:       os.Stdout,
			Separator:    separator,
		}
		return client.Run(context.Background(), cfg)
	},
}

func main() {
	flags := cmd.Flags()
	flags.StringVarP(&queueName, "queue", "q", "", "xqsmee queue to watch")
	flags.StringVarP(&serverHost, "server", "s", "", "server ip or dns address")
	flags.IntVarP(&serverPort, "port", "p", 9443, "server grpc port")
	flags.StringVar(&separator, "ifs", "\n", "record separator")
	flags.BoolVar(&insecure, "insecure", false, "allow grpc without tls")
	for _, flag := range []string{"queue", "server", "port"} {
		err := cmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
	if err := cmd.Execute(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
