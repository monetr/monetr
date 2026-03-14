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
	"log/slog"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
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
	Clock() clock.Clock
	Log() *slog.Logger
	DB() pg.DBI
	Publisher() pubsub.Publisher
	Platypus() platypus.Platypus
	KMS() secrets.KeyManagement
	Storage() storage.Storage
	Billing() billing.Billing
	Email() communication.EmailCommunication
	Processor() Processor

	// RunInTransaction wraps the current context in a transaction, notably the
	// [Context.DB] function here will be an actual postgresql transaction now and
	// any errors or panics in this block will result in a full rollback.
	RunInTransaction(ctx context.Context, callback func(ctx Context) error) error
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

// Processor does not expose any public methods. Instead it is meant to be
// interacted with via the [Enqueue], [EnqueueAt], and [Register] functions so
// that way the jobs can take advantage of generics in order to maintain type
// safety and simplicity.
type Processor interface {
	enqueueAt(
		ctx context.Context,
		queue string,
		at time.Time,
		args any,
	) error
	register(
		ctx context.Context,
		queue string,
		job internalJobWrapper,
	) error
	registerCron(
		ctx context.Context,
		queue string,
		schedule string,
		job internalJobWrapper,
	) error

	Start() error
	Close() error
}

func Enqueue[T any](
	ctx context.Context,
	processor Processor,
	job JobFunction[T],
	args T,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	return processor.enqueueAt(
		span.Context(),
		queueNameFromJobFunction[T](job),
		time.Now(),
		args,
	)
}

func EnqueueAt[T any](
	ctx context.Context,
	processor Processor,
	at time.Time,
	job JobFunction[T],
	args T,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	return processor.enqueueAt(
		span.Context(),
		queueNameFromJobFunction[T](job),
		at,
		args,
	)
}

func Register[T any](
	ctx context.Context,
	processor Processor,
	job JobFunction[T],
) error {
	queue := queueNameFromJobFunction[T](job)

	return processor.register(ctx, queue, func(ctx Context, argBytes []byte) error {
		var args T
		if err := decodeArguments(argBytes, &args); err != nil {
			return errors.Wrapf(err, "failed to decode arguments for job: %s", queue)
		}

		return job(ctx, args)
	})
}

func RegisterCron(
	ctx context.Context,
	processor Processor,
	schedule string,
	job CronFunction,
) error {
	queue := queueNameFromJobFunction[any](job)

	return processor.registerCron(
		ctx,
		queue,
		schedule,
		// Cron jobs don't need their argument bytes, but are represented as the
		// same signature internaly for simplicity.
		func(ctx Context, _ []byte) (err error) {
			span := sentry.SpanFromContext(ctx)
			hub := sentry.GetHubFromContext(ctx)
			hub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetContext("monitor", sentry.Context{"slug": queue})
			})
			defer hub.RecoverWithContext(ctx, nil)
			crontab := strings.SplitN(schedule, " ", 2)[1]
			monitorSchedule := sentry.CrontabSchedule(crontab)
			monitorConfig := &sentry.MonitorConfig{
				Schedule:      monitorSchedule,
				MaxRuntime:    2,
				CheckInMargin: 1,
			}
			checkInId := hub.CaptureCheckIn(&sentry.CheckIn{
				MonitorSlug: queue,
				Status:      sentry.CheckInStatusInProgress,
			}, monitorConfig)
			defer func() {
				status := sentry.CheckInStatusOK
				span.Status = sentry.SpanStatusOK
				if err != nil {
					status = sentry.CheckInStatusError
					span.Status = sentry.SpanStatusInternalError
					hub.CaptureException(err)
					ctx.Log().WarnContext(span.Context(), "cron job finished with an error", "err", err)
				} else {
					ctx.Log().DebugContext(span.Context(), "cron job finished")
				}
				span.Finish()
				if checkInId != nil {
					hub.CaptureCheckIn(&sentry.CheckIn{
						ID:          *checkInId,
						MonitorSlug: queue,
						Status:      status,
						Duration:    span.EndTime.Sub(span.StartTime),
					}, monitorConfig)
				}
			}()

			err = job(ctx)
			return err
		},
	)
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
	return sentryMonitorSlug(name)
}

// encodeArguments is the common function used for marshaling any arguments into
// the actual data that is stored in the queue. At the moment this just wraps
// the JSON library. But in the future if we ever want to replace this with
// something else such as msgpack; then this is how we can do that. This
// function also makes sure that any errors have a proper stack trace.
func encodeArguments(args any) ([]byte, error) {
	result, err := json.Marshal(args)
	return result, errors.WithStack(err)
}

// decodeArguments is the same as [encodeArguments] but will unmarshal the data.
// This currently is implemented as json but can be changed in the future. Must
// match the data type for [encodeArguments]. This function also makes sure that
// errors have stack traces.
func decodeArguments[T any](data []byte, result *T) error {
	return errors.WithStack(json.Unmarshal(data, result))
}

// sentryMonitorSlug converts a queue name into a valid Sentry monitor slug.
// Queue names contain dots, slashes, colons, and other punctuation, so this
// function normalises them: any character that is not alphanumeric is replaced
// with a hyphen, consecutive hyphens are collapsed, and leading/trailing
// hyphens are trimmed.
//
// Examples:
//
//	"background.CleanupFilesCron"          -> "background-CleanupFilesCron"
//	"background.ProcessSpending::struct{}" -> "background-ProcessSpending-struct"
func sentryMonitorSlug(queue string) string {
	slug := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, queue)
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	return strings.Trim(slug, "-")
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
