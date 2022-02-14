package testutils

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gomodule/redigo/redis"
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
