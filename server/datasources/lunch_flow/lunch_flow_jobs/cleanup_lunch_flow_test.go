package lunch_flow_jobs_test

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/datasources/lunch_flow/lunch_flow_jobs"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCleanupLunchFlowCron(t *testing.T) {
	t.Run("lunch flow not enabled", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).
			Return(nil).
			Times(0)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Log().Return(log).MinTimes(1)
		context.EXPECT().Enqueuer().Return(enqueuer).Times(0)
		context.EXPECT().Configuration().Return(config.Configuration{
			LunchFlow: config.LunchFlow{
				Enabled: false,
			},
		}).MinTimes(1)

		err := lunch_flow_jobs.CleanupLunchFlowCron(
			mockqueue.NewMockContext(context),
		)
		assert.NoError(t, err, "should return without error when lunch flow is disabled")
	})

	t.Run("enqueues cleanup for stale link", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		kms := secrets.NewPlaintextKMS()

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
			kms,
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
			ApiUrl:    config.DefaultLunchFlowAPIURL,
			Status:    models.LunchFlowLinkStatusPending,
			CreatedBy: user.UserId,
		}
		assert.NoError(
			t,
			repo.CreateLunchFlowLink(t.Context(), &lunchFlowLink),
			"must be able to create lunch flow link",
		)

		// Move the clock forward past 24 hours so the link becomes stale.
		clock.Add(25 * time.Hour)

		enqueuer := mockgen.NewMockProcessor(ctrl)

		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(lunch_flow_jobs.CleanupLunchFlow),
				gomock.Any(),
				gomock.Eq(lunch_flow_jobs.CleanupLunchFlowArguments{
					AccountId:       user.AccountId,
					LunchFlowLinkId: lunchFlowLink.LunchFlowLinkId,
				}),
			).
			Return(nil).
			Times(1)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).MinTimes(1)
		context.EXPECT().Configuration().Return(config.Configuration{
			LunchFlow: config.LunchFlow{
				Enabled: true,
			},
		}).MinTimes(1)
		context.EXPECT().DB().Return(db).MinTimes(1)
		context.EXPECT().Enqueuer().Return(enqueuer).MinTimes(1)
		context.EXPECT().Log().Return(log).MinTimes(1)

		err := lunch_flow_jobs.CleanupLunchFlowCron(
			mockqueue.NewMockContext(context),
		)
		assert.NoError(t, err, "cleanup lunch flow cron should succeed")
	})

	t.Run("does not enqueue for a link that is not yet stale", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		kms := secrets.NewPlaintextKMS()

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
			kms,
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
			ApiUrl:    config.DefaultLunchFlowAPIURL,
			Status:    models.LunchFlowLinkStatusPending,
			CreatedBy: user.UserId,
		}
		assert.NoError(
			t,
			repo.CreateLunchFlowLink(t.Context(), &lunchFlowLink),
			"must be able to create lunch flow link",
		)

		// Do not advance the clock past 24 hours, so the link is not yet stale.
		clock.Add(1 * time.Hour)

		enqueuer := mockgen.NewMockProcessor(ctrl)

		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).
			Return(nil).
			Times(0)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).MinTimes(1)
		context.EXPECT().Configuration().Return(config.Configuration{
			LunchFlow: config.LunchFlow{
				Enabled: true,
			},
		}).MinTimes(1)
		context.EXPECT().DB().Return(db).MinTimes(1)
		context.EXPECT().Enqueuer().Return(enqueuer).Times(0)
		context.EXPECT().Log().Return(log).MinTimes(1)

		err := lunch_flow_jobs.CleanupLunchFlowCron(
			mockqueue.NewMockContext(context),
		)
		assert.NoError(t, err, "cleanup lunch flow cron should succeed")
	})
}

func TestCleanupLunchFlow(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		kms := secrets.NewPlaintextKMS()

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
			kms,
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
			ApiUrl:    config.DefaultLunchFlowAPIURL,
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

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).MinTimes(1)
		context.EXPECT().DB().Return(db).MinTimes(1)
		context.EXPECT().Log().Return(log).MinTimes(1)
		context.EXPECT().KMS().Return(kms).MinTimes(1)
		context.EXPECT().Configuration().Return(config.Configuration{
			LunchFlow: config.LunchFlow{
				Enabled: true,
			},
		}).MinTimes(1)

		err := lunch_flow_jobs.CleanupLunchFlow(
			mockqueue.NewMockContext(context),
			lunch_flow_jobs.CleanupLunchFlowArguments{
				AccountId:       user.AccountId,
				LunchFlowLinkId: lunchFlowLink.LunchFlowLinkId,
			},
		)
		assert.NoError(t, err, "cleanup lunch flow job should succeed")
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

	t.Run("won't delete a non-pending lunch flow link", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		kms := secrets.NewPlaintextKMS()

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
			kms,
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
			ApiUrl:    config.DefaultLunchFlowAPIURL,
			Status:    models.LunchFlowLinkStatusActive,
			CreatedBy: user.UserId,
		}
		assert.NoError(
			t,
			repo.CreateLunchFlowLink(t.Context(), &lunchFlowLink),
			"must be able to create lunch flow link in the active status",
		)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).MinTimes(1)
		context.EXPECT().DB().Return(db).MinTimes(1)
		context.EXPECT().Log().Return(log).MinTimes(1)
		context.EXPECT().KMS().Return(kms).MinTimes(1)
		context.EXPECT().Configuration().Return(config.Configuration{
			LunchFlow: config.LunchFlow{
				Enabled: true,
			},
		}).MinTimes(1)

		err := lunch_flow_jobs.CleanupLunchFlow(
			mockqueue.NewMockContext(context),
			lunch_flow_jobs.CleanupLunchFlowArguments{
				AccountId:       user.AccountId,
				LunchFlowLinkId: lunchFlowLink.LunchFlowLinkId,
			},
		)
		assert.NoError(t, err, "cleanup lunch flow job should succeed")

		// Make sure that we have not deleted anything, even though we did not
		// return an error.
		testutils.MustDBExist(t, models.Secret{
			SecretId:  secret.SecretId,
			AccountId: user.AccountId,
		})
		testutils.MustDBExist(t, lunchFlowLink)
	})
}
