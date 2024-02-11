package repository_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/stretchr/testify/assert"
)

func TestRepository_CreateTellerLink(t *testing.T) {
	t.Run("without secret", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		repo := repository.NewRepositoryFromSession(clock, user.UserId, user.AccountId, db)

		link := models.TellerLink{
			EnrollmentId:         gofakeit.UUID(),
			UserId:               gofakeit.Generate("user_######"),
			Status:               models.TellerLinkStatusSetup,
			ErrorCode:            nil,
			InstitituionName:     fmt.Sprintf("Bank Of %s", gofakeit.City()),
			LastManualSync:       nil,
			LastSuccessfulUpdate: nil,
			LastAttemptedUpdate:  nil,
		}

		err := repo.CreateTellerLink(context.Background(), &link)
		assert.EqualError(t, err, `failed to create Teller link: ERROR #23502 null value in column "secret_id" of relation "teller_links" violates not-null constraint`)
	})

	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		kms := secrets.NewPlaintextKMS()
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		repo := repository.NewRepositoryFromSession(clock, user.UserId, user.AccountId, db)
		secretsRepo := repository.NewSecretsRepository(log, clock, db, kms, user.AccountId)

		secret := repository.Secret{
			Kind:   models.TellerSecretKind,
			Secret: gofakeit.UUID(),
		}
		assert.NoError(t, secretsRepo.Store(context.Background(), &secret), "should be able to store a secret")

		link := models.TellerLink{
			SecretId:             secret.SecretId,
			EnrollmentId:         gofakeit.UUID(),
			UserId:               gofakeit.Generate("user_######"),
			Status:               models.TellerLinkStatusSetup,
			ErrorCode:            nil,
			InstitituionName:     fmt.Sprintf("Bank Of %s", gofakeit.City()),
			LastManualSync:       nil,
			LastSuccessfulUpdate: nil,
			LastAttemptedUpdate:  nil,
		}

		err := repo.CreateTellerLink(context.Background(), &link)
		assert.NoError(t, err, "must be able to create a teller link")
		assert.NotZero(t, link.TellerLinkId, "must now have the ID set")
		assert.Equal(t, user.AccountId, link.AccountId, "should have set the account Id")
	})
}
