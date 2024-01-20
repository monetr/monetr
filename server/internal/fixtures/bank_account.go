package fixtures

import (
	"context"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/require"
)

func GivenIHaveABankAccount(t *testing.T, clock clock.Clock, link *models.Link, accountType models.BankAccountType, subType models.BankAccountSubType) models.BankAccount {
	require.NotNil(t, link, "link must actually be provided")
	require.NotZero(t, link.LinkId, "link id must be included")
	require.NotZero(t, link.AccountId, "link id must be included")

	db := testutils.GetPgDatabase(t)
	repo := repository.NewRepositoryFromSession(clock, link.CreatedByUserId, link.AccountId, db)

	current := int64(gofakeit.Number(2000, 100000))
	available := current - int64(gofakeit.Number(100, 2000))

	// By doing this as an array, its actually a pointer. And can be updated by reference.
	banks := []*models.BankAccount{
		{
			AccountId:        link.AccountId,
			Account:          link.Account,
			LinkId:           link.LinkId,
			Link:             link,
			AvailableBalance: available,
			CurrentBalance:   current,
			Mask:             gofakeit.Generate("####"),
			Name:             "E-ACCOUNT",
			Type:             accountType,
			SubType:          subType,
			LastUpdated:      clock.Now(),
		},
	}

	err := repo.CreateBankAccounts(context.Background(), banks...)
	require.NoError(t, err, "must seed bank account")
	require.NotZero(t, banks[0].BankAccountId, "bank account Id must have been set")

	return *banks[0]

}
