package redisqueue

import (
	"context"
	"github.com/WillAbides/xqsmee/queue"
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var redisPool = &redis.Pool{
	MaxIdle:     100,
	MaxActive:   100,
	IdleTimeout: 60 * time.Second,
	Wait:        true,
	Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", ":6379", redis.DialDatabase(10)) },
	TestOnBorrow: func(c redis.Conn, t time.Time) error {
		_, err := c.Do("PING")
		return err
	},
}

type testObjects struct {
	queue           *Queue
	webRequest      *queue.WebRequest
	webRequestBytes []byte
	*testing.T
}

func testSetup(t *testing.T) *testObjects {
	t.Helper()

	conn := redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("FLUSHDB")
	require.Nil(t, err)

	webRequest := &queue.WebRequest{
		Body: "foo",
		Header: []*queue.Header{
			{Name: "fakeheader", Value: []string{"hi"}},
			{Name: "fakeheader2", Value: []string{"hi", "bye"}},
		},
		ReceivedAt: 4,
		Host:       "yomamashost",
	}

	webRequestBytes, err := proto.Marshal(webRequest)
	require.Nil(t, err)

	return &testObjects{
		T: t,
		queue: &Queue{
			Prefix: "foo",
			Pool:   redisPool,
		},
		webRequest:      webRequest,
		webRequestBytes: webRequestBytes,
	}
}

func TestQueue_Push(t *testing.T) {
	t.Run("works", func(t *testing.T) {
		tt := testSetup(t)
		pushResponse, err := tt.queue.Push(context.Background(), queue.NewPushRequest("bar", tt.webRequest))
		assert.Empty(t, pushResponse)
		assert.Nil(tt, err)
		conn := redisPool.Get()
		defer conn.Close()
		reply, err := redis.Values(conn.Do("LRANGE", "foo:bar", 0, -1))
		assert.Nil(tt, err)
		assert.Equal(tt, tt.webRequestBytes, reply[0])
	})

	t.Run("errors on validation error", func(t *testing.T) {
		tt := testSetup(t)
		tt.queue.Prefix = ""
		pushResponse, err := tt.queue.Push(context.Background(), queue.NewPushRequest("bar", tt.webRequest))
		assert.Empty(t, pushResponse)
		assert.Equal(tt, ErrEmptyPrefix, err)
	})
}

func TestQueue_Pop(t *testing.T) {
	t.Run("works", func(t *testing.T) {
		tt := testSetup(t)
		conn := redisPool.Get()
		defer conn.Close()
		_, err := conn.Do("RPUSH", "foo:bar", tt.webRequestBytes)
		assert.Nil(tt, err)
		got, err := tt.queue.Pop(context.Background(), &queue.PopRequest{QueueName: "bar"})
		assert.Nil(tt, err)
		assert.Equal(tt, tt.webRequest, got.WebRequest)
	})

	t.Run("empty response when queue is empty", func(t *testing.T) {
		tt := testSetup(t)
		got, err := tt.queue.Pop(context.Background(), &queue.PopRequest{QueueName: "bar"})
		assert.Nil(tt, err)
		assert.Empty(tt, got.GetWebRequest())
	})

	t.Run("errors on validation error", func(t *testing.T) {
		tt := testSetup(t)
		tt.queue.Prefix = ""
		got, err := tt.queue.Pop(context.Background(), &queue.PopRequest{QueueName: "bar"})
		assert.Equal(tt, ErrEmptyPrefix, err)
		assert.Empty(tt, got)
	})
}

func TestQueue_BPop(t *testing.T) {
	t.Run("works", func(t *testing.T) {
		tt := testSetup(t)
		conn := redisPool.Get()
		defer conn.Close()
		_, err := conn.Do("RPUSH", "foo:bar", tt.webRequestBytes)
		assert.Nil(tt, err)
		got, err := tt.queue.BPop(context.Background(), &queue.BPopRequest{QueueName: "bar", Timeout: 0})
		assert.Nil(tt, err)
		assert.Equal(tt, tt.webRequest, got.WebRequest)
	})

	t.Run("blocks", func(t *testing.T) {
		tt := testSetup(t)
		gotChan := make(chan *queue.BPopResponse, 1)
		errChan := make(chan error, 1)
		go func() {
			got, err := tt.queue.BPop(context.Background(), &queue.BPopRequest{QueueName: "bar", Timeout: 0})
			gotChan <- got
			errChan <- err
		}()
		conn := redisPool.Get()
		defer conn.Close()
		_, err := conn.Do("RPUSH", "foo:bar", tt.webRequestBytes)
		assert.Nil(tt, err)
		assert.Nil(tt, <-errChan)
		got := <-gotChan
		assert.Equal(tt, tt.webRequest, got.GetWebRequest())
	})
}

func TestQueue_validate(t *testing.T) {
	t.Run("no error on valid", func(t *testing.T) {
		tt := testSetup(t)
		assert.Nil(tt, tt.queue.validate())
	})

	t.Run("checks for empty prefix", func(t *testing.T) {
		tt := testSetup(t)
		tt.queue.Prefix = ""
		assert.Equal(tt, ErrEmptyPrefix, tt.queue.validate())
	})

	t.Run("checks for nil pool", func(t *testing.T) {
		tt := testSetup(t)
		tt.queue.Pool = nil
		assert.Equal(tt, ErrNilPool, tt.queue.validate())
	})
}
