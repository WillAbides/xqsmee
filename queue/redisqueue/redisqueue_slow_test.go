// +build slowtests

package redisqueue

import (
	"context"
	"github.com/WillAbides/xqsmee/queue"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQueueServer_Pop_Slow(t *testing.T) {
	t.Run("returns empty after timeout", func(t *testing.T) {
		tt := testSetup(t)
		got, err := tt.queueServer.Pop(context.Background(), &queue.PopRequest{QueueName: "bar", Timeout: 1})
		assert.Nil(tt, err)
		assert.Empty(tt, got.WebRequest)
	})
}
