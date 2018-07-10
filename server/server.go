package server

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"

	"github.com/WillAbides/idcheck"
	"github.com/WillAbides/xqsmee/queue"
	"github.com/WillAbides/xqsmee/services/hooks"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type Config struct {
	Queue           queue.Queue
	Httpaddr        string
	Grpcaddr        string
	TLSCertPEMBlock []byte
	TLSKeyPEMBlock  []byte
	UseTLS          bool
	idcheckSalt     string
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
		cert, err = tls.X509KeyPair(config.TLSCertPEMBlock, config.TLSKeyPEMBlock)
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
	errs := make(chan error)

	idChecker := idcheck.NewIDChecker(idcheck.Salt(config.idcheckSalt))

	httpServer := &http.Server{
		Handler: hooks.New(config.Queue, idChecker).Router(),
	}

	go func() {
		errs <- httpServer.Serve(httpListener)
	}()
	defer func() {
		err := httpServer.Close()
		if err != nil {
			log.Println("failed closing httpServer: ", err)
		}
	}()

	grpcServer := grpc.NewServer()
	grpcHandler := queue.NewGRPCHandler(config.Queue)
	queue.RegisterQueueServer(grpcServer, grpcHandler)

	go func() {
		errs <- grpcServer.Serve(grpcListener)
	}()
	defer grpcServer.Stop()

	return <-errs
}
