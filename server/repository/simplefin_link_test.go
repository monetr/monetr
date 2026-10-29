package repository_test

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mock_simplefin"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryBase_CreateSimpleFinLink(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		kms := secrets.NewPlaintextKMS()
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		secretsRepo := repository.NewSecretsRepository(log, clock, db, kms, user.AccountId)

		secret := repository.SecretData{
			Kind:  models.SecretKindSimpleFin,
			Value: mock_simplefin.AccessUrl(t),
		}
		err := secretsRepo.Store(t.Context(), &secret)
		assert.NoError(t, err, "should be able to store the secret successfully")
		assert.NotZero(t, secret.SecretId, "secret Id should now be set")

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		link := models.SimpleFinLink{
			Name:     "Testing",
			SecretId: secret.SecretId,
			ApiUrl:   "https://bridge.simplefin.org/simplefin",
			Status:   models.SimpleFinLinkStatusPending,
		}
		err = repo.CreateSimpleFinLink(t.Context(), &link)
		assert.NoError(t, err, "should be able to create a simplefin link")
		assert.NotZero(t, link.SimpleFinLinkId, "should have generated a simplefin link Id")
		assert.Equal(t, user.AccountId, link.AccountId, "account Id should be set by the repository")
		assert.Equal(t, user.UserId, link.CreatedBy, "created by should be set by the repository")
		assert.Equal(t, clock.Now().UTC(), link.CreatedAt, "created at should come from the clock")
	})
}

func TestRepositoryBase_UpdateSimpleFinLink(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		simpleFinLink := link.SimpleFinLink
		simpleFinLink.Name = "Updated name"
		simpleFinLink.Status = models.SimpleFinLinkStatusError
		simpleFinLink.LastErrors = []models.SimpleFinLinkError{
			{
				Code:         "con.auth",
				Message:      "Lehman Brothers needs your attention",
				ConnectionId: new("CON-1234"),
				SimpleFinId:  nil,
			},
		}
		err := repo.UpdateSimpleFinLink(t.Context(), simpleFinLink)
		assert.NoError(t, err, "should be able to update the simplefin link")

		result, err := repo.GetSimpleFinLink(t.Context(), simpleFinLink.SimpleFinLinkId)
		assert.NoError(t, err, "should be able to read the simplefin link back")
		assert.Equal(t, "Updated name", result.Name, "name should be updated")
		assert.Equal(t, models.SimpleFinLinkStatusError, result.Status, "status should be updated")
		assert.Equal(t, simpleFinLink.LastErrors, result.LastErrors, "errors should survive the round trip")
	})

	t.Run("cant update someone elses link", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		otherUser, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		repo := repository.NewRepositoryFromSession(
			clock,
			otherUser.UserId,
			otherUser.AccountId,
			db,
			log,
		)

		simpleFinLink := *link.SimpleFinLink
		simpleFinLink.Name = "Updated name"
		err := repo.UpdateSimpleFinLink(t.Context(), &simpleFinLink)
		assert.EqualError(t, err, "no rows updated", "should not update another account's link")
	})
}

func TestRepositoryBase_RemoveSimpleFinLink(t *testing.T) {
	t.Run("pending link", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		kms := secrets.NewPlaintextKMS()
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		secretsRepo := repository.NewSecretsRepository(log, clock, db, kms, user.AccountId)

		secret := repository.SecretData{
			Kind:  models.SecretKindSimpleFin,
			Value: mock_simplefin.AccessUrl(t),
		}
		err := secretsRepo.Store(t.Context(), &secret)
		assert.NoError(t, err, "must be able to store the secret")

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		// Pending links don't have a monetr link yet, that only gets created once
		// the user picks their accounts.
		link := models.SimpleFinLink{
			Name:     "Testing",
			SecretId: secret.SecretId,
			ApiUrl:   "https://bridge.simplefin.org/simplefin",
			Status:   models.SimpleFinLinkStatusPending,
		}
		err = repo.CreateSimpleFinLink(t.Context(), &link)
		assert.NoError(t, err, "must be able to create a simplefin link")

		err = repo.RemoveSimpleFinLink(t.Context(), link.SimpleFinLinkId)
		assert.NoError(t, err, "should be able to remove the simplefin link")
		testutils.MustDBNotExist(t, link)
	})

	t.Run("monetr link becomes manual", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		err := repo.RemoveSimpleFinLink(t.Context(), *link.SimpleFinLinkId)
		assert.NoError(t, err, "should be able to remove the simplefin link")
		testutils.MustDBNotExist(t, *link.SimpleFinLink)

		result, err := repo.GetLink(t.Context(), link.LinkId)
		assert.NoError(t, err, "should still be able to retrieve the monetr link")
		assert.Equal(t, models.ManualLinkType, result.LinkType, "link should now be manual")
		assert.Nil(t, result.SimpleFinLinkId, "simplefin link Id should be cleared")
		assert.Nil(t, result.SimpleFinLink, "simplefin link should not be loaded")
	})

	t.Run("cant remove someone elses link", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		otherUser, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		repo := repository.NewRepositoryFromSession(
			clock,
			otherUser.UserId,
			otherUser.AccountId,
			db,
			log,
		)

		err := repo.RemoveSimpleFinLink(t.Context(), *link.SimpleFinLinkId)
		assert.NoError(t, err, "removing a link that isnt ours should not error")
		testutils.MustDBExist(t, *link.SimpleFinLink)

		result := testutils.MustDBRead(t, link)
		assert.Equal(t, models.SimpleFinLinkType, result.LinkType, "link should still be a simplefin link")
		assert.Equal(t, link.SimpleFinLinkId, result.SimpleFinLinkId, "simplefin link Id should not change")
	})
}

