package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/background"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/storage"
	"golang.org/x/net/websocket"
)

func (c *Controller) postTransactionUpload(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	// If sentry is setup, make sure we never send the body for this request to
	// sentry.
	if hub := sentry.GetHubFromContext(c.getContext(ctx)); hub != nil {
		if scope := hub.Scope(); scope != nil {
			scope.SetRequestBody(nil)
		}
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	upload := TransactionUpload{
		BankAccountId: bankAccountId,
		Status:        TransactionUploadStatusPending,
		Error:         nil,
	}

	// Take the body and upload it as a file
	file, err := c.consumeFileUpload(ctx, upload)
	if err != nil {
		return err
	}
	upload.FileId = file.FileId
	upload.File = file

	if !strings.EqualFold(file.ContentType, string(storage.IntuitQFXContentType)) {
		c.getLog(ctx).
			WithField("contentType", file.ContentType).
			Debug("could not create transaction upload because the file is not the expected content type: OFX")
		return c.badRequest(ctx, "File is not a OFX file, and cannot be used for transaction imports")
	}

	if err := repo.CreateTransactionUpload(
		c.getContext(ctx),
		bankAccountId,
		&upload,
	); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Failed to create transaction upload")
	}

	if err := c.JobRunner.EnqueueJob(
		c.getContext(ctx),
		background.ProcessOFXUpload,
		background.ProcessOFXUploadArguments{
			AccountId:           c.mustGetAccountId(ctx),
			BankAccountId:       bankAccountId,
			TransactionUploadId: upload.TransactionUploadId,
		},
	); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Failed to enqueue upload for processing")
	}

	return ctx.JSON(http.StatusOK, upload)
}

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
	listener, err := c.PubSub.Subscribe(c.getContext(ctx), channel)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Failed to subscribe to transaction upload changes")
	}
	defer listener.Close()

	timeout := time.NewTimer(1 * time.Minute)
	defer timeout.Stop()

	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		if err := c.sendWebsocketMessage(ctx, ws, upload); err != nil {
			return
		}

		switch upload.Status {
		case TransactionUploadStatusComplete, TransactionUploadStatusFailed:
			log.WithField("status", upload.Status).Debug("upload is already in a final status, sending message then closing")
			_ = c.sendWebsocketMessage(ctx, ws, map[string]interface{}{
				"status": upload.Status,
			})
			return
		}

	ListenerLoop:
		for {
			select {
			case <-timeout.C:
				log.Warn("transaction upload is taking too long, websocket will be terminated")
				_ = c.sendWebsocketMessage(ctx, ws, map[string]interface{}{
					"status": "timed out",
				})
				break ListenerLoop
			case status := <-listener.Channel():
				log.WithField("status", status).Debug("sending status message for transaction upload")
				if err := c.sendWebsocketMessage(ctx, ws, map[string]interface{}{
					"status": status.Payload(),
				}); err != nil {
					return
				}

				switch TransactionUploadStatus(status.Payload()) {
				case TransactionUploadStatusComplete, TransactionUploadStatusFailed:
					log.WithField("status", status.Payload()).Debug("observed final status, ending socket")
					break ListenerLoop
				}
			}

		}

		log.Trace("final status detected, re-reading transaction upload object")
		upload, err := repo.GetTransactionUpload(
			c.getContext(ctx),
			bankAccountId,
			transactionUploadId,
		)
		if err != nil {
			return
		}

		if err := c.sendWebsocketMessage(ctx, ws, upload); err != nil {
			return
		}
	}).ServeHTTP(ctx.Response(), ctx.Request())

	return nil
}

func (c *Controller) sendWebsocketMessage(ctx echo.Context, ws *websocket.Conn, message any) error {
	log := c.getLog(ctx)
	msg, err := json.Marshal(message)
	if err != nil {
		log.WithField("mesasge", message).WithError(err).Error("failed to encode websocket message")
		return err
	}
	if err := websocket.Message.Send(ws, string(msg)); err != nil {
		log.WithField("mesasge", message).WithError(err).Error("failed to send websocket message")
		return err
	}

	return nil
}
