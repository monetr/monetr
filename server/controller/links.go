package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/validation"
)

func (c *Controller) getLinks(ctx echo.Context) error {
	repo := c.mustGetAuthenticatedRepository(ctx)

	links, err := repo.GetLinks(c.getContext(ctx))
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve links")
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
	link := Link{
		// Since we can only create manual links via the API directly like this,
		// then set this field initially. It cannot be overwritten by the unmarshall
		// anyway.
		LinkType: ManualLinkType,
	}
	switch err := link.UnmarshalRequest(
		c.getContext(ctx),
		ctx.Request().Body,
		link.CreateValidators()...,
	).(type) {
	case validation.Errors:
		return ctx.JSON(http.StatusBadRequest, map[string]any{
			"error":    "Invalid request",
			"problems": err,
		})
	case nil:
		break
	default:
		return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "failed to parse post request")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	if err := repo.CreateLink(c.getContext(ctx), &link); err != nil {
		return c.wrapPgError(ctx, err, "Could not create a manual link")
	}

	return ctx.JSON(http.StatusOK, link)
}

func (c *Controller) putLink(ctx echo.Context) error {
	linkId, err := ParseID[Link](ctx.Param("linkId"))
	if err != nil || linkId.IsZero() {
		return c.badRequest(ctx, "must specify a valid link Id to update")
	}

	var request struct {
		InstituionName string  `json:"instituionName"`
		Description    *string `json:"description"`
	}
	if err := ctx.Bind(&request); err != nil {
		return c.invalidJson(ctx)
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

func (c *Controller) patchLink(ctx echo.Context) error {
	linkId, err := ParseID[Link](ctx.Param("linkId"))
	if err != nil || linkId.IsZero() {
		return c.badRequest(ctx, "must specify a valid link Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	existingLink, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve link")
	}

	switch err := existingLink.UnmarshalRequest(
		c.getContext(ctx),
		ctx.Request().Body,
		existingLink.UpdateValidator()...,
	).(type) {
	case validation.Errors:
		return ctx.JSON(http.StatusBadRequest, map[string]any{
			"error":    "Invalid request",
			"problems": err,
		})
	case nil:
		break
	default:
		return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "failed to parse patch request")
	}

	if err = repo.UpdateLink(c.getContext(ctx), existingLink); err != nil {
		return c.wrapPgError(ctx, err, "failed to update link")
	}

	return ctx.JSON(http.StatusOK, *existingLink)
}

func (c *Controller) convertLink(ctx echo.Context) error {
	linkId, err := ParseID[Link](ctx.Param("linkId"))
	if err != nil || linkId.IsZero() {
		return c.badRequest(ctx, "must specify a valid link Id to convert")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	link, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
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
				"Failed to retrieve access token for plaid link.", "secrets", map[string]any{
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

	if err = background.TriggerRemoveLink(
		c.getContext(ctx),
		c.JobRunner,
		background.RemoveLinkArguments{
			AccountId: link.AccountId,
			LinkId:    link.LinkId,
		},
	); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to enqueue link removal job")
	}

	return ctx.NoContent(http.StatusOK)
}
