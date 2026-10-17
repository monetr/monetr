package cache

import (
	"context"
	"crypto/subtle"
	"log/slog"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"
)

type lruCacheEntry struct {
	Data       []byte    `json:"data"`
	Expiration time.Time `json:"expiration"`
}

type lruCache struct {
	log   *slog.Logger
	cas   sync.RWMutex
	clock clock.Clock
	lru   *expirable.LRU[string, *lruCacheEntry]
}

func NewLRUCache(
	log *slog.Logger,
	clock clock.Clock,
	conf config.MemoryCache,
) Cache {
	return &lruCache{
		log:   log,
		clock: clock,
		cas:   sync.RWMutex{},
		// Basically gives us a max TTL of 1 hour regardless.
		lru: expirable.NewLRU[string, *lruCacheEntry](conf.MaxKeys, nil, time.Hour),
	}
}

// CompareAndSwap implements [Cache].
func (l *lruCache) CompareAndSwap(
	ctx context.Context,
	key string,
	expected []byte,
	next []byte,
	ttl time.Duration,
) (swapped bool, err error) {
	if key == "" {
		return false, errors.WithStack(ErrBlankKey)
	}

	span := sentry.StartSpan(ctx, "cache.compare_and_swap")
	defer span.Finish()
	span.Description = key
	span.Status = sentry.SpanStatusOK
	span.SetData("db.system", "lru")
	span.SetData("cache.key", []string{key})
	span.SetData("cache.success", true) // Can't fail

	start := time.Now()
	// This might seem backwards but i think the logic is sound, basically the
	// normal set operation grabs a read lock because the lru is already thread
	// safe. So having multiple routines holding read locks even for the same key
	// is fine there. Where we need to make sure a single operation only happens
	// one at a time is here where the read and then the write matter. So this one
	// grabs the exclusive lock, preventing the only other write operation from
	// proceeding, but also waiting for the only other write operation to complete
	// before it begins.
	l.cas.Lock()
	defer l.cas.Unlock()
	value, ok := l.lru.Get(key)
	// If the value does exist and is not expired and equals the expected value
	if ok &&
		// If no TTL was set then thats fine too
		(value.Expiration.IsZero() || value.Expiration.After(l.clock.Now())) &&
		subtle.ConstantTimeCompare(value.Data, expected) == 1 {
		// If the current value matches then overwrite it with the new value
		// TODO HAndle the evicted return value?
		expiration := value.Expiration
		// If no TTL is specified then use the TTL of the original entry
		if ttl > 0 {
			expiration = l.clock.Now().Add(ttl)
		}
		l.lru.Add(key, &lruCacheEntry{
			Data:       next,
			Expiration: expiration,
		})
		span.SetData("cache.swapped", true)
		recordOperation("compare_and_swap", "swapped", start)
		return true, nil
	}
	span.SetData("cache.swapped", false)
	recordOperation("compare_and_swap", "not_swapped", start)
	return false, nil
}

// Delete implements [Cache].
func (l *lruCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return errors.WithStack(ErrBlankKey)
	}

	span := sentry.StartSpan(ctx, "cache.remove")
	defer span.Finish()
	span.Description = key
	span.Status = sentry.SpanStatusOK
	span.SetData("db.system", "lru")
	span.SetData("cache.key", []string{key})
	span.SetData("cache.success", true) // Can't fail

	start := time.Now()
	l.cas.RLock()
	defer l.cas.RUnlock()
	l.lru.Remove(key)
	recordOperation("delete", "ok", start)
	return nil
}

// Get implements [Cache].
func (l *lruCache) Get(ctx context.Context, key string) ([]byte, error) {
	span := sentry.StartSpan(ctx, "cache.get")
	defer span.Finish()

	if key == "" {
		span.Status = sentry.SpanStatusInvalidArgument
		return nil, errors.WithStack(ErrBlankKey)
	}

	span.Status = sentry.SpanStatusOK
	span.Description = key
	span.SetData("cache.key", []string{key})
	span.SetData("db.system", "lru")
	span.SetData("cache.success", true) // Can't fail

	start := time.Now()
	result, ok := l.lru.Get(key)

	// If the expiration of the indivudual key is in the future or has no TTL then
	// we can return it safely. This is so we still support individual TTLs
	if ok && (result.Expiration.IsZero() || result.Expiration.After(l.clock.Now())) {
		span.SetData("cache.hit", true)
		span.SetData("cache.item_size", len(result.Data))
		recordOperation("get", "hit", start)
		return result.Data, nil
	}

	span.SetData("cache.hit", false)
	span.Status = sentry.SpanStatusNotFound
	recordOperation("get", "miss", start)

	return nil, nil
}

// GetEz implements [Cache].
func (l *lruCache) GetEz(ctx context.Context, key string, output any) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Status = sentry.SpanStatusOK

	data, err := l.Get(span.Context(), key)
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

// Set implements [Cache].
func (l *lruCache) Set(ctx context.Context, key string, value []byte) error {
	if key == "" {
		return errors.WithStack(ErrBlankKey)
	}

	span := sentry.StartSpan(ctx, "cache.put")
	defer span.Finish()
	span.Description = key
	span.SetData("db.system", "lru")
	span.SetData("cache.key", []string{key})
	span.SetData("cache.item_size", len(value))
	span.SetData("cache.success", true) // Can't fail

	// Because LRU is thread safe, we don't need an exclusive lock to write to it
	// like this. We only need the exclusive lock when we are doing compare and
	// swap operations
	start := time.Now()
	l.cas.RLock()
	defer l.cas.RUnlock()
	l.lru.Add(key, &lruCacheEntry{
		Data:       value,
		Expiration: time.Time{}, // Leave empty for no TTL
	})
	recordOperation("set", "ok", start)
	return nil
}

// SetEz implements [Cache].
func (l *lruCache) SetEz(ctx context.Context, key string, object any) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	data, err := msgpack.Marshal(object)
	if err != nil {
		return errors.Wrap(err, "failed to marshal item to be cached")
	}

	return l.Set(span.Context(), key, data)
}

// SetEzTTL implements [Cache].
func (l *lruCache) SetEzTTL(
	ctx context.Context,
	key string,
	object any,
	lifetime time.Duration,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	data, err := msgpack.Marshal(object)
	if err != nil {
		return errors.Wrap(err, "failed to marshal item to be cached")
	}

	return l.SetTTL(span.Context(), key, data, lifetime)
}

// SetTTL implements [Cache].
func (l *lruCache) SetTTL(
	ctx context.Context,
	key string,
	value []byte,
	lifetime time.Duration,
) error {
	if key == "" {
		return errors.WithStack(ErrBlankKey)
	}

	span := sentry.StartSpan(ctx, "cache.put")
	defer span.Finish()
	span.Description = key
	span.SetData("db.system", "lru")
	span.SetData("cache.key", []string{key})
	span.SetData("cache.item_size", len(value))
	span.SetData("cache.ttl", int64(lifetime.Seconds()))
	span.SetData("cache.success", true) // Can't fail

	start := time.Now()
	l.cas.RLock()
	defer l.cas.RUnlock()
	l.lru.Add(key, &lruCacheEntry{
		Data:       value,
		Expiration: l.clock.Now().Add(lifetime),
	})
	recordOperation("set", "ok", start)
	return nil
}

func (l *lruCache) Close() error {
	l.lru.Purge()
	return nil
}
