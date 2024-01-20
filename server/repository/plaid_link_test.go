package repository_test

import (
	"context"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
)

func TestPlaidRepositoryBase_GetLink(t *testing.T) {
	clock := clock.NewMock()
	db := testutils.GetPgDatabase(t)

	user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
	link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
	plaidLink := link.PlaidLink

	plaidRepo := repository.NewPlaidRepository(db)

	t.Run("simple", func(t *testing.T) {
		readLink, err := plaidRepo.GetLink(context.Background(), link.AccountId, link.LinkId)
		assert.NoError(t, err, "failed to retrieve link")
		assert.NotNil(t, readLink.PlaidLink, "must include plaid link child")
		assert.EqualValues(t, link.LinkId, readLink.LinkId, "link Id must match")
		assert.EqualValues(t, plaidLink.PlaidLinkID, readLink.PlaidLink.PlaidLinkID, "plaid link Id must match")
	})

	t.Run("not found", func(t *testing.T) {
		readLink, err := plaidRepo.GetLink(context.Background(), link.AccountId, link.LinkId+100)
		assert.EqualError(t, err, "failed to retrieve link: pg: no rows in result set")
		assert.Nil(t, readLink, "link must be nil")
	})
}

func TestPlaidRepositoryBase_GetLinkByItemId(t *testing.T) {
	clock := clock.NewMock()
	db := testutils.GetPgDatabase(t)

	user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
	link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
	plaidLink := link.PlaidLink

	plaidRepo := repository.NewPlaidRepository(db)

	t.Run("simple", func(t *testing.T) {
		readLink, err := plaidRepo.GetLinkByItemId(context.Background(), plaidLink.PlaidId)
		assert.NoError(t, err, "failed to retrieve link")
		assert.NotNil(t, readLink.PlaidLink, "must include plaid link child")
		assert.EqualValues(t, link.LinkId, readLink.LinkId, "link Id must match")
		assert.EqualValues(t, plaidLink.PlaidLinkID, readLink.PlaidLink.PlaidLinkID, "plaid link Id must match")
	})

	t.Run("not found", func(t *testing.T) {
		readLink, err := plaidRepo.GetLinkByItemId(context.Background(), "not a real item id")
		assert.EqualError(t, err, "failed to retrieve link by item Id: pg: no rows in result set")
		assert.Nil(t, readLink, "link must be nil")
	})
}
