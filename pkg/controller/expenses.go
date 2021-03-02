package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"net/http"
)

func (c *Controller) handleExpenses(p iris.Party) {
	p.Get("/{bankAccountId:uint64}/expenses", func(ctx *context.Context) {
		bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
		if bankAccountId == 0 {
			c.returnError(ctx, http.StatusBadRequest, "must specify valid bank account Id")
			return
		}

		repo := c.mustGetAuthenticatedRepository(ctx)
		expenses, err := repo.GetExpenses(bankAccountId)
		if err != nil {
			c.wrapPgError(ctx, err, "could not retrieve expenses")
			return
		}

		ctx.JSON(map[string]interface{}{
			"expenses": expenses,
		})
	})

	p.Post("/{bankAccountId:uint64}/expenses", func(ctx *context.Context) {

	})

	p.Put("/{bankAccountId:uint64}/expenses/{expenseId:uint64}", func(ctx *context.Context) {

	})

	p.Delete("/{bankAccountId:uint64}/expenses/{expenseId:uint64}", func(ctx *context.Context) {

	})
}
