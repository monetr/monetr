package controller

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/storage"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

func (c *Controller) postTransactionImport(ctx echo.Context) error {
	linkId, err := ParseID[Link](ctx.Param("linkId"))
	if err != nil || linkId.IsZero() {
		return c.badRequest(ctx, "Must specify a valid link Id")
	}

	// If sentry is setup, make sure we never send the body for this request to
	// sentry.
	if hub := sentry.GetHubFromContext(c.getContext(ctx)); hub != nil {
		if scope := hub.Scope(); scope != nil {
			scope.SetRequestBody(nil)
		}
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	link, err := repo.GetLink(c.getContext(ctx), linkId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve link")
	}

	if link.LinkType != ManualLinkType {
		return c.badRequest(ctx, "Link must be manual to import transactions")
	}

	txnImport := TransactionImport{
		LinkId: linkId,
	}

	file, err := c.consumeFileUpload(ctx, txnImport)
	if err != nil {
		return err
	}

	txnImport.FileId = file.FileId
	txnImport.File = file

	switch {
	case strings.EqualFold(file.ContentType, string(storage.IntuitQFXContentType)):
		break
	case strings.EqualFold(file.ContentType, string(storage.CAMT053ContentType)):
		break
	default:
		c.getLog(ctx).
			WithField("contentType", file.ContentType).
			Debug("could not create transaction upload because the file is not the expected content type: OFX")
		return c.badRequest(ctx, "File is not a recognized file type, and cannot be used for transaction imports")
	}

	// TODO Create the import record in the database, enqueue initial parsing

	return nil
}

func (c *Controller) getTransactionImportProgress(ctx echo.Context) error {
	linkId, err := ParseID[Link](ctx.Param("linkId"))
	if err != nil || linkId.IsZero() {
		return c.badRequest(ctx, "Must specify a valid link Id")
	}

	transactionImportId, err := ParseID[TransactionImport](ctx.Param("transactionImportId"))
	if err != nil || transactionImportId.IsZero() {
		return c.badRequest(ctx, "Must specify a valid transaction import Id")
	}

	log := c.getLog(ctx).WithFields(logrus.Fields{
		"linkId":              linkId,
		"transactionImportId": transactionImportId,
	})
	repo := c.mustGetAuthenticatedRepository(ctx)

	// TODO Read the import from the DB

	channel := fmt.Sprintf(
		"account:%s:transaction_import:%s:progress",
		c.mustGetAccountId(ctx), transactionImportId,
	)
	listener, err := c.PubSub.Subscribe(c.getContext(ctx), channel)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Failed to subscribe to transaction import changes")
	}
	defer listener.Close()

	timeout := time.NewTimer(5 * time.Minute)
	defer timeout.Stop()

	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		log.Trace("transaction import websocket established")

		if err := c.sendWebsocketMessage(ctx, ws, map[string]any{
			"foo": "bar",
		}); err != nil {
			return
		}

		// If the status of the import is terminal, then simply exit the loop now.
		// If the status of the import is confirming, then wait for the user to
		// provide some information with a timeout.
		// If the status of the import is processing then simply send updates.

	}).ServeHTTP(ctx.Response(), ctx.Request())

	return nil
}
