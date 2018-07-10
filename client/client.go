package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"

	"github.com/WillAbides/xqsmee/queue"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

//Config is the config for running a client
type Config struct {
	Host      string
	QueueName string
	Separator string
	Port      int
	Insecure  bool
	UseTLS    bool
	Stdout    io.Writer
}

func dialGRPC(ctx context.Context, config *Config) (*grpc.ClientConn, error) {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	if config.UseTLS {
		tlsConfig := &tls.Config{ServerName: config.Host}
		if config.Insecure {
			tlsConfig.InsecureSkipVerify = true
		}
		creds := credentials.NewTLS(tlsConfig)
		return grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(creds))
	}
	return grpc.DialContext(ctx, addr, grpc.WithInsecure())
}

//Run runs a client
func Run(ctx context.Context, config *Config) error {
	conn, err := dialGRPC(ctx, config)
	if err != nil {
		return err
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			panic(err)
		}
	}()

	c := queue.NewQueueClient(conn)

	for {
		r, err := c.Pop(ctx, &queue.PopRequest{QueueName: config.QueueName})
		if err != nil {
			return err
		}
		webRequest := r.GetWebRequest()
		if webRequest != nil {
			jb, err := json.Marshal(webRequest)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(config.Stdout, "%s%s", string(jb), config.Separator)
			if err != nil {
				return err
			}
		}
	}
}
