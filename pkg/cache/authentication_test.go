package cache_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/1Password/srp"
	"github.com/monetr/monetr/pkg/cache"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestSRPCache_CacheAuthenticationSession(t *testing.T) {
	t.Run("default SRP", func(t *testing.T) {
		cacheBasic := NewTestCache(t)
		log := testutils.GetLog(t)
		srpCache := cache.NewSRPCache(log, cacheBasic)

		session := srp.NewSRPServer(srp.KnownGroups[srp.RFC5054Group8192], big.NewInt(1234), nil)
		sessionId, err := srpCache.CacheAuthenticationSession(context.Background(), session)
		assert.NoError(t, err, "must cache the session")
		assert.NotEmpty(t, sessionId, "must return a valid sessionId")
	})
}
