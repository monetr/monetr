package simplefin_jobs_test

import (
	"errors"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/datasources/simplefin/simplefin_jobs"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mock_simplefin"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCleanupSimpleFinCron(t *testing.T) {
	t.Run("simplefin not enabled", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		enqueuer.EXPECT().
			BulkEnqueueAt(
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
			SimpleFin: config.SimpleFin{
				Enabled: false,
			},
		}).MinTimes(1)

		err := simplefin_jobs.CleanupSimpleFinCron(
			mockqueue.NewMockContext(context),
		)
		assert.NoError(t, err, "should return without error when simplefin is disabled")
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
			Kind:  models.SecretKindSimpleFin,
			Value: mock_simplefin.AccessUrl(t),
		}
		assert.NoError(
			t,
			secretsRepo.Store(t.Context(), &secret),
			"must be able to create simplefin secret",
		)

		simpleFinLink := models.SimpleFinLink{
			AccountId: user.AccountId,
			SecretId:  secret.SecretId,
			Name:      "Test SimpleFIN Link",
			ApiUrl:    "https://bridge.simplefin.org/simplefin",
			Status:    models.SimpleFinLinkStatusPending,
			CreatedBy: user.UserId,
		}
		assert.NoError(
			t,
			repo.CreateSimpleFinLink(t.Context(), &simpleFinLink),
			"must be able to create simplefin link",
		)

		// Move the clock forward past 24 hours so the link becomes stale.
		clock.Add(25 * time.Hour)

		enqueuer := mockgen.NewMockProcessor(ctrl)

		enqueuer.EXPECT().
			BulkEnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(simplefin_jobs.CleanupSimpleFin),
				gomock.Any(),
				gomock.InAnyOrder([]any{
					simplefin_jobs.CleanupSimpleFinArguments{
						AccountId:       user.AccountId,
						SimpleFinLinkId: simpleFinLink.SimpleFinLinkId,
					},
				}),
			).
			Return(nil).
			Times(1)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).MinTimes(1)
		context.EXPECT().Configuration().Return(config.Configuration{
			SimpleFin: config.SimpleFin{
				Enabled: true,
				AllowedHosts: []string{
					config.DefaultSimpleFinHost,
				},
			},
		}).MinTimes(1)
		context.EXPECT().DB().Return(db).MinTimes(1)
		context.EXPECT().Enqueuer().Return(enqueuer).MinTimes(1)
		context.EXPECT().Log().Return(log).MinTimes(1)

		err := simplefin_jobs.CleanupSimpleFinCron(
			mockqueue.NewMockContext(context),
		)
		assert.NoError(t, err, "cleanup simplefin cron should succeed")
	})

	t.Run("link is not stale yet", func(t *testing.T) {
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
			Kind:  models.SecretKindSimpleFin,
			Value: mock_simplefin.AccessUrl(t),
		}
		assert.NoError(
			t,
			secretsRepo.Store(t.Context(), &secret),
			"must be able to create simplefin secret",
		)

		simpleFinLink := models.SimpleFinLink{
			AccountId: user.AccountId,
			SecretId:  secret.SecretId,
			Name:      "Test SimpleFIN Link",
			ApiUrl:    "https://bridge.simplefin.org/simplefin",
			Status:    models.SimpleFinLinkStatusPending,
			CreatedBy: user.UserId,
		}
		assert.NoError(
			t,
			repo.CreateSimpleFinLink(t.Context(), &simpleFinLink),
			"must be able to create simplefin link",
		)

		// Do not advance the clock past 24 hours, so the link is not yet stale.
		clock.Add(1 * time.Hour)

		enqueuer := mockgen.NewMockProcessor(ctrl)

		enqueuer.EXPECT().
			BulkEnqueueAt(
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
			SimpleFin: config.SimpleFin{
				Enabled: true,
				AllowedHosts: []string{
					config.DefaultSimpleFinHost,
				},
			},
		}).MinTimes(1)
		context.EXPECT().DB().Return(db).MinTimes(1)
		context.EXPECT().Enqueuer().Return(enqueuer).Times(0)
		context.EXPECT().Log().Return(log).MinTimes(1)

		err := simplefin_jobs.CleanupSimpleFinCron(
			mockqueue.NewMockContext(context),
		)
		assert.NoError(t, err, "cleanup simplefin cron should succeed")
	})

	t.Run("enqueue fails", func(t *testing.T) {
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
			Kind:  models.SecretKindSimpleFin,
			Value: mock_simplefin.AccessUrl(t),
		}
		assert.NoError(
			t,
			secretsRepo.Store(t.Context(), &secret),
			"must be able to create simplefin secret",
		)

		simpleFinLink := models.SimpleFinLink{
			AccountId: user.AccountId,
			SecretId:  secret.SecretId,
			Name:      "Test SimpleFIN Link",
			ApiUrl:    "https://bridge.simplefin.org/simplefin",
			Status:    models.SimpleFinLinkStatusPending,
			CreatedBy: user.UserId,
		}
		assert.NoError(
			t,
			repo.CreateSimpleFinLink(t.Context(), &simpleFinLink),
			"must be able to create simplefin link",
		)

		// Move the clock forward past 24 hours so the link becomes stale.
		clock.Add(25 * time.Hour)

		enqueuer := mockgen.NewMockProcessor(ctrl)

		enqueuer.EXPECT().
			BulkEnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(simplefin_jobs.CleanupSimpleFin),
				gomock.Any(),
				gomock.Any(),
			).
			Return(errors.New("database is on fire")).
			Times(1)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).MinTimes(1)
		context.EXPECT().Configuration().Return(config.Configuration{
			SimpleFin: config.SimpleFin{
				Enabled: true,
				AllowedHosts: []string{
					config.DefaultSimpleFinHost,
				},
			},
		}).MinTimes(1)
		context.EXPECT().DB().Return(db).MinTimes(1)
		context.EXPECT().Enqueuer().Return(enqueuer).MinTimes(1)
		context.EXPECT().Log().Return(log).MinTimes(1)

		err := simplefin_jobs.CleanupSimpleFinCron(
			mockqueue.NewMockContext(context),
		)
		assert.EqualError(t, err, "database is on fire")
	})
}

