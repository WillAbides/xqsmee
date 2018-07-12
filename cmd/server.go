package cmd

import (
	"errors"
	"io/ioutil"
	"net/url"
	"time"

	"github.com/WillAbides/xqsmee/queue/redisqueue"
	"github.com/WillAbides/xqsmee/server"
	"github.com/gomodule/redigo/redis"
)

type serverCmd struct {
	Redisurl     *url.URL `default:"redis://:6379" short:"r" help:"redis url" env:"XQSMEE_REDISURL"`
	Maxactive    int      `default:"100" help:"max number of active redis connections" env:"XQSMEE_MAXACTIVE"`
	NoTLS        bool     `help:"don't use tls (serve unencrypted http and grpc)" env:"XQSMEE_NOTLS"`
	Httpaddr     string   `default:":8443" help:"tcp address for http connections" env:"XQSMEE_HTTPADDR"`
	Grpcaddr     string   `default:":9443" help:"tcp address for grpc connections" env:"XQSMEE_GRPCADDR"`
	Redisprefix  string   `default:"xqsmee" help:"prefix for redis keys" env:"XQSMEE_REDISPREFIX"`
	Tlskey       string   `type:"existingfile" help:"file containing a tls key" env:"XQSMEE_TLSKEY"`
	Tlscert      string   `type:"existingfile" help:"file containing a tls certificate" env:"XQSMEE_TLSCERT"`
	tlsKeyBlock  []byte
	tlsCertBlock []byte
}

func (c *serverCmd) AfterHook() error {
	if c.NoTLS {
		return nil
	}
	if c.Tlscert == "" || c.Tlskey == "" {
		return errors.New("you must specify both --tlskey and --tlscert unless --no-tls is set")
	}
	var err error
	c.tlsKeyBlock, err = ioutil.ReadFile(c.Tlskey)
	if err != nil {
		return errors.New("failed reading tls key file")
	}
	c.tlsCertBlock, err = ioutil.ReadFile(c.Tlscert)
	if err != nil {
		return errors.New("failed reading tls certificate file")
	}
	return nil
}

func (c *serverCmd) Run() error {
	redisPool := &redis.Pool{
		MaxActive: c.Maxactive,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(c.Redisurl.String())
		},
		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			_, err := conn.Do("PING")
			return err
		},
	}

	redisQueue := redisqueue.New(c.Redisprefix, redisPool)

	cfg := &server.Config{
		Queue:           redisQueue,
		Httpaddr:        c.Httpaddr,
		Grpcaddr:        c.Grpcaddr,
		TLSCertPEMBlock: c.tlsCertBlock,
		TLSKeyPEMBlock:  c.tlsKeyBlock,
		UseTLS:          !c.NoTLS,
	}

	return server.Run(cfg)
}
