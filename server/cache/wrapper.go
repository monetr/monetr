package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gomodule/redigo/redis"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack/v5"
)

type Cache interface {
	Set(ctx context.Context, key string, value []byte) error
	SetTTL(ctx context.Context, key string, value []byte, lifetime time.Duration) error
	SetEz(ctx context.Context, key string, object interface{}) error
	SetEzTTL(ctx context.Context, key string, object interface{}, lifetime time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	GetEz(ctx context.Context, key string, output interface{}) error
	Delete(ctx context.Context, key string) error
}

var (
	_ Cache = &redisCache{}
)

var (
	ErrBlankKey = errors.New("key is blank")
)

type redisCache struct {
	log    *logrus.Entry
	client *redis.Pool
}

func NewCache(log *logrus.Entry, client *redis.Pool) Cache {
	return &redisCache{
		log:    log,
		client: client,
	}
}

func (r *redisCache) send(ctx context.Context, commandName string, args ...interface{}) error {
	conn, err := r.client.GetContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve connection from pool")
	}
	defer func() {
		if err := conn.Close(); err != nil {
			r.log.WithContext(ctx).WithError(err).Warn("failed to close/release redis connection")
		}
	}()

	return conn.Send(commandName, args...)
}

func (r *redisCache) do(ctx context.Context, commandName string, args ...interface{}) (interface{}, error) {
	conn, err := r.client.GetContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve connection from pool")
	}
	defer func() {
		if err := conn.Close(); err != nil {
			r.log.WithContext(ctx).WithError(err).Warn("failed to close/release redis connection")
		}
	}()

	return conn.Do(commandName, args...)
}

func (r *redisCache) Set(ctx context.Context, key string, value []byte) error {
	if key == "" {
		return errors.WithStack(ErrBlankKey)
	}
	span := sentry.StartSpan(ctx, "cache.put")
	defer span.Finish()
	span.Description = key
	span.SetTag("db.system", "redis")
	span.SetTag("cache.key", key)
	span.SetTag("cache.item_size", fmt.Sprint(len(value)))

	return errors.Wrap(r.send(span.Context(), "SET", key, value), "failed to store item in cache")
}

func (r *redisCache) SetTTL(ctx context.Context, key string, value []byte, lifetime time.Duration) error {
	if key == "" {
		return errors.WithStack(ErrBlankKey)
	}
	span := sentry.StartSpan(ctx, "cache.put")
	defer span.Finish()
	span.Description = key
	span.SetTag("db.system", "redis")
	span.SetTag("cache.key", key)
	span.SetTag("cache.item_size", fmt.Sprint(len(value)))
	span.SetTag("cache.ttl", fmt.Sprint(lifetime.Seconds()))

	return errors.Wrap(
		r.send(
			span.Context(),
			"SET", key, value, "EXAT", time.Now().Add(lifetime).Unix(),
		),
		"failed to store item in cache",
	)
}

func (r *redisCache) SetEz(ctx context.Context, key string, object interface{}) error {
	span := sentry.StartSpan(ctx, "function")
	defer span.Finish()
	span.Description = ""

	data, err := msgpack.Marshal(object)
	if err != nil {
		return errors.Wrap(err, "failed to marshal item to be cached")
	}

	return r.Set(span.Context(), key, data)
}

func (r *redisCache) SetEzTTL(ctx context.Context, key string, object interface{}, lifetime time.Duration) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	data, err := msgpack.Marshal(object)
	if err != nil {
		return errors.Wrap(err, "failed to marshal item to be cached")
	}

	return r.SetTTL(span.Context(), key, data, lifetime)
}

func (r *redisCache) Get(ctx context.Context, key string) ([]byte, error) {
	span := sentry.StartSpan(ctx, "cache.get")
	defer span.Finish()

	if key == "" {
		span.Status = sentry.SpanStatusInvalidArgument
		return nil, errors.WithStack(ErrBlankKey)
	}

	span.Status = sentry.SpanStatusOK
	span.Description = key
	span.SetTag("cache.key", key)
	span.SetTag("db.system", "redis")

	result, err := r.do(span.Context(), "GET", key)
	if err != nil {
		span.SetTag("cache", "failure")
		span.SetTag("cache.success", "false")
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve item from cache")
	}
	span.SetTag("cache.success", "true")

	switch raw := result.(type) {
	case nil:
		span.SetTag("cache", "miss")
		span.SetTag("cache.hit", "false")
		span.Status = sentry.SpanStatusNotFound
		return nil, nil
	case string:
		span.SetTag("cache", "hit")
		span.SetTag("cache.hit", "true")
		return []byte(raw), nil
	case []byte:
		span.SetTag("cache", "hit")
		span.SetTag("cache.hit", "true")
		return raw, nil
	default:
		span.Status = sentry.SpanStatusUnimplemented
		panic(fmt.Sprintf("unsupported cache value type: %T", raw))
	}
}

func (r *redisCache) GetEz(ctx context.Context, key string, output interface{}) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Status = sentry.SpanStatusOK

	data, err := r.Get(span.Context(), key)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return err
	}

	if len(data) == 0 {
		return nil
	}

	if err = msgpack.Unmarshal(data, output); err != nil {
		span.Status = sentry.SpanStatusDataLoss
		return errors.Wrap(err, "failed to unmarshal from cache")
	}

	return nil
}

func (r *redisCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return errors.WithStack(ErrBlankKey)
	}

	span := sentry.StartSpan(ctx, "cache.remove")
	defer span.Finish()
	span.Description = key
	span.Status = sentry.SpanStatusOK
	span.SetTag("db.system", "redis")
	span.SetTag("cache.key", key)

	return errors.Wrap(r.send(span.Context(), "DEL", key), "failed to delete item from cache")
}
