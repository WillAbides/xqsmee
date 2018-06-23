package redisqueue

import (
	"context"
	"github.com/WillAbides/xqsmee/queue"
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
)

var (
	ErrEmptyPrefix = errors.New("prefix is empty")
	ErrNilPool     = errors.New("pool is nil")
)

type Queue struct {
	Prefix string
	Pool   *redis.Pool
}

type QueueServer struct {
	q *Queue
}

func (qs *QueueServer) Pop(ctx context.Context, request *queue.PopRequest) (*queue.PopResponse, error) {
	webRequest, err := qs.q.Pop(ctx, request.GetQueueName(), request.GetTimeout())
	return &queue.PopResponse{WebRequest: webRequest}, err
}

func (qs *QueueServer) Peek(ctx context.Context, request *queue.PeekRequest) (*queue.PeekResponse, error) {
	webRequests, err := qs.q.Peek(ctx, request.GetQueueName(), request.GetCount())
	return &queue.PeekResponse{WebRequest: webRequests}, err
}

func (q *Queue) Push(ctx context.Context, queueName string, webRequests []*queue.WebRequest) error {
	if err := q.validate(); err != nil {
		return err
	}
	conn := q.Pool.Get()
	defer conn.Close()
	key := q.key(queueName)
	for _, webRequest := range webRequests {
		protoBytes, err := proto.Marshal(webRequest)
		if err != nil {
			return errors.Wrap(err, "failed marshaling protobuf")
		}
		_, err = conn.Do("RPUSH", key, protoBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

func (q *Queue) Pop(ctx context.Context, queueName string, timeout int64) (*queue.WebRequest, error) {
	if err := q.validate(); err != nil {
		return nil, err
	}
	conn := q.Pool.Get()
	defer conn.Close()
	key := q.key(queueName)
	values, err := redis.ByteSlices(conn.Do("BLPOP", key, timeout))
	switch err {
	case nil:
	case redis.ErrNil:
		return nil, nil
	default:
		return nil, err
	}
	if len(values) < 2 {
		return nil, nil
	}
	webRequestBytes := values[1]
	webRequest := new(queue.WebRequest)
	err = proto.Unmarshal(webRequestBytes, webRequest)
	return webRequest, err
}

func (q *Queue) Peek(ctx context.Context, queueName string, count int64) ([]*queue.WebRequest, error) {
	response := make([]*queue.WebRequest, 0)
	if err := q.validate(); err != nil {
		return response, err
	}
	conn := q.Pool.Get()
	defer conn.Close()
	if count == 0 {
		count = 10
	}
	key := q.key(queueName)
	values, err := redis.ByteSlices(conn.Do("LRANGE", key, 0, count-1))
	switch err {
	case nil:
	case redis.ErrNil:
		return response, nil
	default:
		return response, err
	}
	for _, webRequestBytes := range values {
		webRequest := new(queue.WebRequest)
		err = proto.Unmarshal(webRequestBytes, webRequest)
		if err != nil {
			return response, err
		}
		response = append(response, webRequest)
	}
	return response, nil
}

func New(prefix string, pool *redis.Pool) *Queue {
	return &Queue{
		Prefix: prefix,
		Pool:   pool,
	}
}

func (q *Queue) QueueServer() *QueueServer {
	return &QueueServer{
		q: q,
	}
}

func (q *Queue) key(queueName string) string {
	return q.Prefix + ":" + queueName
}

func (q *Queue) validate() error {
	if q.Prefix == "" {
		return ErrEmptyPrefix
	}
	if q.Pool == nil {
		return ErrNilPool
	}
	return nil
}
