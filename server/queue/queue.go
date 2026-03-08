// Package queue is a concept for a new background job processing framework.
// Intended to simplify the approach used in the [background] package. This
// framework will have a cleaner external interface and will not require as much
// dependency injection as it will take a controller approach where job
// functions are always given the items they need.
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/billing"
	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/secrets"
	"github.com/monetr/monetr/server/storage"
	"github.com/pkg/errors"
)

type Context interface {
	context.Context
	DB() pg.DBI
	Publisher() pubsub.Publisher
	Platypus() platypus.Platypus
	KMS() secrets.KeyManagement
	Storage() storage.Storage
	Billing() billing.Billing
	Email() communication.EmailCommunication
	Processor() Processor
}

// JobFunction is any function (even outside of this package) that takes the job
// context interface and any generic arguments and returns an error.
type JobFunction[A any] func(ctx Context, args A) error

// internalJobWrapper is the function that is actually passed to the processor.
// This function has no generic types and always has the same signature. This
// way the job processer can be agnostic of the actual implementation of the job
// functions. This wrapper function is responsible for tracing, error handling.
// Retry logic should not be implemented here! Retry logic should be implemented
// in the processor layer.
type internalJobWrapper func(ctx Context, args []byte) error

// RetryableError is an error interface that can be returned by a job function.
// If this error interface is returned then the [Retryable] function is called
// with the number of attempts performed already, including the attempt that was
// just performed. If this method returns true, then the job is inserted back
// into the queue with the attempt counter incremented. If the method returns
// false then the job is not re-attempted.
type RetryableError interface {
	error
	Retryable(attempts int) bool
}

// Notifier is an interface that surfaces how the queue will know when to
// consume jobs from the queue. Notifications from this interface do not promise
// that a job can or will be consumed. Just that when a notification is sent on
// this interface, the consumer of the notification should attempt to consume a
// job from the queue. This interface will be implemented differently depending
// on the datastore that is backing the queue. For example, PostgreSQL will
// likely have a timer as well as a LISTEN/NOTIFY flow for job notifications.
// Where as SQLite will likely just use an in memory channel to notify since it
// is always a single process and polling the database won't make sense.
type Notifier interface {
	Channel() chan struct{}
}

type Processor interface {
	enqueue(ctx context.Context, queue string, args any) error
	register(ctx context.Context, queue string, job internalJobWrapper) error
	registerCron(ctx context.Context, queue string, schedule string, job internalJobWrapper) error
}

func Enqueue[T any](
	ctx context.Context,
	processor Processor,
	job JobFunction[T],
	args T,
) error {
	return processor.enqueue(ctx, queueNameFromJobFunction(job), args)
}

func Register[T any](
	ctx context.Context,
	processor Processor,
	job JobFunction[T],
) error {
	queue := queueNameFromJobFunction(job)

	processor.register(ctx, queue, func(ctx Context, argBytes []byte) error {
		var args T
		if err := json.Unmarshal(argBytes, args); err != nil {
			return errors.Wrapf(err, "failed to decode arguments for job: %s", queue)
		}

		// TODO, transaction wrapping?
		return job(ctx, args)
	})

	return nil
}

func RegisterCron[T any](
	ctx context.Context,
	processor Processor,
	schedule string,
	job JobFunction[T],
) error {
	queue := queueNameFromJobFunction(job)

	processor.registerCron(
		ctx,
		queue,
		schedule,
		func(ctx Context, argBytes []byte) error {
			var args T
			if err := json.Unmarshal(argBytes, args); err != nil {
				return errors.Wrapf(err, "failed to decode arguments for job: %s", queue)
			}

			// TODO, transaction wrapping?
			return job(ctx, args)
		},
	)

	return nil
}

// queueNameFromJobFunction takes the job function we want to enqueue or just
// known the name of for processing and derives the job queue name from it which
// is a string.
func queueNameFromJobFunction[T any](job JobFunction[T]) string {
	v := reflect.ValueOf(job)
	if v.Kind() != reflect.Func {
		panic(fmt.Sprintf("Enqueue expected a job function to be provided, instead got: %T", job))
	}
	pc := v.Pointer()
	f := runtime.FuncForPC(pc)
	return f.Name()
}
