package controller

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"net/http"
	"strings"
)

func (c *Controller) handleBankAccounts(p iris.Party) {
	p.Get("/", func(ctx *context.Context) {
		repo := c.mustGetAuthenticatedRepository(ctx)

		bankAccounts, err := repo.GetBankAccounts()
		if err != nil {
			c.wrapPgError(ctx, err, "failed to retrieve bank accounts")
			return
		}

		ctx.JSON(map[string]interface{}{
			"bankAccounts": bankAccounts,
		})
	})

	// Create bank accounts manually.
	p.Post("/", func(ctx *context.Context) {
		var bankAccount models.BankAccount
		if err := ctx.ReadJSON(&bankAccount); err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
			return
		}

		// TODO (elliotcourant) Also verify that the link is a manual link.
		if bankAccount.LinkId == 0 {
			c.returnError(ctx, http.StatusBadRequest, "link Id must be provided")
			return
		}

		bankAccount.BankAccountId = 0
		bankAccount.Name = strings.TrimSpace(bankAccount.Name)
		bankAccount.Mask = strings.TrimSpace(bankAccount.Mask)

		// TODO (elliotcourant) Add proper bank account types that the user can specify. Make them required.
		bankAccount.Type = strings.TrimSpace(bankAccount.Type)
		bankAccount.SubType = strings.TrimSpace(bankAccount.SubType)

		if bankAccount.Name == "" {
			c.returnError(ctx, http.StatusBadRequest, "bank account must have a name")
			return
		}

		repo := c.mustGetAuthenticatedRepository(ctx)

		if err := repo.CreateBankAccounts(bankAccount); err != nil {
			c.wrapPgError(ctx, err, "could not create bank account")
			return
		}

		ctx.JSON(bankAccount)
	})
}
