package testutils

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/benbjohnson/clock"
	"github.com/gomodule/redigo/redis"
	"github.com/monetr/monetr/server/cache"
	"github.com/monetr/monetr/server/config"
	"github.com/stretchr/testify/require"
)

func GetRedisPool(t *testing.T) *redis.Pool {
	mini, err := miniredis.Run()
	require.NoError(t, err, "miniredis must be able to run")

	// Store our "embedded" redis address for use below.
	redisAddress := mini.Server().Addr().String()

	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", redisAddress)
		},
	}

	t.Cleanup(func() {
		require.NoError(t, pool.Close(), "pool should close successfully")
		mini.Close()
	})

	return pool
}

func GetCache(t *testing.T) cache.Cache {
	log := GetLog(t)
	miniRedis := miniredis.NewMiniRedis()
	require.NoError(t, miniRedis.Start())

	dataCache, err := cache.GetCache(t.Context(), log, clock.New(), config.Configuration{
		Cache: config.Cache{
			Kind: "valkey",
			Valkey: config.Valkey{
				Address:  miniRedis.Server().Addr().IP.String(),
				Port:     miniRedis.Server().Addr().Port,
				Database: 0,
			},
		},
	})
	require.NoError(t, err, "must be able to setup miniredis/valkey for test")

	t.Cleanup(func() {
		require.NoError(t, dataCache.Close())
		miniRedis.Close()
	})

	return dataCache
}
