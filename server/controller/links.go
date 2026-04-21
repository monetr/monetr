package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/links/link_jobs"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/schema"
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
	var err error
	link, err = parse(
		c,
		ctx,
		schema.CreateLink,
		&link,
	)
	if err != nil {
		return err
	}

	// TODO Come back to this tomorrow, lunch flow link ID is not being properly
	// set here!
	repo := c.mustGetAuthenticatedRepository(ctx)

	// If the user is creating a lunch flow link then we need to validate that the
	// link is valid and can be activated. If it is then activate it as part of
	// this creation step.
	if link.LunchFlowLinkId != nil {
		if !c.Configuration.LunchFlow.Enabled {
			return c.notFound(ctx, "Lunch Flow is not enabled on this server")
		}

		link.LinkType = LunchFlowLinkType
		lunchFlowLink, err := repo.GetLunchFlowLink(
			c.getContext(ctx),
			*link.LunchFlowLinkId,
		)
		if err != nil {
			return c.wrapPgError(ctx, err, "Failed to retrieve lunch flow link")
		}

		if lunchFlowLink.Status != LunchFlowLinkStatusPending {
			return c.badRequest(ctx, "Cannot create a link from a Lunch Flow link that is not in a pending status")
		}

		lunchFlowLink.Status = LunchFlowLinkStatusActive
		if err := repo.UpdateLunchFlowLink(
			c.getContext(ctx),
			lunchFlowLink,
		); err != nil {
			return c.wrapPgError(ctx, err, "Failed to update Lunch Flow link")
		}
	}

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

	link := *existingLink
	link, err = parse(
		c,
		ctx,
		schema.PatchLink,
		existingLink,
	)
	if err != nil {
		return err
	}

	if err = repo.UpdateLink(c.getContext(ctx), &link); err != nil {
		return c.wrapPgError(ctx, err, "failed to update link")
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

	// This way we don't end up making a call to Plaid twice or enqueuing the job
	// multiple times.
	if link.DeletedAt != nil {
		return c.badRequest(ctx, "Link has already been deleted and cannot be deleted again")
	}

	now := c.Clock.Now().UTC()
	link.DeletedAt = &now
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
					"plaidLinkId": link.PlaidLink.PlaidLinkId,
					"linkId":      link.LinkId,
					"itemId":      link.PlaidLink.PlaidId,
					"secretId":    secret.SecretId,
				},
			)
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve access token for removal")
		}

		{ // Mark the plaid link as soft deleted too!
			link.PlaidLink.DeletedAt = &now
			link.PlaidLink.Status = PlaidLinkStatusDeactivated
			if err := repo.UpdatePlaidLink(
				c.getContext(ctx),
				link.PlaidLink,
			); err != nil {
				crumbs.Error(
					c.getContext(ctx),
					"Failed to mark Plaid Link as deleted prior to deactivation", "links", map[string]any{
						"plaidLinkId": link.PlaidLink.PlaidLinkId,
						"linkId":      link.LinkId,
						"itemId":      link.PlaidLink.PlaidId,
					},
				)
				return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve access token for removal")
			}
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
			crumbs.Error(c.getContext(ctx), "Failed to remove item", "plaid", map[string]any{
				"linkId":   link.LinkId,
				"itemId":   link.PlaidLink.PlaidId,
				"secretId": secret.SecretId,
				"error":    err.Error(),
			})
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to remove item from Plaid")
		}
	} else if link.LunchFlowLink != nil {
		link.LunchFlowLink.Status = LunchFlowLinkStatusDeactivated
		link.LunchFlowLink.DeletedAt = myownsanity.Pointer(c.Clock.Now())
		if err := repo.UpdateLunchFlowLink(
			c.getContext(ctx),
			link.LunchFlowLink,
		); err != nil {
			return c.wrapPgError(ctx, err, "Failed to update Lunch Flow Link")
		}
	}

	if err = enqueueJob(
		c,
		ctx,
		link_jobs.RemoveLink,
		link_jobs.RemoveLinkArguments{
			AccountId: link.AccountId,
			LinkId:    link.LinkId,
		},
	); err != nil {
		return c.wrapAndReturnError(
			ctx,
			err,
			http.StatusInternalServerError,
			"failed to enqueue link removal job",
		)
	}

	return ctx.NoContent(http.StatusOK)
}
