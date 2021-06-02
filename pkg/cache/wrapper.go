package cache

import (
	"context"
	"encoding/json"
	"github.com/getsentry/sentry-go"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, value []byte) error
	SetTTL(ctx context.Context, key string, value []byte, lifetime time.Duration) error
	SetEz(ctx context.Context, key string, object interface{}) error
	SetEzTTL(ctx context.Context, key string, object interface{}, lifetime time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	GetEz(ctx context.Context, key string, output interface{}) error
	Delete(ctx context.Context, key string) error
	Close() error
}

var (
	_ Cache = &redisCache{}
)

type redisCache struct {
	client redis.Conn
}

func (r *redisCache) Set(ctx context.Context, key string, value []byte) error {
	span := sentry.StartSpan(ctx, "Redis - Set")
	defer span.Finish()
	return errors.Wrap(r.client.Send("SET", key, value), "failed to store item in cache")
}

func (r *redisCache) SetTTL(ctx context.Context, key string, value []byte, lifetime time.Duration) error {
	span := sentry.StartSpan(ctx, "Redis - SetTTL")
	defer span.Finish()
	return errors.Wrap(r.client.Send("SET", key, value, "EXAT", time.Now().Add(lifetime)), "failed to store item in cache")
}

func (r *redisCache) SetEz(ctx context.Context, key string, object interface{}) error {
	span := sentry.StartSpan(ctx, "Redis - SetEz")
	defer span.Finish()

	data, err := json.Marshal(object)
	if err != nil {
		return errors.Wrap(err, "failed to marshal item to be cached")
	}

	return r.Set(span.Context(), key, data)
}

func (r *redisCache) SetEzTTL(ctx context.Context, key string, object interface{}, lifetime time.Duration) error {
	panic("implement me")
}

func (r *redisCache) Get(ctx context.Context, key string) ([]byte, error) {
	panic("implement me")
}

func (r *redisCache) GetEz(ctx context.Context, key string, output interface{}) error {
	panic("implement me")
}

func (r *redisCache) Delete(ctx context.Context, key string) error {
	panic("implement me")
}

func (r *redisCache) Close() error {
	panic("implement me")
}
