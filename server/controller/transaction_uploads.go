package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/background"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/storage"
)

func (c *Controller) postTransactionUpload(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	var request struct {
		FileId ID[File] `json:"fileId"`
	}
	if err := ctx.Bind(&request); err != nil {
		return c.invalidJson(ctx)
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	file, err := repo.GetFile(c.getContext(ctx), request.FileId)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve specified file")
	}

	if file.ContentType == string(storage.IntuitQFXContentType) {
		return c.badRequest(ctx, "File is not a QFX/OFX file, and cannot be used for transaction imports")
	}

	upload := TransactionUpload{
		BankAccountId: bankAccountId,
		FileId:        request.FileId,
		Status:        TransactionUploadStatusPending,
		Error:         nil,
	}

	if err := repo.CreateTransactionUpload(
		c.getContext(ctx),
		bankAccountId,
		&upload,
	); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Failed to create transaction upload")
	}

	if err := c.jobRunner.EnqueueJob(
		c.getContext(ctx),
		background.ProcessQFXUpload,
		background.ProcessQFXUploadArguments{
			AccountId:           c.mustGetAccountId(ctx),
			BankAccountId:       bankAccountId,
			TransactionUploadId: upload.TransactionUploadId,
		},
	); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Failed to enqueue upload for processing")
	}

	return ctx.JSON(http.StatusOK, upload)
}
