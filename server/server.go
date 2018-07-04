package server

import (
	"crypto/tls"
	"github.com/WillAbides/xqsmee/queue"
	"github.com/WillAbides/xqsmee/services/hooks"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"net"
	"net/http"
)

type Config struct {
	Queue           queue.Queue
	Httpaddr        string
	Grpcaddr        string
	TLSCertPEMBlock []byte
	TLSKeyPEMBlock  []byte
	UseTLS          bool
}

func (config *Config) buildListeners() (httpListener, grpcListener net.Listener, err error) {
	httpListener, err = net.Listen("tcp", config.Httpaddr)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed starting http listener")
	}

	grpcListener, err = net.Listen("tcp", config.Grpcaddr)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed starting grpc listener")
	}

	if config.UseTLS {
		var cert tls.Certificate
		cert, err = tls.X509KeyPair([]byte(config.TLSCertPEMBlock), []byte(config.TLSKeyPEMBlock))
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed creating tls certificate from key pair")
		}
		tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
		grpcListener = tls.NewListener(grpcListener, tlsConfig)
		httpListener = tls.NewListener(httpListener, tlsConfig)
	}
	return httpListener, grpcListener, err
}

func Run(config *Config) error {
	httpListener, grpcListener, err := config.buildListeners()
	if err != nil {
		return errors.Wrap(err, "failed building listeners")
	}
	grpcErrs := make(chan error)
	httpErrs := make(chan error)

	httpServer := &http.Server{
		Handler: hooks.New(config.Queue).Router(),
	}

	go func() {
		httpErrs <- httpServer.Serve(httpListener)
	}()
	defer httpServer.Close()

	grpcServer := grpc.NewServer()
	grpcHandler := queue.NewGRPCHandler(config.Queue)
	queue.RegisterQueueServer(grpcServer, grpcHandler)

	go func() {
		grpcErrs <- grpcServer.Serve(grpcListener)
	}()
	defer grpcServer.Stop()

	select {
	case err := <-httpErrs:
		return err
	case err := <-grpcErrs:
		return err
	}
}
