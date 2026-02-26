package background_test

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/datasources/lunch_flow"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/stretchr/testify/assert"
)

func TestCleanupLunchFlowJob_Run(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)
		secretsRepo := repository.NewSecretsRepository(
			log,
			clock,
			db,
			secrets.NewPlaintextKMS(),
			user.AccountId,
		)

		secret := repository.SecretData{
			Kind:  models.SecretKindLunchFlow,
			Value: "test-secret",
		}
		assert.NoError(
			t,
			secretsRepo.Store(t.Context(), &secret),
			"must be able to create lunch flow secret",
		)

		lunchFlowLink := models.LunchFlowLink{
			AccountId: user.AccountId,
			SecretId:  secret.SecretId,
			Name:      "Test Lunch Flow Link",
			ApiUrl:    lunch_flow.DefaultAPIURL,
			Status:    models.LunchFlowLinkStatusPending,
			CreatedBy: user.UserId,
		}
		assert.NoError(
			t,
			repo.CreateLunchFlowLink(t.Context(), &lunchFlowLink),
			"must be able to create lunch flow link",
		)

		firstLunchFlowBankAccount := models.LunchFlowBankAccount{
			AccountId:       user.AccountId,
			LunchFlowLinkId: lunchFlowLink.LunchFlowLinkId,
			LunchFlowId:     "1234",
			LunchFlowStatus: models.LunchFlowBankAccountExternalStatusActive,
			Name:            "First Lunch Flow Account",
			InstitutionName: "Lehman Brothers",
			Provider:        "bogus",
			Currency:        "USD",
			Status:          models.LunchFlowBankAccountStatusInactive,
			CurrentBalance:  100,
			CreatedBy:       user.UserId,
		}
		assert.NoError(
			t,
			repo.CreateLunchFlowBankAccount(t.Context(), &firstLunchFlowBankAccount),
			"must be able to create first lunch flow bank account",
		)

		secondLunchFlowBankAccount := models.LunchFlowBankAccount{
			AccountId:       user.AccountId,
			LunchFlowLinkId: lunchFlowLink.LunchFlowLinkId,
			LunchFlowId:     "1235",
			LunchFlowStatus: models.LunchFlowBankAccountExternalStatusActive,
			Name:            "Second Lunch Flow Account",
			InstitutionName: "Lehman Brothers",
			Provider:        "bogus",
			Currency:        "USD",
			Status:          models.LunchFlowBankAccountStatusInactive,
			CurrentBalance:  250,
			CreatedBy:       user.UserId,
		}
		assert.NoError(
			t,
			repo.CreateLunchFlowBankAccount(t.Context(), &secondLunchFlowBankAccount),
			"must be able to create second lunch flow bank account",
		)

		job, err := background.NewCleanupLunchFlowJob(
			log,
			repo,
			secretsRepo,
			clock,
			background.CleanupLunchFlowArguments{
				AccountId:       user.AccountId,
				LunchFlowLinkId: lunchFlowLink.LunchFlowLinkId,
			},
		)
		assert.NoError(t, err, "must create cleanup lunch flow job")

		assert.NoError(t, job.Run(t.Context()), "cleanup lunch flow job should succeed")
		// Make sure that wee have fully deleted all of our objects that are related
		// to the lunch flow link.
		testutils.MustDBNotExist(t, firstLunchFlowBankAccount)
		testutils.MustDBNotExist(t, secondLunchFlowBankAccount)
		testutils.MustDBNotExist(t, models.Secret{
			SecretId:  secret.SecretId,
			AccountId: user.AccountId,
		})
		testutils.MustDBNotExist(t, lunchFlowLink)
	})
}
