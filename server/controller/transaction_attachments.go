package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	. "github.com/monetr/monetr/server/models"
)

func (c *Controller) postTransactionAttachment(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	transactionId, err := ParseID[Transaction](ctx.Param("transactionId"))
	if err != nil || transactionId.IsZero() {
		return c.badRequest(ctx, "must specify a valid transaction Id")
	}

	attachment := TransactionAttachment{
		BankAccountId: bankAccountId,
		TransactionId: transactionId,
	}

	// Take the body and upload it as a file
	file, err := c.consumeFileUpload(ctx, attachment)
	if err != nil {
		return err
	}
	attachment.FileId = file.FileId
	attachment.File = file

	return ctx.JSON(http.StatusOK, attachment)
}
