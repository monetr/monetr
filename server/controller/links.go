package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

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
func (c *Controller) getLinks(ctx echo.Context) error {
	repo := c.mustGetAuthenticatedRepository(ctx)

	links, err := repo.GetLinks(c.getContext(ctx))
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve links")
	}

	return ctx.JSON(http.StatusOK, links)
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
func (c *Controller) getLink(ctx echo.Context) error {
	linkId, err := strconv.ParseUint(ctx.Param("linkId"), 10, 64)
	if err != nil || linkId == 0 {
		return c.badRequest(ctx, "must specify a link Id to retrieve")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	links, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve link")
	}

	return ctx.JSON(http.StatusOK, links)
}

func (c *Controller) postLinks(ctx echo.Context) error {
	var request struct {
		InstitutionName string  `json:"institutionName"`
		Description     *string `json:"description"`
	}
	if err := ctx.Bind(&request); err != nil {
		return c.invalidJson(ctx)
	}

	request.InstitutionName = strings.TrimSpace(request.InstitutionName)
	if request.InstitutionName == "" {
		return c.badRequest(ctx, "link must have an institution name")
	}

	// If a description is provided. Trim the space on the description.
	if request.Description != nil {
		request.Description = myownsanity.StringP(strings.TrimSpace(*request.Description))
	}

	link := models.Link{
		InstitutionName: request.InstitutionName,
		Description:     request.Description,
		LinkType:        models.ManualLinkType,
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	if err := repo.CreateLink(c.getContext(ctx), &link); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not create manual link")
	}

	return ctx.JSON(http.StatusOK, link)
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
func (c *Controller) putLink(ctx echo.Context) error {
	linkId, err := strconv.ParseUint(ctx.Param("linkId"), 10, 64)
	if err != nil {
		return c.returnError(ctx, http.StatusBadRequest, "must specify a link Id to update")
	}

	var link struct {
		Description *string `json:"description"`
	}
	if err := ctx.Bind(&link); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	existingLink, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve existing link for update")
	}

	hasUpdate := false

	if link.Description != nil {
		existingLink.Description = link.Description
		hasUpdate = true
	}

	if !hasUpdate {
		return ctx.NoContent(http.StatusNotModified)
	}

	if err = repo.UpdateLink(c.getContext(ctx), existingLink); err != nil {
		return c.wrapPgError(ctx, err, "could not update link")
	}

	return ctx.JSON(http.StatusOK, existingLink)
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
func (c *Controller) convertLink(ctx echo.Context) error {
	linkId, err := strconv.ParseUint(ctx.Param("linkId"), 10, 64)
	if err != nil {
		return c.returnError(ctx, http.StatusBadRequest, "must specify a link Id to convert")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	link, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		if errors.Is(errors.Cause(err), pg.ErrNoRows) {
			return c.notFound(ctx, "the specified link ID does not exist")
		}

		return c.wrapPgError(ctx, err, "could not retrieve link to convert")
	}

	if link.LinkType == models.ManualLinkType {
		return c.badRequest(ctx, "link is already manual")
	}

	link.LinkType = models.ManualLinkType

	if err = repo.UpdateLink(c.getContext(ctx), link); err != nil {
		return c.wrapPgError(ctx, err, "failed to convert link to manual")
	}

	return ctx.JSON(http.StatusOK, link)
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
func (c *Controller) deleteLink(ctx echo.Context) error {
	linkId, err := strconv.ParseUint(ctx.Param("linkId"), 10, 64)
	if err != nil {
		return c.returnError(ctx, http.StatusBadRequest, "must specify a link Id to delete")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	link, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		if errors.Is(errors.Cause(err), pg.ErrNoRows) {
			return c.notFound(ctx, "the specified link ID does not exist")
		}

		return c.wrapPgError(ctx, err, "failed to retrieve the specified link")
	}

	link.DeletedAt = myownsanity.TimeP(c.clock.Now().UTC())
	if err := repo.UpdateLink(c.getContext(ctx), link); err != nil {
		return c.wrapPgError(ctx, err, "failed to mark the link as deleted")
	}

	secretsRepo := c.mustGetSecretsRepository(ctx)

	if link.PlaidLink != nil {
		secret, err := secretsRepo.Read(c.getContext(ctx), link.PlaidLink.SecretId)
		if err != nil {
			crumbs.Error(c.getContext(ctx), "Failed to retrieve access token for plaid link.", "secrets", map[string]interface{}{
				"linkId":   link.LinkId,
				"itemId":   link.PlaidLink.PlaidId,
				"secretId": secret.SecretId,
			})
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve access token for removal")
		}

		client, err := c.plaid.NewClient(
			c.getContext(ctx),
			link,
			secret.Secret,
			link.PlaidLink.PlaidId,
		)
		if err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create plaid client")
		}

		if err = client.RemoveItem(c.getContext(ctx)); err != nil {
			crumbs.Error(c.getContext(ctx), "Failed to remove item", "plaid", map[string]interface{}{
				"linkId":   link.LinkId,
				"itemId":   link.PlaidLink.PlaidId,
				"secretId": secret.SecretId,
				"error":    err.Error(),
			})
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to remove item from Plaid")
		}
	}
	if link.TellerLink != nil {
		secret, err := secretsRepo.Read(c.getContext(ctx), link.TellerLink.SecretId)
		if err != nil {
			crumbs.Error(c.getContext(ctx), "Failed to retrieve access token for Teller link.", "secrets", map[string]interface{}{
				"linkId":       link.LinkId,
				"enrollmentId": link.TellerLink.EnrollmentId,
				"secretId":     secret.SecretId,
			})
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve access token for removal")
		}

		client := c.teller.GetAuthenticatedClient(secret.Secret)

		bankAccounts, err := repo.GetBankAccountsByLinkId(c.getContext(ctx), link.LinkId)
		if err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve bank accounts")
		}

		for i := range bankAccounts {
			bankAccount := bankAccounts[i]
			if err := client.DeleteAccount(c.getContext(ctx), bankAccount.TellerBankAccount.TellerId); err != nil {
				return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to remove account access")
			}

			bankAccount.TellerBankAccount.Status = models.TellerBankAccountStatusClosed
			if err := repo.UpdateTellerBankAccount(c.getContext(ctx), bankAccount.TellerBankAccount); err != nil {
				return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to update teller bank account")
			}
		}
	}

	if err = background.TriggerRemoveLink(c.getContext(ctx), c.jobRunner, background.RemoveLinkArguments{
		AccountId: link.AccountId,
		LinkId:    link.LinkId,
	}); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to enqueue link removal job")
	}

	return ctx.NoContent(http.StatusOK)
}

