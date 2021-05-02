package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/monetrapp/rest-api/pkg/models"
)

func (c *Controller) handleAccount(p iris.Party) {
	p.Get("/", c.getAccount)
}

type AccountInformationResponse struct {
	Account      models.Account       `json:"account"`
	Subscription *models.Subscription `json:"subscription"`
}

func (c *Controller) getAccount(ctx iris.Context) {
}
