package server

import (
	"github.com/WillAbides/xqsmee/queue"
	"github.com/WillAbides/xqsmee/services/hooks"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"net"
	"net/http"
)

type Config struct {
	Queue    queue.Queue
	Listener net.Listener
}

func Run(config *Config) error {
	hooksSvc := hooks.New(config.Queue)
	router := hooksSvc.Router()

	m := cmux.New(config.Listener)
	grpcListener := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	httpListener := m.Match(cmux.HTTP1Fast())

	grpcServer := grpc.NewServer()
	grpcHandler := queue.NewGRPCHandler(config.Queue)
	queue.RegisterQueueServer(grpcServer, grpcHandler)
	httpServer := &http.Server{
		Handler: router,
	}

	go grpcServer.Serve(grpcListener)
	defer grpcServer.Stop()
	go httpServer.Serve(httpListener)
	defer httpServer.Close()
	return m.Serve()
}
