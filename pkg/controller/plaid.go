package controller

import (
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/core/router"
	"github.com/plaid/plaid-go/plaid"
)

func (c *Controller) handlePlaidLinkEndpoints(p router.Party) {
	p.Get("/token/new", func(ctx *context.Context) {

		c.plaid.CreateLinkToken(plaid.LinkTokenConfigs{
			User:        nil,
			ClientName:  "",
			Products:    nil,
			AccessToken: "",
			CountryCodes: []string{
				"US",
			},
			Webhook:               "",
			AccountFilters:        nil,
			CrossAppItemAdd:       nil,
			PaymentInitiation:     nil,
			Language:              "",
			LinkCustomizationName: "",
			RedirectUri:           "",
			AndroidPackageName:    "",
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
