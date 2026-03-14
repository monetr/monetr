package ofx

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/elliotcourant/gofx"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/currency"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/similar"
	"github.com/monetr/monetr/server/storage/storage_jobs"
	"github.com/pkg/errors"
)

type ProcessOFXUploadArguments struct {
	AccountId           models.ID[models.Account]           `json:"accountId"`
	BankAccountId       models.ID[models.BankAccount]       `json:"bankAccountId"`
	TransactionUploadId models.ID[models.TransactionUpload] `json:"transactionUploadId"`
}

// updateUploadStatus will update the record of the transaction upload in the
// database to have the most up to date status. It will also send out a pubsub
// message for that transaction upload. Keep in mind that the pub sub message is
// not transactional, but the update can be depending on where this function is
// called. If this is called inside of a DB transaction; then the changes that
// this update has will only take effect if the transaction is committed. Where
// as the notification will go out no matter what.
func updateUploadStatus(
	ctx queue.Context,
	args ProcessOFXUploadArguments,
	status models.TransactionUploadStatus,
	errorMessage *string,
) error {
	log := ctx.Log().With(
		"accountId", args.AccountId,
		"bankAccountId", args.BankAccountId,
		"transactionUploadId", args.TransactionUploadId,
	)

	query := ctx.DB().ModelContext(ctx, &models.TransactionUpload{}).
		Where(`"account_id" = ?`, args.AccountId).
		Where(`"bank_account_id" = ?`, args.BankAccountId).
		Where(`"transaction_upload_id" = ?`, args.TransactionUploadId).
		Set(`"status" = ?`, status)

	switch status {
	case models.TransactionUploadStatusProcessing:
		query = query.Set(`"processed_at" = ?`, ctx.Clock().Now())
	case models.TransactionUploadStatusComplete:
		query = query.Set(`"completed_at" = ?`, ctx.Clock().Now())
	case models.TransactionUploadStatusFailed:
		query = query.Set(`"completed_at" = ?`, ctx.Clock().Now())
		if errorMessage != nil {
			query = query.Set(`"error" = ?`, *errorMessage)
		} else {
			query = query.Set(`"error" = ?`, "Unknown failure")
		}
	}

	log.Log(
		ctx,
		logging.LevelTrace,
		"updated transaction upload status",
		"status", status,
	)

	_, err := query.Update()
	if err != nil {
		return errors.Wrap(err, "failed to update upload status")
	}

	channel := fmt.Sprintf(
		"account:%s:transaction_upload:%s:progress",
		args.AccountId, args.TransactionUploadId,
	)
	payload := string(status)
	if err := ctx.Publisher().Notify(
		ctx,
		args.AccountId,
		channel,
		payload,
	); err != nil {
		return errors.Wrap(err, "failed to send progress notification for job")
	}
	log.Log(
		ctx,
		logging.LevelTrace,
		"sent progress notification for file upload",
		"channel", channel,
		"payload", payload,
	)

	return nil
}

type processOfxUpload struct {
	args     ProcessOFXUploadArguments
	log      *slog.Logger
	repo     repository.BaseRepository
	timezone *time.Location

	bankAccount           *models.BankAccount
	upload                *models.TransactionUpload
	file                  *models.File
	data                  *gofx.OFX
	currency              string
	statementTransactions []gofx.StatementTransaction
	existingTransactions  map[string]models.Transaction
}

// loadFile takes the ID of the [models.TrasactionUpload] and reads the file
// from the storage system. If this fails due to a retryable issue then an error
// is returned so the job can be re-attempted. If this fails due to a bad file,
// then this function panics so that the job is not retried.
func (j *processOfxUpload) loadFile(
	ctx queue.Context,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	txnUpload, err := j.repo.GetTransactionUpload(
		span.Context(),
		j.args.BankAccountId,
		j.args.TransactionUploadId,
	)
	if err != nil {
		return errors.Wrap(err, "failed to process OFX file upload")
	}
	j.upload = txnUpload

	file, err := j.repo.GetFile(span.Context(), txnUpload.FileId)
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

	ofxData, err := ParseFile(fileReader)
	if err != nil {
		return queue.FailWithoutRetry(ctx, err)
	}

	j.data = ofxData
	return nil
}

