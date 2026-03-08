package queue

import (
	"fmt"
	"testing"

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

func TestQueueNameFromJobFunction(t *testing.T) {
	t.Run("anonymous function", func(t *testing.T) {
		type Args struct {
		}
		queueName := queueNameFromJobFunction[Args](
			JobFunction[Args](func(ctx Context, args Args) error {
				return nil
			}),
		)
		assert.Equal(t, "queue.TestQueueNameFromJobFunction.func1.1::queue.Args", queueName)
	})

	t.Run("variable function", func(t *testing.T) {
		type Args struct {
		}
		jobFunction := JobFunction[Args](func(ctx Context, args Args) error {
			return nil
		})
		queueName := queueNameFromJobFunction[Args](jobFunction)
		assert.Equal(t, "queue.TestQueueNameFromJobFunction.func2.1::queue.Args", queueName)
	})

	t.Run("regular function", func(t *testing.T) {
		queueName := queueNameFromJobFunction[map[string]any](helloWorld)
		assert.Equal(t, "queue.helloWorld::map[string]interface", queueName)
	})

	t.Run("regular cron function", func(t *testing.T) {
		queueName := queueNameFromJobFunction[any](cronHello)
		assert.Equal(t, "queue.cronHello", queueName)
	})
}
