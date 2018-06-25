package main

import (
	"fmt"
	"github.com/WillAbides/xqsmee/queue/redisqueue"
	"github.com/WillAbides/xqsmee/server"
	"github.com/gomodule/redigo/redis"
	"github.com/spf13/cobra"
	"net"
	"os"
	"time"
)

var (
	redisKeyPrefix string
	redisUrl       string
	tcpAddr        string
	redisMaxActive int
)

var cmd = &cobra.Command{
	Use: "xqsmee",
	RunE: func(cmd *cobra.Command, args []string) error {
		redisPool := &redis.Pool{
			MaxActive: redisMaxActive,
			Wait:      true,
			Dial: func() (redis.Conn, error) {
				return redis.DialURL(redisUrl)
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		}
		listener, err := net.Listen("tcp", tcpAddr)
		if err != nil {
			return err
		}
		redisQueue := redisqueue.New(redisKeyPrefix, redisPool)
		cfg := &server.Config{
			Queue:    redisQueue,
			Listener: listener,
		}
		return server.Run(cfg)
	},
}

func main() {
	flags := cmd.Flags()
	flags.StringVarP(&redisUrl, "redisurl", "r", "redis://:6379", "redis url")
	flags.IntVar(&redisMaxActive, "maxactive", 100, "max number of active redis connections")
	flags.StringVarP(&tcpAddr, "tcp address to listen on", "a", ":8000", "tcp address to listen on")
	flags.StringVar(&redisKeyPrefix, "redisprefix", "xqsmee", "prefix for redis keys")
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
