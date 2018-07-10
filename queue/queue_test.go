package queue_test

import (
	"context"
	"testing"
	"time"

	"github.com/WillAbides/xqsmee/queue"
	"github.com/WillAbides/xqsmee/queue/mockqueue"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testObjects struct {
	queue      *mockqueue.MockQueue
	teardown   func()
	assert     *assert.Assertions
	require    *require.Assertions
	webRequest *queue.WebRequest
	*testing.T
}

func testSetup(t *testing.T) *testObjects {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockQueue := mockqueue.NewMockQueue(ctrl)

	return &testObjects{
		queue: mockQueue,
		teardown: func() {
			ctrl.Finish()
		},
		assert:     assert.New(t),
		require:    require.New(t),
		T:          t,
		webRequest: &queue.WebRequest{Body: "hi"},
	}
}

func TestGRPCHandler_Pop(t *testing.T) {
	tt := testSetup(t)
	defer tt.teardown()
	tt.queue.EXPECT().Pop(gomock.Any(), "asdf", 12*time.Second).Return(tt.webRequest, nil)
	popRequest := &queue.PopRequest{QueueName: "asdf", Timeout: ptypes.DurationProto(12 * time.Second)}
	grpcHandler := queue.NewGRPCHandler(tt.queue)
	response, err := grpcHandler.Pop(context.Background(), popRequest)
	tt.assert.Nil(err)
	tt.assert.Equal(tt.webRequest, response.GetWebRequest())
}

func TestGRPCHandler_Peek(t *testing.T) {
	tt := testSetup(t)
	defer tt.teardown()
	expect := []*queue.WebRequest{tt.webRequest, tt.webRequest, tt.webRequest}
	tt.queue.EXPECT().Peek(gomock.Any(), "asdf", int64(12)).Return(expect, nil)
	peekRequest := &queue.PeekRequest{QueueName: "asdf", Count: 12}
	grpcHandler := queue.NewGRPCHandler(tt.queue)
	response, err := grpcHandler.Peek(context.Background(), peekRequest)
	tt.assert.Nil(err)
	tt.assert.Equal(expect, response.GetWebRequest())
}
