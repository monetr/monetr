package repository_test

import (
	"context"
	"testing"

	"github.com/monetr/monetr/pkg/internal/fixtures"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/stretchr/testify/assert"
)

func TestJobRepository_GetPlaidLinksByAccount(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		jobRepo := repository.NewJobRepository(db)

		user, _ := fixtures.GivenIHaveABasicAccount(t)
		_ = fixtures.GivenIHaveAPlaidLink(t, user)
		_ = fixtures.GivenIHaveAPlaidLink(t, user)

		plaidLinks, err := jobRepo.GetPlaidLinksByAccount(context.Background())
		assert.NoError(t, err, "should be able to retrieve the two links")
		assert.Len(t, plaidLinks, 1, "should retrieve the one account")
		assert.Len(t, plaidLinks[0].LinkIds, 2, "should have two links for the one account")
	})
}
