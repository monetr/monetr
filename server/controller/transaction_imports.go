package controller

import (
	"net/http"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/datasources/csv"
	"github.com/monetr/monetr/server/datasources/csv/csv_jobs"
	. "github.com/monetr/monetr/server/models"
)

func (c *Controller) postTransactionImport(ctx echo.Context) error {
	c.scrubSentryBody(ctx)

	if !c.Configuration.Storage.Enabled {
		return c.notFound(ctx, "File uploads are not enabled on this server")
	}

	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	log := c.getLog(ctx)

	repo := c.mustGetAuthenticatedRepository(ctx)
	transactionImport := TransactionImport{
		BankAccountId: bankAccountId,
		Status:        TransactionImportStatusMapping,
	}

	// Take the body and upload it as a file
	// TODO Add ClamAV anti-virus here
	{
		form, err := ctx.MultipartForm()
		if err != nil {
			return c.badRequestError(ctx, err, "Failed to read file upload")
		}
		if len(form.File) != 1 || len(form.File["data"]) != 1 {
			return c.badRequest(ctx, "exactly one file must be uploaded under field \"data\"")
		}

		reader, header, err := ctx.Request().FormFile("data")
		if err != nil {
			return c.badRequestError(ctx, err, "Failed to read file upload")
		}
		defer reader.Close()

		if header.Size <= 0 {
			return c.badRequest(ctx, "uploaded file must not be empty")
		}

		contentType := header.Header.Get("Content-Type")
		extension := strings.ToLower(path.Ext(header.Filename))
		log = log.With(
			"contentType", contentType,
			"fileName", header.Filename,
			"extension", extension,
		)

		// TODO We only support CSV here, so we need to be able to attempt to parse
		// the file before we store it.
		delimeter, headers, buffer, err := csv.PeekHeader(reader)
		if err != nil {
			return c.badRequestError(ctx, err, "Failed to parse CSV file")
		}

		file := File{
			Name:        header.Filename,
			Kind:        transactionImport.FileKind(),
			ContentType: TextCSVContentType,
			Size:        uint64(header.Size),
			ExpiresAt:   transactionImport.FileExpiration(c.Clock),
		}

		if err := repo.CreateFile(c.getContext(ctx), &file); err != nil {
			return c.wrapPgError(ctx, err, "Failed to create file")
		}

		err = c.FileStorage.Store(
			c.getContext(ctx),
			buffer,
			file,
		)
		if err != nil {
			return c.wrapAndReturnError(
				ctx,
				err,
				http.StatusInternalServerError,
				"Failed to upload file",
			)
		}

		transactionImport.Headers = headers
		transactionImport.Delimeter = string(delimeter)
		transactionImport.FileId = file.FileId
		transactionImport.File = &file
	}

	if err := repo.CreateTransactionImport(
		c.getContext(ctx),
		bankAccountId,
		&transactionImport,
	); err != nil {
		return c.wrapPgError(ctx, err, "Failed to create transaction import")
	}

	return ctx.JSON(http.StatusOK, transactionImport)
}

func (c *Controller) getTransactionImportById(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	transactionImportId, err := ParseID[TransactionImport](ctx.Param("transactionImportId"))
	if err != nil || transactionImportId.IsZero() {
		return c.badRequest(ctx, "must specify a valid transaction import Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	existing, err := repo.GetTransactionImport(
		c.getContext(ctx),
		bankAccountId,
		transactionImportId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve transaction import")
	}

	return ctx.JSON(http.StatusOK, existing)
}

func (c *Controller) patchTransactionImport(ctx echo.Context) error {
	// After the user has created their transaction import, then theyll get the
	// header info back and need to map it. The frontend can look up mappings
	// based on the header signature
	// (?signature=sha256(join(sort(lower(headers)), ","))) to autofill the
	// mappings if a similar file has been created before. The signature is
	// case-insensitive so that "Date" and "date" match the same mapping. If the
	// user modifies the mapping then they do a POST to the mapping endpoint, if
	// the user uses the existing mapping then they can use that ID.
	// The user then PATCH calls the transaction import with the mapping ID and
	// updating the status to `preview`.
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	transactionImportId, err := ParseID[TransactionImport](ctx.Param("transactionImportId"))
	if err != nil || transactionImportId.IsZero() {
		return c.badRequest(ctx, "must specify a valid transaction import Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	existing, err := repo.GetTransactionImport(
		c.getContext(ctx),
		bankAccountId,
		transactionImportId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve transaction import")
	}

	transactionImport, err := parse(
		c,
		ctx,
		existing,
		existing.PatchSchemas()...,
	)
	if err != nil {
		return err
	}

	// We only allow patching in two scenarios, when the new status is pending
	// preview and when the new status is pending processing. And we only allow
	// each to be specified when the current status is also in a specific state.
	switch existing.Status {
	case TransactionImportStatusMapping:
		if transactionImport.Status != TransactionImportStatusPendingPreview {
			return c.badRequest(ctx, "Cannot move a transaction import to any status other than pending preview from mapping")
		}
	case TransactionImportStatusPreview:
		if transactionImport.Status != TransactionImportStatusPendingProcessing {
			return c.badRequest(ctx, "Cannot move a transaction import to any status other than pending processing from preview")
		}
	}

	if err = repo.UpdateTransactionImport(
		c.getContext(ctx),
		bankAccountId,
		&transactionImport,
	); err != nil {
		return c.wrapPgError(ctx, err, "failed to update transaction import")
	}

	// TODO What about when we have other file types in the import path? Like
	// OFX?
	switch transactionImport.Status {
	case TransactionImportStatusPendingPreview:
		if err = enqueueJob(
			c,
			ctx,
			csv_jobs.PreviewCSVImport,
			csv_jobs.PreviewCSVImportArguments{
				AccountId:           c.mustGetAccountId(ctx),
				BankAccountId:       bankAccountId,
				TransactionImportId: transactionImportId,
			},
		); err != nil {
			return c.wrapAndReturnError(
				ctx,
				err,
				http.StatusInternalServerError,
				"failed to enqueue import preview job",
			)
		}
	case TransactionImportStatusPendingProcessing:
		// TODO Kick off processing
	}

	return ctx.JSON(http.StatusOK, transactionImport)
}