// hydrateTransactions takes all of the transactions that were present in the
// OFX file and tries to cross reference them with transactions that already
// exist in the database. It relies on the OFX file having a unique identifier
// for each transaction that is consistent between each download from the FI.
func (j *processOfxUpload) hydrateTransactions(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	externalTransactionIds := make([]string, 0)

	// Gather the bank transactions
	// TODO Clean this up, this is a nested nightmare.
	if bankResponse := j.data.BANKMSGSRSV1; bankResponse != nil {
		for _, statementTransactions := range bankResponse.STMTTRNRS {
			if stmtrs := statementTransactions.STMTRS; stmtrs != nil {
				if stmtrs.BANKTRANLIST != nil {
					for _, transaction := range stmtrs.BANKTRANLIST.STMTTRN {
						// Someday we might need to also consider CORRECTFITID.
						externalTransactionIds = append(externalTransactionIds, transaction.FITID)
						j.statementTransactions = append(j.statementTransactions, *transaction)
					}
				}
				if j.currency == "" {
					j.currency = strings.ToUpper(stmtrs.CURDEF)
				}
			}
		}
	} else if bankResponse := j.data.CREDITCARDMSGSRSV1; bankResponse != nil {
		for _, statementTransactions := range bankResponse.CCSTMTTRNRS {
			if ccstmtrs := statementTransactions.CCSTMTRS; ccstmtrs != nil {
				if ccstmtrs.BANKTRANLIST != nil {
					for _, transaction := range ccstmtrs.BANKTRANLIST.STMTTRN {
						// Someday we might need to also consider CORRECTFITID.
						externalTransactionIds = append(externalTransactionIds, transaction.FITID)
						j.statementTransactions = append(j.statementTransactions, *transaction)
					}
				}
				if j.currency == "" {
					j.currency = strings.ToUpper(ccstmtrs.CURDEF)
				}
			}
		}
	}

	// If we are unable to derive the currency code from the file itself then we
	// should fallback to the bank account's default. This way if we are able to
	// verify the currency from the file but it doesn't match the currency of the
	// account we still throw an error later.
	if j.currency == "" {
		j.log.DebugContext(span.Context(),
			"could not detect currency from OFX file, defaulting to bank account's currency instead",
			"currency", j.bankAccount.Currency,
		)
		j.currency = j.bankAccount.Currency
	} else {
		j.log.DebugContext(span.Context(), "detected currency from OFX file", "currency", j.currency)
	}

	// Reverse the order of the arrray we store such that the order we insert the
	// transactions into the DB matches the order of the transactions in the
	// actual file.
	slices.Reverse(j.statementTransactions)

	// TODO Add others as needed. Not sure what other formats we'll end up seeing
	// over time.

	if len(externalTransactionIds) == 0 {
		j.log.WarnContext(span.Context(), "no external transaction IDs were found in the file, account type may not be supported")
		return nil
	}

	var err error
	j.existingTransactions, err = j.repo.GetTransactonsByUploadIdentifier(
		span.Context(),
		j.args.BankAccountId,
		externalTransactionIds,
	)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve existing transactions for upload processing")
	}

	if count := len(j.existingTransactions); count > 0 {
		j.log.DebugContext(span.Context(), "found existing transactions for upload", "existingTransactions", count)
	}

	return nil
}

