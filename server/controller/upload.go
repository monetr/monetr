package controller

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/storage"
)

func (c *Controller) postUploadTransactions(ctx echo.Context) error {
	if c.configuration.Storage.Enabled {
		return c.notFound(ctx, "file uploads are not enabled on this server")
	}

	bankAccountId, err := strconv.ParseUint(ctx.Param("bankAccountId"), 10, 64)
	if err != nil {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	ok, err := repo.GetLinkIsManualByBankAccountId(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to verify bank account link type")
	}
	if !ok {
		return c.badRequest(ctx, "Cannot import transactions for non-manual link.")
	}

	contentType := ctx.Request().Header.Get("Content-Type")
	valid := storage.GetContentTypeIsValid(contentType)
	if !valid {
		crumbs.Debug(c.getContext(ctx),
			"Unsupported file type was provided!",
			map[string]interface{}{
				"contentType": contentType,
			},
		)
		return c.badRequest(ctx, "Unsupported file type!")
	}

	reader, _, err := ctx.Request().FormFile("data")
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "failed to read file upload")
	}

	c.fileStorage.Store(
		c.getContext(ctx),
		reader,
		storage.ContentType(contentType),
	)

	return nil
}
