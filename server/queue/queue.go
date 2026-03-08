// Package queue is a concept for a new background job processing framework.
// Intended to simplify the approach used in the [background] package. This
// framework will have a cleaner external interface and will not require as much
// dependency injection as it will take a controller approach where job
// functions are always given the items they need.
package queue

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/billing"
	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/crumbs"
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

// CronFunction is similar to the [JobFunction] but this one cannot take any
// dynamic arguments. Instead it is simply provided the job context.
type CronFunction func(ctx Context) error

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
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	return processor.enqueue(span.Context(), queueNameFromJobFunction[T](job), args)
}

func Register[T any](
	ctx context.Context,
	processor Processor,
	job JobFunction[T],
) error {
	queue := queueNameFromJobFunction[T](job)

	processor.register(ctx, queue, func(ctx Context, argBytes []byte) error {
		var args T
		if err := decodeArguments(argBytes, &args); err != nil {
			return errors.Wrapf(err, "failed to decode arguments for job: %s", queue)
		}

		// TODO, transaction wrapping?
		return job(ctx, args)
	})

	return nil
}

func RegisterCron(
	ctx context.Context,
	processor Processor,
	schedule string,
	job CronFunction,
) error {
	queue := queueNameFromJobFunction[any](job)

	processor.registerCron(
		ctx,
		queue,
		schedule,
		// Cron jobs don't need their argument bytes, but are represented as the
		// same signature internaly for simplicity.
		func(ctx Context, _ []byte) error {
			// TODO, transaction wrapping?
			// TODO, add monitor check in with sentry here!
			return job(ctx)
		},
	)

	return nil
}

// queueNameFromJobFunction takes the job function we want to enqueue or just
// known the name of for processing and derives the job queue name from it which
// is a string.
func queueNameFromJobFunction[T any](job any) string {
	// Make sure that the provided job matches one of our expected types otherwise
	// panic.
	var args string
	switch job.(type) {
	case func(ctx Context, args T) error:
		args = fmt.Sprintf("::%T", *new(T))
	case JobFunction[T]:
		args = fmt.Sprintf("::%T", *new(T))
	case func(ctx Context) error:
	case CronFunction:
	default:
		panic(fmt.Sprintf("Expected a job function to be provided, instead got: %T", job))
	}
	pc := reflect.ValueOf(job).Pointer()
	f := runtime.FuncForPC(pc)
	name := strings.TrimPrefix(f.Name(), "github.com/monetr/monetr/server/")
	name = name + args
	name = strings.ReplaceAll(name, "{}", "")
	name = strings.TrimSpace(name)
	return name
}

func encodeArguments(args any) ([]byte, error) {
	result, err := json.Marshal(args)
	return result, errors.WithStack(err)
}

func decodeArguments[T any](data []byte, result *T) error {
	return errors.WithStack(json.Unmarshal(data, result))
}

// jobSignature takes the timestamp that the job is going to be enqueued at as
// well as the arguments for the job (if there are any) and generates a hash
// that is used to prevent an identical job from being enqueued at the same
// time.
func jobSignature(timestamp time.Time, arguments []byte) string {
	truncatedTimestamp := timestamp.Truncate(time.Second)
	signatureBuilder := fnv.New32()
	signatureBuilder.Write(arguments)
	signatureBuilder.Write([]byte(truncatedTimestamp.String()))
	return hex.EncodeToString(signatureBuilder.Sum(nil))
}
