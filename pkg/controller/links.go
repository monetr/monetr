package controller

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"net/http"
	"strings"
	"time"
)

func (c *Controller) linksController(p iris.Party) {
	// GET will list all the links in the current account.
	p.Get("/", func(ctx *context.Context) {
		repo := c.mustGetAuthenticatedRepository(ctx)

		links, err := repo.GetLinks()
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve links")
			return
		}

		ctx.JSON(map[string]interface{}{
			"links": links,
		})
	})

	// POST will create a new link, links created this way are manual only. Plaid links must be created through a plaid
	// workflow.
	p.Post("/", func(ctx *context.Context) {
		var link models.Link
		if err := ctx.ReadJSON(&link); err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
			return
		}

		link.LinkId = 0 // Make sure the link Id is unset.
		link.InstitutionName = strings.TrimSpace(link.InstitutionName)
		link.LinkType = models.ManualLinkType

		repo := c.mustGetAuthenticatedRepository(ctx)
		if err := repo.CreateLink(&link); err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not create manual link")
			return
		}

		ctx.JSON(link)
	})

	p.Put("/{linkId:uint64}", func(ctx *context.Context) {
		linkId := ctx.Params().GetUint64Default("linkId", 0)
		if linkId == 0 {
			c.returnError(ctx, http.StatusBadRequest, "must specify a link Id to update")
			return
		}

		var link models.Link
		if err := ctx.ReadJSON(&link); err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
			return
		}

		link.LinkId = linkId

		// We are not going to update default value or null fields. So we can simply clear these fields out to make sure
		// the user does not overwrite them somehow.
		link.CreatedByUserId = 0 // Make sure they don't change the created by userId.
		link.CreatedAt = time.Time{}
		link.InstitutionName = "" // This cannot be changed. If the user wants to set a name then they need to change the custom one.
		link.LinkType = 0         // Make sure they don't change the link type. This can be changed, but not by the user.
		link.PlaidLinkId = nil    // Make sure they don't change the plaidLink.

		repo := c.mustGetAuthenticatedRepository(ctx)

		if err := repo.UpdateLink(&link); err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not update link")
			return
		}

		ctx.JSON(link)
	})
}
