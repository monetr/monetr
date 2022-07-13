package repository_test

import (
	"context"
	"testing"

	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryBase_GetSettings(t *testing.T) {
	t.Run("will safe defaults if settings dont exist", func(t *testing.T) {
		repo := GetTestAuthenticatedRepository(t)
		// Make sure the test account does not already have settings.
		testutils.MustDBNotExist(t, models.Settings{AccountId: repo.AccountId()})

		settings, err := repo.GetSettings(context.Background())
		assert.NoError(t, err, "must retrieve settings successfully")
		assert.NotNil(t, settings, "settings should not be nil")
		assert.False(t, settings.MaxSafeToSpend.Enabled, "should not have max safe to spend enabled")
		assert.Zero(t, settings.MaxSafeToSpend.Maximum, "should not have a max safe to spend")
	})

	t.Run("will read existing settings", func(t *testing.T) {
		repo := GetTestAuthenticatedRepository(t)
		// Make sure the test account does not already have settings.
		testutils.MustDBNotExist(t, models.Settings{AccountId: repo.AccountId()})
		testutils.MustDBInsert(t, &models.Settings{
			AccountId: repo.AccountId(),
			MaxSafeToSpend: struct {
				Enabled bool  `json:"enabled"`
				Maximum int64 `json:"maximum"`
			}{
				Enabled: true,
				Maximum: 10000,
			},
		})

		settings, err := repo.GetSettings(context.Background())
		assert.NoError(t, err, "must retrieve settings successfully")
		assert.NotNil(t, settings, "settings should not be nil")
		assert.True(t, settings.MaxSafeToSpend.Enabled, "should have max safe to spend enabled")
		assert.EqualValues(t, 10000, settings.MaxSafeToSpend.Maximum, "should have max safe to spend configured")
	})
}
