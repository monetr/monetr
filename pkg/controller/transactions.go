package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/core/router"
	"net/http"
)

func (c *Controller) handleTransactions(p iris.Party) {
	p.PartyFunc("/{bankAccountId:uint64}/transactions", func(p router.Party) {
		p.Get("/", func(ctx *context.Context) {
			bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
			if bankAccountId == 0 {
				c.returnError(ctx, http.StatusBadRequest, "must specify valid bank account Id")
				return
			}

			// TODO Enforce a max limit for the number of transactions that can be requested.
			limit := ctx.URLParamIntDefault("limit", 25)
			offset := ctx.URLParamIntDefault("offset", 0)

			repo := c.mustGetAuthenticatedRepository(ctx)

			transactions, err := repo.GetTransactions(bankAccountId, limit, offset)
			if err != nil {
				c.wrapPgError(ctx, err, "failed to retrieve transactions")
				return
			}

			ctx.JSON(map[string]interface{}{
				"transactions": transactions,
			})
		})

		p.Post("/", func(ctx *context.Context) {
			bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
			if bankAccountId == 0 {
				c.returnError(ctx, http.StatusBadRequest, "must specify valid bank account Id")
				return
			}
		})

		p.Put("/{transactionId:uint64}", func(ctx *context.Context) {
			bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
			if bankAccountId == 0 {
				c.returnError(ctx, http.StatusBadRequest, "must specify valid bank account Id")
				return
			}

		})

		p.Delete("/{transactionId:uint64}", func(ctx *context.Context) {
			bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
			if bankAccountId == 0 {
				c.returnError(ctx, http.StatusBadRequest, "must specify valid bank account Id")
				return
			}

		})
	})
}
