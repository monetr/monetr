package fixtures

import (
	"context"
	"fmt"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
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
func GivenIHaveAPlaidLink(t *testing.T, clock clock.Clock, user models.User) models.Link {
	log := testutils.GetLog(t)
	db := testutils.GetPgDatabase(t)

	repo := repository.NewRepositoryFromSession(clock, user.UserId, user.AccountId, db)
	secretsRepo := repository.NewSecretsRepository(
		log,
		clock,
		db,
		secrets.NewPlaintextKMS(),
		user.AccountId,
	)

	secret := repository.Secret{
		Kind:   models.PlaidSecretKind,
		Secret: gofakeit.UUID(),
	}
	err := secretsRepo.Store(context.Background(), &secret)
	require.NoError(t, err, "must be able to see plaid token secret")

	plaidLink := models.PlaidLink{
		AccountId:            user.AccountId,
		SecretId:             secret.SecretId,
		PlaidId:              gofakeit.Generate("???????????????????????????????????"),
		Products:             consts.PlaidProductStrings(),
		Status:               models.PlaidLinkStatusSetup,
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
		CreatedByUserId:      user.UserId,
	}
	err = repo.CreatePlaidLink(context.Background(), &plaidLink)
	require.NoError(t, err, "must be able to seed plaid link")

	link := models.Link{
		AccountId:       user.AccountId,
		Account:         user.Account,
		LinkType:        models.PlaidLinkType,
		PlaidLinkId:     &plaidLink.PlaidLinkId, // To be filled in later.
		PlaidLink:       &plaidLink,
		InstitutionName: plaidLink.InstitutionName,
		CreatedAt:       clock.Now(),
		CreatedByUserId: user.UserId,
		CreatedByUser:   &user,
		UpdatedAt:       clock.Now(),
	}

	err = repo.CreateLink(context.Background(), &link)
	require.NoError(t, err, "must be able to seed link")

	return link
}

func GivenIHaveAManualLink(t *testing.T, clock clock.Clock, user models.User) models.Link {
	db := testutils.GetPgDatabase(t)

	repo := repository.NewRepositoryFromSession(clock, user.UserId, user.AccountId, db)

	link := models.Link{
		AccountId:       user.AccountId,
		Account:         user.Account,
		LinkType:        models.ManualLinkType,
		InstitutionName: "Manual Link",
		CreatedAt:       clock.Now(),
		CreatedByUserId: user.UserId,
		CreatedByUser:   &user,
		UpdatedAt:       clock.Now(),
	}

	err := repo.CreateLink(context.Background(), &link)
	require.NoError(t, err, "must be able to seed link")

	return link
}
