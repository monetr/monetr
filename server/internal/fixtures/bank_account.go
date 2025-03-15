package fixtures

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/require"
)

func ReadBankAccounts(
	t *testing.T,
	clock clock.Clock,
	link models.Link,
) []models.BankAccount {
	require.NotZero(t, link.LinkId, "link id must be included")
	require.NotZero(t, link.AccountId, "link id must be included")

	log := testutils.GetLog(t)
	db := testutils.GetPgDatabase(t)
	repo := repository.NewRepositoryFromSession(
		clock,
		link.CreatedBy,
		link.AccountId,
		db,
		log,
	)

	bankAccounts, err := repo.GetBankAccountsByLinkId(
		t.Context(),
		link.LinkId,
	)
	require.NoError(t, err, "must be able to read bank accounts")

	return bankAccounts
}

func GivenIHaveABankAccount(
	t *testing.T,
	clock clock.Clock,
	link *models.Link,
	accountType models.BankAccountType,
	subType models.BankAccountSubType,
) models.BankAccount {
	require.NotNil(t, link, "link must actually be provided")
	require.NotZero(t, link.LinkId, "link id must be included")
	require.NotZero(t, link.AccountId, "link id must be included")

	log := testutils.GetLog(t)
	db := testutils.GetPgDatabase(t)
	repo := repository.NewRepositoryFromSession(
		clock,
		link.CreatedBy,
		link.AccountId,
		db,
		log,
	)

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
			Currency:         "USD",
			Mask:             gofakeit.Generate("####"),
			Name:             "E-ACCOUNT",
			Type:             accountType,
			SubType:          subType,
			LastUpdated:      clock.Now(),
		},
	}

	err := repo.CreateBankAccounts(t.Context(), banks...)
	require.NoError(t, err, "must seed bank account")
	require.NotZero(t, banks[0].BankAccountId, "bank account Id must have been set")

	return *banks[0]
}

func GivenIHaveAPlaidBankAccount(
	t *testing.T,
	clock clock.Clock,
	link *models.Link,
	accountType models.BankAccountType,
	subType models.BankAccountSubType,
) models.BankAccount {
	require.NotNil(t, link, "link must actually be provided")
	require.NotZero(t, link.LinkId, "link id must be included")
	require.NotZero(t, link.AccountId, "link id must be included")
	require.NotZero(t, link.PlaidLinkId, "link plaid link id must be included")

	log := testutils.GetLog(t)
	db := testutils.GetPgDatabase(t)
	repo := repository.NewRepositoryFromSession(
		clock,
		link.CreatedBy,
		link.AccountId,
		db,
		log,
	)

	current := int64(gofakeit.Number(2000, 100000))
	available := current - int64(gofakeit.Number(100, 2000))

	plaidBankAccount := models.PlaidBankAccount{
		AccountId:    link.AccountId,
		PlaidLinkId:  *link.PlaidLinkId,
		PlaidId:      gofakeit.UUID(),
		Name:         "E-ACCOUNT",
		OfficialName: "E-ACCOUNT",
		Mask:         gofakeit.Generate("####"),
		CreatedAt:    clock.Now(),
		CreatedBy:    link.CreatedBy,
	}
	require.NoError(
		t,
		repo.CreatePlaidBankAccount(t.Context(), &plaidBankAccount),
		"must be able to create the plaid bank account record",
	)

	bankAccount := models.BankAccount{
		AccountId:          link.AccountId,
		Account:            link.Account,
		LinkId:             link.LinkId,
		Link:               link,
		PlaidBankAccountId: &plaidBankAccount.PlaidBankAccountId,
		PlaidBankAccount:   &plaidBankAccount,
		AvailableBalance:   available,
		CurrentBalance:     current,
		Mask:               gofakeit.Generate("####"),
		Name:               "E-ACCOUNT",
		Type:               accountType,
		SubType:            subType,
		LastUpdated:        clock.Now(),
		CreatedAt:          clock.Now(),
	}

	require.NoError(
		t,
		repo.CreateBankAccounts(t.Context(), &bankAccount),
		"must seed bank account",
	)
	require.NotZero(t, bankAccount.BankAccountId, "bank account Id must have been set")

	return bankAccount
}
