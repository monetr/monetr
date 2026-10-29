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

func TestRepositoryBase_CreateSimpleFinTransactions(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		transactions := []models.SimpleFinTransaction{
			{
				SimpleFinBankAccountId: *bankAccount.SimpleFinBankAccountId,
				SimpleFinId:            "TRN-1234",
				Description:            "Wendy's",
				Amount:                 -1594,
				Posted:                 clock.Now(),
			},
		}
		err := repo.CreateSimpleFinTransactions(t.Context(), transactions)
		assert.NoError(t, err, "should be able to create simplefin transactions")
		assert.NotZero(t, transactions[0].SimpleFinTransactionId, "should have generated a simplefin transaction Id")
		assert.Equal(t, user.AccountId, transactions[0].AccountId, "account Id should be set by the repository")
		assert.Equal(t, clock.Now().UTC(), transactions[0].CreatedAt, "created at should come from the clock")
		testutils.MustDBExist(t, transactions[0])
	})

	t.Run("duplicate simplefin id", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		err := repo.CreateSimpleFinTransactions(t.Context(), []models.SimpleFinTransaction{
			{
				SimpleFinBankAccountId: *bankAccount.SimpleFinBankAccountId,
				SimpleFinId:            "TRN-1234",
				Description:            "Wendy's",
				Amount:                 -1594,
				Posted:                 clock.Now(),
			},
		})
		assert.NoError(t, err, "must be able to create the first simplefin transaction")

		err = repo.CreateSimpleFinTransactions(t.Context(), []models.SimpleFinTransaction{
			{
				SimpleFinBankAccountId: *bankAccount.SimpleFinBankAccountId,
				SimpleFinId:            "TRN-1234",
				Description:            "Wendy's",
				Amount:                 -1594,
				Posted:                 clock.Now(),
			},
		})
		assert.EqualError(t, err, `failed to create SimpleFIN transactions: ERROR: duplicate key value violates unique constraint "uq_simplefin_transactions_simplefin_id" (SQLSTATE=23505)`)
	})
}