func TestRepositoryBase_GetSimpleFinLinks(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		first := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		second := fixtures.GivenIHaveASimpleFinLink(t, clock, user)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		result, err := repo.GetSimpleFinLinks(t.Context())
		assert.NoError(t, err, "should be able to list simplefin links")
		assert.Len(t, result, 2, "should return both links")
		ids := []models.ID[models.SimpleFinLink]{
			result[0].SimpleFinLinkId,
			result[1].SimpleFinLinkId,
		}
		assert.ElementsMatch(t, []models.ID[models.SimpleFinLink]{
			*first.SimpleFinLinkId,
			*second.SimpleFinLinkId,
		}, ids, "should return the links we created")
	})

	t.Run("excludes deleted", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		link.SimpleFinLink.Status = models.SimpleFinLinkStatusDeactivated
		link.SimpleFinLink.DeletedAt = new(clock.Now())
		err := repo.UpdateSimpleFinLink(t.Context(), link.SimpleFinLink)
		assert.NoError(t, err, "must be able to soft delete the simplefin link")

		result, err := repo.GetSimpleFinLinks(t.Context())
		assert.NoError(t, err, "should be able to list simplefin links")
		assert.Empty(t, result, "deleted links should not be returned")
	})

	t.Run("no cross account reads", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		otherUser, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		repo := repository.NewRepositoryFromSession(
			clock,
			otherUser.UserId,
			otherUser.AccountId,
			db,
			log,
		)

		result, err := repo.GetSimpleFinLinks(t.Context())
		assert.NoError(t, err, "should be able to list simplefin links")
		assert.Empty(t, result, "should not see another account's links")
	})
}

func TestRepositoryBase_GetSimpleFinLink(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		result, err := repo.GetSimpleFinLink(t.Context(), *link.SimpleFinLinkId)
		assert.NoError(t, err, "should be able to retrieve the simplefin link")
		assert.Equal(t, *link.SimpleFinLinkId, result.SimpleFinLinkId, "should return the right link")
		assert.Equal(t, link.SimpleFinLink.SecretId, result.SecretId, "secret Id should be stored")
		assert.Equal(t, "https://bridge.simplefin.org/simplefin", result.ApiUrl, "api url should not have credentials")
	})

	t.Run("cant read someone elses link", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		otherUser, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		repo := repository.NewRepositoryFromSession(
			clock,
			otherUser.UserId,
			otherUser.AccountId,
			db,
			log,
		)

		result, err := repo.GetSimpleFinLink(t.Context(), *link.SimpleFinLinkId)
		assert.EqualError(t, err, "failed to retrieve SimpleFIN link: sql: no rows in result set", "should not find another account's link")
		assert.Nil(t, result, "should not return a link")
	})

	t.Run("get link includes simplefin link", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		result, err := repo.GetLink(t.Context(), link.LinkId)
		assert.NoError(t, err, "should be able to retrieve the link")
		assert.Equal(t, models.SimpleFinLinkType, result.LinkType, "should be a simplefin link")
		if assert.NotNil(t, result.SimpleFinLink, "simplefin link should be loaded") {
			assert.Equal(t, *link.SimpleFinLinkId, result.SimpleFinLink.SimpleFinLinkId, "should load the right simplefin link")
		}
	})
}
