package fixtures

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/internal/consts"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/stretchr/testify/require"
)

func GivenIHaveAPlaidLink(t *testing.T, user models.User) models.Link {
	db := testutils.GetPgDatabase(t)

	repo := repository.NewRepositoryFromSession(user.UserId, user.AccountId, db)

	plaidLink := models.PlaidLink{
		ItemId:          gofakeit.Generate("???????????????????????????????????"),
		Products:        consts.PlaidProductStrings(),
		WebhookUrl:      "https://monetr.mini/api/plaid/webhook",
		InstitutionId:   gofakeit.Generate("ins_######"),
		InstitutionName: fmt.Sprintf("Bank Of %s", gofakeit.City()),
	}
	err := repo.CreatePlaidLink(context.Background(), &plaidLink)
	require.NoError(t, err, "must be able to seed plaid link")

	link := models.Link{
		AccountId:             user.AccountId,
		Account:               user.Account,
		LinkType:              models.PlaidLinkType,
		PlaidLinkId:           &plaidLink.PlaidLinkID, // To be filled in later.
		PlaidLink:             &plaidLink,
		LinkStatus:            models.LinkStatusSetup,
		InstitutionName:       plaidLink.InstitutionName,
		PlaidInstitutionId:    myownsanity.StringP(gofakeit.Generate("ins_####")),
		CustomInstitutionName: "",
		CreatedAt:             time.Now(),
		CreatedByUserId:       user.UserId,
		CreatedByUser:         &user,
		UpdatedAt:             time.Now(),
		LastSuccessfulUpdate:  myownsanity.TimeP(time.Now()),
		BankAccounts:          nil,
	}

	err = repo.CreateLink(context.Background(), &link)
	require.NoError(t, err, "must be able to seed link")

	return link
}
