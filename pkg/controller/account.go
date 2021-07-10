package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/monetr/rest-api/pkg/models"
)

func (c *Controller) handleAccount(p iris.Party) {
	p.Get("/", c.getAccount)
}

type AccountInformationResponse struct {
	Account      models.Account       `json:"account"`
}

func (c *Controller) getAccount(ctx iris.Context) {
}
