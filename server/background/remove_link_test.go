package background_test

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestRemoveLinkJob_Run(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		clock := clock.New()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		transactions := fixtures.GivenIHaveNTransactions(t, clock, bankAccount, 100)

		{ // Deactivate the plaid link!
			link.PlaidLink.Status = models.PlaidLinkStatusDeactivated
			link.PlaidLink.DeletedAt = myownsanity.Pointer(clock.Now())
			testutils.MustDBUpdate(t, link.PlaidLink)
		}

		job, err := background.NewRemoveLinkJob(
			log,
			db,
			clock,
			publisher,
			background.RemoveLinkArguments{
				AccountId: user.AccountId,
				LinkId:    bankAccount.LinkId,
			},
		)
		assert.NoError(t, err, "should not return an error creating the job")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		assert.NotPanics(t, func() {
			assert.NoError(t, job.Run(ctx), "remove link job should succeed")
		})

		{ // Make sure all data has been removed
			for i := range transactions {
				transaction := transactions[i]
				testutils.MustDBNotExist(t, *transaction.PlaidTransaction)
				testutils.MustDBNotExist(t, transaction)
			}

			testutils.MustDBNotExist(t, *bankAccount.PlaidBankAccount)
			testutils.MustDBNotExist(t, bankAccount)
			testutils.MustDBNotExist(t, *link.PlaidLink)
			testutils.MustDBNotExist(t, link)
		}
	})

	t.Run("no transactions", func(t *testing.T) {
		clock := clock.New()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)

		bankAccount := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		job, err := background.NewRemoveLinkJob(
			log,
			db,
			clock,
			publisher,
			background.RemoveLinkArguments{
				AccountId: user.AccountId,
				LinkId:    bankAccount.LinkId,
			},
		)
		assert.NoError(t, err, "should not return an error creating the job")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		assert.NotPanics(t, func() {
			assert.NoError(t, job.Run(ctx), "remove link job should succeed")
		})

		{ // Make sure all data has been removed
			testutils.MustDBNotExist(t, *bankAccount.PlaidBankAccount)
			testutils.MustDBNotExist(t, bankAccount)
			testutils.MustDBNotExist(t, *link.PlaidLink)
			testutils.MustDBNotExist(t, link)
		}
	})

	t.Run("still active plaid link", func(t *testing.T) {
		clock := clock.New()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		publisher := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveAPlaidBankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		transactions := fixtures.GivenIHaveNTransactions(t, clock, bankAccount, 100)

		job, err := background.NewRemoveLinkJob(
			log,
			db,
			clock,
			publisher,
			background.RemoveLinkArguments{
				AccountId: user.AccountId,
				LinkId:    bankAccount.LinkId,
			},
		)
		assert.NoError(t, err, "should not return an error creating the job")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		assert.NotPanics(t, func() {
			assert.EqualError(t, job.Run(ctx), "cannot remove data for link that is not in a deactivated status")
		})

		{ // Make sure all data has not been removed if the check fails
			for i := range transactions {
				transaction := transactions[i]
				testutils.MustDBExist(t, *transaction.PlaidTransaction)
				testutils.MustDBExist(t, transaction)
			}

			testutils.MustDBExist(t, *bankAccount.PlaidBankAccount)
			testutils.MustDBExist(t, bankAccount)
			testutils.MustDBExist(t, *link.PlaidLink)
			testutils.MustDBExist(t, link)
		}
	})
}
