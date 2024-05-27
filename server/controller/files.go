package controller

import (
	"net/http"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/storage"
	"github.com/sirupsen/logrus"
)

func (c *Controller) postFile(ctx echo.Context) error {
	if !c.configuration.Storage.Enabled {
		return c.notFound(ctx, "File uploads are not enabled on this server")
	}

	log := c.getLog(ctx)

	repo := c.mustGetAuthenticatedRepository(ctx)

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
			Name:        header.Filename,
			Kind:        "transactions/import", // TODO What should this be?
			AccountId:   c.mustGetAccountId(ctx),
			ContentType: storage.ContentType(contentType),
		},
	)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Failed to upload file")
	}

	file := File{
		AccountId:   c.mustGetAccountId(ctx),
		Name:        header.Filename,
		ContentType: contentType,
		Size:        uint64(header.Size),
		BlobUri:     fileUri,
	}

	if err := repo.CreateFile(c.getContext(ctx), &file); err != nil {
		return c.wrapPgError(ctx, err, "Failed to create file")
	}

	return ctx.JSON(http.StatusOK, file)
}

func (c *Controller) getFiles(ctx echo.Context) error {
	repo := c.mustGetAuthenticatedRepository(ctx)

	files, err := repo.GetFiles(c.getContext(ctx))
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to list files")
	}

	return ctx.JSON(http.StatusOK, files)
}