func (j *processOfxUpload) syncTransactions(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	transactionsToUpdate := make([]*models.Transaction, 0)
	transactionsToCreate := make([]models.Transaction, 0)
	for y := range j.statementTransactions {
		externalTransaction := j.statementTransactions[y]
		uploadIdentifier := externalTransaction.FITID
		tlog := j.log.With("uploadIdentifier", uploadIdentifier)

		// Parse the amount in the specified currency.
		amount, err := ParseTransactionAmount(
			externalTransaction,
			j.currency,
		)
		if err != nil {
			tlog.ErrorContext(span.Context(), "failed to parse transaction amount",
				"err", err,
				"trntype", externalTransaction.TRNTYPE,
				"trnamt", externalTransaction.TRNAMT,
			)
			continue
		}

		// Still need to figure out something to do with memo, but for now we can
		// take the name and trim it. Memo seems to behave a bit differently from
		// FI to FI. At NFCU for example it contains a larger more un-santized
		// version of the transaction name, but at US Bank it seems to contain
		// reference numbers that might be useful internally? But are definitely
		// not helpful here.
		name := strings.TrimSpace(externalTransaction.NAME)
		originalName := strings.TrimSpace(externalTransaction.MEMO)

		// Make sure that the original name and name are set. This way if name is
		// blank it will use the original name. And if original name is blank it
		// will use the name.
		name = myownsanity.CoalesceStrings(name, originalName)
		originalName = myownsanity.CoalesceStrings(originalName, name)

		// TODO Also parse DTAVAIL at some point
		date, err := ParseDate(externalTransaction.DTPOSTED, j.timezone)
		if err != nil {
			tlog.ErrorContext(span.Context(), "failed to parse transaction date posted",
				"err", err,
				"dtposted", externalTransaction.DTPOSTED,
			)
			continue
		}

		transaction, ok := j.existingTransactions[uploadIdentifier]
		if !ok {
			transaction = models.Transaction{
				TransactionId:        models.NewID[models.Transaction](),
				AccountId:            j.args.AccountId,
				BankAccountId:        j.args.BankAccountId,
				Amount:               amount,
				Date:                 date,
				Name:                 name,
				OriginalName:         originalName,
				OriginalMerchantName: name,
				IsPending:            false, // OFX files don't show pending?
				UploadIdentifier:     &uploadIdentifier,
				Source:               models.TransactionSourceUpload,
			}
			transactionsToCreate = append(transactionsToCreate, transaction)
			continue
		}

		// TODO Process changes to an existing transaction.
		_ = transaction
	}

	// Persist any new transactions.
	if count := len(transactionsToCreate); count > 0 {
		j.log.InfoContext(span.Context(), "creating new transactions from import", "new", count)
		if err := j.repo.InsertTransactions(span.Context(), transactionsToCreate); err != nil {
			return errors.Wrap(err, "failed to persist new transactions")
		}
	}

	// If there are any updated transactions persist those as well.
	if count := len(transactionsToUpdate); count > 0 {
		j.log.InfoContext(span.Context(), "updating transactions from import", "updated", count)
		if err := j.repo.UpdateTransactions(span.Context(), transactionsToUpdate); err != nil {
			return errors.Wrap(err, "failed to update transactions")
		}
	}

	return nil
}