func (c *Controller) waitForDeleteLink(ctx echo.Context) error {
	// TODO Just deprecate this endpoint entirely and instead use soft delete?
	linkId, err := strconv.ParseUint(ctx.Param("linkId"), 10, 64)
	if err != nil {
		return c.returnError(ctx, http.StatusBadRequest, "must specify a link Id to wait for")
	}

	log := c.getLog(ctx).WithFields(logrus.Fields{
		"linkId": linkId,
	})
	repo := c.mustGetAuthenticatedRepository(ctx)
	// link, err := repo.GetLink(c.getContext(ctx), linkId)
	// if err != nil {
	// 	return c.wrapPgError(ctx, err, "failed to retrieve link")
	// }

	// If the link is done just return.
	// TODO This is all wrong, why are we checking for link status setup for deleting?
	//      Just going to have it return nothing for now.
	// if link.LinkStatus == models.LinkStatusSetup {
	// 	crumbs.Debug(c.getContext(ctx), "Link is setup, no need to poll.", nil)
	// 	return ctx.NoContent(http.StatusNoContent)
	// }

	channelName := fmt.Sprintf("link:remove:%d:%d", repo.AccountId(), linkId)

	listener, err := c.ps.Subscribe(c.getContext(ctx), channelName)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to listen on channel")
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
		log.Trace("timed out waiting for link to be removed")
		return ctx.NoContent(http.StatusRequestTimeout)
	case <-listener.Channel():
		// Just exit successfully, any message on this channel is considered a success.
		log.Trace("link removed successfully")
		return ctx.NoContent(http.StatusOK)
	}
}
