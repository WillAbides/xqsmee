package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/WillAbides/xqsmee/queue"
	"google.golang.org/grpc"
	"io"
)

type Config struct {
	Address      string
	WithInsecure bool
	QueueName    string
	Stdout       io.Writer
	Separator    string
}

func dialGRPC(ctx context.Context, config *Config) (*grpc.ClientConn, error) {
	dialOptions := make([]grpc.DialOption, 0)
	if config.WithInsecure == true {
		dialOptions = append(dialOptions, grpc.WithInsecure())
	}
	return grpc.DialContext(ctx, config.Address, dialOptions...)
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

	return nil
}
