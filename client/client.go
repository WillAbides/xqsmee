package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/WillAbides/xqsmee/queue"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
)

type Config struct {
	Host         string
	Port         int
	WithInsecure bool
	QueueName    string
	Stdout       io.Writer
	Separator    string
}

func dialGRPC(ctx context.Context, config *Config) (*grpc.ClientConn, error) {
	dialOptions := make([]grpc.DialOption, 0)
	if config.WithInsecure {
		dialOptions = append(dialOptions, grpc.WithInsecure())
	} else {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	return grpc.DialContext(ctx, addr, dialOptions...)
}

func Run(ctx context.Context, config *Config) error {
	conn, err := dialGRPC(ctx, config)
	if err != nil {
		return err
	}
	defer conn.Close()

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
