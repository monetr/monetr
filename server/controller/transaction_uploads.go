package controller

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/background"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/storage"
	"golang.org/x/net/websocket"
)

func (c *Controller) getTransactionUploadById(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	transactionUploadId, err := ParseID[TransactionUpload](ctx.Param("transactionUploadId"))
	if err != nil || transactionUploadId.IsZero() {
		return c.badRequest(ctx, "must specify a valid transaction upload Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	upload, err := repo.GetTransactionUpload(
		c.getContext(ctx),
		bankAccountId,
		transactionUploadId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve transaction upload by ID")
	}

	return ctx.JSON(http.StatusOK, upload)
}

func (c *Controller) getTransactionUploadProgress(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	transactionUploadId, err := ParseID[TransactionUpload](ctx.Param("transactionUploadId"))
	if err != nil || transactionUploadId.IsZero() {
		return c.badRequest(ctx, "must specify a valid transaction upload Id")
	}

	log := c.getLog(ctx)

	repo := c.mustGetAuthenticatedRepository(ctx)
	upload, err := repo.GetTransactionUpload(
		c.getContext(ctx),
		bankAccountId,
		transactionUploadId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve transaction upload by ID")
	}

	channel := fmt.Sprintf(
		"account:%s:transaction_upload:%s:progress",
		c.mustGetAccountId(ctx), transactionUploadId,
	)
	listener, err := c.ps.Subscribe(c.getContext(ctx), channel)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Failed to subscribe to transaction upload changes")
	}
	defer listener.Close()

	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		if err := websocket.Message.Send(ws, upload); err != nil {
			log.WithError(err).Error("failed to send transaction upload initial payload")
			return
		}

		switch upload.Status {
		case TransactionUploadStatusComplete, TransactionUploadStatusFailed:
			log.WithField("status", upload.Status).Debug("upload is already in a final status, sending message then closing")
			if err := websocket.Message.Send(ws, map[string]interface{}{
				"status": upload.Status,
			}); err != nil {
				log.WithError(err).Error("failed to send transaction upload status")
				return
			}
			return
		}

		for status := range listener.Channel() {
			log.WithField("status", status).Debug("sending status message for transaction upload")
			if err := websocket.Message.Send(ws, map[string]interface{}{
				"status": status.Payload(),
			}); err != nil {
				log.WithError(err).Error("failed to send transaction upload status")
				return
			}

			switch TransactionUploadStatus(status.Payload()) {
			case TransactionUploadStatusComplete, TransactionUploadStatusFailed:
				log.WithField("status", status.Payload()).Debug("observed final status, ending socket")
				return
			}
		}
	}).ServeHTTP(ctx.Response(), ctx.Request())

	return nil
}

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
