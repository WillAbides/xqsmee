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

func (q *Queue) Push(ctx context.Context, pushRequest *queue.PushRequest) (*queue.PushResponse, error) {
	response := new(queue.PushResponse)
	var err error
	if err = q.validate(); err != nil {
		return response, err
	}
	conn := q.Pool.Get()
	defer conn.Close()
	key := q.key(pushRequest.GetQueueName())
	for _, webRequest := range pushRequest.WebRequest {
		protoBytes, err := proto.Marshal(webRequest)
		if err != nil {
			return response, errors.Wrap(err, "failed marshaling protobuf")
		}
		_, err = conn.Do("RPUSH", key, protoBytes)
		if err != nil {
			return response, err
		}
	}
	return response, nil
}

func (q *Queue) Pop(ctx context.Context, popRequest *queue.PopRequest) (*queue.PopResponse, error) {
	response := new(queue.PopResponse)
	var err error
	if err = q.validate(); err != nil {
		return response, err
	}
	conn := q.Pool.Get()
	defer conn.Close()
	key := q.key(popRequest.GetQueueName())
	replyBytes, err := redis.Bytes(conn.Do("LPOP", key))
	switch err {
	case nil:
	case redis.ErrNil:
		return response, nil
	default:
		return response, err
	}

	webRequest := new(queue.WebRequest)
	err = proto.Unmarshal(replyBytes, webRequest)
	if err != nil {
		return response, err
	}
	response.WebRequest = webRequest
	return response, nil
}

func (q *Queue) BPop(ctx context.Context, popRequest *queue.BPopRequest) (*queue.BPopResponse, error) {
	response := new(queue.BPopResponse)
	if err := q.validate(); err != nil {
		return response, err
	}
	conn := q.Pool.Get()
	defer conn.Close()
	key := q.key(popRequest.GetQueueName())
	values, err := redis.ByteSlices(conn.Do("BLPOP", key, popRequest.GetTimeout()))
	switch err {
	case nil:
	case redis.ErrNil:
		return response, nil
	default:
		return response, err
	}
	if len(values) < 2 {
		return response, nil
	}
	webRequestBytes := values[1]
	webRequest := new(queue.WebRequest)
	err = proto.Unmarshal(webRequestBytes, webRequest)
	response.WebRequest = webRequest
	return response, err
}

func (q *Queue) Peek(ctx context.Context, peekRequest *queue.PeekRequest) (*queue.PeekResponse, error) {
	panic("implement me")
}

func New(prefix string, pool *redis.Pool) *Queue {
	return &Queue{
		Prefix: prefix,
		Pool:   pool,
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