func (j *processOfxUpload) syncBalances(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	// TODO Somehow keep track of the as of timestamp? This way if someone is
	// importing files out of order we could potentially avoid updating the
	// balance to an old value.
	var currentBalance, availableBalance, limitBalance int64
	var err error
	if j.data.BANKMSGSRSV1 != nil {
		for i := range j.data.BANKMSGSRSV1.STMTTRNRS {
			statementTransactions := j.data.BANKMSGSRSV1.STMTTRNRS[i]
			if statementTransactions.STMTRS.LEDGERBAL != nil {
				currentBalance, err = currency.ParseFriendlyToAmount(
					statementTransactions.STMTRS.LEDGERBAL.BALAMT,
					j.currency,
				)
				// EOF means the amount is blank, we can treat this as zero
				switch errors.Cause(err) {
				case nil, io.EOF:
					break
				default:
					return errors.Wrap(err, "failed to parse ledger balance amount")
				}
			}

			if statementTransactions.STMTRS.AVAILBAL != nil {
				availableBalance, err = currency.ParseFriendlyToAmount(
					statementTransactions.STMTRS.AVAILBAL.BALAMT,
					j.currency,
				)
				// EOF means the amount is blank, we can treat this as zero
				switch errors.Cause(err) {
				case nil, io.EOF:
					break
				default:
					return errors.Wrap(err, "failed to parse available balance amount")
				}
			}
		}
	} else if j.data.CREDITCARDMSGSRSV1 != nil {
		for i := range j.data.CREDITCARDMSGSRSV1.CCSTMTTRNRS {
			statementTransactions := j.data.CREDITCARDMSGSRSV1.CCSTMTTRNRS[i]
			if statementTransactions.CCSTMTRS.LEDGERBAL != nil {
				currentBalance, err = currency.ParseFriendlyToAmount(
					statementTransactions.CCSTMTRS.LEDGERBAL.BALAMT,
					j.currency,
				)
				// EOF means the amount is blank, we can treat this as zero
				switch errors.Cause(err) {
				case nil, io.EOF:
					break
				default:
					return errors.Wrap(err, "failed to parse ledger balance amount")
				}
			}

			if statementTransactions.CCSTMTRS.AVAILBAL != nil {
				availableBalance, err = currency.ParseFriendlyToAmount(
					statementTransactions.CCSTMTRS.AVAILBAL.BALAMT,
					j.currency,
				)
				// EOF means the amount is blank, we can treat this as zero
				switch errors.Cause(err) {
				case nil, io.EOF:
					break
				default:
					return errors.Wrap(err, "failed to parse available balance amount")
				}
			}
			// The limit for credit cards is equal to the amount currrently available
			// plus the inverse of any amount currently used.
			currentBalance = -1 * currentBalance
			limitBalance = availableBalance + currentBalance
		}
	}

	if j.currency != j.bankAccount.Currency {
		return errors.Errorf(
			"Currency of OFX file does not match currency of bank account, file: [%s], account: [%s]",
			j.currency,
			j.bankAccount.Currency,
		)
	}

	// TODO Log the previous value and the new one?
	j.bankAccount.CurrentBalance = currentBalance
	j.bankAccount.AvailableBalance = availableBalance
	j.bankAccount.LimitBalance = limitBalance

	if err := j.repo.UpdateBankAccount(span.Context(), j.bankAccount); err != nil {
		return errors.Wrap(err, "failed to update bank account balances")
	}

	return nil
}

func ProcessOFXUpload(
	ctx queue.Context,
	args ProcessOFXUploadArguments,
) error {
	crumbs.IncludeUserInScope(ctx, args.AccountId)
	defer func() {
		// Only mark the file as failed if we panic. This means the job won't be
		// retried.
		if recovery := recover(); recovery != nil {
			ctx.Log().ErrorContext(
				ctx,
				"panic processing OFX file upload",
				"err", recovery,
			)
			updateUploadStatus(ctx, args, models.TransactionUploadStatusFailed, nil)
			panic(recovery)
		}
	}()

	return ctx.RunInTransaction(ctx, func(ctx queue.Context) error {
		updateUploadStatus(ctx, args, models.TransactionUploadStatusProcessing, nil)
		log := ctx.Log().With(
			"accountId", args.AccountId,
			"transactionUploadId", args.TransactionUploadId,
			"bankAccountId", args.BankAccountId,
		)
		j := &processOfxUpload{
			args: args,
			log:  log,
			repo: repository.NewRepositoryFromSession(
				ctx.Clock(),
				"user_system",
				args.AccountId,
				ctx.DB(),
				log,
			),
			statementTransactions: []gofx.StatementTransaction{},
			existingTransactions:  map[string]models.Transaction{},
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

		// Pull all of the transactions that already exist in our system from the file
		// so we can compare.
		if err := j.hydrateTransactions(ctx); err != nil {
			return err
		}

		// Push new and updated transactions to the database.
		if err := j.syncTransactions(ctx); err != nil {
			return err
		}

		if err := j.syncBalances(ctx); err != nil {
			return err
		}

		// Queue up the other jobs that we want after this one.
		return myownsanity.FirstError(
			queue.Enqueue(
				ctx,
				ctx.Enqueuer(),
				similar.CalculateTransactionClusters,
				similar.CalculateTransactionClustersArguments{
					AccountId:     args.AccountId,
					BankAccountId: args.BankAccountId,
				},
			),
			queue.Enqueue(
				ctx,
				ctx.Enqueuer(),
				storage_jobs.RemoveFile,
				storage_jobs.RemoveFileArguments{
					AccountId: j.file.AccountId,
					FileId:    j.file.FileId,
				},
			),
		)
	})
}
