package redisqueue

import (
	"context"
	"github.com/WillAbides/xqsmee/queue"
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
	"os"
)

var redisPool = &redis.Pool{
	Dial: func() (redis.Conn, error) { return redis.Dial("tcp", ":6379", redis.DialDatabase(10)) },
}

type testObjects struct {
	queueServer     *QueueServer
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

	webRequest, webRequestBytes := newWebRequestAndBytes(t, "foo")

	q := &Queue{
		Prefix: "foo",
		Pool:   redisPool,
	}

	return &testObjects{
		T:               t,
		queue:           q,
		queueServer:     &QueueServer{q: q},
		webRequest:      webRequest,
		webRequestBytes: webRequestBytes,
	}
}

func (tt *testObjects) slowTest() *testObjects {
	tt.Helper()
	if os.Getenv("SLOW_TESTS") == "" {
		tt.Skip("skipping slow test")
	}
	return tt
}

func newWebRequestAndBytes(t *testing.T, body string) (*queue.WebRequest, []byte) {
	t.Helper()
	wr := &queue.WebRequest{
		Body: body,
		Header: []*queue.Header{
			{Name: "fakeheader", Value: []string{"hi"}},
			{Name: "fakeheader2", Value: []string{"hi", "bye"}},
		},
		ReceivedAt: 4,
		Host:       "yomamashost",
	}
	wrb, err := proto.Marshal(wr)
	require.Nil(t, err)
	return wr, wrb
}

func TestQueue_Push(t *testing.T) {
	t.Run("works", func(t *testing.T) {
		tt := testSetup(t)
		err := tt.queue.Push(context.Background(), "bar", []*queue.WebRequest{tt.webRequest})
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
		err := tt.queue.Push(context.Background(), "bar", []*queue.WebRequest{tt.webRequest})
		assert.Equal(tt, ErrEmptyPrefix, err)
	})
}

func TestQueueServer_Pop(t *testing.T) {
	t.Run("works", func(t *testing.T) {
		tt := testSetup(t)
		conn := redisPool.Get()
		defer conn.Close()
		_, err := conn.Do("RPUSH", "foo:bar", tt.webRequestBytes)
		assert.Nil(tt, err)
		got, err := tt.queueServer.Pop(context.Background(), &queue.PopRequest{QueueName: "bar", Timeout: 0})
		assert.Nil(tt, err)
		assert.Equal(tt, tt.webRequest, got.WebRequest)
	})

	t.Run("blocks", func(t *testing.T) {
		tt := testSetup(t)
		gotChan := make(chan *queue.PopResponse, 1)
		errChan := make(chan error, 1)
		go func() {
			got, err := tt.queueServer.Pop(context.Background(), &queue.PopRequest{QueueName: "bar", Timeout: 0})
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

	t.Run("returns empty after timeout", func(t *testing.T) {
		tt := testSetup(t).slowTest()
		got, err := tt.queueServer.Pop(context.Background(), &queue.PopRequest{QueueName: "bar", Timeout: 1})
		assert.Nil(tt, err)
		assert.Empty(tt, got.WebRequest)
	})
}

func TestQueueServer_Peek(t *testing.T) {
	t.Run("works", func(t *testing.T) {
		tt := testSetup(t)
		conn := redisPool.Get()
		defer conn.Close()
		for i := 0; i < 20; i++ {
			body := strconv.Itoa(i)
			_, wrb := newWebRequestAndBytes(t, body)
			_, err := conn.Do("RPUSH", "foo:bar", wrb)
			require.Nil(tt, err)
		}
		response, err := tt.queueServer.Peek(context.Background(), &queue.PeekRequest{QueueName: "bar", Count: 15})
		assert.Nil(t, err)
		for i := 0; i < 15; i++ {
			exbody := strconv.Itoa(i)
			body := response.WebRequest[i].GetBody()
			assert.Equal(t, exbody, body)
		}
	})

	t.Run("works on empty queue", func(t *testing.T) {
		tt := testSetup(t)
		response, err := tt.queueServer.Peek(context.Background(), &queue.PeekRequest{QueueName: "bar", Count: 15})
		assert.Nil(t, err)
		assert.Equal(tt, 0, len(response.GetWebRequest()))
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
