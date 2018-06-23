// +build slowtests

package redisqueue

import (
	"context"
	"github.com/WillAbides/xqsmee/queue"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQueueServer_BPop_Slow(t *testing.T) {
	t.Run("returns empty after timeout", func(t *testing.T) {
		tt := testSetup(t)
		got, err := tt.queueServer.BPop(context.Background(), &queue.BPopRequest{QueueName: "bar", Timeout: 1})
		assert.Nil(tt, err)
		assert.Empty(tt, got.WebRequest)
	})
}
