package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gomodule/redigo/redis"
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
	span := sentry.StartSpan(ctx, "Redis - Set")
	defer span.Finish()

	if key == "" {
		return errors.WithStack(ErrBlankKey)
	}

	return errors.Wrap(r.send(span.Context(), "SET", key, value), "failed to store item in cache")
}

func (r *redisCache) SetTTL(ctx context.Context, key string, value []byte, lifetime time.Duration) error {
	span := sentry.StartSpan(ctx, "Redis - SetTTL")
	defer span.Finish()

	if key == "" {
		return errors.WithStack(ErrBlankKey)
	}

	return errors.Wrap(
		r.send(
			span.Context(),
			"SET", key, value, "EXAT", time.Now().Add(lifetime).Unix(),
		),
		"failed to store item in cache",
	)
}

func (r *redisCache) SetEz(ctx context.Context, key string, object interface{}) error {
	span := sentry.StartSpan(ctx, "Redis - SetEz")
	defer span.Finish()

	data, err := msgpack.Marshal(object)
	if err != nil {
		return errors.Wrap(err, "failed to marshal item to be cached")
	}

	return r.Set(span.Context(), key, data)
}

func (r *redisCache) SetEzTTL(ctx context.Context, key string, object interface{}, lifetime time.Duration) error {
	span := sentry.StartSpan(ctx, "Redis - SetEzTTL")
	defer span.Finish()

	data, err := msgpack.Marshal(object)
	if err != nil {
		return errors.Wrap(err, "failed to marshal item to be cached")
	}

	return r.SetTTL(span.Context(), key, data, lifetime)
}

func (r *redisCache) Get(ctx context.Context, key string) ([]byte, error) {
	span := sentry.StartSpan(ctx, "Redis - Get")
	defer span.Finish()

	if key == "" {
		return nil, errors.WithStack(ErrBlankKey)
	}

	span.SetTag("key", key)

	result, err := r.do(span.Context(), "GET", key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve item from cache")
	}

	switch raw := result.(type) {
	case nil:
		return nil, nil
	case string:
		return []byte(raw), nil
	case []byte:
		return raw, nil
	default:
		panic(fmt.Sprintf("unsupported cache value type: %T", raw))
	}
}

func (r *redisCache) GetEz(ctx context.Context, key string, output interface{}) error {
	span := sentry.StartSpan(ctx, "Redis - GetEz")
	defer span.Finish()

	data, err := r.Get(span.Context(), key)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	return errors.Wrap(msgpack.Unmarshal(data, output), "failed to unmarshal from cache")
}

func (r *redisCache) Delete(ctx context.Context, key string) error {
	span := sentry.StartSpan(ctx, "Redis - Delete")
	defer span.Finish()

	if key == "" {
		return errors.WithStack(ErrBlankKey)
	}

	span.SetTag("key", key)

	return errors.Wrap(r.send(span.Context(), "DEL", key), "failed to delete item from cache")
}
