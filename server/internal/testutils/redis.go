package testutils

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/cache"
	"github.com/monetr/monetr/server/config"
	"github.com/stretchr/testify/require"
)

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
