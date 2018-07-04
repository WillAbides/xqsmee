package main

import (
	"fmt"
	"github.com/WillAbides/xqsmee/queue/redisqueue"
	"github.com/WillAbides/xqsmee/server"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
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

		var tlsCert []byte
		if c.TLSCert != "" {
			tlsCert, err = ioutil.ReadFile(c.TLSCert)
			if err != nil {
				return errors.Wrap(err, "failed reading tls certificate file")
			}
		}

		var tlsKey []byte
		if c.TLSKey != "" {
			tlsKey, err = ioutil.ReadFile(c.TLSKey)
			if err != nil {
				return errors.Wrap(err, "failed reading tls key file")
			}
		}

		cfg := &server.Config{
			Queue:           redisQueue,
			Httpaddr:        c.Httpaddr,
			Grpcaddr:        c.Grpcaddr,
			TLSKeyPEMBlock:  tlsKey,
			TLSCertPEMBlock: tlsCert,
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
	flags.String("tlskey", "", "file containing a tls key")
	flags.String("tlscert", "", "file containing a tls certificate")
	flags.Bool("no-tls", false, "don't use tls (serve unencrypted http and grpc)")
	must(viper.BindPFlags(flags))
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
