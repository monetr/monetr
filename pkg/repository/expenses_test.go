package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/monetr/monetr/pkg/internal/fixtures"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryBase_GetSpendingById(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		user, _ := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAManualLink(t, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		db := testutils.GetPgDatabase(t)
		repo := repository.NewRepositoryFromSession(link.CreatedByUserId, link.AccountId, db)
		spending := models.Spending{
			Name:           "Testing",
			CurrentAmount:  0,
			TargetAmount:   100,
			NextRecurrence: time.Now().AddDate(0, 0, 1),
			SpendingType:   models.SpendingTypeGoal,
			BankAccountId:  bankAccount.BankAccountId,
		}
		assert.NoError(t, repo.CreateSpending(context.Background(), &spending), "must create goal")

	})

}
