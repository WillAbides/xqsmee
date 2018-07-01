package server

import (
	"github.com/WillAbides/xqsmee/queue"
	"github.com/WillAbides/xqsmee/services/hooks"
	"google.golang.org/grpc"
	"net"
	"net/http"
)

type Config struct {
	Queue        queue.Queue
	HttpListener net.Listener
	GrpcListener net.Listener
}

func Run(config *Config) error {
	hooksSvc := hooks.New(config.Queue)
	router := hooksSvc.Router()

	grpcServer := grpc.NewServer()
	grpcHandler := queue.NewGRPCHandler(config.Queue)
	queue.RegisterQueueServer(grpcServer, grpcHandler)
	httpServer := &http.Server{
		Handler: router,
	}
	grpcErrs := make(chan error)
	go func() {
		grpcErrs <- grpcServer.Serve(config.GrpcListener)
	}()
	defer grpcServer.Stop()
	httpErrs := make(chan error)
	go func() {
		httpErrs <- httpServer.Serve(config.HttpListener)
	}()
	defer httpServer.Close()
	select {
	case err := <-httpErrs:
		return err
	case err := <-grpcErrs:
		return err
	}
}
