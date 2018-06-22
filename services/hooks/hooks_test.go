package hooks

import (
	"github.com/WillAbides/xqsmee/queue"
	"github.com/WillAbides/xqsmee/queue/mockqueue"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/jsonpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testObjects struct {
	queue    *mockqueue.MockQueueServer
	service  *Service
	teardown func()
	assert   *assert.Assertions
	require  *require.Assertions
	*testing.T
}

func testSetup(t *testing.T) *testObjects {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockQueue := mockqueue.NewMockQueueServer(ctrl)

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
		expectedPushRequest := queue.NewPushRequest("asdf", &queue.WebRequest{
			Body:       "hi",
			ReceivedAt: 12,
			Header:     []*queue.Header{},
		})
		tt.queue.EXPECT().Push(gomock.Any(), expectedPushRequest).Return(&queue.PushResponse{}, nil)
		res := tt.doRequest(http.MethodPost, "hi", "/asdf")
		tt.assert.Equal(http.StatusOK, res.Code)
	})

	t.Run("500 on queue error", func(t *testing.T) {
		tt := testSetup(t)
		defer tt.teardown()
		tt.service.receivedAtOverride = 12
		expectedPushRequest := queue.NewPushRequest("asdf", &queue.WebRequest{
			Body:       "hi",
			ReceivedAt: 12,
			Header:     []*queue.Header{},
		})
		tt.queue.EXPECT().Push(gomock.Any(), expectedPushRequest).Return(&queue.PushResponse{}, assert.AnError)
		res := tt.doRequest(http.MethodPost, "hi", "/asdf")
		tt.assert.Equal(http.StatusInternalServerError, res.Code)
	})

	t.Run("empty body", func(t *testing.T) {
		tt := testSetup(t)
		defer tt.teardown()
		tt.service.receivedAtOverride = 12
		expectedPushRequest := queue.NewPushRequest("asdf", &queue.WebRequest{
			Body:       "",
			ReceivedAt: 12,
			Header:     []*queue.Header{},
		})
		tt.queue.EXPECT().Push(gomock.Any(), expectedPushRequest).Return(&queue.PushResponse{}, nil)
		res := tt.doRequest(http.MethodPost, "", "/asdf")
		tt.assert.Equal(http.StatusOK, res.Code)
	})
}

func TestService_popHandler(t *testing.T) {
	tt := testSetup(t)
	defer tt.teardown()
	popRequest := &queue.PopRequest{
		QueueName: "asdf",
	}

	webRequest := &queue.WebRequest{
		Body: "hi",
	}

	popResponse := &queue.PopResponse{
		WebRequest: webRequest,
	}
	tt.queue.EXPECT().Pop(gomock.Any(), popRequest).Return(popResponse, nil)
	res := tt.doRequest(http.MethodGet, "", "/asdf/pop")
	tt.assert.Equal(http.StatusOK, res.Code)
	got := new(queue.WebRequest)
	err := jsonpb.Unmarshal(res.Body, got)
	tt.assert.Nil(err)
	tt.assert.Equal(webRequest, got)
}
