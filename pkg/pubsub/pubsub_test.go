package pubsub

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestPostgresPubSub_Notify(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		db := testutils.GetPgDatabase(t)
		channelName := gofakeit.UUID()
		log := testutils.GetLog(t).WithField("channel", channelName)

		ps := NewPostgresPubSub(log, db)

		listener, err := ps.Subscribe(context.Background(), channelName)
		assert.NoError(t, err, "must not receive an error just trying to subscribe to a channel")

		var wg sync.WaitGroup
		wg.Add(1)

		deadline := time.NewTimer(10 * time.Second)
		var counter int64
		go func() {
			defer wg.Done()
			time.Sleep(1 * time.Second)
			log.Info("sending test notification")
			err = ps.Notify(context.Background(), channelName, "test")
		}()

		select {
		case <-deadline.C:
			log.Fatal("pubsub deadline was reached before a notification was received")
			t.FailNow()
			return
		case <-listener.Channel():
			log.Info("NOTIFICATION RECEIVED")
			atomic.AddInt64(&counter, 1)
		}

		assert.NoError(t, listener.Close(), "must close listener gracefully")
		assert.Equal(t, int64(1), atomic.LoadInt64(&counter), "counter should be incremented")
		wg.Wait() // Wait for go routine to exit.
		assert.NoError(t, err, "must be able to notify on channel")
	})
}
