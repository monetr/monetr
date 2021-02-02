package controller

import (
	"net/url"

	"github.com/kataras/iris/v12/context"
)

func (c *Controller) handlePlaidWebhook(ctx *context.Context) {

}

func (c *Controller) getWebhookUrl() string {
	if !c.configuration.EnableWebhooks {
		return ""
	}

	uri, err := url.Parse(c.configuration.APIDomainName)
	if err != nil {
		panic(err)
	}

	return uri.String()
}
