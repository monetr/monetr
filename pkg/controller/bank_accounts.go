package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
)

func (c *Controller) handleBankAccounts(p iris.Party) {
	p.Get("/", func(ctx *context.Context) {

	})

	// Create transactions manually. Check to see if the bank account is in manual mode.
	p.Post("/", func(ctx *context.Context) {

	})

	// Update transactions
	p.Put("/{transactionId:uint64}", func(ctx *context.Context) {

	})
}
