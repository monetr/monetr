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
		l.lru.Add(key, &lruCacheEntry{
			Data:       next,
			Expiration: l.clock.Now().Add(ttl),
		})
		return true, nil
	}
	return false, nil
}

// Delete implements [Cache].
func (l *lruCache) Delete(ctx context.Context, key string) error {
	l.cas.RLock()
	defer l.cas.RUnlock()
	l.lru.Remove(key)
	return nil
}

// Get implements [Cache].
func (l *lruCache) Get(ctx context.Context, key string) ([]byte, error) {
	span := sentry.StartSpan(ctx, "cache.get")
	defer span.Finish()

	span.Status = sentry.SpanStatusOK
	span.Description = key
	span.SetData("cache.key", []string{key})
	span.SetData("db.system", "lru")
	span.SetData("cache.success", true) // Can't fail

	result, ok := l.lru.Get(key)

	// If the expiration of the indivudual key is in the future or has no TTL then
	// we can return it safely. This is so we still support individual TTLs
	if ok && result.Expiration.IsZero() || result.Expiration.After(l.clock.Now()) {
		span.SetData("cache.hit", true)
		span.SetData("cache.item_size", len(result.Data))
		return result.Data, nil
	}

	span.SetData("cache.hit", false)

	return nil, nil
}

// GetEz implements [Cache].
func (l *lruCache) GetEz(ctx context.Context, key string, output any) error {
	data, err := l.Get(ctx, key)
	if err != nil {
		return nil
	}

	if len(data) == 0 {
		return nil
	}

	if err = msgpack.Unmarshal(data, output); err != nil {
		return errors.Wrap(err, "failed to unmarshal from cache")
	}

	return nil
}

// Set implements [Cache].
func (l *lruCache) Set(ctx context.Context, key string, value []byte) error {
	// Because LRU is thread safe, we don't need an exclusive lock to write to it
	// like this. We only need the exclusive lock when we are doing compare and
	// swap operations
	l.cas.RLock()
	defer l.cas.RUnlock()
	l.lru.Add(key, &lruCacheEntry{
		Data:       value,
		Expiration: time.Time{}, // Leave empty for no TTL
	})
	return nil
}

// SetEz implements [Cache].
func (l *lruCache) SetEz(ctx context.Context, key string, object any) error {
	data, err := msgpack.Marshal(object)
	if err != nil {
		return errors.Wrap(err, "failed to marshal item to be cached")
	}

	return l.Set(ctx, key, data)
}

// SetEzTTL implements [Cache].
func (l *lruCache) SetEzTTL(
	ctx context.Context,
	key string,
	object any,
	lifetime time.Duration,
) error {
	data, err := msgpack.Marshal(object)
	if err != nil {
		return errors.Wrap(err, "failed to marshal item to be cached")
	}

	return l.SetTTL(ctx, key, data, lifetime)
}

// SetTTL implements [Cache].
func (l *lruCache) SetTTL(
	ctx context.Context,
	key string,
	value []byte,
	lifetime time.Duration,
) error {
	l.cas.RLock()
	defer l.cas.RUnlock()
	l.lru.Add(key, &lruCacheEntry{
		Data:       value,
		Expiration: l.clock.Now().Add(lifetime),
	})
	return nil
}

func (l *lruCache) Close() error {
	l.lru.Purge()
	return nil
}
