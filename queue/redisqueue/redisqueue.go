package redisqueue

import (
	"context"
	"github.com/WillAbides/xqsmee/queue"
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"sync"
	"time"
)

var (
	ErrEmptyPrefix = errors.New("prefix is empty")
	ErrNilPool     = errors.New("pool is nil")
)

type Queue struct {
	Prefix string
	Pool   *redis.Pool
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
		_, err = conn.Do("PUBLISH", key, "new")
		if err != nil {
			return err
		}
	}
	return nil
}

// listenPubSubChannels listens for messages on Redis pubsub channels. The
// onStart function is called after the channels are subscribed. The onMessage
// function is called for each message.
func listenPubSubChannel(ctx context.Context, pool *redis.Pool,
	doPop func() error,
	channel string) error {
	// A ping is set to the server with this period to test for the health of
	// the connection and server.
	const healthCheckPeriod = time.Minute

	c, err := pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	psc := redis.PubSubConn{Conn: c}

	if err := psc.Subscribe(channel); err != nil {
		return err
	}

	done := make(chan error, 1)

	// Start a goroutine to receive notifications from the server.
	go func() {
		for {
			switch n := psc.Receive().(type) {
			case error:
				done <- n
				return
			case redis.Message:
				if err := doPop(); err != nil {
					done <- err
					return
				}
			case redis.Subscription:
				switch n.Count {
				case 1:
					if err := doPop(); err != nil {
						done <- err
						return
					}
				case 0:
					// Return from the goroutine when all channels are unsubscribed.
					done <- nil
					return
				}
			}
		}
	}()

	ticker := time.NewTicker(healthCheckPeriod)
	defer ticker.Stop()
loop:
	for err == nil {
		select {
		case <-ticker.C:
			// Send ping to test health of connection and server. If
			// corresponding pong is not received, then receive on the
			// connection will timeout and the receive goroutine will exit.
			if err = psc.Ping(""); err != nil {
				break loop
			}
		case <-ctx.Done():
			break loop
		case err := <-done:
			// Return error from the receive goroutine.
			return err
		}
	}

	// Signal the receiving goroutine to exit by unsubscribing
	_ = psc.Unsubscribe(channel)

	// Wait for goroutine to complete.
	return <-done
}

func (q *Queue) Pop(ctx context.Context, queueName string, timeout time.Duration) (*queue.WebRequest, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	if err := q.validate(); err != nil {
		return nil, err
	}
	conn := q.Pool.Get()
	defer conn.Close()
	key := q.key(queueName)

	cancelChan := make(chan struct{})

	go func() {
		select {
		case <-cancelChan:
			cancel()
		case <-ctx.Done():
		}
	}()

	var requestMux = &sync.Mutex{}
	var webRequest *queue.WebRequest

	doPop := func() error {
		requestMux.Lock()
		defer requestMux.Unlock()
		if webRequest != nil {
			close(cancelChan)
			return nil
		}
		var err error
		conn := q.Pool.Get()
		defer conn.Close()
		webRequest, err = lpop(key, conn)
		if err != nil {
			return err
		}
		if webRequest != nil {
			close(cancelChan)
		}
		return nil
	}

	err := listenPubSubChannel(ctx, q.Pool, doPop, key)
	return webRequest, err
}

func lpop(key string, conn redis.Conn) (*queue.WebRequest, error) {
	value, err := redis.Bytes(conn.Do("LPOP", key))
	switch err {
	case nil:
	case redis.ErrNil:
		return nil, nil
	default:
		return nil, err
	}
	webRequest := new(queue.WebRequest)
	err = proto.Unmarshal(value, webRequest)
	return webRequest, err
}

func (q *Queue) blpop(ctx context.Context, queueName string, timeout int64) (*queue.WebRequest, error) {
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

func (q *Queue) Pop2(ctx context.Context, queueName string, timeout int64) (*queue.WebRequest, error) {
	return q.blpop(ctx, queueName, timeout)
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
