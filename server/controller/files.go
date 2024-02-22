package controller

import (
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/storage"
	"github.com/sirupsen/logrus"
)

func (c *Controller) postFile(ctx echo.Context) error {
	if !c.configuration.Storage.Enabled {
		return c.notFound(ctx, "File uploads are not enabled on this server")
	}

	log := c.getLog(ctx)

	bankAccountId, err := strconv.ParseUint(ctx.Param("bankAccountId"), 10, 64)
	if err != nil {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	log = log.WithField("bankAccountId", bankAccountId)

	repo := c.mustGetAuthenticatedRepository(ctx)

	ok, err := repo.GetLinkIsManualByBankAccountId(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to verify bank account link type")
	}
	if !ok {
		return c.badRequest(ctx, "Cannot import transactions for non-manual link.")
	}

	reader, header, err := ctx.Request().FormFile("data")
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "Failed to read file upload")
	}
	defer reader.Close()

	contentType := header.Header.Get("Content-Type")
	extension := strings.ToLower(path.Ext(header.Filename))
	log = log.WithFields(logrus.Fields{
		"contentType": contentType,
		"fileName":    header.Filename,
		"extension":   extension,
	})
	// If we only received an octet-stream then we need to try to interpret the
	// file format using the extension. We can validate the file more later.
	if contentType == "application/octet-stream" {
		log.Debug("upload content type is an octet stream, detecting file type by extension")
		switch extension {
		case ".qfx":
			log.Debug("detected QFX file format")
			contentType = string(storage.IntuitQFXContentType)
		default:
			log.Warn("could not determine file format by file extension")
		}
	}
	valid := storage.GetContentTypeIsValid(contentType)
	if !valid {
		crumbs.Debug(c.getContext(ctx),
			"Unsupported file type was provided!",
			map[string]interface{}{
				"fileName":    header.Filename,
				"contentType": contentType,
				"extension":   extension,
			},
		)
		return c.badRequest(ctx, "Unsupported file type!")
	}

	fileUri, err := c.fileStorage.Store(
		c.getContext(ctx),
		reader,
		storage.FileInfo{
			Name:          header.Filename,
			AccountId:     c.mustGetAccountId(ctx),
			BankAccountId: bankAccountId,
			ContentType:   storage.ContentType(contentType),
		},
	)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Failed to upload file")
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
		return c.wrapPgError(ctx, err, "Failed to create file")
	}

	return ctx.JSON(http.StatusOK, file)
}

func (c *Controller) getFiles(ctx echo.Context) error {
	bankAccountId, err := strconv.ParseUint(ctx.Param("bankAccountId"), 10, 64)
	if err != nil {
		return c.badRequest(ctx, "Must specify a valid bank account Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	files, err := repo.GetFiles(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to list files")
	}

	return ctx.JSON(http.StatusOK, files)
}
