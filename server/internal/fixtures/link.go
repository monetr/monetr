package fixtures

import (
	"fmt"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/internal/testutils"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/stretchr/testify/require"
)

// GivenIHaveAPlaidLink will seed the following models and assoc them and return
// the parent Link model:
//   - Secret
//   - PlaidLink
//   - Link
//
// Note: The secret subobject will be nil on the plaid link.
func GivenIHaveAPlaidLink(t *testing.T, clock clock.Clock, user User) Link {
	log := testutils.GetLog(t)
	db := testutils.GetPgDatabase(t)

	repo := repository.NewRepositoryFromSession(
		clock,
		user.UserId,
		user.AccountId,
		db,
		log,
	)
	secretsRepo := repository.NewSecretsRepository(
		log,
		clock,
		db,
		secrets.NewPlaintextKMS(),
		user.AccountId,
	)

	secret := repository.SecretData{
		Kind:  SecretKindPlaid,
		Value: gofakeit.UUID(),
	}
	err := secretsRepo.Store(t.Context(), &secret)
	require.NoError(t, err, "must be able to see plaid token secret")

	plaidLink := PlaidLink{
		AccountId:            user.AccountId,
		SecretId:             secret.SecretId,
		PlaidId:              gofakeit.Generate("???????????????????????????????????"),
		Products:             consts.PlaidProductStrings(),
		Status:               PlaidLinkStatusSetup,
		ErrorCode:            nil,
		ExpirationDate:       nil,
		NewAccountsAvailable: false,
		WebhookUrl:           "https://monetr.mini/api/plaid/webhook",
		InstitutionId:        gofakeit.Generate("ins_######"),
		InstitutionName:      fmt.Sprintf("Bank Of %s", gofakeit.City()),
		LastManualSync:       nil,
		LastSuccessfulUpdate: nil,
		LastAttemptedUpdate:  nil,
		UpdatedAt:            clock.Now().UTC(),
		CreatedAt:            clock.Now().UTC(),
		CreatedBy:            user.UserId,
	}
	err = repo.CreatePlaidLink(t.Context(), &plaidLink)
	require.NoError(t, err, "must be able to seed plaid link")

	link := Link{
		AccountId:       user.AccountId,
		Account:         user.Account,
		LinkType:        PlaidLinkType,
		PlaidLinkId:     &plaidLink.PlaidLinkId, // To be filled in later.
		PlaidLink:       &plaidLink,
		InstitutionName: plaidLink.InstitutionName,
		CreatedAt:       clock.Now(),
		CreatedBy:       user.UserId,
		CreatedByUser:   &user,
		UpdatedAt:       clock.Now(),
	}

	err = repo.CreateLink(t.Context(), &link)
	require.NoError(t, err, "must be able to seed link")

	return link
}

func GivenIHaveAManualLink(t *testing.T, clock clock.Clock, user User) Link {
	log := testutils.GetLog(t)
	db := testutils.GetPgDatabase(t)
	repo := repository.NewRepositoryFromSession(
		clock,
		user.UserId,
		user.AccountId,
		db,
		log,
	)

	link := Link{
		AccountId:       user.AccountId,
		Account:         user.Account,
		LinkType:        ManualLinkType,
		InstitutionName: "Manual Link",
		CreatedAt:       clock.Now(),
		CreatedBy:       user.UserId,
		CreatedByUser:   &user,
		UpdatedAt:       clock.Now(),
	}

	err := repo.CreateLink(t.Context(), &link)
	require.NoError(t, err, "must be able to seed link")

	return link
}
