package main

import (
	"context"
	"github.com/WillAbides/xqsmee/queue"
	"google.golang.org/grpc"
	"log"
)

const address = "localhost:8089"

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := queue.NewQueueClient(conn)

	queueName := "asdf"
	ctx := context.Background()
	popRequest := &queue.BPopRequest{
		QueueName: queueName,
	}
	r, err := c.BPop(ctx, popRequest)
	if err != nil {
		log.Fatal(err)
	}
	wr := r.GetWebRequest()
	if wr != nil {
		log.Println(wr.String())
	}
}
