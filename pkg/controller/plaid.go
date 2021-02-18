package controller

import (
	"fmt"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"net/http"
	"strconv"
	"time"

	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/core/router"
	"github.com/plaid/plaid-go/plaid"
)

func (c *Controller) handlePlaidLinkEndpoints(p router.Party) {
	p.Get("/token/new", func(ctx *context.Context) {
		// Retrieve the user's details. We need to pass some of these along to
		// plaid as part of the linking process.
		me, err := c.mustGetAuthenticatedRepository(ctx).GetMe()
		if err != nil {
			c.wrapAndReturnError(ctx, err,
				http.StatusInternalServerError, "failed to get user details for link",
			)
		}

		userId := c.mustGetUserId(ctx)
		plaidProducts := []string{
			"transactions",
		}

		legalName := ""
		if len(me.LastName) > 0 {
			legalName = fmt.Sprintf("%s %s", me.FirstName, me.LastName)
		}

		var phoneNumber string
		if me.Login.PhoneNumber != nil {
			phoneNumber = me.Login.PhoneNumber.E164()
		}

		token, err := c.plaid.CreateLinkToken(plaid.LinkTokenConfigs{
			User: &plaid.LinkTokenUser{
				ClientUserID: strconv.FormatUint(userId, 10),
				LegalName:    legalName,
				PhoneNumber:  phoneNumber,
				EmailAddress: me.Login.Email,
				// TODO (elliotcourant) I'm going to leave these be for now but we need
				//  to loop back and add this once email/phone verification is working.
				PhoneNumberVerifiedTime:  time.Time{},
				EmailAddressVerifiedTime: time.Time{},
			},
			ClientName: "Hard",
			Products:   plaidProducts,
			CountryCodes: []string{
				"US",
			},
			// TODO (elliotcourant) Implement webhook once we are running in kube.
			Webhook:               "",
			AccountFilters:        nil,
			CrossAppItemAdd:       nil,
			PaymentInitiation:     nil,
			Language:              "en",
			LinkCustomizationName: "",
			RedirectUri:           "",
		})
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create link token")
			return
		}

		ctx.JSON(map[string]interface{}{
			"linkToken": token.LinkToken,
		})
	})

	p.Post("/token/callback", func(ctx *context.Context) {
		var callbackRequest struct {
			PublicToken     string   `json:"publicToken"`
			InstitutionId   string   `json:"institutionId"`
			InstitutionName string   `json:"institutionName"`
			AccountIds      []string `json:"accountIds"`
		}
		if err := ctx.ReadJSON(&callbackRequest); err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed json")
			return
		}

		if len(callbackRequest.AccountIds) == 0 {
			c.returnError(ctx, http.StatusBadRequest, "must select at least one account")
			return
		}

		result, err := c.plaid.ExchangePublicToken(callbackRequest.PublicToken)
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to exchange token")
			return
		}

		plaidAccounts, err := c.plaid.GetAccountsWithOptions(result.AccessToken, plaid.GetAccountsOptions{
			AccountIDs: callbackRequest.AccountIds,
		})
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve accounts")
			return
		}

		if len(plaidAccounts.Accounts) == 0 {
			c.returnError(ctx, http.StatusInternalServerError, "could not retrieve details for any accounts")
			return
		}

		repo := c.mustGetAuthenticatedRepository(ctx)

		plaidLink := models.PlaidLink{
			ItemId:      result.ItemID,
			AccessToken: result.AccessToken,
			Products: []string{
				// TODO (elliotcourant) Make this based on what product's we sent in the create link token request.
				"transactions",
			},
			WebhookUrl:      "",
			InstitutionId:   callbackRequest.InstitutionId,
			InstitutionName: callbackRequest.InstitutionName,
		}
		if err := repo.CreatePlaidLink(&plaidLink); err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to store credentials")
			return
		}

		link := models.Link{
			AccountId:       repo.AccountId(),
			PlaidLinkId:     &plaidLink.PlaidLinkID,
			LinkType:        models.PlaidLinkType,
			InstitutionName: callbackRequest.InstitutionName,
			CreatedByUserId: repo.UserId(),
		}
		if err = repo.CreateLink(&link); err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create link")
			return
		}

		accounts := make([]models.BankAccount, len(plaidAccounts.Accounts))
		for i, plaidAccount := range plaidAccounts.Accounts {
			accounts[i] = models.BankAccount{
				AccountId:         repo.AccountId(),
				LinkId:            link.LinkId,
				PlaidAccountId:    plaidAccount.AccountID,
				AvailableBalance:  int64(plaidAccount.Balances.Available * 100),
				CurrentBalance:    int64(plaidAccount.Balances.Current * 100),
				Mask:              plaidAccount.Mask,
				PlaidName:         plaidAccount.Name,
				PlaidOfficialName: plaidAccount.OfficialName,
				Type:              plaidAccount.Type,
				SubType:           plaidAccount.Subtype,
			}
		}
		if err = repo.CreateBankAccounts(accounts); err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create bank accounts")
			return
		}

		ctx.JSON(map[string]interface{}{
			"success": true,
		})
	})
}
