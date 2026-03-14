package queue

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func helloWorld(ctx Context, args map[string]any) error {
	fmt.Println("Hello world!")
	return nil
}

func cronHello(ctx Context) error {
	fmt.Println("Hello world!")
	return nil
}

func TestJobSignature(t *testing.T) {
	t.Run("same args and timestamp produce same signature", func(t *testing.T) {
		ts := time.Now()
		args := []byte(`{"foo":"bar"}`)
		assert.Equal(t, jobSignature(ts, args), jobSignature(ts, args))
	})

	t.Run("different args produce different signatures", func(t *testing.T) {
		ts := time.Now()
		assert.NotEqual(t,
			jobSignature(ts, []byte(`{"a":1}`)),
			jobSignature(ts, []byte(`{"a":2}`)),
		)
	})

	t.Run("different timestamps produce different signatures", func(t *testing.T) {
		args := []byte(`{"foo":"bar"}`)
		t1 := time.Now().Truncate(time.Second)
		t2 := t1.Add(time.Second)
		assert.NotEqual(t, jobSignature(t1, args), jobSignature(t2, args))
	})

	t.Run("timestamp is truncated to the second", func(t *testing.T) {
		args := []byte(`{"foo":"bar"}`)
		base := time.Now().Truncate(time.Second)
		// Two timestamps within the same second must produce the same signature
		assert.Equal(t,
			jobSignature(base, args),
			jobSignature(base.Add(500*time.Millisecond), args),
		)
	})
}

func TestQueueNameFromJobFunction(t *testing.T) {
	t.Run("anonymous function", func(t *testing.T) {
		type Args struct {
		}
		queueName := QueueNameFromJobFunction[Args](
			JobFunction[Args](func(ctx Context, args Args) error {
				return nil
			}),
		)
		assert.Equal(t, "queue-TestQueueNameFromJobFunction-func1-1", queueName)
	})

	t.Run("variable function", func(t *testing.T) {
		type Args struct {
		}
		jobFunction := JobFunction[Args](func(ctx Context, args Args) error {
			return nil
		})
		queueName := QueueNameFromJobFunction[Args](jobFunction)
		assert.Equal(t, "queue-TestQueueNameFromJobFunction-func2-1", queueName)
	})

	t.Run("regular function", func(t *testing.T) {
		queueName := QueueNameFromJobFunction[map[string]any](helloWorld)
		assert.Equal(t, "queue-helloWorld", queueName)
	})

	t.Run("regular cron function", func(t *testing.T) {
		queueName := QueueNameFromJobFunction[any](cronHello)
		assert.Equal(t, "queue-cronHello", queueName)
	})
}
