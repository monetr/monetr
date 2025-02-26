package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/repository"
)

type CreateAPIKeyRequest struct {
	Name      string     `json:"name" validate:"required"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

type CreateAPIKeyResponse struct {
	Key    string      `json:"key"`
	APIKey interface{} `json:"apiKey"`
}

func (c *Controller) RegisterAPIKeyRoutes(g *echo.Group) {
	g.POST("/security/api-keys", c.createAPIKey)
	g.GET("/security/api-keys", c.listAPIKeys)
	g.DELETE("/security/api-keys/:apiKeyId", c.revokeAPIKey)
}

func (c *Controller) createAPIKey(ctx echo.Context) error {
	var request CreateAPIKeyRequest
	if err := ctx.Bind(&request); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed request")
	}

	userId := c.mustGetUserId(ctx)
	repo := repository.NewRepositoryFromSession(c.Clock, userId, c.mustGetAccountId(ctx), c.DB)
	key, apiKey, err := repo.CreateAPIKey(ctx.Request().Context(), string(userId), request.Name, request.ExpiresAt)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create API key")
	}

	return ctx.JSON(http.StatusOK, CreateAPIKeyResponse{
		Key:    key,
		APIKey: apiKey,
	})
}

func (c *Controller) listAPIKeys(ctx echo.Context) error {
	userId := c.mustGetUserId(ctx)
	repo := repository.NewRepositoryFromSession(c.Clock, userId, c.mustGetAccountId(ctx), c.DB)
	keys, err := repo.ListAPIKeys(ctx.Request().Context(), string(userId))
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to list API keys")
	}

	return ctx.JSON(http.StatusOK, keys)
}

func (c *Controller) revokeAPIKey(ctx echo.Context) error {
	userId := c.mustGetUserId(ctx)
	apiKeyIdStr := ctx.Param("apiKeyId")
	if apiKeyIdStr == "" {
		return c.badRequest(ctx, "apiKeyId is required")
	}
	
	apiKeyId, err := strconv.ParseInt(apiKeyIdStr, 10, 64)
	if err != nil {
		return c.badRequest(ctx, "invalid apiKeyId")
	}

	repo := repository.NewRepositoryFromSession(c.Clock, userId, c.mustGetAccountId(ctx), c.DB)
	if err := repo.RevokeAPIKey(ctx.Request().Context(), string(userId), apiKeyId); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to revoke API key")
	}

	return ctx.NoContent(http.StatusOK)
}
