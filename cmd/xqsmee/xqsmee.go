package main

import (
	"fmt"
	"github.com/WillAbides/xqsmee/queue/redisqueue"
	"github.com/WillAbides/xqsmee/server"
	"github.com/gomodule/redigo/redis"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"time"
)

var cmd = &cobra.Command{
	Use: "xqsmee",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := new(struct {
			RedisURL    string
			MaxActive   int
			Httpaddr    string
			Grpcaddr    string
			RedisPrefix string
			TLSCert     string
			TLSKey      string
			NoTLS       bool `mapstructure:"no-tls"`
		})

		err := viper.Unmarshal(c)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		redisPool := &redis.Pool{
			MaxActive: c.MaxActive,
			Wait:      true,
			Dial: func() (redis.Conn, error) {
				return redis.DialURL(c.RedisURL)
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		}
		redisQueue := redisqueue.New(c.RedisPrefix, redisPool)

		cfg := &server.Config{
			Queue:           redisQueue,
			Httpaddr:        c.Httpaddr,
			Grpcaddr:        c.Grpcaddr,
			TLSKeyPEMBlock:  []byte(c.TLSKey),
			TLSCertPEMBlock: []byte(c.TLSCert),
			UseTLS:          !c.NoTLS,
		}
		return server.Run(cfg)
	},
}

func init() {
	cobra.OnInitialize(func() {
		viper.SetEnvPrefix("XQSMEE")
		viper.AutomaticEnv()
	})
	flags := cmd.Flags()
	flags.StringP("redisurl", "r", "redis://:6379", "redis url")
	flags.Int("maxactive", 100, "max number of active redis connections")
	flags.String("httpaddr", ":8443", "tcp address to listen on")
	flags.String("grpcaddr", ":9443", "tcp address to listen on")
	flags.String("redisprefix", "xqsmee", "prefix for redis key")
	flags.Bool("no-tls", false, "don't use tls (serve unencrypted http and grpc)")
	must(viper.BindPFlags(flags))
	must(viper.BindEnv("TLSCERT"))
	must(viper.BindEnv("TLSKEY"))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
