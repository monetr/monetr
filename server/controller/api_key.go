package controller

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/powchallenge"
	"github.com/monetr/monetr/server/schemas"
)

func (c *Controller) postApiKey(ctx *echo.Context) error {
	repo := c.mustGetAuthenticatedRepository(ctx)

	// When proof of work is enabled the request must include a solved challenge,
	// so validate against the schema that knows about those fields. When it is
	// disabled the plain schema is used which rejects them outright.
	createSchema := schemas.CreateApiKey
	if c.Configuration.ProofOfWork.Enabled {
		createSchema = schemas.CreateApiKeyChallenge
	}
	request, err := parse(
		c,
		ctx,
		new(schemas.CreateApiKeyRequest),
		createSchema,
	)
	if err != nil {
		return err
	}

	// The cheap fail-fast before we generate a key. No-op when disabled.
	if err := c.validateProofOfWork(
		ctx,
		powchallenge.PurposeCreateApiKey,
		request.Challenge,
		request.Nonce,
	); err != nil {
		return err // validateProofOfWork returns a valid http error.
	}

	key, secret, err := models.NewApiKey()
	if err != nil {
		return c.wrapAndReturnError(
			ctx,
			err,
			http.StatusInternalServerError,
			"Failed to generate API key",
		)
	}
	key.Name = request.Name

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

	// When proof of work is enabled the request must carry a solved challenge in
	// its body. When it is disabled the delete request has no body at all.
	if c.Configuration.ProofOfWork.Enabled {
		request, err := parse(
			c,
			ctx,
			new(schemas.DeleteApiKeyRequest),
			schemas.DeleteApiKeyChallenge,
		)
		if err != nil {
			return err
		}

		if err := c.validateProofOfWork(
			ctx,
			powchallenge.PurposeDeleteApiKey,
			request.Challenge,
			request.Nonce,
		); err != nil {
			return err // validateProofOfWork returns a valid http error.
		}
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
