package link_jobs_test

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/links/link_jobs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRemoveLink(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
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

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()

		assert.NotPanics(t, func() {
			err := link_jobs.RemoveLink(
				mockqueue.NewMockContext(context),
				link_jobs.RemoveLinkArguments{
					AccountId: user.AccountId,
					LinkId:    bankAccount.LinkId,
				},
			)
			assert.NoError(t, err, "remove link job should succeed")
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
			testutils.MustDBNotExist(t, models.Secret{
				SecretId:  link.PlaidLink.SecretId,
				AccountId: link.AccountId,
			})
		}
	})

	t.Run("no transactions", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
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

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Publisher().Return(publisher).AnyTimes()

		assert.NotPanics(t, func() {
			err := link_jobs.RemoveLink(
				mockqueue.NewMockContext(context),
				link_jobs.RemoveLinkArguments{
					AccountId: user.AccountId,
					LinkId:    bankAccount.LinkId,
				},
			)
			assert.NoError(t, err, "remove link job should succeed")
		})

		{ // Make sure all data has been removed
			testutils.MustDBNotExist(t, *bankAccount.PlaidBankAccount)
			testutils.MustDBNotExist(t, bankAccount)
			testutils.MustDBNotExist(t, *link.PlaidLink)
			testutils.MustDBNotExist(t, link)
		}
	})
}
