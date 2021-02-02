package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/core/router"
	"github.com/plaid/plaid-go/plaid"
)

func (c *Controller) handlePlaidLinkEndpoints(p router.Party) {
	p.Get("/token/new", func(ctx *context.Context) {
		me, err := c.mustGetAuthenticatedRepository(ctx).GetMe()
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to get user details for link")
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
				ClientUserID:             strconv.FormatUint(userId, 10),
				LegalName:                legalName,
				PhoneNumber:              "",
				EmailAddress:             me.Login.Email,
				PhoneNumberVerifiedTime:  time.Time{},
				EmailAddressVerifiedTime: time.Time{},
			},
			ClientName:  "Hard",
			Products:    plaidProducts,
			AccessToken: "",
			CountryCodes: []string{
				"US",
			},
			Webhook:               "",
			AccountFilters:        nil,
			CrossAppItemAdd:       nil,
			PaymentInitiation:     nil,
			Language:              "en-US",
			LinkCustomizationName: "",
			RedirectUri:           "",
		})
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create link token")
			return
		}

		ctx.JSON(map[string]interface{}{
			"linkToken": token,
		})
	})

	p.Post("/token/callback", func(ctx *context.Context) {
		var callbackRequest struct {
			PublicToken     string `json:"publicToken"`
			InstitutionId   string `json:"institutionId"`
			InstitutionName string `json:"institutionName"`
		}
	})
}
