package controller

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/kataras/iris/v12"
	"github.com/monetr/monetr/pkg/background"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/swag"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (c *Controller) linksController(p iris.Party) {
	// GET will list all the links in the current account.
	p.Get("/", c.getLinks)
	p.Get("/{linkId:uint64}", c.getLink)
	p.Post("/", c.postLinks)
	p.Put("/{linkId:uint64}", c.putLink)
	p.Put("/convert/{linkId:uint64}", c.convertLink)
	p.Delete("/{linkId:uint64}", c.deleteLink)
	p.Get("/wait/{linkId:uint64}", c.waitForDeleteLink)
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

// Get Link
// @Summary Get Link
// @id get-link
// @tags Links
// @description Retrieve a single specific link using the link's unique Id.
// @Produce json
// @Security ApiKeyAuth
// @Router /links/{linkId} [get]
// @Param linkId path int true "Link ID"
// @Success 200 {object} swag.LinkResponse
// @Failure 400 {object} InvalidLinkIdError The provided link Id is not valid.
// @Failure 402 {object} SubscriptionNotActiveError The user's subscription is not active.
// @Failure 404 {object} LinkNotFoundError The link could not be found.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) getLink(ctx iris.Context) {
	linkId := ctx.Params().GetUint64Default("linkId", 0)
	if linkId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify a link Id to retrieve")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	links, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve link")
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

// Update Link
// @Summary Update Link
// @id update-link
// @tags Links
// @description Update an existing link.
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Router /links/{linkId} [put]
// @Param linkId path int true "Link ID"
// @Param newLink body swag.UpdateLinkRequest true "Updated Link"
// @Success 200 {object} swag.LinkResponse "Updated link object after changes."
// @Success 304 {object} swag.LinkResponse "If no updates were made then the link object is returned unchanged."
// @Failure 402 {object} SubscriptionNotActiveError The user's subscription is not active.
// @Failure 404 {object} LinkNotFoundError The link could not be found.
// @Failure 500 {object} ApiError "Something went wrong on our end."
func (c *Controller) putLink(ctx iris.Context) {
	linkId := ctx.Params().GetUint64Default("linkId", 0)
	if linkId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify a link Id to update")
		return
	}

	var link swag.UpdateLinkRequest
	if err := ctx.ReadJSON(&link); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	existingLink, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve existing link for update")
		return
	}

	hasUpdate := false

	if link.CustomInstitutionName != "" {
		existingLink.CustomInstitutionName = link.CustomInstitutionName
		hasUpdate = true
	}

	if !hasUpdate {
		ctx.JSON(&existingLink)
		ctx.StatusCode(http.StatusNotModified)
		return
	}

	if err = repo.UpdateLink(c.getContext(ctx), existingLink); err != nil {
		c.wrapPgError(ctx, err, "could not update link")
		return
	}

	ctx.JSON(existingLink)
}

// Convert A Link To Manual
// @Summary Convert A Link To Manual
// @id convert-link
// @tags Links
// @description Convert an existing link into a manual one.
// @Produce json
// @Security ApiKeyAuth
// @Router /links/convert/{linkId} [put]
// @Param linkId path int true "Link ID"
// @Success 200 {object} swag.LinkResponse "New link object after being converted to a manual link."
// @Failure 400 {object} ApiError "The link specified is already a manual link."
// @Failure 402 {object} SubscriptionNotActiveError The user's subscription is not active.
// @Failure 404 {object} LinkNotFoundError A not found status code and an error is returned if the provided link ID does not exist.
// @Failure 500 {object} ApiError "Something went wrong on our end."
func (c *Controller) convertLink(ctx iris.Context) {
	linkId := ctx.Params().GetUint64Default("linkId", 0)
	if linkId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify a link Id to convert")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	link, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		if errors.Is(errors.Cause(err), pg.ErrNoRows) {
			c.notFound(ctx, "the specified link ID does not exist")
			return
		}

		c.wrapPgError(ctx, err, "could not retrieve link to convert")
		return
	}

	if link.LinkType == models.ManualLinkType {
		c.badRequest(ctx, "link is already manual")
		return
	}

	link.LinkType = models.ManualLinkType
	link.LinkStatus = models.LinkStatusSetup

	if err = repo.UpdateLink(c.getContext(ctx), link); err != nil {
		c.wrapPgError(ctx, err, "failed to convert link to manual")
		return
	}

	ctx.JSON(link)
}