func TestCleanupSimpleFin(t *testing.T) {
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
			Kind:  models.SecretKindSimpleFin,
			Value: mock_simplefin.AccessUrl(t),
		}
		assert.NoError(
			t,
			secretsRepo.Store(t.Context(), &secret),
			"must be able to create simplefin secret",
		)

		simpleFinLink := models.SimpleFinLink{
			AccountId: user.AccountId,
			SecretId:  secret.SecretId,
			Name:      "Test SimpleFIN Link",
			ApiUrl:    "https://bridge.simplefin.org/simplefin",
			Status:    models.SimpleFinLinkStatusPending,
			CreatedBy: user.UserId,
		}
		assert.NoError(
			t,
			repo.CreateSimpleFinLink(t.Context(), &simpleFinLink),
			"must be able to create simplefin link",
		)

		firstSimpleFinBankAccount := models.SimpleFinBankAccount{
			AccountId:       user.AccountId,
			SimpleFinLinkId: simpleFinLink.SimpleFinLinkId,
			SimpleFinId:     "ACT-1234",
			ConnectionId:    "CON-1234",
			Name:            "First SimpleFIN Account",
			InstitutionName: "Lehman Brothers",
			Currency:        "USD",
			Status:          models.SimpleFinBankAccountStatusInactive,
			CurrentBalance:  100,
			CreatedBy:       user.UserId,
		}
		assert.NoError(
			t,
			repo.CreateSimpleFinBankAccount(t.Context(), &firstSimpleFinBankAccount),
			"must be able to create first simplefin bank account",
		)

		secondSimpleFinBankAccount := models.SimpleFinBankAccount{
			AccountId:       user.AccountId,
			SimpleFinLinkId: simpleFinLink.SimpleFinLinkId,
			SimpleFinId:     "ACT-1235",
			ConnectionId:    "CON-1234",
			Name:            "Second SimpleFIN Account",
			InstitutionName: "Lehman Brothers",
			Currency:        "USD",
			Status:          models.SimpleFinBankAccountStatusInactive,
			CurrentBalance:  250,
			CreatedBy:       user.UserId,
		}
		assert.NoError(
			t,
			repo.CreateSimpleFinBankAccount(t.Context(), &secondSimpleFinBankAccount),
			"must be able to create second simplefin bank account",
		)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).MinTimes(1)
		context.EXPECT().DB().Return(db).MinTimes(1)
		context.EXPECT().Log().Return(log).MinTimes(1)
		context.EXPECT().KMS().Return(kms).MinTimes(1)
		context.EXPECT().Configuration().Return(config.Configuration{
			SimpleFin: config.SimpleFin{
				Enabled: true,
				AllowedHosts: []string{
					config.DefaultSimpleFinHost,
				},
			},
		}).MinTimes(1)

		err := simplefin_jobs.CleanupSimpleFin(
			mockqueue.NewMockContext(context),
			simplefin_jobs.CleanupSimpleFinArguments{
				AccountId:       user.AccountId,
				SimpleFinLinkId: simpleFinLink.SimpleFinLinkId,
			},
		)
		assert.NoError(t, err, "cleanup simplefin job should succeed")
		// Make sure that we have fully deleted all of our objects that are related
		// to the simplefin link.
		testutils.MustDBNotExist(t, firstSimpleFinBankAccount)
		testutils.MustDBNotExist(t, secondSimpleFinBankAccount)
		testutils.MustDBNotExist(t, models.Secret{
			SecretId:  secret.SecretId,
			AccountId: user.AccountId,
		})
		testutils.MustDBNotExist(t, simpleFinLink)
	})

	t.Run("wont delete a non-pending link", func(t *testing.T) {
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
			Kind:  models.SecretKindSimpleFin,
			Value: mock_simplefin.AccessUrl(t),
		}
		assert.NoError(
			t,
			secretsRepo.Store(t.Context(), &secret),
			"must be able to create simplefin secret",
		)

		simpleFinLink := models.SimpleFinLink{
			AccountId: user.AccountId,
			SecretId:  secret.SecretId,
			Name:      "Test SimpleFIN Link",
			ApiUrl:    "https://bridge.simplefin.org/simplefin",
			Status:    models.SimpleFinLinkStatusActive,
			CreatedBy: user.UserId,
		}
		assert.NoError(
			t,
			repo.CreateSimpleFinLink(t.Context(), &simpleFinLink),
			"must be able to create simplefin link in the active status",
		)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).MinTimes(1)
		context.EXPECT().DB().Return(db).MinTimes(1)
		context.EXPECT().Log().Return(log).MinTimes(1)
		context.EXPECT().KMS().Return(kms).MinTimes(1)
		context.EXPECT().Configuration().Return(config.Configuration{
			SimpleFin: config.SimpleFin{
				Enabled: true,
				AllowedHosts: []string{
					config.DefaultSimpleFinHost,
				},
			},
		}).MinTimes(1)

		err := simplefin_jobs.CleanupSimpleFin(
			mockqueue.NewMockContext(context),
			simplefin_jobs.CleanupSimpleFinArguments{
				AccountId:       user.AccountId,
				SimpleFinLinkId: simpleFinLink.SimpleFinLinkId,
			},
		)
		assert.NoError(t, err, "cleanup simplefin job should succeed")

		// Make sure that we have not deleted anything, even though we did not
		// return an error.
		testutils.MustDBExist(t, models.Secret{
			SecretId:  secret.SecretId,
			AccountId: user.AccountId,
		})
		testutils.MustDBExist(t, simpleFinLink)
	})
}
