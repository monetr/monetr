package fixtures

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/util"
	"github.com/stretchr/testify/require"
)

func GivenIHaveAFundingSchedule(t *testing.T, bankAccount *models.BankAccount, ruleString string, excludeWeekends bool) *models.FundingSchedule {
	require.NotNil(t, bankAccount, "must provide a valid bank account")
	require.NotZero(t, bankAccount.BankAccountId, "bank account must have a valid Id")
	require.NotZero(t, bankAccount.AccountId, "bank account must have a valid account Id")
	require.NotZero(t, bankAccount.Link.CreatedByUserId, "bank account must have a valid created by user Id")

	if excludeWeekends {
		panic("sorry I haven't implemented this yet")
	}

	db := testutils.GetPgDatabase(t)
	repo := repository.NewRepositoryFromSession(bankAccount.Link.CreatedByUserId, bankAccount.AccountId, db)

	timezone := testutils.MustEz(t, bankAccount.Account.GetTimezone)
	rule := testutils.RuleToSet(t, timezone, ruleString)
	nextOccurrence := util.Midnight(rule.After(time.Now(), false), timezone)

	fundingSchedule := models.FundingSchedule{
		AccountId:              bankAccount.AccountId,
		Account:                bankAccount.Account,
		BankAccountId:          bankAccount.BankAccountId,
		BankAccount:            bankAccount,
		Name:                   gofakeit.Generate("Payday {uuid}"),
		Description:            gofakeit.Generate("{sentence:5}"),
		RuleSet:                rule,
		ExcludeWeekends:        excludeWeekends,
		LastOccurrence:         nil,
		NextOccurrence:         nextOccurrence,
		NextOccurrenceOriginal: nextOccurrence,
	}

	require.NoError(t, repo.CreateFundingSchedule(context.Background(), &fundingSchedule), "must be able to create funding schedule")

	return &fundingSchedule
}
