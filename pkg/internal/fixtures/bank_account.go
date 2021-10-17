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

func GivenIHaveABankAccount(t *testing.T, link *models.Link, accountType models.BankAccountType, subType models.BankAccountSubType) models.BankAccount {
	require.NotNil(t, link, "link must actually be provided")
	require.NotZero(t, link.LinkId, "link id must be included")
	require.NotZero(t, link.AccountId, "link id must be included")

	if link.BankAccounts == nil {
		link.BankAccounts = make([]models.BankAccount, 0, 1)
	}

	db := testutils.GetPgDatabase(t)
	repo := repository.NewRepositoryFromSession(link.CreatedByUserId, link.AccountId, db)

	current := int64(gofakeit.Number(2000, 100000))
	available := current - int64(gofakeit.Number(100, 2000))

	// By doing this as an array, its actually a pointer. And can be updated by reference.
	banks := []models.BankAccount{
		{
			AccountId:         link.AccountId,
			Account:           link.Account,
			LinkId:            link.LinkId,
			Link:              link,
			PlaidAccountId:    gofakeit.UUID(),
			AvailableBalance:  available,
			CurrentBalance:    current,
			Mask:              gofakeit.Generate("####"),
			Name:              "E-ACCOUNT",
			PlaidName:         "EACCOUNT",
			PlaidOfficialName: "EACCOUNT",
			Type:              accountType,
			SubType:           subType,
			LastUpdated:       time.Now(),
		},
	}

	err := repo.CreateBankAccounts(context.Background(), banks...)
	require.NoError(t, err, "must seed bank account")
	require.NotZero(t, banks[0].BankAccountId, "bank account Id must have been set")

	link.BankAccounts = append(link.BankAccounts, banks...)

	return banks[0]

}
