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

func (c *Controller) consumeFileUpload(ctx echo.Context, kind Uploadable) (*File, error) {
	if !c.Configuration.Storage.Enabled {
		return nil, c.notFound(ctx, "File uploads are not enabled on this server")
	}

	log := c.getLog(ctx)

	repo := c.mustGetAuthenticatedRepository(ctx)

	reader, header, err := ctx.Request().FormFile("data")
	if err != nil {
		return nil, c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "Failed to read file upload")
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
		case ".qfx", ".ofx":
			log.Debug("detected OFX file format")
			contentType = string(storage.IntuitQFXContentType)
		case ".pdf":
			contentType = "application/pdf"
		case ".png":
			contentType = "image/png"
		case ".jpeg", ".jpg":
			contentType = "image/jpeg"
		case ".heic", ".heif": // Apple's image format.
			contentType = "image/heif"
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
		return nil, c.badRequest(ctx, "Unsupported file type!")
	}

	fileUri, err := c.FileStorage.Store(
		c.getContext(ctx),
		reader,
		storage.FileInfo{
			Name:        header.Filename,
			Kind:        kind.FileKind(),
			AccountId:   c.mustGetAccountId(ctx),
			ContentType: storage.ContentType(contentType),
		},
	)
	if err != nil {
		return nil, c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Failed to upload file")
	}

	file := File{
		AccountId:   c.mustGetAccountId(ctx),
		Name:        header.Filename,
		ContentType: contentType,
		Size:        uint64(header.Size),
		BlobUri:     fileUri,
		// May be nil, if it is nil then it never expires.
		ExpiresAt: kind.FileExpiration(c.Clock),
	}

	if err := repo.CreateFile(c.getContext(ctx), &file); err != nil {
		return nil, c.wrapPgError(ctx, err, "Failed to create file")
	}

	return &file, nil
}
