package csv_jobs

import (
	"encoding/csv"
	"io"
	"log/slog"
	"time"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/datasources/table"
	"github.com/monetr/monetr/server/id"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
)

type PreviewCSVImportArguments struct {
	AccountId           models.ID[models.Account]           `json:"accountId"`
	BankAccountId       models.ID[models.BankAccount]       `json:"bankAccountId"`
	TransactionImportId models.ID[models.TransactionImport] `json:"transactionImportId"`
}

type previewCSVImport struct {
	args     PreviewCSVImportArguments
	log      *slog.Logger
	repo     repository.BaseRepository
	timezone *time.Location

	bankAccount          *models.BankAccount
	transactionImport    *models.TransactionImport
	file                 *models.File
	currency             string
	rows                 []models.TransactionImportPreviewItem
	existingTransactions map[string]models.Transaction
}

// loadFile takes the ID of the [models.TrasactionUpload] and reads the file
// from the storage system. If this fails due to a retryable issue then an error
// is returned so the job can be re-attempted. If this fails due to a bad file,
// then this function panics so that the job is not retried.
func (j *previewCSVImport) loadFile(
	ctx queue.Context,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	transactionImport, err := j.repo.GetTransactionImport(
		span.Context(),
		j.args.BankAccountId,
		j.args.TransactionImportId,
	)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve transaction import")
	}
	j.transactionImport = transactionImport

	// GetTransactionImport does not eager load the mapping relation, but we
	// need its [table.Mapping] data to parse the file. Fetch it directly.
	if j.transactionImport.TransactionImportMappingId == nil {
		return errors.New("transaction import is missing a mapping")
	}
	mapping, err := j.repo.GetTransactionImportMapping(
		span.Context(),
		*j.transactionImport.TransactionImportMappingId,
	)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve transaction import mapping")
	}
	j.transactionImport.TransactionImportMapping = mapping

	file, err := j.repo.GetFile(span.Context(), transactionImport.FileId)
	if err != nil {
		return errors.Wrap(err, "could not get file for processing")
	}
	j.file = file

	if file.DeletedAt != nil {
		return errors.New("cannot import transactions from a deleted file")
	}

	fileReader, err := ctx.Storage().Read(span.Context(), *file)
	if err != nil {
		return errors.Wrap(err, "failed to access file from storage")
	}
	defer fileReader.Close()

	// TODO Store the delimiter on the import so we can use it here.
	csvReader := csv.NewReader(fileReader)
	csvReader.TrimLeadingSpace = true

	tableReader := table.NewTable(
		csvReader,
		&j.transactionImport.TransactionImportMapping.Mapping,
		true, // Store this too?
	)
	// TODO, make arbitrary max number of rows supported a const
	for i := 0; i < 100000; i++ {
		row, err := tableReader.Read(span.Context())
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return errors.Wrap(err, "failed to parse csv file")
		}
		j.rows = append(j.rows, models.TransactionImportPreviewItem{
			TransactionImportPreviewItemId: id.New(),
			Data:                           *row,
			ExistingTransactionIds:         []models.ID[models.Transaction]{},
		})
	}

	j.log.DebugContext(
		span.Context(),
		"parsed rows from CSV file for preview",
	)

	return nil
}

func (j *previewCSVImport) hydrateTransactions(
	ctx queue.Context,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var err error
	j.existingTransactions, err = j.repo.GetTransactonsByUploadIdentifier(
		span.Context(),
		j.args.BankAccountId,
		myownsanity.Map(j.rows, func(row models.TransactionImportPreviewItem) string {
			return row.Data.ID
		}),
	)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve existing transactions")
	}

	if len(j.existingTransactions) > 0 {
		j.log.DebugContext(
			span.Context(),
			"found existing transactions",
			"count", len(j.existingTransactions),
		)
	} else {
		j.log.DebugContext(
			span.Context(),
			"did not find any existing transactions",
		)
	}

	return nil
}

func (j *previewCSVImport) syncTransactions(
	ctx queue.Context,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	for i := range j.rows {
		uploadIdentifier := j.rows[i].Data.ID
		transaction, ok := j.existingTransactions[uploadIdentifier]
		if ok {
			j.rows[i].ExistingTransactionIds = append(j.rows[i].ExistingTransactionIds, transaction.TransactionId)
		}

		// TODO Process changes to an existing transaction.
		_ = transaction
	}

	return nil
}

func PreviewCSVImport(
	ctx queue.Context,
	args PreviewCSVImportArguments,
) error {
	crumbs.IncludeUserInScope(ctx, args.AccountId)

	return ctx.RunInTransaction(ctx, func(ctx queue.Context) error {
		log := ctx.Log().With(
			"accountId", args.AccountId,
			"transactionImportId", args.TransactionImportId,
			"bankAccountId", args.BankAccountId,
		)
		j := &previewCSVImport{
			args: args,
			log:  log,
			repo: repository.NewRepositoryFromSession(
				ctx.Clock(),
				"user_system",
				args.AccountId,
				ctx.DB(),
				log,
			),
			existingTransactions: map[string]models.Transaction{},
		}

		account, err := j.repo.GetAccount(ctx)
		if err != nil {
			j.log.ErrorContext(ctx, "failed to retrieve account for job", "err", err)
			return err
		}

		j.timezone, err = account.GetTimezone()
		if err != nil {
			j.log.WarnContext(ctx, "failed to get account's time zone, defaulting to UTC", "err", err)
			j.timezone = time.UTC
		}

		// Load the bank account ahead of processing the file, we need this for
		// currency data and will use it for balance updates later.
		j.bankAccount, err = j.repo.GetBankAccount(ctx, args.BankAccountId)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve bank account for file import sync")
		}

		// Load the file and its data into memory.
		if err := j.loadFile(ctx); err != nil {
			return err
		}

		if err := j.hydrateTransactions(ctx); err != nil {
			return err
		}

		if err := j.syncTransactions(ctx); err != nil {
			return err
		}

		preview := models.TransactionImportPreview{
			TransactionImportId: args.TransactionImportId,
			Rows:                j.rows,
			AvailableBalance:    0,
			CurrentBalance:      0,
		}

		if err := j.repo.CreateTransactionImportPreview(
			ctx,
			args.BankAccountId,
			&preview,
		); err != nil {
			return err
		}

		// Move to the preview status
		j.transactionImport.Status = models.TransactionImportStatusPreview
		if err := j.repo.UpdateTransactionImport(
			ctx,
			j.args.BankAccountId,
			j.transactionImport,
		); err != nil {
			return err
		}

		return nil
	})
}
