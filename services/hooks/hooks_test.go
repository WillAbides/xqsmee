package hooks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/WillAbides/idcheck"
	"github.com/WillAbides/xqsmee/queue"
	"github.com/WillAbides/xqsmee/queue/mockqueue"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testObjects struct {
	queue     *mockqueue.MockQueue
	service   *Service
	teardown  func()
	assert    *assert.Assertions
	require   *require.Assertions
	timestamp *timestamp.Timestamp
	now       *time.Time
	*testing.T
}

const testQueue = "deoQcZVCBM6UC1OIbTXWeg"

func testSetup(t *testing.T) *testObjects {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockQueue := mockqueue.NewMockQueue(ctrl)
	now := time.Now()
	ts, err := ptypes.TimestampProto(now)
	require.Nil(t, err)
	return &testObjects{
		service: New(mockQueue, idcheck.NewIDChecker()),
		queue:   mockQueue,
		teardown: func() {
			ctrl.Finish()
		},
		assert:    assert.New(t),
		require:   require.New(t),
		T:         t,
		timestamp: ts,
		now:       &now,
	}
}

func (tt *testObjects) doRequest(method, body, url string) *httptest.ResponseRecorder {
	tt.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	tt.require.Nil(err)
	res := httptest.NewRecorder()
	tt.service.Router().ServeHTTP(res, req)
	return res
}

func TestService_pingHandler(t *testing.T) {
	t.Run("pongs", func(t *testing.T) {
		tt := testSetup(t)
		defer tt.teardown()
		res := tt.doRequest(http.MethodGet, "", "/_ping")
		tt.assert.Equal(http.StatusOK, res.Code)
		tt.assert.Equal("pong", res.Body.String())
	})
}

func TestService_peekHandler(t *testing.T) {
	t.Run("works", func(t *testing.T) {
		tt := testSetup(t)
		defer tt.teardown()
		ret := []*queue.WebRequest{}
		for i := 0; i < 10; i++ {
			ret = append(ret, &queue.WebRequest{
				Body:       "hi",
				ReceivedAt: tt.timestamp,
				Header:     []*queue.Header{},
			})
		}
		exJSON, err := json.Marshal(ret)
		tt.require.Nil(err)
		tt.queue.EXPECT().Peek(gomock.Any(), testQueue, int64(0)).Return(ret, nil)
		res := tt.doRequest(http.MethodGet, "", "/q/"+testQueue)
		tt.assert.Equal(http.StatusOK, res.Code)
		body := strings.TrimSpace(res.Body.String())
		tt.assert.Equal(string(exJSON), body)
	})

	t.Run("subqueue", func(t *testing.T) {
		tt := testSetup(t)
		defer tt.teardown()
		ret := []*queue.WebRequest{}
		for i := 0; i < 10; i++ {
			ret = append(ret, &queue.WebRequest{
				Body:       "hi",
				ReceivedAt: tt.timestamp,
				Header:     []*queue.Header{},
			})
		}
		exJSON, err := json.Marshal(ret)
		tt.require.Nil(err)
		tt.queue.EXPECT().Peek(gomock.Any(), testQueue+"/subqueue", int64(0)).Return(ret, nil)
		res := tt.doRequest(http.MethodGet, "", "/q/"+testQueue+"/subqueue")
		tt.assert.Equal(http.StatusOK, res.Code)
		body := strings.TrimSpace(res.Body.String())
		tt.assert.Equal(string(exJSON), body)
	})
}

func TestService_postHandler(t *testing.T) {
	t.Run("works", func(t *testing.T) {
		tt := testSetup(t)
		defer tt.teardown()
		tt.service.receivedAtOverride = tt.now
		exWebRequest := &queue.WebRequest{
			Body:       "hi",
			ReceivedAt: tt.timestamp,
			Header:     []*queue.Header{},
		}
		tt.queue.EXPECT().Push(gomock.Any(), testQueue, []*queue.WebRequest{exWebRequest}).Return(nil)
		res := tt.doRequest(http.MethodPost, "hi", "/q/"+testQueue)
		tt.assert.Equal(http.StatusOK, res.Code)
	})

	t.Run("subqueue", func(t *testing.T) {
		tt := testSetup(t)
		defer tt.teardown()
		tt.service.receivedAtOverride = tt.now
		exWebRequest := &queue.WebRequest{
			Body:       "hi",
			ReceivedAt: tt.timestamp,
			Header:     []*queue.Header{},
		}
		tt.queue.EXPECT().Push(gomock.Any(), testQueue+"/foo", []*queue.WebRequest{exWebRequest}).Return(nil)
		res := tt.doRequest(http.MethodPost, "hi", "/q/"+testQueue+"/foo")
		tt.assert.Equal(http.StatusOK, res.Code)
	})

	t.Run("500 on queue error", func(t *testing.T) {
		tt := testSetup(t)
		defer tt.teardown()
		tt.service.receivedAtOverride = tt.now
		exWebRequest := &queue.WebRequest{
			Body:       "hi",
			ReceivedAt: tt.timestamp,
			Header:     []*queue.Header{},
		}
		tt.queue.EXPECT().Push(gomock.Any(), testQueue, []*queue.WebRequest{exWebRequest}).Return(assert.AnError)
		res := tt.doRequest(http.MethodPost, "hi", "/q/"+testQueue)
		tt.assert.Equal(http.StatusInternalServerError, res.Code)
	})

	t.Run("empty body", func(t *testing.T) {
		tt := testSetup(t)
		defer tt.teardown()
		tt.service.receivedAtOverride = tt.now
		exWebRequest := &queue.WebRequest{
			Body:       "",
			ReceivedAt: tt.timestamp,
			Header:     []*queue.Header{},
		}
		tt.queue.EXPECT().Push(gomock.Any(), testQueue, []*queue.WebRequest{exWebRequest}).Return(nil)
		res := tt.doRequest(http.MethodPost, "", "/q/"+testQueue)
		tt.assert.Equal(http.StatusOK, res.Code)
	})
}
