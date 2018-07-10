package redisqueue

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/WillAbides/xqsmee/queue"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var redisPool = &redis.Pool{
	MaxActive: 100,
	Wait:      true,
	Dial: func() (redis.Conn, error) {
		return redis.DialURL("redis://:6379/10")
	},
	TestOnBorrow: func(c redis.Conn, t time.Time) error {
		_, err := c.Do("PING")
		return err
	},
}

type testObjects struct {
	queue           *Queue
	webRequest      *queue.WebRequest
	webRequestBytes []byte
	assert          *assert.Assertions
	require         *require.Assertions
	timestamp       *timestamp.Timestamp
	*testing.T
}

func testSetup(t *testing.T) *testObjects {
	t.Helper()

	conn := redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("FLUSHDB")
	require.Nil(t, err)

	now := time.Now()
	ts, err := ptypes.TimestampProto(now)
	require.Nil(t, err)

	webRequest, webRequestBytes := newWebRequestAndBytes(t, "foo", ts)

	q := &Queue{
		Prefix: "foo",
		Pool:   redisPool,
	}

	return &testObjects{
		T:               t,
		queue:           q,
		webRequest:      webRequest,
		webRequestBytes: webRequestBytes,
		assert:          assert.New(t),
		require:         require.New(t),
		timestamp:       ts,
	}
}

func newWebRequestAndBytes(t *testing.T, body string, receivedAt *timestamp.Timestamp) (*queue.WebRequest, []byte) {
	t.Helper()
	wr := &queue.WebRequest{
		Body: body,
		Header: []*queue.Header{
			{Name: "fakeheader", Value: []string{"hi"}},
			{Name: "fakeheader2", Value: []string{"hi", "bye"}},
		},
		ReceivedAt: receivedAt,
		Host:       "yomamashost",
	}
	wrb, err := proto.Marshal(wr)
	require.Nil(t, err)
	return wr, wrb
}

func TestQueue_Push(t *testing.T) {
	t.Run("works", func(t *testing.T) {
		tt := testSetup(t)
		done := make(chan struct{})
		psc := redis.PubSubConn{Conn: redisPool.Get()}
		defer psc.Conn.Close()
		tt.require.Nil(psc.Subscribe("foo:bar"))
		go func() {
			psc.ReceiveWithTimeout(100 * time.Millisecond)
			msg, ok := psc.ReceiveWithTimeout(100 * time.Millisecond).(redis.Message)
			tt.assert.True(ok)
			tt.assert.Equal("new", string(msg.Data))
			close(done)
		}()
		err := tt.queue.Push(context.Background(), "bar", []*queue.WebRequest{tt.webRequest})
		tt.assert.Nil(err)
		conn := redisPool.Get()
		defer conn.Close()
		reply, err := redis.Values(conn.Do("LRANGE", "foo:bar", 0, -1))
		tt.assert.Nil(err)
		tt.assert.Equal(tt.webRequestBytes, reply[0])
		<-done
	})

	t.Run("errors on validation error", func(t *testing.T) {
		tt := testSetup(t)
		tt.queue.Prefix = ""
		err := tt.queue.Push(context.Background(), "bar", []*queue.WebRequest{tt.webRequest})
		tt.assert.Equal(ErrEmptyPrefix, err)
	})
}

func TestQueue_Pop(t *testing.T) {
	t.Run("works", func(t *testing.T) {
		tt := testSetup(t)
		conn := redisPool.Get()
		defer conn.Close()
		_, err := conn.Do("RPUSH", "foo:bar", tt.webRequestBytes)
		tt.assert.Nil(err)
		got, err := tt.queue.Pop(context.Background(), "bar", 100*time.Millisecond)
		tt.assert.Nil(err)
		tt.assert.True(proto.Equal(tt.webRequest, got))
	})

	t.Run("blocks", func(t *testing.T) {
		tt := testSetup(t)
		gotChan := make(chan *queue.WebRequest, 1)
		errChan := make(chan error, 1)
		go func() {
			got, err := tt.queue.Pop(context.Background(), "bar", time.Second)
			gotChan <- got
			errChan <- err
		}()
		conn := redisPool.Get()
		defer conn.Close()
		time.Sleep(10 * time.Millisecond)
		_, err := conn.Do("RPUSH", "foo:bar", tt.webRequestBytes)
		tt.assert.Nil(err)
		_, err = conn.Do("PUBLISH", "foo:bar", "1")
		tt.assert.Nil(err)
		tt.assert.Nil(<-errChan)
		got := <-gotChan
		tt.assert.True(proto.Equal(tt.webRequest, got))
	})

	t.Run("returns empty after timeout", func(t *testing.T) {
		tt := testSetup(t)
		got, err := tt.queue.Pop(context.Background(), "bar", 100*time.Millisecond)
		tt.assert.Nil(err)
		tt.assert.Empty(got)
	})
}

func TestQueue_Peek(t *testing.T) {
	t.Run("works", func(t *testing.T) {
		tt := testSetup(t)
		conn := redisPool.Get()
		defer conn.Close()
		for i := 0; i < 20; i++ {
			body := strconv.Itoa(i)
			_, wrb := newWebRequestAndBytes(t, body, tt.timestamp)
			_, err := conn.Do("RPUSH", "foo:bar", wrb)
			tt.require.Nil(err)
		}
		response, err := tt.queue.Peek(context.Background(), "bar", 15)
		assert.Nil(t, err)
		for i := 0; i < 15; i++ {
			exbody := strconv.Itoa(i)
			body := response[i].GetBody()
			assert.Equal(t, exbody, body)
		}
	})

	t.Run("works on empty queue", func(t *testing.T) {
		tt := testSetup(t)
		response, err := tt.queue.Peek(context.Background(), "bar", 15)
		tt.assert.Nil(err)
		tt.assert.Equal(0, len(response))
	})
}

func TestQueue_validate(t *testing.T) {
	t.Run("no error on valid", func(t *testing.T) {
		tt := testSetup(t)
		tt.assert.Nil(tt.queue.validate())
	})

	t.Run("checks for empty prefix", func(t *testing.T) {
		tt := testSetup(t)
		tt.queue.Prefix = ""
		tt.assert.Equal(ErrEmptyPrefix, tt.queue.validate())
	})

	t.Run("checks for nil pool", func(t *testing.T) {
		tt := testSetup(t)
		tt.queue.Pool = nil
		tt.assert.Equal(ErrNilPool, tt.queue.validate())
	})
}
