package main

import (
	"github.com/WillAbides/xqsmee/queue"
	"github.com/WillAbides/xqsmee/queue/redisqueue"
	"github.com/WillAbides/xqsmee/services/hooks"
	"github.com/gomodule/redigo/redis"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"time"
)

var redisPool = &redis.Pool{
	MaxIdle:     100,
	MaxActive:   100,
	IdleTimeout: 60 * time.Second,
	Wait:        true,
	Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", ":6379") },
	TestOnBorrow: func(c redis.Conn, t time.Time) error {
		_, err := c.Do("PING")
		return err
	},
}

func main() {
	redisQueue := redisqueue.New("xqsmee", redisPool)
	hooksSvc := hooks.New(redisQueue)
	router := hooksSvc.Router()

	l, err := net.Listen("tcp", "localhost:8089")
	if err != nil {
		log.Fatal(err)
	}

	m := cmux.New(l)
	grpcListener := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	httpListener := m.Match(cmux.HTTP1Fast())

	grpcServer := grpc.NewServer()

	queue.RegisterQueueServer(grpcServer, redisQueue.QueueServer())

	httpServer := &http.Server{
		Handler: router,
	}

	go grpcServer.Serve(grpcListener)
	go httpServer.Serve(httpListener)

	err = m.Serve()
	if err != nil {
		log.Fatal(err)
	}
}
