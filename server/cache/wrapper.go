package cache

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/getsentry/sentry-go"
	"github.com/gomodule/redigo/redis"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"
)

type Cache interface {
	Set(ctx context.Context, key string, value []byte) error
	SetTTL(ctx context.Context, key string, value []byte, lifetime time.Duration) error
	SetEz(ctx context.Context, key string, object any) error
	SetEzTTL(ctx context.Context, key string, object any, lifetime time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	GetEz(ctx context.Context, key string, output any) error
	Delete(ctx context.Context, key string) error
	// CompareAndSwap atomically sets key to next only if its current value equals
	// expected; a missing key or mismatch writes nothing and returns false. On a
	// swap the expiry is refreshed to ttl, or kept (KEEPTTL) when ttl <= 0.
	CompareAndSwap(ctx context.Context, key string, expected, next []byte, ttl time.Duration) (swapped bool, err error)
}

var (
	_ Cache = &redisCache{}
)

var (
	ErrBlankKey = errors.New("key is blank")
)

type redisCache struct {
	log    *slog.Logger
	client *redis.Pool
}

func NewCache(log *slog.Logger, client *redis.Pool) Cache {
	return &redisCache{
		log:    log,
		client: client,
	}
}

func (r *redisCache) send(ctx context.Context, commandName string, args ...any) error {
	conn, err := r.client.GetContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve connection from pool")
	}
	defer func() {
		if err := conn.Close(); err != nil {
			r.log.WarnContext(ctx, "failed to close/release redis connection", "err", err)
		}
	}()

	return conn.Send(commandName, args...)
}

func (r *redisCache) do(ctx context.Context, commandName string, args ...any) (any, error) {
	conn, err := r.client.GetContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve connection from pool")
	}
	defer func() {
		if err := conn.Close(); err != nil {
			r.log.WarnContext(ctx, "failed to close/release redis connection", "err", err)
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
	span.SetData("db.system", "redis")
	span.SetData("cache.key", []string{key})
	span.SetData("cache.item_size", len(value))

	if err := errors.Wrap(
		r.send(span.Context(), "SET", key, value),
		"failed to store item in cache",
	); err != nil {
		span.SetData("cache.success", false)
		return err
	}

	span.SetData("cache.success", true)
	return nil
}

func (r *redisCache) SetTTL(ctx context.Context, key string, value []byte, lifetime time.Duration) error {
	if key == "" {
		return errors.WithStack(ErrBlankKey)
	}
	span := sentry.StartSpan(ctx, "cache.put")
	defer span.Finish()
	span.Description = key
	span.SetData("db.system", "redis")
	span.SetData("cache.key", []string{key})
	span.SetData("cache.item_size", len(value))
	span.SetData("cache.ttl", int64(lifetime.Seconds()))

	if err := errors.Wrap(
		r.send(
			span.Context(),
			"SET", key, value, "EXAT", time.Now().Add(lifetime).Unix(),
		),
		"failed to store item in cache",
	); err != nil {
		span.SetData("cache.success", false)
		return err
	}

	span.SetData("cache.success", true)
	return nil
}

func (r *redisCache) SetEz(ctx context.Context, key string, object any) error {
	span := sentry.StartSpan(ctx, "function")
	defer span.Finish()
	span.Description = ""

	data, err := msgpack.Marshal(object)
	if err != nil {
		return errors.Wrap(err, "failed to marshal item to be cached")
	}

	return r.Set(span.Context(), key, data)
}

func (r *redisCache) SetEzTTL(ctx context.Context, key string, object any, lifetime time.Duration) error {
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
	span.SetData("cache.key", []string{key})
	span.SetData("db.system", "redis")

	result, err := r.do(span.Context(), "GET", key)
	if err != nil {
		span.SetData("cache.success", false)
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve item from cache")
	}
	span.SetData("cache.success", true)

	switch raw := result.(type) {
	case nil:
		span.SetData("cache.hit", false)
		span.Status = sentry.SpanStatusNotFound
		return nil, nil
	case string:
		span.SetData("cache.hit", true)
		span.SetData("cache.item_size", len(raw))
		return []byte(raw), nil
	case []byte:
		span.SetData("cache.hit", true)
		span.SetData("cache.item_size", len(raw))
		return raw, nil
	default:
		span.Status = sentry.SpanStatusUnimplemented
		panic(fmt.Sprintf("unsupported cache value type: %T", raw))
	}
}

func (r *redisCache) GetEz(ctx context.Context, key string, output any) error {
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

// Lua for CompareAndSwap, run atomically by redis (no read-then-write race).
// KEYS[1] key, ARGV[1] expected, ARGV[2] next, ARGV[3] ttl ms. Returns 1 on
// swap, 0 otherwise (no write on a mismatch or missing key).
const compareAndSwapScript = `
local current = redis.call('GET', KEYS[1])
if current == false then
	return 0
end
if current == ARGV[1] then
	local ttl = tonumber(ARGV[3])
	if ttl and ttl > 0 then
		redis.call('SET', KEYS[1], ARGV[2], 'PX', ttl)
	else
		redis.call('SET', KEYS[1], ARGV[2], 'KEEPTTL')
	end
	return 1
end
return 0
`

func (r *redisCache) CompareAndSwap(ctx context.Context, key string, expected, next []byte, ttl time.Duration) (bool, error) {
	if key == "" {
		return false, errors.WithStack(ErrBlankKey)
	}

	span := sentry.StartSpan(ctx, "cache.compare_and_swap")
	defer span.Finish()
	span.Description = key
	span.Status = sentry.SpanStatusOK
	span.SetData("db.system", "redis")
	span.SetData("cache.key", []string{key})

	result, err := r.do(
		span.Context(),
		"EVAL", compareAndSwapScript, 1, key, expected, next, ttl.Milliseconds(),
	)
	if err != nil {
		span.SetData("cache.success", false)
		span.Status = sentry.SpanStatusInternalError
		return false, errors.Wrap(err, "failed to compare and swap item in cache")
	}
	span.SetData("cache.success", true)

	swapped, err := redis.Int64(result, nil)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return false, errors.Wrap(err, "failed to interpret compare and swap result")
	}

	span.SetData("cache.swapped", swapped == 1)
	return swapped == 1, nil
}

func (r *redisCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return errors.WithStack(ErrBlankKey)
	}

	span := sentry.StartSpan(ctx, "cache.remove")
	defer span.Finish()
	span.Description = key
	span.Status = sentry.SpanStatusOK
	span.SetData("db.system", "redis")
	span.SetData("cache.key", []string{key})

	if err := errors.Wrap(r.send(span.Context(), "DEL", key), "failed to delete item from cache"); err != nil {
		span.SetData("cache.success", false)
		return err
	}

	span.SetData("cache.success", true)
	return nil
}
