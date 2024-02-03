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
	"github.com/stretchr/testify/require"
)

func GivenIHaveAPlaidLink(t *testing.T, clock clock.Clock, user models.User) models.Link {
	db := testutils.GetPgDatabase(t)

	repo := repository.NewRepositoryFromSession(clock, user.UserId, user.AccountId, db)

	itemId := gofakeit.Generate("???????????????????????????????????")
	plaidToken := models.PlaidToken{
		ItemId:      itemId,
		AccountId:   user.AccountId,
		KeyID:       nil,
		Version:     nil,
		AccessToken: gofakeit.UUID(),
	}
	testutils.MustDBInsert(t, &plaidToken)

	plaidLink := models.PlaidLink{
		AccountId:            user.AccountId,
		PlaidId:              itemId,
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
	err := repo.CreatePlaidLink(context.Background(), &plaidLink)
	require.NoError(t, err, "must be able to seed plaid link")

	link := models.Link{
		AccountId:       user.AccountId,
		Account:         user.Account,
		LinkType:        models.PlaidLinkType,
		PlaidLinkId:     &plaidLink.PlaidLinkID, // To be filled in later.
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
