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
	serverAddr string
	insecure   bool
	separator  string
)

var cmd = &cobra.Command{
	Use: "xqsmee-client",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := &client.Config{
			Address:      serverAddr,
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
	flags.StringVarP(&serverAddr, "server", "s", "", "address of xqsmee server")
	flags.StringVar(&separator, "ifs", "\n", "record separator")
	flags.BoolVar(&insecure, "insecure", false, "ignore ssl warnings")
	cmd.MarkFlagRequired("queue")
	cmd.MarkFlagRequired("server")
	if err := cmd.Execute(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
