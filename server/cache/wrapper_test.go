package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/monetr/monetr/server/cache"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	TestValue = []byte("Hello World!")
)

func NewTestRedisPool(t *testing.T) *redis.Pool {
	miniRedis := miniredis.NewMiniRedis()
	require.NoError(t, miniRedis.Start())
	redisPool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", miniRedis.Server().Addr().String())
		},
	}

	t.Cleanup(func() {
		require.NoError(t, redisPool.Close(), "must close miniredis pool successfully")
		miniRedis.Close()
	})

	return redisPool
}

func NewTestCache(t *testing.T) cache.Cache {
	return cache.NewCache(testutils.GetLog(t), NewTestRedisPool(t))
}

// NewTestCacheWithRedis is like NewTestCache but it also hands back the
// miniredis instance so a test can poke at things like TTLs directly. This is
// useful for proving that our Lua script behaves the way we expect against
// miniredis specifically.
func NewTestCacheWithRedis(t *testing.T) (cache.Cache, *miniredis.Miniredis) {
	miniRedis := miniredis.NewMiniRedis()
	require.NoError(t, miniRedis.Start())
	redisPool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", miniRedis.Server().Addr().String())
		},
	}

	t.Cleanup(func() {
		require.NoError(t, redisPool.Close(), "must close miniredis pool successfully")
		miniRedis.Close()
	})

	return cache.NewCache(testutils.GetLog(t), redisPool), miniRedis
}

func TestRedisCache_Set(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		memoryCache := NewTestCache(t)

		err := memoryCache.Set(context.Background(), "test:data", TestValue)
		assert.NoError(t, err, "should successfully set value")
	})

	t.Run("nil value", func(t *testing.T) {
		memoryCache := NewTestCache(t)

		err := memoryCache.Set(context.Background(), "test:data", nil)
		assert.NoError(t, err, "should successfully set value")
	})

	t.Run("no key", func(t *testing.T) {
		memoryCache := NewTestCache(t)

		err := memoryCache.Set(context.Background(), "", TestValue)
		assert.Equal(t, cache.ErrBlankKey, errors.Cause(err), "should be blank key error")
	})
}

func TestRedisCache_CompareAndSwap(t *testing.T) {
	t.Run("swaps when the current value matches", func(t *testing.T) {
		memoryCache := NewTestCache(t)
		ctx := t.Context()

		// Seed the key with our "expected" value. CompareAndSwap relies on the key
		// already existing, that way an absent key cannot be swapped.
		require.NoError(t, memoryCache.SetTTL(ctx, "test:cas", []byte("unused"), time.Minute))

		swapped, err := memoryCache.CompareAndSwap(ctx, "test:cas", []byte("unused"), []byte("consumed"), time.Minute)
		assert.NoError(t, err, "should be able to compare and swap")
		assert.True(t, swapped, "should swap because the current value matched expected")

		// The value should now be the new value.
		value, err := memoryCache.Get(ctx, "test:cas")
		assert.NoError(t, err, "should be able to read back the swapped value")
		assert.Equal(t, []byte("consumed"), value, "value must have been swapped")
	})

	t.Run("does not swap when the current value does not match", func(t *testing.T) {
		memoryCache := NewTestCache(t)
		ctx := t.Context()

		require.NoError(t, memoryCache.SetTTL(ctx, "test:cas", []byte("consumed"), time.Minute))

		// We expect "unused" but the value is already "consumed", so nothing should
		// happen. This is what stops a proof of work challenge from being used a
		// second time.
		swapped, err := memoryCache.CompareAndSwap(ctx, "test:cas", []byte("unused"), []byte("consumed"), time.Minute)
		assert.NoError(t, err, "comparison itself should not error")
		assert.False(t, swapped, "must not swap when the current value does not match expected")
	})

	t.Run("does not swap when the key is absent", func(t *testing.T) {
		memoryCache := NewTestCache(t)
		ctx := t.Context()

		// The key was never set, so there is nothing to swap. We want this to fail
		// closed (no write) so that an expired or never-issued challenge cannot be
		// consumed.
		swapped, err := memoryCache.CompareAndSwap(ctx, "test:cas", []byte("unused"), []byte("consumed"), time.Minute)
		assert.NoError(t, err, "comparison itself should not error")
		assert.False(t, swapped, "must not swap when the key does not exist")

		value, err := memoryCache.Get(ctx, "test:cas")
		assert.NoError(t, err, "should be able to read back the key")
		assert.Empty(t, value, "no value should have been written")
	})

	t.Run("blank key", func(t *testing.T) {
		memoryCache := NewTestCache(t)

		_, err := memoryCache.CompareAndSwap(t.Context(), "", []byte("unused"), []byte("consumed"), time.Minute)
		assert.Equal(t, cache.ErrBlankKey, errors.Cause(err), "should be blank key error")
	})

	// These two subtests prove that the Lua script's TTL handling actually works
	// against miniredis. miniredis runs our EVAL script the same way real redis
	// does, including the SET options (KEEPTTL and PX) we use inside the script.
	t.Run("keeps the existing TTL when ttl is zero", func(t *testing.T) {
		memoryCache, miniRedis := NewTestCacheWithRedis(t)
		ctx := t.Context()

		require.NoError(t, memoryCache.SetTTL(ctx, "test:cas", []byte("unused"), time.Minute))

		// A ttl of zero means KEEPTTL, the script should leave whatever expiry the
		// key already had in place.
		swapped, err := memoryCache.CompareAndSwap(ctx, "test:cas", []byte("unused"), []byte("consumed"), 0)
		assert.NoError(t, err, "should be able to compare and swap")
		assert.True(t, swapped, "should swap because the current value matched expected")

		ttl := miniRedis.TTL("test:cas")
		assert.Greater(t, ttl, time.Duration(0), "the key must still have an expiry after a KEEPTTL swap")
		assert.LessOrEqual(t, ttl, time.Minute, "the expiry should not have been extended")
	})

	t.Run("refreshes the TTL when one is provided", func(t *testing.T) {
		memoryCache, miniRedis := NewTestCacheWithRedis(t)
		ctx := t.Context()

		// Start with a short expiry.
		require.NoError(t, memoryCache.SetTTL(ctx, "test:cas", []byte("unused"), 10*time.Second))

		// Swapping with a longer ttl should refresh the expiry to the new value.
		swapped, err := memoryCache.CompareAndSwap(ctx, "test:cas", []byte("unused"), []byte("consumed"), time.Hour)
		assert.NoError(t, err, "should be able to compare and swap")
		assert.True(t, swapped, "should swap because the current value matched expected")

		ttl := miniRedis.TTL("test:cas")
		assert.Greater(t, ttl, 10*time.Second, "the expiry should have been refreshed to the longer ttl")
	})
}
