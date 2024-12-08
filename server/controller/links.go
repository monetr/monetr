package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (c *Controller) getLinks(ctx echo.Context) error {
	repo := c.mustGetAuthenticatedRepository(ctx)

	links, err := repo.GetLinks(c.getContext(ctx))
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve links")
	}

	return ctx.JSON(http.StatusOK, links)
}

func (c *Controller) getLink(ctx echo.Context) error {
	linkId, err := ParseID[Link](ctx.Param("linkId"))
	if err != nil || linkId.IsZero() {
		return c.badRequest(ctx, "must specify a valid link Id to retrieve")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	links, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve link")
	}

	return ctx.JSON(http.StatusOK, links)
}

func (c *Controller) postLinks(ctx echo.Context) error {
	var err error
	var request struct {
		InstitutionName string  `json:"institutionName"`
		Description     *string `json:"description"`
	}
	if err = ctx.Bind(&request); err != nil {
		return c.invalidJson(ctx)
	}

	request.InstitutionName, err = c.cleanString(ctx, "Institution Name", request.InstitutionName)
	if err != nil {
		return err
	}
	if request.InstitutionName == "" {
		return c.badRequest(ctx, "link must have an institution name")
	}

	// If a description is provided. Trim the space on the description.
	if request.Description != nil {
		desc, err := c.cleanString(ctx, "Description", *request.Description)
		if err != nil {
			return err
		}
		request.Description = &desc
	}

	link := Link{
		InstitutionName: request.InstitutionName,
		Description:     request.Description,
		LinkType:        ManualLinkType,
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	if err := repo.CreateLink(c.getContext(ctx), &link); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not create manual link")
	}

	return ctx.JSON(http.StatusOK, link)
}

func (c *Controller) putLink(ctx echo.Context) error {
	linkId, err := ParseID[Link](ctx.Param("linkId"))
	if err != nil || linkId.IsZero() {
		return c.badRequest(ctx, "must specify a valid link Id to update")
	}

	var request struct {
		Description *string `json:"description"`
	}
	if err := ctx.Bind(&request); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
	}

	// If a description is provided. Trim the space on the description.
	if request.Description != nil {
		desc, err := c.cleanString(ctx, "Description", *request.Description)
		if err != nil {
			return err
		}
		request.Description = &desc
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	existingLink, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve existing link for update")
	}

	hasUpdate := false

	if request.Description != nil {
		existingLink.Description = request.Description
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

func (c *Controller) convertLink(ctx echo.Context) error {
	linkId, err := ParseID[Link](ctx.Param("linkId"))
	if err != nil || linkId.IsZero() {
		return c.badRequest(ctx, "must specify a valid link Id to convert")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	link, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		if errors.Is(errors.Cause(err), pg.ErrNoRows) {
			return c.notFound(ctx, "the specified link ID does not exist")
		}

		return c.wrapPgError(ctx, err, "could not retrieve link to convert")
	}

	if link.LinkType == ManualLinkType {
		return c.badRequest(ctx, "link is already manual")
	}

	link.LinkType = ManualLinkType

	if err = repo.UpdateLink(c.getContext(ctx), link); err != nil {
		return c.wrapPgError(ctx, err, "failed to convert link to manual")
	}

	return ctx.JSON(http.StatusOK, link)
}

func (c *Controller) deleteLink(ctx echo.Context) error {
	linkId, err := ParseID[Link](ctx.Param("linkId"))
	if err != nil || linkId.IsZero() {
		return c.badRequest(ctx, "must specify a valid link Id to delete")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	link, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		if errors.Is(errors.Cause(err), pg.ErrNoRows) {
			return c.notFound(ctx, "the specified link ID does not exist")
		}

		return c.wrapPgError(ctx, err, "failed to retrieve the specified link")
	}

	link.DeletedAt = myownsanity.TimeP(c.Clock.Now().UTC())
	if err := repo.UpdateLink(c.getContext(ctx), link); err != nil {
		return c.wrapPgError(ctx, err, "failed to mark the link as deleted")
	}

	secretsRepo := c.mustGetSecretsRepository(ctx)

	if link.PlaidLink != nil {
		secret, err := secretsRepo.Read(c.getContext(ctx), link.PlaidLink.SecretId)
		if err != nil {
			crumbs.Error(
				c.getContext(ctx),
				"Failed to retrieve access token for plaid link.", "secrets", map[string]interface{}{
					"linkId":   link.LinkId,
					"itemId":   link.PlaidLink.PlaidId,
					"secretId": secret.SecretId,
				},
			)
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve access token for removal")
		}

		client, err := c.Plaid.NewClient(
			c.getContext(ctx),
			link,
			secret.Value,
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

	if err = background.TriggerRemoveLink(c.getContext(ctx), c.JobRunner, background.RemoveLinkArguments{
		AccountId: link.AccountId,
		LinkId:    link.LinkId,
	}); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to enqueue link removal job")
	}

	return ctx.NoContent(http.StatusOK)
}

func (c *Controller) waitForDeleteLink(ctx echo.Context) error {
	// TODO Just deprecate this endpoint entirely and instead use soft delete?
	linkId, err := ParseID[Link](ctx.Param("linkId"))
	if err != nil || linkId.IsZero() {
		return c.badRequest(ctx, "must specify a valid link Id to wait for")
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
	// if link.LinkStatus == LinkStatusSetup {
	// 	crumbs.Debug(c.getContext(ctx), "Link is setup, no need to poll.", nil)
	// 	return ctx.NoContent(http.StatusNoContent)
	// }

	channelName := fmt.Sprintf("link:remove:%s:%s", repo.AccountId(), linkId)

	listener, err := c.PubSub.Subscribe(c.getContext(ctx), channelName)
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
