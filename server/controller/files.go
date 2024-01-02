package controller

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/storage"
)

func (c *Controller) postFile(ctx echo.Context) error {
	if !c.configuration.Storage.Enabled {
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

	reader, header, err := ctx.Request().FormFile("data")
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "failed to read file upload")
	}
	defer reader.Close()

	contentType := header.Header.Get("Content-Type")
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

	fileUri, err := c.fileStorage.Store(
		c.getContext(ctx),
		reader,
		storage.FileInfo{
			Name:          "", // TODO
			AccountId:     c.mustGetAccountId(ctx),
			BankAccountId: bankAccountId,
			ContentType:   storage.ContentType(contentType),
		},
	)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to upload file")
	}

	file := models.File{
		AccountId:     c.mustGetAccountId(ctx),
		BankAccountId: bankAccountId,
		Name:          header.Filename,
		ContentType:   contentType,
		Size:          uint64(header.Size),
		ObjectUri:     fileUri,
	}

	if err := repo.CreateFile(c.getContext(ctx), &file); err != nil {
		return c.wrapPgError(ctx, err, "failed to create file")
	}

	return ctx.JSON(http.StatusOK, file)
}

func (c *Controller) getFiles(ctx echo.Context) error {
	bankAccountId, err := strconv.ParseUint(ctx.Param("bankAccountId"), 10, 64)
	if err != nil {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	files, err := repo.GetFiles(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to list files")
	}

	return ctx.JSON(http.StatusOK, files)
}
