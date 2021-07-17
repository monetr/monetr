package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/monetr/rest-api/pkg/models"
	"net/http"
	"strings"
	"time"
)

func (c *Controller) linksController(p iris.Party) {
	// GET will list all the links in the current account.
	p.Get("/", c.getLinks)
	p.Post("/", c.postLinks)
	p.Put("/{linkId:uint64}", c.putLinks)
	p.Delete("/{linkId:uint64}", c.deleteLink)
}

// List all links
// @Summary List All Links
// @id list-all-links
// @tags Links
// @description Lists all of the links for the currently authenticated user.
// @Produce json
// @Security ApiKeyAuth
// @Router /links [get]
// @Success 200 {array} swag.LinkResponse
// @Failure 402 {object} SubscriptionNotActiveError The user's subscription is not active.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) getLinks(ctx iris.Context) {
	repo := c.mustGetAuthenticatedRepository(ctx)

	links, err := repo.GetLinks(c.getContext(ctx))
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve links")
		return
	}

	ctx.JSON(links)
}

// Create A Link
// @Summary Create A Link
// @id create-link
// @tags Links
// @description Create a manual link.
// @Produce json
// @Accept json
// @Security ApiKeyAuth
// @Router /links [post]
// @Param newLink body swag.CreateLinkRequest true "New Manual Link"
// @Success 200 {object} swag.LinkResponse "Newly created manual link"
// @Failure 400 {object} MalformedJSONError "Malformed JSON."
// @Failure 402 {object} SubscriptionNotActiveError The user's subscription is not active.
// @Failure 500 {object} ApiError "Something went wrong on our end."
func (c *Controller) postLinks(ctx iris.Context) {
	var link models.Link
	if err := ctx.ReadJSON(&link); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
		return
	}

	link.LinkId = 0 // Make sure the link Id is unset.
	link.InstitutionName = strings.TrimSpace(link.InstitutionName)
	link.LinkType = models.ManualLinkType
	link.LinkStatus = models.LinkStatusSetup

	repo := c.mustGetAuthenticatedRepository(ctx)
	if err := repo.CreateLink(c.getContext(ctx), &link); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not create manual link")
		return
	}

	ctx.JSON(link)
}

func (c *Controller) putLinks(ctx iris.Context) {
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

	if err := repo.UpdateLink(c.getContext(ctx), &link); err != nil {
		c.wrapPgError(ctx, err, "could not update link")
		return
	}

	ctx.JSON(link)
}

// Delete Manual Link
// @Summary Delete Manual Link
// @id delete-manual-link
// @tags Links
// @description Remove a manual link from your account. This will remove
//  - All bank accounts associated with this link.
//  - All spending objects associated with each of those bank accounts.
//  - All transactions for the those bank accounts.
//  This cannot be undone and data cannot be recovered.
// @Security ApiKeyAuth
// @Produce json
// @Param linkId path int true "Link ID"
// @Router /links/{linkId} [delete]
// @Success 200
// @Failure 400 {object} ApiError A bad request can be returned if you attempt to delete a link that is not manual.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) deleteLink(ctx iris.Context) {
	linkId := ctx.Params().GetUint64Default("linkId", 0)
	if linkId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify a link Id to update")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	link, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve the specified link")
		return
	}

	if link.LinkType != models.ManualLinkType {
		c.badRequest(ctx, "cannot delete a non-manual link")
		return
	}

	// TODO Queue the link and its sub-objects for deletion.
}
