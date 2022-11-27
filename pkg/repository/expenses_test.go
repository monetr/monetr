package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/monetr/monetr/pkg/internal/fixtures"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
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

		rule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")

		fundingSchedule := models.FundingSchedule{
			BankAccountId:    bankAccount.BankAccountId,
			Name:             "Payday",
			Description:      "Elliot's Payday",
			Rule:             rule,
			ExcludeWeekends:  true,
			WaitForDeposit:   false,
			EstimatedDeposit: myownsanity.Int64P(360000),
			LastOccurrence:   nil,
			NextOccurrence:   rule.After(time.Now(), false),
		}
		assert.NoError(t, repo.CreateFundingSchedule(context.Background(), &fundingSchedule), "must create funding schedule")

		spending := models.Spending{
			Name:           "Testing",
			CurrentAmount:  0,
			TargetAmount:   100,
			NextRecurrence: time.Now().AddDate(0, 0, 1),
			SpendingType:   models.SpendingTypeGoal,
			BankAccountId:  bankAccount.BankAccountId,
		}
		assert.NoError(t, repo.CreateSpending(context.Background(), &spending), "must create goal")

		spendingFunding := models.SpendingFunding{
			BankAccountId:          bankAccount.BankAccountId,
			SpendingId:             spending.SpendingId,
			FundingScheduleId:      fundingSchedule.FundingScheduleId,
			NextContributionAmount: 0,
		}
		assert.NoError(t, repo.CreateSpendingFunding(context.Background(), &spendingFunding), "must create spending funding")
		assert.NotZero(t, spendingFunding.SpendingFundingId, "must have a spending funding ID")
		assert.NotZero(t, spendingFunding.AccountId, "must have a spending funding ID")

		fullSpending, err := repo.GetSpendingById(context.Background(), bankAccount.BankAccountId, spending.SpendingId)
		assert.NoError(t, err, "must not return an error reading a valid spending")
		assert.Equal(t, spending.SpendingId, fullSpending.SpendingId, "should retrieve the same spending")
		assert.Len(t, fullSpending.SpendingFunding, 1, "should have one spending funding item")
	})

}
