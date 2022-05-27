package fixtures

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/stretchr/testify/require"
)

func GivenIHaveAFundingSchedule(t *testing.T, bankAccount *models.BankAccount, ruleString string, excludeWeekends bool) *models.FundingSchedule {
	require.NotNil(t, bankAccount, "must provide a valid bank account")
	require.NotZero(t, bankAccount.BankAccountId, "bank account must have a valid Id")
	require.NotZero(t, bankAccount.AccountId, "bank account must have a valid account Id")
	require.NotZero(t, bankAccount.Link.CreatedByUserId, "bank account must have a valid created by user Id")

	db := testutils.GetPgDatabase(t)
	repo := repository.NewRepositoryFromSession(bankAccount.Link.CreatedByUserId, bankAccount.AccountId, db)
	rule := testutils.Must(t, models.NewRule, ruleString)
	tz := testutils.MustEz(t, bankAccount.Account.GetTimezone)
	rule.DTStart(time.Now().In(tz).Add(-30 * 24 * time.Hour))

	fundingSchedule := models.FundingSchedule{
		AccountId:         bankAccount.AccountId,
		Account:           bankAccount.Account,
		BankAccountId:     bankAccount.BankAccountId,
		BankAccount:       bankAccount,
		Name:              gofakeit.Generate("Payday {uuid}"),
		Description:       gofakeit.Generate("{sentence:5}"),
		Rule:              rule,
		ExcludeWeekends:   excludeWeekends,
		LastOccurrence:    nil,
		NextOccurrence:    rule.After(time.Now(), false),
	}

	require.NoError(t, repo.CreateFundingSchedule(context.Background(), &fundingSchedule), "must be able to create funding schedule")

	return &fundingSchedule
}
