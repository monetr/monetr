package repository_test

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryBase_CreateLunchFlowLink(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		kms := secrets.NewPlaintextKMS()
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		secretsRepo := repository.NewSecretsRepository(log, clock, db, kms, user.AccountId)

		secret := repository.SecretData{
			Kind:  models.SecretKindLunchFlow,
			Value: "Bogus secret for Lunch Flow",
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

		link := models.LunchFlowLink{
			SecretId:  secret.SecretId,
			ApiUrl:    "https://example.com/",
			Status:    models.LunchFlowLinkStatusActive,
			CreatedBy: user.UserId,
		}
		err = repo.CreateLunchFlowLink(t.Context(), &link)
		assert.NoError(t, err, "Must be able to create a lunch flow link without error")
		assert.NotEmpty(t, link.LunchFlowLinkId, "Must have generated a new lunch flow link ID")
	})
}
