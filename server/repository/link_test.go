package repository_test

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryBase_DeleteLink(t *testing.T) {
	t.Run("link does not exist", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		repo := repository.NewRepositoryFromSession(
			clock,
			"user_bogus",
			"acct_bogus",
			db,
			log,
		)
		err := repo.DeleteLink(
			t.Context(),
			models.ID[models.Link]("link_bogus"),
		)
		assert.EqualError(t, err, repository.ErrLinkNotFound.Error())
	})

	t.Run("happy path", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)
		err := repo.DeleteLink(
			t.Context(),
			link.LinkId,
		)
		assert.NoError(t, err, "must be able to delete link")
	})

	t.Run("can't delete a plaid link", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)
		err := repo.DeleteLink(
			t.Context(),
			link.LinkId,
		)
		assert.EqualError(t, err, repository.ErrLinkIsPlaidLink.Error())
	})
}
