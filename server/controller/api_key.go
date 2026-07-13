package controller

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/schemas"
)

func (c *Controller) postApiKey(ctx *echo.Context) error {
	repo := c.mustGetAuthenticatedRepository(ctx)

	key, secret, err := models.NewApiKey()
	if err != nil {
		return c.wrapAndReturnError(
			ctx,
			err,
			http.StatusInternalServerError,
			"Failed to generate API key",
		)
	}

	key, err = parse(
		c,
		ctx,
		key,
		schemas.CreateApiKey,
	)
	if err != nil {
		return err
	}

	if err := repo.CreateApiKey(c.getContext(ctx), key); err != nil {
		return c.wrapPgError(ctx, err, "Failed to create API key")
	}

	var result struct {
		models.ApiKey
		Secret string `json:"secret"`
	}
	result.ApiKey = *key
	result.Secret = secret

	return ctx.JSON(http.StatusOK, result)
}

func (c *Controller) getApiKeys(ctx *echo.Context) error {
	repo := c.mustGetAuthenticatedRepository(ctx)

	keys, err := repo.GetApiKeys(c.getContext(ctx))
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to list API keys")
	}

	return ctx.JSON(http.StatusOK, keys)
}

func (c *Controller) deleteApiKey(ctx *echo.Context) error {
	apiKeyId, err := models.ParseID[models.ApiKey](ctx.Param("apiKeyId"))
	if err != nil {
		return c.badRequest(ctx, "Must specify a valid API key Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	key, err := repo.GetApiKeyById(c.getContext(ctx), apiKeyId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve API key")
	}

	if key.DeletedAt != nil {
		return c.badRequest(ctx, "API key has already been revoked")
	}

	if err := repo.DeleteApiKey(c.getContext(ctx), apiKeyId); err != nil {
		return c.wrapPgError(ctx, err, "failed to revoke API key")
	}

	return ctx.NoContent(http.StatusOK)
}
