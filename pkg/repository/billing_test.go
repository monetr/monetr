package repository

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRepositoryBase_GetActiveSubscription(t *testing.T) {
	t.Run("no subscription", func(t *testing.T) {
		repo := GetTestAuthenticatedRepository(t)

		// Make sure that if we try to retrieve the active subscription, if there is not one then it just returns nil.
		subscription, err := repo.GetActiveSubscription(context.Background())
		assert.NoError(t, err, "should not return an error")
		assert.Nil(t, subscription, "subscription should be nil")
	})
}