// Delete Link
// @Summary Delete Link
// @id delete-link
// @tags Links
// @description Remove a link from your account. This will remove
// @description - All bank accounts associated with this link.
// @description - All spending objects associated with each of those bank accounts.
// @description - All transactions for the those bank accounts.
// @description This cannot be undone and data cannot be recovered.
// @description If the link specified is a Plaid link, then the access_token associated with that link will also be
// @description revoked. Link data is deleted in the background, so if you need to "wait" for all of the link's data to
// @description be properly deleted. Then you should poll the `/link/wait` endpoint.
// @Security ApiKeyAuth
// @Produce json
// @Param linkId path int true "Link ID for the plaid link that is being setup. NOTE: Not Plaid's ID, this is a numeric ID we assign to the object that is returned from the callback endpoint."
// @Router /links/{linkId} [delete]
// @Success 200
// @Failure 400 {object} ApiError A bad request can be returned if the link you specified is not valid.
// @Failure 402 {object} SubscriptionNotActiveError The user's subscription is not active.
// @Failure 404 {object} LinkNotFoundError A not found status code and an error is returned if the provided link ID does not exist.
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
		if errors.Is(errors.Cause(err), pg.ErrNoRows) {
			c.notFound(ctx, "the specified link ID does not exist")
			return
		}

		c.wrapPgError(ctx, err, "failed to retrieve the specified link")
		return
	}

	if link.LinkType == models.PlaidLinkType {
		if link.PlaidLink == nil {
			crumbs.Error(c.getContext(ctx), "BUG: Plaid Link object was missing on the link object", "bug", map[string]interface{}{
				"linkId": link.LinkId,
			})
			c.returnError(ctx, http.StatusInternalServerError, "missing plaid data to remove link")
			return
		}

		accessToken, err := c.plaidSecrets.GetAccessTokenForPlaidLinkId(c.getContext(ctx), repo.AccountId(), link.PlaidLink.ItemId)
		if err != nil {
			crumbs.Error(c.getContext(ctx), "Failed to retrieve access token for plaid link.", "secrets", map[string]interface{}{
				"linkId": link.LinkId,
				"itemId": link.PlaidLink.ItemId,
			})
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve access token for removal")
			return
		}

		client, err := c.plaid.NewClient(c.getContext(ctx), link, accessToken, link.PlaidLink.ItemId)
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create plaid client")
			return
		}

		if err = client.RemoveItem(c.getContext(ctx)); err != nil {
			crumbs.Error(c.getContext(ctx), "Failed to remove item", "plaid", map[string]interface{}{
				"linkId": link.LinkId,
				"itemId": link.PlaidLink.ItemId,
				"error":  err.Error(),
			})
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to remove item from Plaid")
			return
		}

		if err = c.plaidSecrets.RemoveAccessTokenForPlaidLink(c.getContext(ctx), repo.AccountId(), link.PlaidLink.ItemId); err != nil {
			crumbs.Error(c.getContext(ctx), "Failed to remove access token", "secrets", map[string]interface{}{
				"linkId": link.LinkId,
				"itemId": link.PlaidLink.ItemId,
				"error":  err.Error(),
			})
			// We don't want to stop the request here, it does suck that we weren't able to remove the access token, but
			// at this point the user could not retry this request. So we have to commit to it. If a stray access token
			// is left over then that is okay. We could add a job later to do periodic cleanup if this becomes an issue.
		}
	}

	if err = background.TriggerRemoveLink(c.getContext(ctx), c.jobRunner, background.RemoveLinkArguments{
		AccountId: link.AccountId,
		LinkId:    link.LinkId,
	}); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to enqueue link removal job")
		return
	}
}

// Wait For Link Deletion
// @Summary Wait For Link Deletion
// @id wait-for-link-deletion
// @tags Links
// @description This endpoint is used to "wait" for all of the data associated with a link to be deleted. If the link is
// @description is already deleted then a simple **200** is returned to the caller. If the link is not deleted then this
// @description endpoint will block for up to 30 seconds at a time while it waits for the link to be removed. If it is
// @description removed while the endpoint is blocking then it will return 200 at that time.
// @Security ApiKeyAuth
// @Param linkId path int true "Link ID for the link that was/is being removed."
// @Router /link/wait/{linkId:uint64} [get]
// @Success 200
// @Success 408
// @Failure 402 {object} SubscriptionNotActiveError The user's subscription is not active.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) waitForDeleteLink(ctx iris.Context) {
	linkId := ctx.Params().GetUint64Default("linkId", 0)
	if linkId == 0 {
		c.badRequest(ctx, "must specify a job Id")
		return
	}

	log := c.getLog(ctx).WithFields(logrus.Fields{
		"linkId": linkId,
	})
	repo := c.mustGetAuthenticatedRepository(ctx)
	link, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve link")
		return
	}

	// If the link is done just return.
	if link.LinkStatus == models.LinkStatusSetup {
		crumbs.Debug(c.getContext(ctx), "Link is setup, no need to poll.", nil)
		return
	}

	channelName := fmt.Sprintf("link:remove:%d:%d", repo.AccountId(), linkId)

	listener, err := c.ps.Subscribe(c.getContext(ctx), channelName)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to listen on channel")
		return
	}
	defer func() {
		if err = listener.Close(); err != nil {
			log.WithFields(logrus.Fields{
				"accountId": c.mustGetAccountId(ctx),
				"linkId":    linkId,
			}).WithError(err).Error("failed to gracefully close listener")
		}
	}()

	crumbs.Debug(c.getContext(ctx), "Waiting for notification on channel", map[string]interface{}{
		"channel": channelName,
	})

	log.Debugf("waiting for link to be removed on channel: %s", channelName)

	span := sentry.StartSpan(c.getContext(ctx), "Wait For Notification")
	defer span.Finish()

	deadLine := time.NewTimer(30 * time.Second)
	defer deadLine.Stop()

	select {
	case <-deadLine.C:
		ctx.StatusCode(http.StatusRequestTimeout)
		log.Trace("timed out waiting for link to be removed")
		return
	case <-listener.Channel():
		// Just exit successfully, any message on this channel is considered a success.
		log.Trace("link removed successfully")
		return
	}
}
