package cmd

import (
	"context"
	"log"
	"os"

	"github.com/WillAbides/xqsmee/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	clientCfg *client.Config

	clientCmd = &cobra.Command{
		Use: "client",
		Short: "Run xqsmee client",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			cc := new(clientCmdCfg)
			err := viper.Unmarshal(cc)
			if err != nil {
				return err
			}
			clientCfg = cc.clientConfig()
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := client.Run(context.Background(), clientCfg)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(clientCmd)

	flags := clientCmd.Flags()
	flags.StringP("queue", "q", "", "xqsmee queue to watch")
	flags.StringP("server", "s", "", "server ip or dns address")
	flags.IntP("port", "p", 9443, "server grpc port")
	flags.String("ifs", "\n", "record separator")
	flags.Bool("insecure", false, "don't check for valid certificate")
	flags.Bool("no-tls", false, "don't use tls (insecure)")
	for _, flag := range []string{"queue", "server", "port"} {
		err := clientCmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
	err := viper.BindPFlags(flags)
	if err != nil {
		panic(err)
	}
}

type clientCmdCfg struct {
	QueueName  string `mapstructure:"queue"`
	ServerHost string `mapstructure:"server"`
	ServerPort int `mapstructure:"port"`
	Insecure   bool
	Separator  string `mapstructure:"ifs"`
	NoTLS      bool   `mapstructure:"no-tls"`
}

func (c *clientCmdCfg) clientConfig() *client.Config {
	return &client.Config{
		Host:      c.ServerHost,
		Port:      c.ServerPort,
		Insecure:  c.Insecure,
		QueueName: c.QueueName,
		Stdout:    os.Stdout,
		Separator: c.Separator,
		UseTLS:    !c.NoTLS,
	}
}
