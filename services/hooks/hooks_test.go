package hooks

import (
	"github.com/WillAbides/xqsmee/queue"
	"github.com/WillAbides/xqsmee/queue/mockqueue"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testObjects struct {
	queue    *mockqueue.MockQueue
	service  *Service
	teardown func()
	assert   *assert.Assertions
	require  *require.Assertions
	*testing.T
}

func testSetup(t *testing.T) *testObjects {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockQueue := mockqueue.NewMockQueue(ctrl)

	return &testObjects{
		service: New(mockQueue),
		queue:   mockQueue,
		teardown: func() {
			ctrl.Finish()
		},
		assert:  assert.New(t),
		require: require.New(t),
		T:       t,
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

func TestService_postHandler(t *testing.T) {
	t.Run("works", func(t *testing.T) {
		tt := testSetup(t)
		defer tt.teardown()
		tt.service.receivedAtOverride = 12
		exWebRequest := &queue.WebRequest{
			Body:       "hi",
			ReceivedAt: 12,
			Header:     []*queue.Header{},
		}
		tt.queue.EXPECT().Push(gomock.Any(), "asdf", []*queue.WebRequest{exWebRequest}).Return(nil)
		res := tt.doRequest(http.MethodPost, "hi", "/asdf")
		tt.assert.Equal(http.StatusOK, res.Code)
	})

	t.Run("500 on queue error", func(t *testing.T) {
		tt := testSetup(t)
		defer tt.teardown()
		tt.service.receivedAtOverride = 12
		exWebRequest := &queue.WebRequest{
			Body:       "hi",
			ReceivedAt: 12,
			Header:     []*queue.Header{},
		}
		tt.queue.EXPECT().Push(gomock.Any(), "asdf", []*queue.WebRequest{exWebRequest}).Return(assert.AnError)
		res := tt.doRequest(http.MethodPost, "hi", "/asdf")
		tt.assert.Equal(http.StatusInternalServerError, res.Code)
	})

	t.Run("empty body", func(t *testing.T) {
		tt := testSetup(t)
		defer tt.teardown()
		tt.service.receivedAtOverride = 12
		exWebRequest := &queue.WebRequest{
			Body:       "",
			ReceivedAt: 12,
			Header:     []*queue.Header{},
		}
		tt.queue.EXPECT().Push(gomock.Any(), "asdf", []*queue.WebRequest{exWebRequest}).Return(nil)
		res := tt.doRequest(http.MethodPost, "", "/asdf")
		tt.assert.Equal(http.StatusOK, res.Code)
	})
}
