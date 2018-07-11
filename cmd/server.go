package cmd

import (
	"io/ioutil"
	"log"
	"time"

	"github.com/WillAbides/xqsmee/queue/redisqueue"
	"github.com/WillAbides/xqsmee/server"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	srvCfg *server.Config

	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Run xqsmee server",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			cc := new(srvCmdCfg)
			err := viper.Unmarshal(cc)
			if err != nil {
				return err
			}
			srvCfg, err = cc.serverConfig()
			return err
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := server.Run(srvCfg)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func init() {
	clientCmd.AddCommand(serverCmd)

	flags := serverCmd.Flags()
	flags.StringP("redisurl", "r", "redis://:6379", "redis url")
	flags.Int("maxactive", 100, "max number of active redis connections")
	flags.String("httpaddr", ":8443", "tcp address to listen on")
	flags.String("grpcaddr", ":9443", "tcp address to listen on")
	flags.String("redisprefix", "xqsmee", "prefix for redis key")
	flags.String("tlskey", "", "file containing a tls key")
	flags.String("tlscert", "", "file containing a tls certificate")
	flags.Bool("no-tls", false, "don't use tls (serve unencrypted http and grpc)")
	err := viper.BindPFlags(flags)
	if err != nil {
		panic(err)
	}
}

type srvCmdCfg struct {
	RedisURL    string
	MaxActive   int
	Httpaddr    string
	Grpcaddr    string
	RedisPrefix string
	TLSCert     string
	TLSKey      string
	NoTLS       bool `mapstructure:"no-tls"`
}

func tlsData(noTLS bool, tlsCertFile, tlsKeyFile string) (tlsCert, tlsKey []byte, err error) {
	if noTLS {
		return nil, nil, nil
	}
	if tlsKeyFile == "" || tlsCertFile == "" {
		return nil, nil, errors.New("you must specify both --tlskey and --tlscert unless --no-tls is set")
	}
	tlsCert, err = ioutil.ReadFile(tlsCertFile) //nolint: gas
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed reading tls certificate file")
	}
	tlsKey, err = ioutil.ReadFile(tlsKeyFile) //nolint: gas
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed reading tls key file")
	}
	return tlsCert, tlsKey, nil
}

func (c *srvCmdCfg) serverConfig() (*server.Config, error) {
	var err error
	tlsCert, tlsKey, err := tlsData(c.NoTLS, c.TLSCert, c.TLSKey)
	if err != nil {
		return nil, err
	}

	redisPool := &redis.Pool{
		MaxActive: c.MaxActive,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(c.RedisURL)
		},
		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			_, err := conn.Do("PING")
			return err
		},
	}
	redisQueue := redisqueue.New(c.RedisPrefix, redisPool)

	sCfg := &server.Config{
		Queue:           redisQueue,
		Httpaddr:        c.Httpaddr,
		Grpcaddr:        c.Grpcaddr,
		TLSKeyPEMBlock:  tlsKey,
		TLSCertPEMBlock: tlsCert,
		UseTLS:          !c.NoTLS,
	}

	return sCfg, nil
}
