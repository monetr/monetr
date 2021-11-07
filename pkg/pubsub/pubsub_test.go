package pubsub

import (
	"context"
	"sync"
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

		go func() {
			defer wg.Done()
			for {
				select {
				case _ = <-deadline.C:
					t.Fatalf("pubsub deadline was reached before a notification was received")
					return
				case _ = <-listener.Channel():
					log.Info("NOTIFICATION RECEIVED")
					return
				}
			}
		}()

		log.Info("sending test notification")
		err = ps.Notify(context.Background(), channelName, "test")
		assert.NoError(t, err, "must be able to notify the channel")
		wg.Wait()
		assert.NoError(t, listener.Close(), "must close listener gracefully")
	})
}
