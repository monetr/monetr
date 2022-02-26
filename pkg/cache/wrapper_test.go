package cache_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/monetr/monetr/pkg/cache"
	"github.com/monetr/monetr/pkg/internal/testutils"
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
